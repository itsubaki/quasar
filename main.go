package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/itsubaki/quasar/handler"
	"github.com/itsubaki/quasar/logger"
	"github.com/itsubaki/quasar/tracer"
)

var (
	timeout = 5 * time.Second
	port    = os.Getenv("PORT")
	gae     = os.Getenv("GAE_APPLICATION") // https://cloud.google.com/appengine/docs/standard/go/runtime#environment_variables
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(gae) > 0 {
		// enable profiler on Google AppEngine
		// https://cloud.google.com/profiler/docs/about-profiler#environment_and_languages
		if err := profiler.Start(profiler.Config{
			MutexProfiling: true,
		}); err != nil {
			log.Fatalf("profiler start: %v", err)
		}
	}

	defer logger.Factory.Close()

	f, err := tracer.Setup(timeout)
	if err != nil {
		log.Fatalf("tracer setup: %v", err)
	}
	defer f()

	if port == "" {
		port = "8080"
	}

	s := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: handler.New(),
	}

	go func() {
		log.Println("http server listen and serve")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s", err)
		}
	}()

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
