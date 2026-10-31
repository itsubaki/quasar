package handler_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/handler"
	"github.com/itsubaki/quasar/store"
)

func TestCodeSizeLimits(t *testing.T) {
	svc := &handler.QuasarService{}
	const prefix = "qubit q; //"
	for _, size := range []int{64 * 1024, 64*1024 + 1} {
		code := prefix + strings.Repeat("x", size-len(prefix))
		_, validateErr := svc.Validate(t.Context(), connect.NewRequest(&quasarv1.ValidateRequest{Code: code}))
		_, simulateErr := svc.Simulate(t.Context(), connect.NewRequest(&quasarv1.SimulateRequest{Code: code}))
		for _, err := range []error{validateErr, simulateErr} {
			if size == 64*1024 {
				if err != nil {
					t.Fatalf("valid source rejected: %v", err)
				}
			} else if err == nil || connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Fatalf("oversized source accepted: %v", err)
			}
		}
	}
}

func TestRequestSizeLimits(t *testing.T) {
	h, err := handler.New(1, 1, &store.MemoryStore{})
	if err != nil {
		t.Fatal(err)
	}

	for _, method := range []string{"Simulate", "Validate", "Share", "Edit"} {
		for _, compressed := range []bool{false, true} {
			body, err := json.Marshal(map[string]string{"code": strings.Repeat("x", 512*1024)})
			if err != nil {
				t.Fatal(err)
			}

			if compressed {
				var zipped bytes.Buffer
				w := gzip.NewWriter(&zipped)
				if _, err := w.Write(body); err != nil {
					t.Fatal(err)
				}
				if err := w.Close(); err != nil {
					t.Fatal(err)
				}
				body = zipped.Bytes()
			}

			req := httptest.NewRequest(http.MethodPost, "/quasar.v1.QuasarService/"+method, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if compressed {
				req.Header.Set("Content-Encoding", "gzip")
			}

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			var result struct {
				Code string `json:"code"`
			}

			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}

			if result.Code != "resource_exhausted" {
				t.Fatalf("status=%d response=%s", rec.Code, rec.Body.String())
			}
		}
	}

	// Preserve the full 64 KiB source allowance even with six-byte JSON escapes.
	body, _ := json.Marshal(map[string]string{"code": strings.Repeat("\x01", 64*1024)})
	req := httptest.NewRequest(http.MethodPost, "/quasar.v1.QuasarService/Share", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid size rejected: %s", rec.Body.String())
	}
}
