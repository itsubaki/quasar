package handler_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/handler"
	"github.com/itsubaki/quasar/store"
)

func ExampleQuasarService_Simulate() {
	code := `
	OPENQASM 3.0;

gate x q { U(pi, 0, pi) q; }
gate h q { U(pi/2.0, 0, pi) q; }
gate cx c, t { ctrl @ U(pi, 0, pi) c, t; }
gate crz(theta) c, t { ctrl @ U(0, 0, theta) c, t; }

def qft(qubit[3] q) {
    h q[0];
    crz(pi/2) q[0], q[1];
    crz(pi/4) q[0], q[2];

    h q[1];
    crz(pi/2) q[1], q[2];

    h q[2];

    // swap
    cx q[0], q[2];
    cx q[2], q[0];
    cx q[0], q[2];
}

qubit[3] q;
x q[2];
qft(q);
	`

	service := &handler.QuasarService{}
	resp, err := service.Simulate(context.Background(), connect.NewRequest(&quasarv1.SimulateRequest{
		Code: code,
	}))
	if err != nil {
		panic(err)
	}

	for _, s := range resp.Msg.States {
		fmt.Printf("%s (%+.4f %+.4f): %.4f\n",
			s.BinaryString,
			s.Amplitude.Real,
			s.Amplitude.Imag,
			s.Probability,
		)
	}

	// Output:
	// [000] (+0.3536 +0.0000): 0.1250
	// [001] (+0.2500 +0.2500): 0.1250
	// [010] (+0.0000 +0.3536): 0.1250
	// [011] (-0.2500 +0.2500): 0.1250
	// [100] (-0.3536 +0.0000): 0.1250
	// [101] (-0.2500 -0.2500): 0.1250
	// [110] (+0.0000 -0.3536): 0.1250
	// [111] (+0.2500 -0.2500): 0.1250
}

func TestQuasarService_Simulate(t *testing.T) {
	cases := []struct {
		code   string
		errMsg string
	}{
		{
			code:   "invalid",
			errMsg: "invalid_argument: 1:7: no viable alternative at input 'invalid'",
		},
		{
			code:   "qubit[12] q;",
			errMsg: "invalid_argument: need=12, max=10: too many qubits",
		},
		{
			code:   "",
			errMsg: "invalid_argument: code not found",
		},
		{
			code:   "qubit[ q;",
			errMsg: `invalid_argument: 1:8: mismatched input ';' expecting ']'`,
		},
		{
			code:   "OPENQASM 3.0;",
			errMsg: "invalid_argument: qubits not found",
		},
	}

	svc := &handler.QuasarService{
		MaxQubits: 10,
		Store:     &store.MemoryStore{},
	}

	for _, c := range cases {
		resp, err := svc.Simulate(t.Context(), connect.NewRequest(&quasarv1.SimulateRequest{
			Code: c.code,
		}))
		if err != nil && err.Error() == c.errMsg {
			continue
		}

		t.Errorf("expected error but got response: resp=%+v, err=%v", resp, err)
	}
}

func TestQuasarService_Edit(t *testing.T) {
	cases := []struct {
		id     string
		errMsg string
	}{
		{
			id:     "", // empty
			errMsg: "invalid_argument: id not found",
		},
		{
			id:     "example",
			errMsg: "not_found: no such entity",
		},
	}

	svc := &handler.QuasarService{
		MaxQubits: 10,
		Store:     &store.MemoryStore{},
	}

	for _, c := range cases {
		resp, err := svc.Edit(t.Context(), connect.NewRequest(&quasarv1.EditRequest{
			Id: c.id,
		}))
		if err != nil && err.Error() == c.errMsg {
			continue
		}

		t.Errorf("expected error but got response: resp=%+v, err=%v", resp, err)
	}
}

