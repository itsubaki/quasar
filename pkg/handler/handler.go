package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/itsubaki/quasar/pkg/handler/qasm"
	"github.com/itsubaki/quasar/pkg/handler/shor"
	"github.com/itsubaki/quasar/pkg/logger"
)

var logf = logger.Factory

func New() *gin.Engine {
	g := gin.New()

	g.Use(SetTraceID)

	g.Use(gin.Recovery())
	if gin.IsDebugging() {
		g.Use(gin.Logger())
	}

	Root(g)
	Status(g)

	g.GET("/shor/:N", shor.Func)
	g.POST("/", qasm.Func)

	return g
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
		c.Next()
		return
	}

	// https://cloud.google.com/trace/docs/setup
	// The header specification is:
	// "X-Cloud-Trace-Context: TRACE_ID/SPAN_ID;o=TRACE_TRUE"
	ids := strings.Split(strings.Split(value, ";")[0], "/")
	c.Set("trace_id", ids[0])

	// https://cloud.google.com/trace/docs/setup
	// SPAN_ID is the decimal representation of the (unsigned) span ID.
	i, err := strconv.ParseUint(ids[1], 10, 64)
	if err != nil {
		logf.New(ids[0], c.Request).ErrorReport("parse %v: %v", ids[1], err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/api.md#retrieving-the-traceid-and-spanid
	// MUST be a 16-hex-character lowercase string
	c.Set("span_id", fmt.Sprintf("%016x", i))

	// https://cloud.google.com/trace/docs/setup
	// TRACE_TRUE must be 1 to trace this request. Specify 0 to not trace the request.
	c.Set("trace_true", false)
	if len(strings.Split(value, ";")) > 1 && strings.Split(value, ";")[1] == "o=1" {
		c.Set("trace_true", true)
	}

	c.Next()
}
