package handler

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

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
	))

	return h2c.NewHandler(mux, &http2.Server{}), nil
}
