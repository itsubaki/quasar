package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/itsubaki/q"
	"github.com/itsubaki/q/pkg/math/number"
	"github.com/itsubaki/q/pkg/math/rand"
)

func Shor(c *gin.Context) {
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
			"message": fmt.Sprintf("N=%v. N must be integer.", N),
		})
		return
	}

	t, err := strconv.Atoi(tstr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("t=%v. t must be integer.", t),
		})
		return
	}

	a, err := strconv.Atoi(astr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("a=%v. a must be integer.", a),
		})
		return
	}

	// number check
	if N < 2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("N=%d. N must be greater than 1.", N),
		})
		return
	}

	if number.IsPrime(N) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("N=%d is prime.", N),
		})
		return
	}

	if number.IsEven(N) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("N=%d is even. p=%d, q=%d.", N, 2, N/2),
		})
		return
	}

	if a, b, ok := number.BaseExp(N); ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("N=%d. N is exponentiation. %d^%d.", N, a, b),
		})
		return
	}

	if a < 0 {
		a = rand.Coprime(N)
	}

	if a < 2 || a > N-1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("N=%d, a=%d. a must be 1 < a < N.", N, a),
		})
		return
	}

	if number.GCD(N, a) != 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("N=%d, a=%d. a is not coprime. a is non-trivial factor.", N, a),
		})
		return
	}

	// quantum algorithm
	qsim := q.New()
	r0 := qsim.ZeroWith(t)
	r1 := qsim.ZeroLog2(N)

	qsim.X(r1[len(r1)-1])
	qsim.H(r0...)
	qsim.CModExp2(a, N, r0, r1)
	qsim.InvQFT(r0...)

	for _, state := range qsim.State(r0) {
		_, m := state.Value()
		_, r, _, ok := number.FindOrder(a, N, fmt.Sprintf("0.%s", m))
		if !ok || number.IsOdd(r) {
			continue
		}

		p0 := number.GCD(number.Pow(a, r/2)-1, N)
		p1 := number.GCD(number.Pow(a, r/2)+1, N)
		if number.IsTrivial(N, p0, p1) {
			continue
		}

		c.JSON(http.StatusOK, gin.H{
			"N": N, "a": a, "t": t,
			"p": p0,
			"q": p1,
		})
		return
	}
}
