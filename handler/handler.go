package handler

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"cloud.google.com/go/firestore"
	"github.com/itsubaki/quasar/gen/quasar/v1/quasarv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func New(projectID, databaseID string, maxQubits int) (http.Handler, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"ok": true}`)
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"ok": true}`)
	})

	fs, err := firestore.NewClientWithDatabase(context.Background(), projectID, databaseID)
	if err != nil {
		return nil, fmt.Errorf("new firestore client: %v", err)
	}

	mux.Handle(quasarv1connect.NewQuasarServiceHandler(
		&QuasarService{
			MaxQubits: maxQubits,
			Firestore: fs,
		},
	))

	return h2c.NewHandler(mux, &http2.Server{}), nil
}
