package tracer

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	gcp "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	// https://cloud.google.com/appengine/docs/standard/go/runtime#environment_variables
	// https://cloud.google.com/run/docs/container-contract?hl=ja#services-env-vars
	projectID    = os.Getenv("GOOGLE_CLOUD_PROJECT")
	serviceName  = os.Getenv("K_SERVICE")
	revision     = os.Getenv("K_REVISION")
)

func Must(f func(), err error) func() {
	if err != nil {
		panic(err)
	}

	return f
}

func Setup(timeout time.Duration) (func(), error) {
	exporter, err := gcp.New(gcp.WithProjectID(projectID))
	if err != nil {
		return nil, fmt.Errorf("new exporter: %v", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
				semconv.ServiceVersionKey.String(revision),
			),
		),
	)

	otel.SetTracerProvider(provider)

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := provider.ForceFlush(ctx); err != nil {
			log.Printf("provider force flush: %v", err)
		}

		if err := provider.Shutdown(ctx); err != nil {
			log.Printf("provider shutdown: %v", err)
		}
	}, nil
}

func NewContext(ctx context.Context, traceID, spanID string, traceTrue bool) (context.Context, error) {
	tID, err := trace.TraceIDFromHex(traceID)
	if err != nil {
		return nil, fmt.Errorf("traceID from hex(%v): %v", traceID, err)
	}

	sID, err := trace.SpanIDFromHex(spanID)
	if err != nil {
		return nil, fmt.Errorf("spanID from hex(%v): %v", spanID, err)
	}

	flags := trace.TraceFlags(00)
	if traceTrue {
		flags = 01
	}

	return trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tID,
		SpanID:     sID,
		TraceFlags: flags,
		Remote:     false,
	})), nil
}
