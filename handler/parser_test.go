package handler_test

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/handler"
)

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
