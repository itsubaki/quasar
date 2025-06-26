package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/itsubaki/q"
	"github.com/itsubaki/q/math/number"
	"github.com/itsubaki/q/math/rand"
	"github.com/itsubaki/q/quantum/qubit"
	qctx "github.com/itsubaki/quasar/context"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var ErrTraceNotFound = errors.New("trace not found")

type QuasarService struct{}

func (s *QuasarService) Factorize(
	ctx context.Context,
	req *connect.Request[quasarv1.FactorizeRequest],
) (*connect.Response[quasarv1.FactorizeResponse], error) {
	trace, _ := qctx.GetTrace(ctx)
	parent, err := tracer.Context(ctx, trace.TraceID, trace.SpanID, trace.TraceTrue)
	if err != nil {
		slog.ErrorContext(ctx, "new context", slog.Any("trace", trace), slog.Any("error", err))
		return nil, connect.NewError(connect.CodeInternal, ErrTraceNotFound)
	}

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

	var tr = otel.Tracer("quasar/factorize")
	if msg, ok := func() (string, bool) {
		_, s := tr.Start(parent, "primality test")
		defer s.End()

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
		qa, s := tr.Start(parent, "quantum algorithm")
		defer s.End()

		qsim := q.New()
		if seed > 0 {
			qsim.Rand = rand.Const(seed)
		}

		r0 := func() []q.Qubit {
			_, s := tr.Start(qa, "qsim.Zeros(t)")
			defer s.End()

			return qsim.Zeros(t)
		}()

		r1 := func() []q.Qubit {
			_, s := tr.Start(qa, "qsim.ZeroLog2(N)")
			defer s.End()

			return qsim.ZeroLog2(N)
		}()

		span(qa, tr, "qsim.X(r1[len(r1)-1])", func() { qsim.X(r1[len(r1)-1]) })
		span(qa, tr, "qsim.H(r0...)", func() { qsim.H(r0...) })
		span(qa, tr, "qsim.CModExp2(a, N, r0, r1)", func() { qsim.CModExp2(a, N, r0, r1) })
		span(qa, tr, "qsim.InvQFT(r0...)", func() { qsim.InvQFT(r0...) })
		span(qa, tr, "qsim.Measure()", func() { qsim.Measure() })

		s0 := qsim.State(r0)
		if len(s0) != 1 {
			return nil, fmt.Errorf("len(qsim.State(r0)) must be 1. qsim.State(r0)=%v", s0)
		}

		return s0, nil
	}()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("quantum algorithm: %w", err))
	}

	_, span := tr.Start(parent, "find non-trivial factor")
	defer span.End()

	m := qs[0].BinaryString()
	ss, r, _, ok := number.FindOrder(a, N, fmt.Sprintf("0.%s", m))
	if !ok || number.IsOdd(r) {
		return connect.NewResponse(&quasarv1.FactorizeResponse{
			N:    uint64(N),
			A:    uint64(a),
			T:    uint64(t),
			Seed: uint64(seed),
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
			Seed: uint64(seed),
			M:    m,
			S:    uint64(ss),
			R:    uint64(r),
		}), nil
	}

	return connect.NewResponse(&quasarv1.FactorizeResponse{
		N:    uint64(N),
		A:    uint64(a),
		T:    uint64(t),
		Seed: uint64(seed),
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
	return connect.NewResponse(&quasarv1.SimulateResponse{}), nil
}

func defaultValue[T any](v *T, w T) T {
	if v != nil {
		return *v
	}

	return w
}

func span(parent context.Context, tr trace.Tracer, spanName string, f func()) {
	_, s := tr.Start(parent, spanName)
	defer s.End()

	f()
}
