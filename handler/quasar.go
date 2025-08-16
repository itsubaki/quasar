package handler

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/antlr4-go/antlr/v4"
	"github.com/itsubaki/q"
	"github.com/itsubaki/qasm/gen/parser"
	"github.com/itsubaki/qasm/visitor"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
)

type QuasarService struct{}

func (s *QuasarService) Simulate(
	ctx context.Context,
	req *connect.Request[quasarv1.SimulateRequest],
) (resp *connect.Response[quasarv1.SimulateResponse], err error) {
	defer func() {
		if r := recover(); r != nil {
			err = connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%v", r))
		}
	}()

	lexer := parser.Newqasm3Lexer(antlr.NewInputStream(req.Msg.Code))
	p := parser.Newqasm3Parser(antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel))

	qsim := q.New()
	env := visitor.NewEnviron()

	v := visitor.New(qsim, env)
	if err := v.Run(p.Program()); err != nil {
		switch {
		case
			errors.Is(err, visitor.ErrAlreadyDeclared),
			errors.Is(err, visitor.ErrIdentifierNotFound),
			errors.Is(err, visitor.ErrQubitNotFound),
			errors.Is(err, visitor.ErrClassicalBitNotFound),
			errors.Is(err, visitor.ErrVariableNotFound),
			errors.Is(err, visitor.ErrFunctionNotFound),
			errors.Is(err, visitor.ErrUnexpected),
			errors.Is(err, visitor.ErrNotImplemented):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// quantum state
	qstate := qsim.Underlying().State()

	// quantum state for json encoding
	state := make([]*quasarv1.SimulateResponse_State, len(qstate))
	for i, s := range qstate {
		state[i] = &quasarv1.SimulateResponse_State{
			Probability: s.Probability(),
			Amplitude: &quasarv1.SimulateResponse_Amplitude{
				Real: real(s.Amplitude()),
				Imag: imag(s.Amplitude()),
			},
			Int: []uint64{
				uint64(s.Int()),
			},
			BinaryString: []string{
				s.BinaryString(),
			},
		}
	}

	return connect.NewResponse(&quasarv1.SimulateResponse{
		State: state,
	}), nil
}
