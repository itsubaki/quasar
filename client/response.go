package client

type FactorizeResponse struct {
	// parameters
	N    int    `json:"N,omitempty"`
	T    int    `json:"t,omitempty"`
	A    int    `json:"a,omitempty"`
	Seed uint64 `json:"seed,omitempty"`

	// results
	P  int    `json:"p,omitempty"`
	Q  int    `json:"q,omitempty"`
	M  string `json:"m,omitempty"`
	SR string `json:"s/r,omitempty"`

	// message
	Message string `json:"message,omitempty"`
}

type RunResponse struct {
	State []State `json:"state"`
}

type State struct {
	Amplitude    Amplitude `json:"amplitude"`
	Probability  float64   `json:"probability"`
	Int          []int64   `json:"int"`
	BinaryString []string  `json:"binary_string"`
}

type Amplitude struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
}
