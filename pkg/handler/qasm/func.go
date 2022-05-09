package qasm

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/itsubaki/q"
	"github.com/itsubaki/qasm/pkg/ast"
	"github.com/itsubaki/qasm/pkg/evaluator"
	"github.com/itsubaki/qasm/pkg/lexer"
	"github.com/itsubaki/qasm/pkg/parser"
	"github.com/itsubaki/quasar/pkg/logger"
	"github.com/itsubaki/quasar/pkg/tracer"
	"go.opentelemetry.io/otel"
)

var (
	logf = logger.Factory
	tra  = otel.Tracer("handler/qasm")
)

func Func(c *gin.Context) {
	traceID := c.GetString("trace_id")
	spanID := c.GetString("span_id")
	traceTrue := c.GetBool("trace_true")

	// logger, tracer
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

	// file read
	file, r, err := func() (*multipart.FileHeader, []byte, error) {
		_, s := tra.Start(parent, "file read")
		defer s.End()

		file, err := c.FormFile("file")
		if err != nil {
			return nil, nil, fmt.Errorf("form file: %v", err)
		}

		f, err := file.Open()
		if err != nil {
			return nil, nil, fmt.Errorf("file open: %v", err)
		}

		r, err := io.ReadAll(f)
		if err != nil {
			return nil, nil, fmt.Errorf("read all: %v", err)
		}

		return file, r, nil
	}()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":  "something went wrong",
			"trace_id": traceID,
		})
		return
	}

	// compute
	state, err := func() ([]State, error) {
		_, s := tra.Start(parent, "compute")
		defer s.End()

		// eval
		l := lexer.New(strings.NewReader(string(r)))
		p := parser.New(l)

		a := p.Parse()
		if errs := p.Errors(); len(errs) != 0 {
			return nil, fmt.Errorf("parse: %v", err)
		}

		e := evaluator.Default()
		if err := e.Eval(a); err != nil {
			return nil, fmt.Errorf("eval: %v", err)
		}

		// quantum state index
		var index [][]int
		for _, n := range e.Env.Qubit.Name {
			qb, ok := e.Env.Qubit.Get(&ast.IdentExpr{Name: n})
			if !ok {
				return nil, fmt.Errorf("qubit(%v) not found", n)
			}

			index = append(index, q.Index(qb...))
		}

		// quantum state for json encoding
		state := make([]State, 0, len(e.Q.Raw().State()))
		for _, s := range e.Q.Raw().State(index...) {
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

		return state, nil
	}()
	if err != nil {
		log.SpanOf(spanID).ErrorReport("compute: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message":  "something went wrong",
			"trace_id": traceID,
		})
		return
	}

	// response
	c.JSON(http.StatusOK, Response{
		Filename: file.Filename,
		Content:  string(r),
		State:    state,
	})
}
