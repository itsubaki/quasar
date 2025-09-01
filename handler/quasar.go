package handler

import (
	"context"
	"fmt"
	"math"

	"cloud.google.com/go/firestore"
	"connectrpc.com/connect"
	"github.com/antlr4-go/antlr/v4"
	"github.com/itsubaki/q"
	"github.com/itsubaki/qasm/gen/parser"
	"github.com/itsubaki/qasm/visitor"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
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
	if len(req.Msg.Code) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("code is empty"))
	}

	ref, _, err := s.Firestore.Collection("qasm").Add(ctx, map[string]any{
		"code": req.Msg.Code,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}

	return connect.NewResponse(&quasarv1.SaveResponse{
		Id: ref.ID,
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
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}

	code, ok := ref.Data()["code"]
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("code is empty"))
	}

	scode, ok := code.(string)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid type(%T)", code))
	}

	return connect.NewResponse(&quasarv1.LoadResponse{
		Code: scode,
	}), nil
}
