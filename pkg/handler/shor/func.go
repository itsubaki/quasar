package shor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/itsubaki/q"
	"github.com/itsubaki/q/pkg/math/number"
	"github.com/itsubaki/q/pkg/math/rand"
	"github.com/itsubaki/q/pkg/quantum/qubit"
	"github.com/itsubaki/quasar/pkg/logger"
	"github.com/itsubaki/quasar/pkg/tracer"
	"go.opentelemetry.io/otel"
)

var (
	logf = logger.Factory
	tra  = otel.Tracer("handler/shor")
)

func Func(c *gin.Context) {
	traceID := c.GetString("trace_id")
	spanID := c.GetString("span_id")
	traceTrue := c.GetBool("trace_true")

	// logger and tracer
	log := logf.New(traceID, c.Request)
	parent, err := tracer.NewContext(context.Background(), traceID, spanID, traceTrue)
	if err != nil {
		log.SpanOf(spanID).ErrorReport("new context: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":  "something went wrong",
			"trace_id": traceID,
		})
		return
	}

	// inputs
	Nq := c.Param("N")
	tq := DefaultValue(c.Query("t"), "3")
	aq := DefaultValue(c.Query("a"), "-1")
	sq := DefaultValue(c.Query("seed"), "-1")

	log.SpanOf(spanID).Debug("param(N)=%v, query(a)=%v, query(t)=%v, query(seed)=%v", Nq, aq, tq, sq)

	// validation
	N, err := strconv.Atoi(Nq)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("N=%v. N must be integer.", Nq),
		})
		return
	}

	t, err := strconv.Atoi(tq)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("t=%v. t must be integer.", tq),
		})
		return
	}

	a, err := strconv.Atoi(aq)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("a=%v. a must be integer.", aq),
		})
		return
	}

	seed, err := strconv.Atoi(sq)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("seed=%v. seed must be integer.", sq),
		})
		return
	}

	// primality test
	if msg, ok := func() (string, bool) {
		_, s := tra.Start(parent, "primality test")
		defer s.End()

		if N < 2 {
			return fmt.Sprintf("N=%d. N must be greater than 1.", N), true
		}

		if number.IsPrime(N) {
			return fmt.Sprintf("N=%d is prime.", N), true
		}

		if number.IsEven(N) {
			return fmt.Sprintf("N=%d is even. p=%d, q=%d.", N, 2, N/2), true
		}

		if a, b, ok := number.BaseExp(N); ok {
			return fmt.Sprintf("N=%d. N is exponentiation. %d^%d.", N, a, b), true
		}

		if a < 0 {
			a = rand.Coprime(N)
			log.Span(s).Debug("rand.Coprime(%v)=%v", N, a)
		}

		if a < 2 || a > N-1 {
			return fmt.Sprintf("N=%d, a=%d. a must be 1 < a < N.", N, a), true
		}

		if number.GCD(N, a) != 1 {
			return fmt.Sprintf("N=%d, a=%d. a is not coprime. a is non-trivial factor.", N, a), true
		}

		return "", false
	}(); ok {
		c.JSON(http.StatusOK, gin.H{
			"message": msg,
		})
		return
	}

	// quantum algorithm
	qs, err := func() ([]qubit.State, error) {
		qa, s := tra.Start(parent, "quantum algorithm")
		defer s.End()

		log.Span(s).Debug("N=%v, a=%v, t=%v, seed=%v", N, a, t, seed)

		qsim := q.New()
		if seed > 0 {
			qsim.Seed = []int{seed}
			qsim.Rand = rand.Math
			log.Span(s).Debug("set seed=%v", seed)
		}

		r0 := func() []q.Qubit {
			_, s := tra.Start(qa, "qsim.ZeroWith(t)")
			defer s.End()

			return qsim.ZeroWith(t)
		}()

		r1 := func() []q.Qubit {
			_, s := tra.Start(qa, "qsim.ZeroLog2(N)")
			defer s.End()

			return qsim.ZeroLog2(N)
		}()

		Span(qa, "qsim.X(r1[len(r1)-1])", func() { qsim.X(r1[len(r1)-1]) })
		Span(qa, "qsim.H(r0...)", func() { qsim.H(r0...) })
		Span(qa, "qsim.CModExp2(a, N, r0, r1)", func() { qsim.CModExp2(a, N, r0, r1) })
		Span(qa, "qsim.InvQFT(r0...)", func() { qsim.InvQFT(r0...) })
		Span(qa, "qsim.Measure()", func() { qsim.Measure() })

		if len(qsim.State(r0)) != 1 {
			return nil, fmt.Errorf("len(qsim.State(r0)) must be 1. qsim.State(r0)=%v", qsim.State(r0))
		}

		return qsim.State(r0), nil
	}()
	if err != nil {
		log.SpanOf(spanID).ErrorReport("quantum algorithm: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":  "something went wrong",
			"trace_id": traceID,
		})
		return
	}

	// find non-trivial factors (classical algorithm)
	out, ok := func() (gin.H, bool) {
		_, s := tra.Start(parent, "find non-trivial factors")
		defer s.End()

		for _, state := range qs {
			_, m := state.Value()
			s, r, _, ok := number.FindOrder(a, N, fmt.Sprintf("0.%s", m))
			if !ok || number.IsOdd(r) {
				return gin.H{
					"N": N, "a": a, "t": t,
					"m":   fmt.Sprintf("0.%s", m),
					"s/r": fmt.Sprintf("%v/%v", s, r),
				}, false
			}

			p0 := number.GCD(number.Pow(a, r/2)-1, N)
			p1 := number.GCD(number.Pow(a, r/2)+1, N)
			if number.IsTrivial(N, p0, p1) {
				return gin.H{
					"N": N, "a": a, "t": t,
					"m":   fmt.Sprintf("0.%s", m),
					"s/r": fmt.Sprintf("%v/%v", s, r),
				}, false
			}

			return gin.H{
				"N": N, "a": a, "t": t,
				"m":   fmt.Sprintf("0.%s", m),
				"s/r": fmt.Sprintf("%v/%v", s, r),
				"p":   p0,
				"q":   p1,
			}, true
		}

		return gin.H{
			"N": N, "a": a, "t": t,
		}, false
	}()

	log.SpanOf(spanID).Debug("out: %v, ok: %v", out, ok)
	c.JSON(http.StatusOK, out)
}

func DefaultValue(v, w string) string {
	if v == "" {
		return w
	}

	return v
}

func Span(parent context.Context, spanName string, f func()) {
	_, s := tra.Start(parent, spanName)
	defer s.End()

	f()
}
