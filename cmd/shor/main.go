package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/itsubaki/quasar/client"
)

var (
	TargetURL     = os.Getenv("TARGET_URL")
	IdentityToken = os.Getenv("IDENTITY_TOKEN")
)

func main() {
	var N, t, a int
	var seed uint64
	flag.IntVar(&N, "N", 15, "positive integer")
	flag.IntVar(&t, "t", 3, "precision bits")
	flag.IntVar(&a, "a", -1, "coprime number of N")
	flag.Uint64Var(&seed, "seed", 0, "PRNG seed for measurements")
	flag.Parse()

	resp, err := client.
		NewWithIdentityToken(TargetURL, IdentityToken).
		Factorize(context.Background(), N, t, a, seed)
	if err != nil {
		panic(err)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
}
