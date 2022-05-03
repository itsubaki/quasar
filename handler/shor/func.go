package shor

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/itsubaki/q"
	"github.com/itsubaki/q/pkg/math/number"
	"github.com/itsubaki/q/pkg/math/rand"
	"github.com/itsubaki/quasar/tracer"
	"go.opentelemetry.io/otel"
)

var tra = otel.Tracer("handler/shor")

func Func(c *gin.Context) {
	traceID := c.GetString("trace_id")
	spanID := c.GetString("span_id")
	traceTrue := c.GetBool("trace_true")

	ctx := context.Background()
	parent, err := tracer.NewContext(ctx, traceID, spanID, traceTrue)
	if err != nil {
		log.Printf("new context. traceID=%v spanID=%v: %v", traceID, spanID, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	Nstr := c.Param("N")
	tstr := c.Query("t")
	astr := c.Query("a")

	// set default value
	if tstr == "" {
		tstr = "3"
	}

	if astr == "" {
		astr = "-1"
	}

	// validation
	N, err := strconv.Atoi(Nstr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("N=%v. N must be integer.", Nstr),
		})
		return
	}

	t, err := strconv.Atoi(tstr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("t=%v. t must be integer.", tstr),
		})
		return
	}

	a, err := strconv.Atoi(astr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("a=%v. a must be integer.", astr),
		})
		return
	}

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
		}

		if a < 2 || a > N-1 {
			return fmt.Sprintf("N=%d, a=%d. a must be 1 < a < N.", N, a), true
		}

		if number.GCD(N, a) != 1 {
			return fmt.Sprintf("N=%d, a=%d. a is not coprime. a is non-trivial factor.", N, a), true
		}

		return "", false
	}(); ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": msg,
		})
		return
	}

	// quantum algorithm
	qsim := q.New()
	r0 := func() []q.Qubit {
		_, s := tra.Start(parent, "qsim.ZeroWith(t)")
		defer s.End()

		return qsim.ZeroWith(t)
	}()

	r1 := func() []q.Qubit {
		_, s := tra.Start(parent, "qsim.ZeroLog2(N)")
		defer s.End()

		return qsim.ZeroLog2(N)
	}()

	Span(parent, "qsim.X(r1[len(r1)-1])", func() { qsim.X(r1[len(r1)-1]) })
	Span(parent, "qsim.H(r0...)", func() { qsim.H(r0...) })
	Span(parent, "qsim.CModExp2(a, N, r0, r1)", func() { qsim.CModExp2(a, N, r0, r1) })
	Span(parent, "qsim.InvQFT(r0...)", func() { qsim.InvQFT(r0...) })
	Span(parent, "qsim.Measure(r0...)", func() { qsim.Measure(r0...) })

	out := func() gin.H {
		_, s := tra.Start(parent, "find non-trivial factors")
		defer s.End()

		for _, state := range qsim.State(r0) {
			_, m := state.Value()
			s, r, _, ok := number.FindOrder(a, N, fmt.Sprintf("0.%s", m))
			if !ok || number.IsOdd(r) {
				continue
			}

			p0 := number.GCD(number.Pow(a, r/2)-1, N)
			p1 := number.GCD(number.Pow(a, r/2)+1, N)
			if number.IsTrivial(N, p0, p1) {
				continue
			}

			return gin.H{
				"N": N, "a": a, "t": t,
				"m":   fmt.Sprintf("0.%s", m),
				"s/r": fmt.Sprintf("%v/%v", s, r),
				"p":   p0,
				"q":   p1,
			}
		}

		return gin.H{
			"N": N, "a": a, "t": t,
		}
	}()

	c.JSON(http.StatusOK, out)
}

func Span(parent context.Context, spanName string, f func()) {
	_, s := tra.Start(parent, spanName)
	defer s.End()

	f()
}
