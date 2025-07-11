package client

type FactorizeResponse struct {
	// parameters
	N    uint64 `json:"N,omitempty"`
	T    uint64 `json:"t,omitempty"`
	A    uint64 `json:"a,omitempty"`
	Seed uint64 `json:"seed,omitempty"`

	// results
	P uint64 `json:"p,omitempty"`
	Q uint64 `json:"q,omitempty"`
	M string `json:"m,omitempty"`
	S uint64 `json:"s,omitempty"`
	R uint64 `json:"r,omitempty"`

	// message
	Message *string `json:"message,omitempty"`
}

type RunResponse struct {
	State []State `json:"state"`
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
