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
	var N, t, a, seed uint64
	flag.Uint64Var(&N, "N", 15, "positive integer")
	flag.Uint64Var(&t, "t", 3, "precision bits")
	flag.Uint64Var(&a, "a", 0, "coprime number of N")
	flag.Uint64Var(&seed, "seed", 0, "PRNG seed for measurements")
	flag.Parse()

	resp, err := client.
		New(TargetURL, client.NewWithIdentityToken(IdentityToken)).
		Factorize(context.Background(), N, &t, &a, &seed)
	if err != nil {
		panic(err)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
}
