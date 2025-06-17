package handler

import (
	"fmt"
	"math/rand/v2"
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/itsubaki/logger"
	"github.com/itsubaki/quasar/handler/factorize"
	"github.com/itsubaki/quasar/handler/qasm"
	"github.com/itsubaki/tracer"
)

func New() *gin.Engine {
	g := gin.New()

	g.Use(SetTraceID)

	g.Use(gin.Recovery())
	if gin.IsDebugging() {
		g.Use(gin.Logger())
	}

	Root(g)
	Status(g)

	g.GET("/factorize/:N", factorize.Func)
	g.POST("/", qasm.Func)

	return g
}

func UsePProf(g *gin.Engine) {
	pprof.Register(g)
}

func Root(g *gin.Engine) {
	g.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ok": true,
		})
	})
}

func Status(g *gin.Engine) {
	g.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ok": true,
		})
	})
}

func SetTraceID(c *gin.Context) {
	value := c.GetHeader("X-Cloud-Trace-Context")
	if value == "" {
		// new trace id, span id for test
		value = fmt.Sprintf("%016x%016x/%d;o=0", rand.Int64(), rand.Int64(), rand.Int64())
		c.Request.Header.Add("X-Cloud-Trace-Context", value)
	}

	xc, err := tracer.Parse(value)
	if err != nil {
		logger.New(c.Request, "", "").ErrorReport("parse %v: %v", value, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Set("trace_id", xc.TraceID)
	c.Set("span_id", xc.SpanID)
	c.Set("trace_true", xc.TraceTrue)

	c.Next()
}
