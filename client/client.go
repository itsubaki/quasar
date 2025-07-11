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

func (c *Client) Factorize(ctx context.Context, N uint64, t, a, seed *uint64) (*FactorizeResponse, error) {
	resp, err := c.quasarClient.Factorize(ctx, connect.NewRequest(&quasarv1.FactorizeRequest{
		N:    N,
		A:    a,
		T:    t,
		Seed: seed,
	}))
	if err != nil {
		return nil, fmt.Errorf("factorize: %w", err)
	}

	return &FactorizeResponse{
		N:       resp.Msg.N,
		T:       resp.Msg.T,
		A:       resp.Msg.A,
		Seed:    resp.Msg.Seed,
		M:       resp.Msg.M,
		S:       resp.Msg.S,
		R:       resp.Msg.R,
		P:       resp.Msg.P,
		Q:       resp.Msg.Q,
		Message: resp.Msg.Message,
	}, nil
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
