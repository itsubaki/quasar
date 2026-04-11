package handler_test

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	quasarv1 "github.com/itsubaki/quasar/gen/quasar/v1"
	"github.com/itsubaki/quasar/handler"
)

func ExampleQuasarService_Simulate() {
	code := `
	OPENQASM 3.0;

gate x q { U(pi, 0, pi) q; }
gate h q { U(pi/2.0, 0, pi) q; }
gate cx c, t { ctrl @ U(pi, 0, pi) c, t; }
gate crz(theta) c, t { ctrl @ U(0, 0, theta) c, t; }

def qft(qubit[3] q) {
    h q[0];
    crz(pi/2) q[0], q[1];
    crz(pi/4) q[0], q[2];

    h q[1];
    crz(pi/2) q[1], q[2];

    h q[2];

    // swap
    cx q[0], q[2];
    cx q[2], q[0];
    cx q[0], q[2];
}

qubit[3] q;
x q[2];
qft(q);
	`

	service := &handler.QuasarService{}
	resp, err := service.Simulate(context.Background(), connect.NewRequest(&quasarv1.SimulateRequest{
		Code: code,
	}))
	if err != nil {
		panic(err)
	}

	for _, s := range resp.Msg.States {
		fmt.Printf("%s %2d: %+.4f %+.4f: %.4f\n",
			s.BinaryString,
			s.Int,
			s.Amplitude.Real,
			s.Amplitude.Imag,
			s.Probability,
		)
	}

	// Output:
	// [000] [ 0]: +0.3536 +0.0000: 0.1250
	// [001] [ 1]: +0.2500 +0.2500: 0.1250
	// [010] [ 2]: +0.0000 +0.3536: 0.1250
	// [011] [ 3]: -0.2500 +0.2500: 0.1250
	// [100] [ 4]: -0.3536 +0.0000: 0.1250
	// [101] [ 5]: -0.2500 -0.2500: 0.1250
	// [110] [ 6]: +0.0000 -0.3536: 0.1250
	// [111] [ 7]: +0.2500 -0.2500: 0.1250
}
