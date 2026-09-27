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
	if TargetURL == "" {
		panic("environment variable TARGET_URL is required")
	}

	if IdentityToken == "" {
		panic("environment variable IDENTITY_TOKEN is required")
	}

	var filepath, snippetID string
	var simulate, validate, share, edit bool
	flag.StringVar(&filepath, "f", "", "filepath")
	flag.StringVar(&snippetID, "id", "", "snippet ID")
	flag.BoolVar(&simulate, "simulate", false, "")
	flag.BoolVar(&share, "share", false, "")
	flag.BoolVar(&validate, "validate", false, "")
	flag.BoolVar(&edit, "edit", false, "")
	flag.Parse()

	switch {
	case simulate:
		contents, err := Read(filepath)
		if err != nil {
			panic(err)
		}

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
	case validate:
		contents, err := Read(filepath)
		if err != nil {
			panic(err)
		}

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
	case share:
		contents, err := Read(filepath)
		if err != nil {
			panic(err)
		}
		resp, err := client.
			New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
			Share(context.Background(), string(contents))
		if err != nil {
			panic(err)
		}

		snippet, err := client.
			New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
			Edit(context.Background(), resp.ID)
		if err != nil {
			panic(err)
		}

		bytes, err := json.Marshal(snippet)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(bytes))
	case edit:
		resp, err := client.
			New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
			Edit(context.Background(), snippetID)
		if err != nil {
			panic(err)
		}

		bytes, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(bytes))
	default:
		panic("no valid action specified")
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
