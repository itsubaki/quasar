package qasm

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/itsubaki/qasm/pkg/evaluator"
	"github.com/itsubaki/qasm/pkg/lexer"
	"github.com/itsubaki/qasm/pkg/parser"
	"github.com/itsubaki/quasar/pkg/logger"
	"go.opentelemetry.io/otel"
)

var (
	logf = logger.Factory
	tra  = otel.Tracer("handler/qasm")
)

func Func(c *gin.Context) {
	traceID := c.GetString("trace_id")
	spanID := c.GetString("span_id")

	// logger
	log := logf.New(traceID, c.Request)

	// file upload
	file, err := c.FormFile("file")
	if err != nil {
		log.SpanOf(spanID).ErrorReport("form file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message":  "something went wrong",
			"trace_id": traceID,
		})
		return
	}

	log.SpanOf(spanID).Debug("filename=%v", file.Filename)

	f, err := file.Open()
	if err != nil {
		log.SpanOf(spanID).ErrorReport("file open: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message":  "something went wrong",
			"trace_id": traceID,
		})
		return
	}

	r, err := io.ReadAll(f)
	if err != nil {
		log.SpanOf(spanID).ErrorReport("read all: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message":  "something went wrong",
			"trace_id": traceID,
		})
		return
	}

	log.SpanOf(spanID).Debug("content=%v", string(r))

	// compute
	l := lexer.New(strings.NewReader(string(r)))
	p := parser.New(l)

	a := p.Parse()
	if errs := p.Errors(); len(errs) != 0 {
		log.SpanOf(spanID).ErrorReport("parse: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message":  "something went wrong",
			"trace_id": traceID,
		})
	}

	log.SpanOf(spanID).Debug("ast=%v", a)

	e := evaluator.Default()
	if err := e.Eval(a); err != nil {
		log.SpanOf(spanID).ErrorReport("eval: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message":  "something went wrong",
			"trace_id": traceID,
		})
	}

	log.SpanOf(spanID).Debug("state=%v", e.Q.State())

	// response
	state := make([]State, 0, len(e.Q.State()))
	for _, s := range e.Q.State() {
		state = append(state, State{
			Amplitude: Amplitude{
				Real: real(s.Amplitude),
				Imag: imag(s.Amplitude),
			},
			Probability:  s.Probability,
			Int:          s.Int,
			BinaryString: s.BinaryString,
		})
	}

	c.JSON(http.StatusOK, Response{
		TraceID:  traceID,
		Filename: file.Filename,
		Content:  string(r),
		State:    state,
	})
}
