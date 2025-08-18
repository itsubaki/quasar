package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"

	"connectrpc.com/connect"
	"github.com/itsubaki/quasar/gen/quasar/v1/quasarv1connect"
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
			Recover(),
		),
	))

	return h2c.NewHandler(mux, &http2.Server{}), nil
}

func Recover() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
			defer func() {
				if r := recover(); r != nil {
					slog.DebugContext(ctx, "recovered from panic", slog.Any("recover", r))
					resp, err = nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unexpected: %v", r))
				}
			}()

			return next(ctx, req)
		}
	}
}
