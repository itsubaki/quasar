package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/gen/quasar/v1/quasarv1connect"
)

type States struct {
	States []State `json:"states"`
}

type State struct {
	Probability  float64   `json:"probability"`
	Amplitude    Amplitude `json:"amplitude"`
	BinaryString []string  `json:"binary_string"`
}

type Amplitude struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
}

type Snippet struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type ValidationResult struct {
	Valid   bool    `json:"valid"`
	Line    *int32  `json:"line,omitempty"`
	Column  *int32  `json:"column,omitempty"`
	Message *string `json:"message,omitempty"`
}

type Client struct {
	quasarClient quasarv1connect.QuasarServiceClient
}

func New(targetURL string, client *http.Client) *Client {
	return &Client{
		quasarClient: quasarv1connect.NewQuasarServiceClient(
			client,
			targetURL,
		),
	}
}

func (c *Client) Simulate(ctx context.Context, code string) (*States, error) {
	resp, err := c.quasarClient.Simulate(ctx, connect.NewRequest(&quasarv1.SimulateRequest{
		Code: code,
	}))
	if err != nil {
		return nil, fmt.Errorf("simulate: %w", err)
	}

	states := make([]State, len(resp.Msg.States))
	for i, s := range resp.Msg.States {
		states[i] = State{
			Probability: s.Probability,
			Amplitude: Amplitude{
				Real: s.Amplitude.Real,
				Imag: s.Amplitude.Imag,
			},
			BinaryString: s.BinaryString,
		}
	}

	return &States{
		States: states,
	}, nil
}

func (c *Client) Share(ctx context.Context, code string) (*Snippet, error) {
	resp, err := c.quasarClient.Share(ctx, connect.NewRequest(&quasarv1.ShareRequest{
		Code: code,
	}))
	if err != nil {
		return nil, fmt.Errorf("share: %w", err)
	}

	return &Snippet{
		ID:        resp.Msg.Id,
		Code:      code,
		CreatedAt: resp.Msg.CreatedAt.AsTime(),
	}, nil
}

func (c *Client) Edit(ctx context.Context, id string) (*Snippet, error) {
	resp, err := c.quasarClient.Edit(ctx, connect.NewRequest(&quasarv1.EditRequest{
		Id: id,
	}))
	if err != nil {
		return nil, fmt.Errorf("edit: %w", err)
	}

	return &Snippet{
		ID:        resp.Msg.Id,
		Code:      resp.Msg.Code,
		CreatedAt: resp.Msg.CreatedAt.AsTime(),
	}, nil
}

func (c *Client) Validate(ctx context.Context, code string) (*ValidationResult, error) {
	resp, err := c.quasarClient.Validate(ctx, connect.NewRequest(&quasarv1.ValidateRequest{
		Code: code,
	}))
	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	return &ValidationResult{
		Valid:   resp.Msg.Valid,
		Line:    resp.Msg.Line,
		Column:  resp.Msg.Column,
		Message: resp.Msg.Message,
	}, nil
}
