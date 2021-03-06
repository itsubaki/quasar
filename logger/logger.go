package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/errorreporting"
	"go.opentelemetry.io/otel/trace"
)

const (
	DEFAULT = iota
	DEBUG
	INFO
	NOTICE
	WARNING
	ERROR
	CRITICAL
	ALERT
	EMERGENCY
)

var (
	projectID   = os.Getenv("PROJECT_ID")
	serviceName = os.Getenv("K_SERVICE")  // https://cloud.google.com/run/docs/container-contract?hl=ja#services-env-vars
	revision    = os.Getenv("K_REVISION") // https://cloud.google.com/run/docs/container-contract?hl=ja#services-env-vars
	loglevel    = LogLevel(os.Getenv("LOG_LEVEL"), "0")
	Factory     = Must(New(context.Background(), projectID))
)

func LogLevel(v, w string) string {
	if v == "" {
		return w
	}

	return v
}

type LoggerFactory struct {
	level     int
	projectID string
	errC      *errorreporting.Client
}

func Must(f *LoggerFactory, err error) *LoggerFactory {
	if err != nil {
		panic(err)
	}

	return f
}

func New(ctx context.Context, projectID string) (*LoggerFactory, error) {
	c, err := errorreporting.NewClient(ctx, projectID, errorreporting.Config{
		ServiceName:    serviceName,
		ServiceVersion: revision,
	})
	if err != nil {
		return nil, fmt.Errorf("new error reporting client: %v", err)
	}

	l, err := strconv.Atoi(loglevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level=%v: %v", loglevel, err)
	}

	return &LoggerFactory{
		level:     l,
		projectID: projectID,
		errC:      c,
	}, nil
}

func (f *LoggerFactory) New(req *http.Request, traceID, spanID string) *Logger {
	trace := ""
	if len(traceID) > 0 {
		trace = fmt.Sprintf("projects/%v/traces/%v", f.projectID, traceID)
	}

	return &Logger{
		level:   f.level,
		errC:    f.errC,
		traceID: trace,
		spanID:  spanID,
		req:     req,
	}
}

func (f *LoggerFactory) Close() {
	f.errC.Flush()
	if err := f.errC.Close(); err != nil {
		log.Printf("close errorreporing client: %v", err)
	}
}

type Logger struct {
	level   int
	traceID string
	spanID  string
	errC    *errorreporting.Client
	req     *http.Request
}

type LogEntry struct {
	Severity string    `json:"severity"`
	Message  string    `json:"message"`
	Time     time.Time `json:"time"`
	TraceID  string    `json:"logging.googleapis.com/trace"`
	SpanID   string    `json:"logging.googleapis.com/spanId,omitempty"`
}

func (l *Logger) Log(severity, format string, a ...interface{}) {
	if err := json.NewEncoder(os.Stdout).Encode(&LogEntry{
		Severity: severity,
		Time:     time.Now(),
		Message:  fmt.Sprintf(format, a...),
		TraceID:  l.traceID,
		SpanID:   l.spanID,
	}); err != nil {
		log.Printf("encode log entry: %v", err)
	}
}

func (l *Logger) Report(a ...interface{}) {
	for _, aa := range a {
		switch err := aa.(type) {
		case error:
			l.errC.Report(errorreporting.Entry{
				Error: err,
				Req:   l.req,
			})
		}
	}
}

func (l *Logger) Debug(format string, a ...interface{}) {
	if l.level > DEBUG {
		return
	}

	l.Log("Debug", format, a...)
}

func (l *Logger) Info(format string, a ...interface{}) {
	if l.level > INFO {
		return
	}

	l.Log("Info", format, a...)
}

func (l *Logger) Error(format string, a ...interface{}) {
	if l.level > ERROR {
		return
	}

	l.Log("Error", format, a...)
}

func (l *Logger) ErrorReport(format string, a ...interface{}) {
	if l.level > ERROR {
		return
	}

	l.Error(format, a...)
	l.Report(a...)
}

func (l *Logger) Span(span trace.Span) *Logger {
	return &Logger{
		level:   l.level,
		traceID: l.traceID,
		spanID:  span.SpanContext().SpanID().String(),
		errC:    l.errC,
		req:     l.req,
	}
}
