package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/itsubaki/quasar/client"
)

var (
	TargetURL     = os.Getenv("TARGET_URL")
	IdentityToken = os.Getenv("IDENTITY_TOKEN")
)

func main() {
	var filepath string
	flag.StringVar(&filepath, "f", "", "filepath")
	flag.Parse()

	if filepath == "" {
		fmt.Printf("Usage: %s -f filepath\n", os.Args[0])
		return
	}

	contents, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	id, createdAt, err := client.
		New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
		Save(context.Background(), string(contents))
	if err != nil {
		panic(err)
	}
	fmt.Println("saved:", id, createdAt)

	code, createdAt, err := client.
		New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
		Load(context.Background(), id)
	if err != nil {
		panic(err)
	}
	fmt.Println("loaded:", id, code, createdAt)
}
