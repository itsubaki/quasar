package handler_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

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
