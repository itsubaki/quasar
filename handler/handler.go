package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsubaki/quasar/handler/qasm"
	"github.com/itsubaki/quasar/handler/shor"
)

func New() *gin.Engine {
	g := gin.New()

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
