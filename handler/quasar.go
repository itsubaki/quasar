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

	"cloud.google.com/go/firestore"
	"connectrpc.com/connect"
	"github.com/antlr4-go/antlr/v4"
	"github.com/itsubaki/q"
	"github.com/itsubaki/qasm/gen/parser"
	"github.com/itsubaki/qasm/visitor"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
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
	ErrSomethingWentWrong = errors.New("something went wrong")
)

type QuasarService struct {
	MaxQubits int
	Firestore *firestore.Client
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

	// quantum state for json encoding
	truncate := func(v float64, n int) float64 {
		factor := math.Pow(10, float64(n))
		return math.Trunc(v*factor) / factor
	}

	states := make([]*quasarv1.SimulateResponse_State, len(qstate))
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

	// id
	id, err := GenID(code, 16)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrSomethingWentWrong)
	}

	// save to firestore
	createdAt := time.Now()
	if _, err := s.Firestore.Collection("qasm").Doc(id).Set(ctx, map[string]any{
		"code":       code,
		"created_at": createdAt,
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

	// load from firestore
	ref, err := s.Firestore.Collection("qasm").Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, connect.NewError(connect.CodeNotFound, ErrIDNotFound)
		}

		return nil, connect.NewError(connect.CodeInternal, ErrSomethingWentWrong)
	}

	code, err := Get[string](ref.Data(), "code")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	createdAt, err := Get[time.Time](ref.Data(), "created_at")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&quasarv1.EditResponse{
		Id:        id,
		Code:      code,
		CreatedAt: timestamppb.New(createdAt),
	}), nil
}

func GenID(code string, length int) (string, error) {
	hash := sha256.New()
	if _, err := io.WriteString(hash, salt); err != nil {
		return "", fmt.Errorf("write salt: %w", err)
	}
	if _, err := hash.Write([]byte(code)); err != nil {
		return "", fmt.Errorf("write code: %w", err)
	}

	sum := hash.Sum(nil)
	b := make([]byte, base64.URLEncoding.EncodedLen(len(sum)))
	base64.URLEncoding.Encode(b, sum)

	hashLen := length
	for hashLen <= len(b) && b[hashLen-1] == '_' {
		hashLen++
	}

	return string(b)[:hashLen], nil
}

func Get[T any](data map[string]any, key string) (T, error) {
	v, ok := data[key]
	if !ok {
		var zero T
		return zero, fmt.Errorf("field %s not found", key)
	}

	typed, ok := v.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("invalid type(%T)", v)
	}

	return typed, nil
}
