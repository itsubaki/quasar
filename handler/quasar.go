package handler

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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

type QuasarService struct {
	MaxQubits int
	Firestore *firestore.Client
}

func (s *QuasarService) Simulate(
	ctx context.Context,
	req *connect.Request[quasarv1.SimulateRequest],
) (resp *connect.Response[quasarv1.SimulateResponse], err error) {
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

func (s *QuasarService) Save(
	ctx context.Context,
	req *connect.Request[quasarv1.SaveRequest],
) (resp *connect.Response[quasarv1.SaveResponse], err error) {
	code := req.Msg.Code
	if len(code) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("code is empty"))
	}

	hash := sha256.Sum256([]byte(code))
	id := base64.RawURLEncoding.EncodeToString(hash[:])[:16]

	createdAt := time.Now()
	if _, err := s.Firestore.Collection("qasm").Doc(id).Set(ctx, map[string]any{
		"code":       code,
		"created_at": createdAt,
	}); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("something went wrong"))
	}

	return connect.NewResponse(&quasarv1.SaveResponse{
		Id:        id,
		CreatedAt: timestamppb.New(createdAt),
	}), nil
}

func (s *QuasarService) Load(
	ctx context.Context,
	req *connect.Request[quasarv1.LoadRequest],
) (resp *connect.Response[quasarv1.LoadResponse], err error) {
	id := req.Msg.Id
	if len(id) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is empty"))
	}

	ref, err := s.Firestore.Collection("qasm").Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("id=%v not found", id))
		}

		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("something went wrong"))
	}

	code, ok := ref.Data()["code"]
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("code is empty"))
	}

	scode, ok := code.(string)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid type(%T)", code))
	}

	createdAt, ok := ref.Data()["created_at"]
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("created_at is empty"))
	}

	tcreatedAt, ok := createdAt.(time.Time)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid type(%T)", createdAt))
	}

	return connect.NewResponse(&quasarv1.LoadResponse{
		Id:        id,
		Code:      scode,
		CreatedAt: timestamppb.New(tcreatedAt),
	}), nil
}
