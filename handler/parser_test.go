package handler_test

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/handler"
)

func TestIncludeNotAllowed(t *testing.T) {
	svc := &handler.QuasarService{
		MaxQubits: 1,
	}

	for _, code := range []string{
		`include "private.qasm"; qubit q;`,
		`OPENQASM 2.0; include "qelib1.inc"; qreg q[1];`,
		`qubit q; if (false) { include "private.qasm"; }`,
		`def unused(qubit q) { include "private.qasm"; } qubit q;`,
		`qubit q; while (true) {} include "private.qasm";`,
		"include /* comment */ \n\"private.qasm\"; qubit q;",
	} {
		_, simulateErr := svc.Simulate(t.Context(), connect.NewRequest(&quasarv1.SimulateRequest{Code: code}))
		_, validateErr := svc.Validate(t.Context(), connect.NewRequest(&quasarv1.ValidateRequest{Code: code}))
		for _, err := range []error{simulateErr, validateErr} {
			if connect.CodeOf(err) != connect.CodeInvalidArgument || !errors.Is(err, handler.ErrIncludeNotAllowed) {
				t.Fatalf("got %v, want invalid_argument: include is not allowed", err)
			}
		}
	}
}

func TestIncludeTextIsAllowed(t *testing.T) {
	svc := &handler.QuasarService{
		MaxQubits: 1,
	}

	for _, code := range []string{
		"// include \"private.qasm\";\nqubit q;",
		`/* include "private.qasm"; */ qubit q;`,
		`int include_count = 1; qubit q;`,
	} {
		if _, err := svc.Simulate(t.Context(), connect.NewRequest(&quasarv1.SimulateRequest{Code: code})); err != nil {
			t.Fatal(err)
		}

		resp, err := svc.Validate(t.Context(), connect.NewRequest(&quasarv1.ValidateRequest{Code: code}))
		if err != nil || !resp.Msg.Valid {
			t.Fatalf("resp=%v err=%v", resp, err)
		}
	}

	// A string literal in another statement must not trigger the include policy.
	resp, err := svc.Validate(t.Context(), connect.NewRequest(&quasarv1.ValidateRequest{Code: `defcalgrammar "include"; qubit q;`}))
	if err != nil || !resp.Msg.Valid {
		t.Fatalf("resp=%v err=%v", resp, err)
	}
}
