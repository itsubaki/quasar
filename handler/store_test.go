package handler_test

import (
	"context"
	"fmt"
	"time"

	"github.com/itsubaki/quasar/handler"
)

func ExampleMemoryStore() {
	s := &handler.MemoryStore{}
	if err := s.Put(context.TODO(), "foo", &handler.Snippet{
		Code:      "bar",
		CreatedAt: time.Now(),
	}); err != nil {
		panic(err)
	}

	snippet, err := s.Get(context.TODO(), "foo")
	if err != nil {
		panic(err)
	}

	fmt.Println(snippet.Code)

	// Output:
	// bar
}
