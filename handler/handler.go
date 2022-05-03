package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/itsubaki/quasar/handler/qasm"
	"github.com/itsubaki/quasar/handler/shor"
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

	g.GET("/shor/:N", shor.Func)
	g.POST("/qasm", qasm.Func)

	return g
}

func Root(g *gin.Engine) {
	g.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
}

func Status(g *gin.Engine) {
	g.GET("/status", func(c *gin.Context) {
		c.Status(http.StatusOK)
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
	c.Set("span_id", ids[1])

	c.Set("trace_true", false)
	if len(strings.Split(value, ";")) > 1 && strings.Split(value, ";")[1] == "o=1" {
		c.Set("trace_true", true)
	}

	c.Next()
}
