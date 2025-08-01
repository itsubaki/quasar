package handler

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/antlr4-go/antlr/v4"
	"github.com/itsubaki/q"
	"github.com/itsubaki/q/math/number"
	"github.com/itsubaki/q/math/rand"
	"github.com/itsubaki/q/quantum/qubit"
	"github.com/itsubaki/qasm/gen/parser"
	"github.com/itsubaki/qasm/visitor"
	"github.com/itsubaki/quasar/gate"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
)

type QuasarService struct{}

func (s *QuasarService) Factorize(
	ctx context.Context,
	req *connect.Request[quasarv1.FactorizeRequest],
) (*connect.Response[quasarv1.FactorizeResponse], error) {
	N := int(req.Msg.N)
	a := int(defaultValue(req.Msg.A, 0))
	t := int(defaultValue(req.Msg.T, 3))
	seed := defaultValue(req.Msg.Seed, 0)

	slog.DebugContext(ctx, "parameter",
		slog.Int("N", N),
		slog.Int("a", a),
		slog.Int("t", t),
		slog.Uint64("seed", seed),
	)

	if msg, ok := func() (string, bool) {
		if N < 2 {
			return fmt.Sprintf("N=%d. N must be greater than 1.", N), true
		}

		if number.IsEven(N) {
			return fmt.Sprintf("N=%d is even. p=%d, q=%d.", N, 2, N/2), true
		}

		if a, b, ok := number.BaseExp(N); ok {
			return fmt.Sprintf("N=%d. N is exponentiation. %d^%d.", N, a, b), true
		}

		if number.IsPrime(N) {
			return fmt.Sprintf("N=%d is prime.", N), true
		}

		if a < 1 {
			a = rand.Coprime(N)
		}

		if a < 2 || a > N-1 {
			return fmt.Sprintf("N=%d, a=%d. a must be 1 < a < N.", N, a), true
		}

		if number.GCD(N, a) != 1 {
			return fmt.Sprintf("N=%d, a=%d. a is not coprime. a is non-trivial factor.", N, a), true
		}

		return "", false
	}(); ok {
		return connect.NewResponse(&quasarv1.FactorizeResponse{
			Message: &msg,
		}), nil
	}

	qs, err := func() ([]qubit.State, error) {
		qsim := q.New()
		if seed > 0 {
			qsim.Rand = rand.Const(seed)
		}

		r0 := qsim.Zeros(t)
		r1 := qsim.ZeroLog2(N)

		qsim.X(r1[len(r1)-1])
		qsim.H(r0...)
		gate.CModExp2(qsim, a, N, r0, r1)
		qsim.InvQFT(r0...)
		qsim.Measure()

		s0 := qsim.State(r0)
		if len(s0) != 1 {
			return nil, fmt.Errorf("len(qsim.State(r0)) must be 1. qsim.State(r0)=%v", s0)
		}

		return s0, nil
	}()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("quantum algorithm: %w", err))
	}

	m := fmt.Sprintf("0.%s", qs[0].BinaryString())
	ss, r, _, ok := number.FindOrder(a, N, m)
	if !ok || number.IsOdd(r) {
		return connect.NewResponse(&quasarv1.FactorizeResponse{
			N:    uint64(N),
			A:    uint64(a),
			T:    uint64(t),
			Seed: seed,
			M:    m,
			S:    uint64(ss),
			R:    uint64(r),
		}), nil
	}

	p0 := number.GCD(number.Pow(a, r/2)-1, N)
	p1 := number.GCD(number.Pow(a, r/2)+1, N)
	if number.IsTrivial(N, p0, p1) {
		return connect.NewResponse(&quasarv1.FactorizeResponse{
			N:    uint64(N),
			A:    uint64(a),
			T:    uint64(t),
			Seed: seed,
			M:    m,
			S:    uint64(ss),
			R:    uint64(r),
		}), nil
	}

	return connect.NewResponse(&quasarv1.FactorizeResponse{
		N:    uint64(N),
		A:    uint64(a),
		T:    uint64(t),
		Seed: seed,
		M:    m,
		S:    uint64(ss),
		R:    uint64(r),
		P:    uint64(p0),
		Q:    uint64(p1),
	}), nil
}

func (s *QuasarService) Simulate(
	ctx context.Context,
	req *connect.Request[quasarv1.SimulateRequest],
) (*connect.Response[quasarv1.SimulateResponse], error) {
	lexer := parser.Newqasm3Lexer(antlr.NewInputStream(req.Msg.Code))
	p := parser.Newqasm3Parser(antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel))

	qsim := q.New()
	env := visitor.NewEnviron()

	v := visitor.New(qsim, env)
	if err := v.Run(p.Program()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("visitor run: %w", err))
	}

	// quantum state
	var index [][]int
	for _, qb := range env.Qubit {
		index = append(index, q.Index(qb...))
	}

	qstate := qsim.Underlying().State(index...)

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

func defaultValue[T any](v *T, w T) T {
	if v != nil {
		return *v
	}

	return w
}
