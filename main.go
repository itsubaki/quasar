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

	"github.com/itsubaki/quasar/handler"
	"github.com/itsubaki/quasar/logger"
	"github.com/itsubaki/quasar/tracer"
)

var (
	timeout = 5 * time.Second
	pprof   = os.Getenv("USE_PPROF")
	port    = os.Getenv("PORT")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// logger, tracer
	defer logger.Factory.Close()
	defer tracer.Must(tracer.Setup(timeout))()

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
			log.Fatalf("listen: %s", err)
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
