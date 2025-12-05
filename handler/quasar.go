package handler

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"connectrpc.com/connect"
	"github.com/antlr4-go/antlr/v4"
	"github.com/itsubaki/q"
	"github.com/itsubaki/qasm/gen/parser"
	"github.com/itsubaki/qasm/visitor"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	salt    = "quasar salt\n"
	maxSize = 64 * 1024
)

var (
	ErrQubitsNotFound     = errors.New("qubits not found")
	ErrCodeNotFound       = errors.New("code not found")
	ErrIDNotFound         = errors.New("id not found")
	ErrNoSuchEntity       = errors.New("no such entity")
	ErrSomethingWentWrong = errors.New("something went wrong")
)

var (
	_ Store = (*store.MemoryStore)(nil)
	_ Store = (*store.FireStore)(nil)
)

type Store interface {
	Put(ctx context.Context, id string, snippet *store.Snippet) error
	Get(ctx context.Context, id string) (*store.Snippet, error)
}

type QuasarService struct {
	MaxQubits int
	Store     Store
}

func (s *QuasarService) Simulate(
	ctx context.Context,
	req *connect.Request[quasarv1.SimulateRequest],
) (*connect.Response[quasarv1.SimulateResponse], error) {
	lexer := parser.Newqasm3Lexer(antlr.NewInputStream(req.Msg.Code))
	p := parser.Newqasm3Parser(antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel))

	qsim := q.New()
	env := visitor.NewEnviron()
	v := visitor.New(qsim, env,
		visitor.WithMaxQubits(s.MaxQubits),
	)

	if err := v.Run(p.Program()); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if len(env.Qubit) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrQubitsNotFound)
	}

	// quantum state
	var index [][]int
	for _, n := range env.QubitOrder {
		index = append(index, q.Index(env.Qubit[n]...))
	}

	qstate := qsim.Underlying().State(index...)
	states := make([]*quasarv1.SimulateResponse_State, len(qstate))

	// quantum state for json encoding
	truncate := func(v float64, n int) float64 {
		factor := math.Pow(10, float64(n))
		return math.Trunc(v*factor) / factor
	}

	// build response
	for i, s := range qstate {
		binaryString, intValue := make([]string, len(index)), make([]uint64, len(index))
		for j := range index {
			binaryString[j], intValue[j] = s.BinaryString(j), uint64(s.Int(j))
		}

		states[i] = &quasarv1.SimulateResponse_State{
			BinaryString: binaryString,
			Int:          intValue,
			Probability:  truncate(s.Probability(), 6),
			Amplitude: &quasarv1.SimulateResponse_Amplitude{
				Real: truncate(real(s.Amplitude()), 6),
				Imag: truncate(imag(s.Amplitude()), 6),
			},
		}
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
	id, createdAt := GenID(code, 16), time.Now()
	if err := s.Store.Put(ctx, id, &store.Snippet{
		Code:      code,
		CreatedAt: createdAt,
	}); err != nil {
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
		if status.Code(err) == codes.NotFound {
			return nil, connect.NewError(connect.CodeNotFound, ErrNoSuchEntity)
		}

		return nil, connect.NewError(connect.CodeInternal, ErrSomethingWentWrong)
	}

	return connect.NewResponse(&quasarv1.EditResponse{
		Id:        id,
		Code:      snippet.Code,
		CreatedAt: timestamppb.New(snippet.CreatedAt),
	}), nil
}

func GenID(code string, length int) string {
	hash := sha256.New()
	io.WriteString(hash, salt)
	hash.Write([]byte(code))

	sum := hash.Sum(nil)
	b := make([]byte, base64.URLEncoding.EncodedLen(len(sum)))
	base64.URLEncoding.Encode(b, sum)

	hashLen := length
	for hashLen <= len(b) && b[hashLen-1] == '_' {
		hashLen++
	}

	return string(b)[:hashLen]
}
