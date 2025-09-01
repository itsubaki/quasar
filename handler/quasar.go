package handler

import (
	"context"
	"math"

	"connectrpc.com/connect"
	"github.com/antlr4-go/antlr/v4"
	"github.com/itsubaki/q"
	"github.com/itsubaki/qasm/gen/parser"
	"github.com/itsubaki/qasm/visitor"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
)

type QuasarService struct {
	MaxQubits int
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
	state := make([]*quasarv1.SimulateResponse_State, len(qstate))
	for i, s := range qstate {
		binaryString, intValue := make([]string, len(index)), make([]uint64, len(index))
		for j := range index {
			binaryString[j], intValue[j] = s.BinaryString(j), uint64(s.Int(j))
		}

		state[i] = &quasarv1.SimulateResponse_State{
			BinaryString: binaryString,
			Int:          intValue,
			Probability: truncate(s.Probability(), 6),
			Amplitude: &quasarv1.SimulateResponse_Amplitude{
				Real: truncate(real(s.Amplitude()), 6),
				Imag: truncate(imag(s.Amplitude()), 6),
			},
		}
	}

	return connect.NewResponse(&quasarv1.SimulateResponse{
		State: state,
	}), nil
}

func truncate(v float64, n int) float64 {
	factor := math.Pow(10, float64(n))
	return math.Trunc(v*factor) / factor
}
