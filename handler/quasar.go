package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/netip"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/antlr4-go/antlr/v4"
	"github.com/itsubaki/q"
	"github.com/itsubaki/qasm/environ"
	"github.com/itsubaki/qasm/gen/parser"
	"github.com/itsubaki/qasm/listener"
	xparser "github.com/itsubaki/qasm/parser"
	"github.com/itsubaki/qasm/visitor"
	"github.com/itsubaki/quasar/execlog"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/store"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	salt            = "quasar salt\n"
	maxSize         = 64 * 1024
	maxRequestBytes = 6*maxSize + 1024 // JSON may encode each source byte as a six-byte Unicode escape.
	execLogTimeout  = 2 * time.Second
)

var (
	scale   = math.Pow(10, 6)
	epsilon = 1e-6
)

var (
	ErrQubitsNotFound     = errors.New("qubits not found")
	ErrCodeNotFound       = errors.New("code not found")
	ErrIDNotFound         = errors.New("id not found")
	ErrNoSuchEntity       = errors.New("no such entity")
	ErrSomethingWentWrong = errors.New("something went wrong")
	ErrIncludeNotAllowed  = errors.New("include is not allowed")
	ErrBusy               = errors.New("another computation is running")
)

var (
	_ Store      = (*store.MemoryStore)(nil)
	_ Store      = (*store.Firestore)(nil)
	_ ExecLogger = (*execlog.StdOutLogger)(nil)
	_ ExecLogger = (*execlog.BigQueryLogger)(nil)
)

type Store interface {
	Put(ctx context.Context, id string, snippet *store.Snippet) error
	Get(ctx context.Context, id string) (*store.Snippet, error)
}

type ExecLogger interface {
	Log(ctx context.Context, entry *execlog.Entry) error
}

type QuasarService struct {
	Revision      string
	MaxQubits     int
	MaxConcurrent int // Set before first use; values below 1 default to 1.
	Store         Store
	ExecLogger    ExecLogger
	computeOnce   sync.Once
	compute       chan struct{}
}

func (s *QuasarService) Simulate(
	ctx context.Context,
	req *connect.Request[quasarv1.SimulateRequest],
) (resp *connect.Response[quasarv1.SimulateResponse], err error) {
	if err := checkCode(req.Msg.Code); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if !s.acquireCompute() {
		return nil, connect.NewError(connect.CodeResourceExhausted, ErrBusy)
	}

	var startedAt time.Time
	defer func() {
		panicValue := recover()
		// Logging must not occupy a compute slot while waiting for its sink.
		<-s.compute

		if !startedAt.IsZero() {
			s.logExecution(ctx, req, resp, err, panicValue, startedAt)
		}

		if panicValue != nil {
			panic(panicValue)
		}
	}()

	program, err := parseCode(req.Msg.Code)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// quantum simulator
	qsim := q.New()
	env := environ.New()
	v := visitor.New(qsim, env,
		visitor.WithMaxQubits(s.MaxQubits),
	)

	startedAt = time.Now()
	if err := v.Run(program); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if len(env.Qubit) == 0 && len(env.QubitArray) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrQubitsNotFound)
	}

	round := func(v float64) float64 {
		return math.Round(v*scale) / scale
	}

	// quantum state
	qstate := qsim.Qubit().State(env.Index()...)

	// build response
	states := make([]*quasarv1.SimulateResponse_State, 0)
	for _, s := range qstate {
		prob := round(s.Probability())
		if prob < epsilon {
			continue
		}

		// prob > 0.000000
		states = append(states, &quasarv1.SimulateResponse_State{
			BinaryString: s.BinaryString(),
			Probability:  prob,
			Amplitude: &quasarv1.SimulateResponse_Amplitude{
				Real: round(real(s.Amplitude())),
				Imag: round(imag(s.Amplitude())),
			},
		})
	}

	return connect.NewResponse(&quasarv1.SimulateResponse{
		States: states,
	}), nil
}

func (s *QuasarService) Share(
	ctx context.Context,
	req *connect.Request[quasarv1.ShareRequest],
) (resp *connect.Response[quasarv1.ShareResponse], err error) {
	code := req.Msg.Code
	if len(code) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrCodeNotFound)
	}

	if len(code) > maxSize {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("code size exceeds %d bytes", maxSize))
	}

	// put
	id, err := genID(code, 16)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrSomethingWentWrong)
	}

	// Truncate to microsecond precision to match Firestore Timestamp precision.
	createdAt := time.Now().Truncate(time.Microsecond)

	if err := s.Store.Put(ctx, id, &store.Snippet{
		Code:      code,
		CreatedAt: createdAt,
	}); err != nil {
		slog.ErrorContext(ctx, "store put", slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, ErrSomethingWentWrong)
	}

	return connect.NewResponse(&quasarv1.ShareResponse{
		Id:        id,
		CreatedAt: timestamppb.New(createdAt),
	}), nil
}

