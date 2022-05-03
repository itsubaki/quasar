package qasm

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Func(c *gin.Context) {
	c.Status(http.StatusOK)
}
