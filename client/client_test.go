package client_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"connectrpc.com/connect"
	"github.com/itsubaki/quasar/client"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/gen/quasar/v1/quasarv1connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mock struct {
	quasarv1connect.QuasarServiceHandler
}

func newMock() *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle(quasarv1connect.NewQuasarServiceHandler(&mock{}))
	return httptest.NewServer(mux)
}

func (m *mock) Simulate(
	ctx context.Context,
	req *connect.Request[quasarv1.SimulateRequest],
) (*connect.Response[quasarv1.SimulateResponse], error) {
	return connect.NewResponse(&quasarv1.SimulateResponse{
		States: []*quasarv1.SimulateResponse_State{
			{
				BinaryString: []string{"101"},
				Int:          []uint64{1, 0, 1},
				Probability:  0.5,
				Amplitude: &quasarv1.SimulateResponse_Amplitude{
					Real: 1.0,
					Imag: -1.0,
				},
			},
		},
	}), nil
}

func (m *mock) Share(
	ctx context.Context,
	req *connect.Request[quasarv1.ShareRequest],
) (*connect.Response[quasarv1.ShareResponse], error) {
	return connect.NewResponse(&quasarv1.ShareResponse{
		Id:        "abcd1234",
		CreatedAt: &timestamppb.Timestamp{Seconds: 1234},
	}), nil
}

func (m *mock) Edit(
	ctx context.Context,
	req *connect.Request[quasarv1.EditRequest],
) (*connect.Response[quasarv1.EditResponse], error) {
	return connect.NewResponse(&quasarv1.EditResponse{
		Id:        "abcd1234",
		Code:      "qubit[3] q;",
		CreatedAt: &timestamppb.Timestamp{Seconds: 1234},
	}), nil
}

func ExampleClient_Simulate() {
	srv := newMock()
	defer srv.Close()

	states, err := client.New(srv.URL, srv.Client()).Simulate(
		context.Background(),
		"qubit[3] q;",
	)
	if err != nil {
		panic(err)
	}

	for _, state := range states.States {
		fmt.Println(state.BinaryString, state.Int, state.Probability, state.Amplitude.Real, state.Amplitude.Imag)
	}

	// Output:
	// [101] [1 0 1] 0.5 1 -1
}

func ExampleClient_Share() {
	srv := newMock()
	defer srv.Close()

	snippet, err := client.New(srv.URL, srv.Client()).Share(
		context.Background(),
		"qubit[3] q;",
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(snippet.ID)
	fmt.Println(snippet.CreatedAt.Unix())

	// Output:
	// abcd1234
	// 1234
}

func ExampleClient_Edit() {
	srv := newMock()
	defer srv.Close()

	snippet, err := client.New(srv.URL, srv.Client()).Edit(
		context.Background(),
		"abcd1234",
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(snippet.ID)
	fmt.Println(snippet.Code)
	fmt.Println(snippet.CreatedAt.Unix())

	// Output:
	// abcd1234
	// qubit[3] q;
	// 1234
}
