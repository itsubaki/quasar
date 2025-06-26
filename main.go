package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/itsubaki/quasar/handler"
)

var (
	projectID   = os.Getenv("PROJECT_ID")
	serviceName = os.Getenv("K_SERVICE")  // https://cloud.google.com/run/docs/container-contract?hl=ja#services-env-vars
	revision    = os.Getenv("K_REVISION") // https://cloud.google.com/run/docs/container-contract?hl=ja#services-env-vars
	cprof       = os.Getenv("USE_CPROF")
	port        = os.Getenv("PORT")
	timeout     = 5 * time.Second
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// profiler
	if strings.ToLower(cprof) == "true" {
		if err := profiler.Start(profiler.Config{
			ProjectID:      projectID,
			Service:        serviceName,
			ServiceVersion: revision,
		}); err != nil {
			log.Fatalf("profiler: %v", err)
		}
	}

	// handler
	h, err := handler.New()
	if err != nil {
		log.Fatalf("new handler: %v", err)
	}

	// server
	addr := ":8080"
	if port != "" {
		addr = fmt.Sprintf(":%s", port)
	}

	s := &http.Server{
		Addr:    addr,
		Handler: h,
	}

	go func() {
		log.Printf("http server listen and serve. addr=%v\n", addr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// shutdown
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("http server shutdown: %v", err)
	}

	log.Println("shutdown finished")
}