func TestQuasarService_Share(t *testing.T) {
	cases := []struct {
		code   string
		errMsg string
	}{
		{
			code:   "", // empty
			errMsg: "invalid_argument: code not found",
		},
		{
			code:   strings.Repeat("qubit[2] q;", 2<<12),
			errMsg: "invalid_argument: code size exceeds 65536 bytes",
		},
	}

	svc := &handler.QuasarService{
		MaxQubits: 10,
		Store:     &store.MemoryStore{},
	}

	for _, c := range cases {
		resp, err := svc.Share(t.Context(), connect.NewRequest(&quasarv1.ShareRequest{
			Code: c.code,
		}))
		if err != nil && err.Error() == c.errMsg {
			continue
		}

		t.Errorf("expected error but got response: resp=%+v, err=%v", resp, err)
	}
}

func TestQuasarService_Validate(t *testing.T) {
	cases := []struct {
		code   string
		want   bool
		line   int32
		column int32
		errMsg string
	}{
		{
			code: "OPENQASM 3.0;",
			want: true,
		},
		{
			code:   "",
			errMsg: "invalid_argument: code not found",
		},
		{
			code:   "qubit[ q;",
			want:   false,
			line:   1,
			column: 8,
		},
	}

	svc := &handler.QuasarService{
		MaxQubits: 10,
		Store:     &store.MemoryStore{},
	}

	for _, c := range cases {
		resp, err := svc.Validate(t.Context(), connect.NewRequest(&quasarv1.ValidateRequest{
			Code: c.code,
		}))
		if err != nil {
			if err.Error() != c.errMsg {
				t.Errorf("got=%v, want=%v", err.Error(), c.errMsg)
			}

			continue
		}

		if resp.Msg.Valid != c.want {
			t.Errorf("got=%v, want=%v", resp.Msg.Valid, c.want)
		}

		if resp.Msg.Line != nil && *resp.Msg.Line != c.line {
			t.Errorf("got=%v, want=%v", *resp.Msg.Line, c.line)
		}

		if resp.Msg.Column != nil && *resp.Msg.Column != c.column {
			t.Errorf("got=%v, want=%v", *resp.Msg.Column, c.column)
		}
	}
}

func TestCodeSizeLimits(t *testing.T) {
	cases := []struct {
		name   string
		size   int
		errMsg string
	}{
		{
			name: "at limit",
			size: 64 * 1024,
		},
		{
			name: "below limit",
			size: 64*1024 - 1,
		},
		{
			name:   "above limit",
			size:   64*1024 + 1,
			errMsg: "invalid_argument: code size exceeds 65536 bytes",
		},
	}

	svc := &handler.QuasarService{}
	const prefix = "qubit q; //"

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code := prefix + strings.Repeat("x", c.size-len(prefix))
			_, validateErr := svc.Validate(t.Context(), connect.NewRequest(&quasarv1.ValidateRequest{Code: code}))
			_, simulateErr := svc.Simulate(t.Context(), connect.NewRequest(&quasarv1.SimulateRequest{Code: code}))

			for _, err := range []error{validateErr, simulateErr} {
				if err != nil {
					if err.Error() != c.errMsg {
						t.Errorf("got=%v, want=%v", err, c.errMsg)
					}

					continue
				}

				if c.errMsg != "" {
					t.Errorf("got=nil, want=%v", c.errMsg)
				}
			}
		})
	}
}