func (s *QuasarService) Edit(
	ctx context.Context,
	req *connect.Request[quasarv1.EditRequest],
) (resp *connect.Response[quasarv1.EditResponse], err error) {
	id := req.Msg.Id
	if len(id) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrIDNotFound)
	}

	// get
	snippet, err := s.Store.Get(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNoSuchEntity) {
			return nil, connect.NewError(connect.CodeNotFound, ErrNoSuchEntity)
		}

		slog.ErrorContext(ctx, "store get", slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, ErrSomethingWentWrong)
	}

	return connect.NewResponse(&quasarv1.EditResponse{
		Id:        id,
		Code:      snippet.Code,
		CreatedAt: timestamppb.New(snippet.CreatedAt),
	}), nil
}

func (s *QuasarService) Validate(
	ctx context.Context,
	req *connect.Request[quasarv1.ValidateRequest],
) (*connect.Response[quasarv1.ValidateResponse], error) {
	if err := checkCode(req.Msg.Code); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if !s.acquireCompute() {
		return nil, connect.NewError(connect.CodeResourceExhausted, ErrBusy)
	}
	defer func() { <-s.compute }()

	if _, err := parseCode(req.Msg.Code); err != nil {
		if syntaxErr, ok := errors.AsType[*listener.SyntaxError](err); ok {
			return connect.NewResponse(&quasarv1.ValidateResponse{
				Valid:   false,
				Line:    new(int32(syntaxErr.Line)),
				Column:  new(int32(syntaxErr.Column)),
				Message: &syntaxErr.Message,
			}), nil
		}

		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&quasarv1.ValidateResponse{
		Valid: true,
	}), nil
}

func (s *QuasarService) acquireCompute() bool {
	s.computeOnce.Do(func() {
		s.compute = make(chan struct{}, max(1, s.MaxConcurrent))
	})

	select {
	case s.compute <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *QuasarService) logExecution(
	ctx context.Context,
	req *connect.Request[quasarv1.SimulateRequest],
	resp *connect.Response[quasarv1.SimulateResponse],
	runErr error,
	panicValue any,
	startedAt time.Time,
) {
	if s.ExecLogger == nil || s.Revision == "" {
		return
	}

	entry := &execlog.Entry{
		ExecutionID:     rand.Text(),
		StartedAt:       startedAt.UTC().Truncate(time.Microsecond),
		Code:            req.Msg.Code,
		CodeSHA256:      fmt.Sprintf("%x", sha256.Sum256([]byte(req.Msg.Code))),
		DurationMS:      time.Since(startedAt).Milliseconds(),
		ResultCode:      "ok",
		ServiceRevision: s.Revision,
		ClientIP:        clientIP(req.Header().Get("X-Quasar-Client-IP"), req.Peer().Addr),
		UserAgent:       req.Header().Get("User-Agent"),
	}

	switch {
	case panicValue != nil:
		entry.ResultCode = connect.CodeInternal.String()
		entry.ErrorMessage = "unexpected internal error"
	case runErr != nil:
		entry.ResultCode = connect.CodeOf(runErr).String()
		entry.ErrorMessage = runErr.Error()
	case resp != nil:
		// Match Connect's protobuf JSON encoding, including its field names.
		result, err := protojson.Marshal(resp.Msg)
		if err != nil {
			slog.ErrorContext(ctx, "marshal execution result",
				slog.String("execution_id", entry.ExecutionID),
				slog.Any("error", err),
			)

			break
		}

		entry.Result = result
	}

	// Complete the bounded write even if the requesting client disconnects.
	logCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), execLogTimeout)
	defer cancel()

	if err := s.ExecLogger.Log(logCtx, entry); err != nil {
		slog.ErrorContext(logCtx, "write execution log",
			slog.String("execution_id", entry.ExecutionID),
			slog.Any("error", err),
		)
	}
}

func clientIP(forwarded, peer string) string {
	if addr, err := netip.ParseAddr(strings.TrimSpace(forwarded)); err == nil {
		return addr.Unmap().String()
	}

	if addr, err := netip.ParseAddrPort(peer); err == nil {
		return addr.Addr().Unmap().String()
	}

	if addr, err := netip.ParseAddr(peer); err == nil {
		return addr.Unmap().String()
	}

	return ""
}

func genID(code string, length int) (string, error) {
	hash := sha256.New()
	if _, err := io.WriteString(hash, salt); err != nil {
		return "", err
	}
	hash.Write([]byte(code))

	sum := hash.Sum(nil)
	b := make([]byte, base64.URLEncoding.EncodedLen(len(sum)))
	base64.URLEncoding.Encode(b, sum)

	hashLen := length
	for hashLen <= len(b) && b[hashLen-1] == '_' {
		hashLen++
	}

	return string(b)[:hashLen], nil
}

func checkCode(code string) error {
	if len(code) > maxSize {
		return fmt.Errorf("code size exceeds %d bytes", maxSize)
	}

	if len(strings.TrimSpace(code)) == 0 {
		return ErrCodeNotFound
	}

	return nil
}

func parseCode(code string) (parser.IProgramContext, error) {
	lexer := parser.Newqasm3Lexer(antlr.NewInputStream(code))
	lexer.RemoveErrorListeners()

	for token := lexer.NextToken(); token.GetTokenType() != antlr.TokenEOF; token = lexer.NextToken() {
		if lexer.GetSymbolicNames()[token.GetTokenType()] == "INCLUDE" {
			return nil, ErrIncludeNotAllowed
		}
	}

	return xparser.Parse(code)
}
