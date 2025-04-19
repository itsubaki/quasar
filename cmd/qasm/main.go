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
	BaseURL       = os.Getenv("BASE_URL")
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

	text, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	resp, err := client.
		New(BaseURL, IdentityToken).
		Run(context.Background(), string(text))
	if err != nil {
		panic(err)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
}
