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
	Amplitude    Amplitude `json:"amplitude"`
	Probability  float64   `json:"probability"`
	Int          []uint64  `json:"int"`
	BinaryString []string  `json:"binary_string"`
}

type Amplitude struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
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
			Int:          s.Int,
			BinaryString: s.BinaryString,
		}
	}

	return &States{States: states}, nil
}

func (c *Client) Save(ctx context.Context, code string) (string, time.Time, error) {
	resp, err := c.quasarClient.Save(ctx, connect.NewRequest(&quasarv1.SaveRequest{
		Code: code,
	}))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("save: %w", err)
	}

	return resp.Msg.Id, resp.Msg.CreatedAt.AsTime(), nil
}

func (c *Client) Load(ctx context.Context, id string) (string, time.Time, error) {
	resp, err := c.quasarClient.Load(ctx, connect.NewRequest(&quasarv1.LoadRequest{
		Id: id,
	}))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("load: %w", err)
	}

	return resp.Msg.Code, resp.Msg.CreatedAt.AsTime(), nil
}
