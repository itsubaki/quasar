package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/itsubaki/logger"
	"github.com/itsubaki/quasar/handler"
	"github.com/itsubaki/tracer"
)

var (
	projectID   = os.Getenv("PROJECT_ID")
	serviceName = os.Getenv("K_SERVICE")  // https://cloud.google.com/run/docs/container-contract?hl=ja#services-env-vars
	revision    = os.Getenv("K_REVISION") // https://cloud.google.com/run/docs/container-contract?hl=ja#services-env-vars
	pprof       = os.Getenv("USE_PPROF")
	cprof       = os.Getenv("USE_CPROF")
	port        = os.Getenv("PORT")
	timeout     = 5 * time.Second
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// tracer, logger
	close := []func() error{
		logger.MustSetup(projectID, serviceName, revision),
		tracer.MustSetup(projectID, serviceName, revision, timeout),
	}
	defer func() {
		for _, c := range close {
			if err := c(); err != nil {
				log.Printf("defer: %v", err)
			}
		}
	}()

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
	if port == "" {
		port = "8080"
	}

	h := handler.New()
	if strings.ToLower(pprof) == "true" {
		// profiler
		handler.UsePProf(h)
	}

	s := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: h,
	}

	go func() {
		log.Printf("http server listen and serve. port: %v\n", port)
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
