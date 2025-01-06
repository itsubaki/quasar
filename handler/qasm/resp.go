package qasm

type State struct {
	Amplitude    Amplitude `json:"amplitude"`
	Probability  float64   `json:"probability"`
	Int          int64     `json:"int"`
	BinaryString string    `json:"binary_string"`
}

type Amplitude struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
}

type Response struct {
	State []State `json:"state"`
}
