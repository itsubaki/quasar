package handler_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/handler"
	"github.com/itsubaki/quasar/store"
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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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

func TestQuasarService_Simulate(t *testing.T) {
	cases := []struct {
		code   string
		errMsg string
	}{
		{
			code:   "invalid",
			errMsg: "invalid_argument: statement=: unexpected",
		},
		{
			code:   "qubit[12] q;",
			errMsg: "invalid_argument: need=12, max=10: too many qubits",
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

		t.Errorf("expected error but got response: %+v, %v", resp, err)
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

		t.Errorf("expected error but got response: %+v, %v", resp, err)
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

		t.Errorf("expected error but got response: %+v, %v", resp, err)
	}
}
