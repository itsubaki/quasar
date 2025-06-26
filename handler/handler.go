package handler

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	_ "net/http/pprof"

	qctx "github.com/itsubaki/quasar/context"
	"github.com/itsubaki/quasar/gen/quasar/v1/quasarv1connect"

	"connectrpc.com/connect"
	"github.com/itsubaki/tracer"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func New() (http.Handler, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"ok": true}`)
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"ok": true}`)
	})

	mux.Handle(quasarv1connect.NewQuasarServiceHandler(
		&QuasarService{},
		connect.WithInterceptors(
			Trace(),
		),
	))

	return h2c.NewHandler(mux, &http2.Server{}), nil
}

func Trace() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			value := req.Header().Get("X-Cloud-Trace-Context")
			if value == "" {
				value = fmt.Sprintf("%016x%016x/%d;o=0", rand.Int64(), rand.Int64(), rand.Int64())
				req.Header().Add("X-Cloud-Trace-Context", value)
			}

			trace, err := tracer.Parse(value)
			if err != nil {
				return nil, fmt.Errorf("parse x-cloud-trace-context=%v: %w", value, err)
			}

			return next(
				qctx.SetTrace(ctx, trace.TraceID, trace.SpanID, trace.TraceTrue),
				req,
			)
		}
	}
}
