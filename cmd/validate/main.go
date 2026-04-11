package main

import (
	"context"
	"encoding/json"
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

	// validate
	resp, err := client.
		New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
		Validate(context.Background(), string(contents))
	if err != nil {
		panic(err)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
}
