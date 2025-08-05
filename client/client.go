package client

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/gen/quasar/v1/quasarv1connect"
)

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

func (c *Client) Simulate(ctx context.Context, code string) (*RunResponse, error) {
	resp, err := c.quasarClient.Simulate(ctx, connect.NewRequest(&quasarv1.SimulateRequest{
		Code: code,
	}))
	if err != nil {
		return nil, fmt.Errorf("simulate: %w", err)
	}

	state := make([]State, len(resp.Msg.State))
	for i, s := range resp.Msg.State {
		state[i] = State{
			Probability: s.Probability,
			Amplitude: Amplitude{
				Real: s.Amplitude.Real,
				Imag: s.Amplitude.Imag,
			},
			Int:          s.Int,
			BinaryString: s.BinaryString,
		}
	}

	return &RunResponse{State: state}, nil

}
