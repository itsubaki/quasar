package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/itsubaki/quasar/client"
	"github.com/urfave/cli/v3"
)

var (
	targetURL     string
	identityToken string
)

func main() {
	app := &cli.Command{
		Name:  "quasar",
		Usage: "Quasar CLI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "target-url",
				Usage:       "URL of the target Google Cloud Run service",
				Sources:     cli.EnvVars("TARGET_URL"),
				Destination: &targetURL,
			},
			&cli.StringFlag{
				Name:        "identity-token",
				Usage:       "Identity token for authenticating with Cloud Run",
				Sources:     cli.EnvVars("IDENTITY_TOKEN"),
				Destination: &identityToken,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "validate",
				Usage: "Validate OpenQASM code",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "path to an OpenQASM file (default: stdin)",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					contents, err := read(cmd.String("file"))
					if err != nil {
						return err
					}

					resp, err := newClient().Validate(ctx, contents)
					if err != nil {
						return err
					}

					return print(resp)
				},
			},
			{
				Name:  "simulate",
				Usage: "Simulate OpenQASM code",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "path to an OpenQASM file (default: stdin)",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					contents, err := read(cmd.String("file"))
					if err != nil {
						return err
					}

					resp, err := newClient().Simulate(ctx, contents)
					if err != nil {
						return err
					}

					return print(resp)
				},
			},
			{
				Name:  "share",
				Usage: "Share OpenQASM code",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "path to an OpenQASM file (default: stdin)",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					contents, err := read(cmd.String("file"))
					if err != nil {
						return err
					}

					resp, err := newClient().Share(ctx, contents)
					if err != nil {
						return err
					}

					return print(resp)
				},
			},
			{
				Name:  "edit",
				Usage: "Edit a shared snippet",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "id",
						Usage:    "snippet ID to edit",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					resp, err := newClient().Edit(ctx, cmd.String("id"))
					if err != nil {
						return err
					}

					return print(resp)
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newClient() *client.Client {
	return client.New(
		targetURL,
		client.NewWithIdentityToken(identityToken),
	)
}

func print(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func read(filepath string) (string, error) {
	if filepath != "" {
		read, err := os.ReadFile(filepath)
		if err != nil {
			return "", fmt.Errorf("read file %s: %w", filepath, err)
		}

		return string(read), nil
	}

	text, err := scan(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	return text, nil
}

func scan(r io.Reader) (string, error) {
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
