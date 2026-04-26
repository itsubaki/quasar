package handler_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"connectrpc.com/connect"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/handler"
)

func ExampleNew_root() {
	h, err := handler.New(5, nil)
	if err != nil {
		panic(err)
	}

	s := httptest.NewServer(h)
	defer s.Close()

	resp, err := http.Get(s.URL)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		panic("invalid status code")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))

	// Output:
	// {"ok": true}
}

func ExampleNew_status() {
	h, err := handler.New(5, nil)
	if err != nil {
		panic(err)
	}

	s := httptest.NewServer(h)
	defer s.Close()

	resp, err := http.Get(fmt.Sprintf("%s/status", s.URL))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		panic("invalid status code")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))

	// Output:
	// {"ok": true}
}

func TestRecover(t *testing.T) {
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	cases := []struct {
		panicVal any
		errMsg   string
	}{
		{
			panicVal: "string panic",
			errMsg:   "unexpected: string panic",
		},
		{
			panicVal: errors.New("new error panic"),
			errMsg:   "unexpected: new error panic",
		},
	}

	for _, c := range cases {
		next := handler.Recover()(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
			panic(c.panicVal)
		})

		resp, err := next(context.Background(), connect.NewRequest(&quasarv1.ShareRequest{}))
		if err == nil {
			t.Fatal("expected error")
		}

		if resp != nil {
			t.Fatalf("expected nil response, got=%+v", resp)
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected *connect.Error, got=%T", err)
		}

		if got := connectErr.Code(); got != connect.CodeInternal {
			t.Fatalf("code=%v, want=%v", got, connect.CodeInternal)
		}

		if got := connectErr.Message(); got != c.errMsg {
			t.Fatalf("message=%v, want=%v", got, c.errMsg)
		}
	}
}
