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

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nEnvironment variables:")
		fmt.Fprintln(os.Stderr, "  TARGET_URL       URL of the target Google Cloud Run service")
		fmt.Fprintln(os.Stderr, "  IDENTITY_TOKEN   Identity token for authenticating with Cloud Run")
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	var filepath, snippetID string
	var simulate, validate, share, edit bool
	flag.StringVar(&filepath, "f", "", "path to an OpenQASM file (default: stdin)")
	flag.StringVar(&snippetID, "id", "", "snippet ID to edit")
	flag.BoolVar(&simulate, "simulate", false, "simulate the OpenQASM code")
	flag.BoolVar(&validate, "validate", false, "validate the OpenQASM code")
	flag.BoolVar(&share, "share", false, "share the OpenQASM code")
	flag.BoolVar(&edit, "edit", false, "edit a shared snippet")
	flag.Parse()

	switch {
	case simulate:
		contents, err := Read(filepath)
		if err != nil {
			return err
		}

		resp, err := client.
			New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
			Simulate(context.Background(), string(contents))
		if err != nil {
			return err
		}

		bytes, err := json.Marshal(resp)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		fmt.Println(string(bytes))
	case validate:
		contents, err := Read(filepath)
		if err != nil {
			return err
		}

		resp, err := client.
			New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
			Validate(context.Background(), string(contents))
		if err != nil {
			return err
		}

		bytes, err := json.Marshal(resp)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		fmt.Println(string(bytes))
	case share:
		contents, err := Read(filepath)
		if err != nil {
			return err
		}

		resp, err := client.
			New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
			Share(context.Background(), string(contents))
		if err != nil {
			return err
		}

		bytes, err := json.Marshal(resp)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		fmt.Println(string(bytes))
	case edit:
		resp, err := client.
			New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
			Edit(context.Background(), snippetID)
		if err != nil {
			return err
		}

		bytes, err := json.Marshal(resp)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		fmt.Println(string(bytes))
	default:
		return fmt.Errorf("no valid action specified")
	}

	return nil
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
