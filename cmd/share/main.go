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

	// share
	resp, err := client.
		New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
		Share(context.Background(), string(contents))
	if err != nil {
		panic(err)
	}

	fmt.Println("shared: ", resp.ID, resp.CreatedAt)

	// edit
	snippet, err := client.
		New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
		Edit(context.Background(), resp.ID)
	if err != nil {
		panic(err)
	}

	fmt.Println("edited:", snippet.ID, snippet.CreatedAt)
	fmt.Println(snippet.Code)
}
