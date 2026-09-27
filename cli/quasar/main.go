package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/itsubaki/quasar/client"
)

var (
	TargetURL     = os.Getenv("TARGET_URL")
	IdentityToken = os.Getenv("IDENTITY_TOKEN")
)

func main() {
	var filepath string
	var validate, share bool
	flag.StringVar(&filepath, "f", "", "filepath")
	flag.BoolVar(&share, "share", false, "")
	flag.BoolVar(&validate, "validate", false, "")
	flag.Parse()

	contents, err := Read(filepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch {
	case share:
		resp, err := client.
			New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
			Share(context.Background(), string(contents))
		if err != nil {
			panic(err)
		}

		fmt.Println("shared: ", resp.ID, resp.CreatedAt)

		snippet, err := client.
			New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
			Edit(context.Background(), resp.ID)
		if err != nil {
			panic(err)
		}

		fmt.Println("edited:", snippet.ID, snippet.CreatedAt)
		fmt.Println(snippet.Code)
	case validate:
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
	default:
		resp, err := client.
			New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
			Simulate(context.Background(), string(contents))
		if err != nil {
			panic(err)
		}

		bytes, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(bytes))
	}
}

func Read(filepath string) (string, error) {
	if filepath != "" {
		read, err := os.ReadFile(filepath)
		if err != nil {
			return "", fmt.Errorf("read file %s: %w", filepath, err)
		}

		return string(read), nil
	}

	text, err := Scan(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	return text, nil
}

func Scan(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)

	var text strings.Builder
	for scanner.Scan() {
		text.WriteString(scanner.Text())
		text.WriteByte('\n')
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan: %w", err)
	}

	return text.String(), nil
}
