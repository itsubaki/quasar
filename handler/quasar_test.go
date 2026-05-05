package handler_test

import (
	"context"
	"fmt"
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
