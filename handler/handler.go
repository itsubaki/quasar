package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"

	"connectrpc.com/connect"
	"github.com/itsubaki/quasar/gen/quasar/v1/quasarv1connect"
)

func New(maxQubits, maxConcurrent int, store Store) (http.Handler, error) {
	if maxConcurrent < 1 {
		return nil, errors.New("max concurrent must be positive")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprintf(w, `{"ok": true}`); err != nil {
			slog.WarnContext(r.Context(), "write response", slog.Any("error", err))
		}
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprintf(w, `{"ok": true}`); err != nil {
			slog.WarnContext(r.Context(), "write response", slog.Any("error", err))
		}
	})

	mux.Handle(quasarv1connect.NewQuasarServiceHandler(
		&QuasarService{
			MaxQubits:     maxQubits,
			MaxConcurrent: maxConcurrent,
			Store:         store,
		},
		connect.WithInterceptors(
			Recover(),
		),
		connect.WithReadMaxBytes(maxRequestBytes),
	))

	return http.MaxBytesHandler(mux, maxRequestBytes), nil
}

func Recover() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
			defer func() {
				if r := recover(); r != nil {
					slog.ErrorContext(ctx,
						"recovered",
						slog.Any("panic", r),
						slog.String("procedure", req.Spec().Procedure),
					)

					resp, err = nil, connect.NewError(connect.CodeInternal, errors.New("unexpected internal error"))
				}
			}()

			return next(ctx, req)
		}
	}
}