func TestRequestSizeLimits(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		code       string
		compressed bool
		statusCode int
		errCode    string
	}{
		{
			name:       "Simulate uncompressed",
			method:     "Simulate",
			code:       strings.Repeat("x", 512*1024),
			compressed: false,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Simulate gzip",
			method:     "Simulate",
			code:       strings.Repeat("x", 512*1024),
			compressed: true,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Validate uncompressed",
			method:     "Validate",
			code:       strings.Repeat("x", 512*1024),
			compressed: false,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Validate gzip",
			method:     "Validate",
			code:       strings.Repeat("x", 512*1024),
			compressed: true,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Share uncompressed",
			method:     "Share",
			code:       strings.Repeat("x", 512*1024),
			compressed: false,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Share gzip",
			method:     "Share",
			code:       strings.Repeat("x", 512*1024),
			compressed: true,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Edit uncompressed",
			method:     "Edit",
			code:       strings.Repeat("x", 512*1024),
			compressed: false,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			name:       "Edit gzip",
			method:     "Edit",
			code:       strings.Repeat("x", 512*1024),
			compressed: true,
			statusCode: http.StatusTooManyRequests,
			errCode:    "resource_exhausted",
		},
		{
			// Preserve the full 64 KiB source allowance even with six-byte JSON escapes.
			name:       "Share escaped source at limit",
			method:     "Share",
			code:       strings.Repeat("\x01", 64*1024),
			statusCode: http.StatusOK,
		},
	}

	h, err := handler.New(1, 1, &store.MemoryStore{})
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{
				"code": c.code,
			})
			if err != nil {
				t.Fatal(err)
			}

			if c.compressed {
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

			req := httptest.NewRequest(http.MethodPost, "/quasar.v1.QuasarService/"+c.method, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if c.compressed {
				req.Header.Set("Content-Encoding", "gzip")
			}

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != c.statusCode {
				t.Fatalf("status=%d, want=%d response=%s", rec.Code, c.statusCode, rec.Body.String())
			}

			var result struct {
				Code string `json:"code"`
			}

			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}

			if result.Code != c.errCode {
				t.Errorf("code=%s, want=%s response=%s", result.Code, c.errCode, rec.Body.String())
			}
		})
	}
}

func TestIncludeNotAllowed(t *testing.T) {
	cases := []struct {
		name string
		code string
	}{
		{
			name: "top level",
			code: `include "private.qasm"; qubit q;`,
		},
		{
			name: "OpenQASM 2",
			code: `OPENQASM 2.0; include "qelib1.inc"; qreg q[1];`,
		},
		{
			name: "inactive branch",
			code: `qubit q; if (false) { include "private.qasm"; }`,
		},
		{
			name: "unused function",
			code: `def unused(qubit q) { include "private.qasm"; } qubit q;`,
		},
		{
			name: "after infinite loop",
			code: `qubit q; while (true) {} include "private.qasm";`,
		},
		{
			name: "intervening comment",
			code: "include /* comment */ \n\"private.qasm\"; qubit q;",
		},
	}

	svc := &handler.QuasarService{
		MaxQubits: 1,
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, simulateErr := svc.Simulate(t.Context(), connect.NewRequest(&quasarv1.SimulateRequest{Code: c.code}))
			_, validateErr := svc.Validate(t.Context(), connect.NewRequest(&quasarv1.ValidateRequest{Code: c.code}))

			for _, err := range []error{simulateErr, validateErr} {
				if connect.CodeOf(err) != connect.CodeInvalidArgument || !errors.Is(err, handler.ErrIncludeNotAllowed) {
					t.Errorf("got %v, want invalid_argument: include is not allowed", err)
				}
			}
		})
	}
}

func TestIncludeTextIsAllowed(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		simulate bool
	}{
		{
			name:     "line comment",
			code:     "// include \"private.qasm\";\nqubit q;",
			simulate: true,
		},
		{
			name:     "block comment",
			code:     `/* include "private.qasm"; */ qubit q;`,
			simulate: true,
		},
		{
			name:     "identifier",
			code:     `int include_count = 1; qubit q;`,
			simulate: true,
		},
		{
			name: "string literal",
			code: `defcalgrammar "include"; qubit q;`,
		},
	}

	svc := &handler.QuasarService{
		MaxQubits: 1,
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.simulate {
				if _, err := svc.Simulate(t.Context(), connect.NewRequest(&quasarv1.SimulateRequest{Code: c.code})); err != nil {
					t.Fatal(err)
				}
			}

			resp, err := svc.Validate(t.Context(), connect.NewRequest(&quasarv1.ValidateRequest{Code: c.code}))
			if err != nil || !resp.Msg.Valid {
				t.Fatalf("resp=%v err=%v", resp, err)
			}
		})
	}
}
