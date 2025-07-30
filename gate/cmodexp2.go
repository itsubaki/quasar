package gate

import (
	"github.com/itsubaki/q"
	"github.com/itsubaki/q/math/number"
	"github.com/itsubaki/q/math/vector"
	"github.com/itsubaki/q/quantum/qubit"
)

// CModExp2 applies controlled modular exponentiation.
func CModExp2(qsim *q.Q, a, N int, control, target []q.Qubit) {
	for j := range control {
		ControlledModExp2(qsim.Underlying(), a, j, N, control[j].Index(), q.Index(target...))
	}
}

// ControlledModExp2 applies the controlled modular exponentiation operation.
// |j>|k> -> |j>|a**(2**j) * k mod N>.
func ControlledModExp2(qb *qubit.Qubit, a, j, N, control int, target []int) {
	n := qb.NumQubits()
	state := qb.Amplitude()
	a2jModN := number.ModExp2(a, j, N)
	cmask := 1 << (n - 1 - control)

	newState := make([]complex128, qb.Dim())
	for i := range qb.Dim() {
		if (i & cmask) == 0 {
			newState[i] += state[i]
			continue
		}

		// binary to integer
		var k int
		for j, t := range target {
			k |= ((i >> (n - 1 - t)) & 1) << (len(target) - 1 - j)
		}

		// a**(2**j) * k mod N
		a2jkModN := a2jModN * k % N

		// integer to binary
		newIdx := i
		for j, t := range target {
			bit := (a2jkModN >> (len(target) - 1 - j)) & 1
			pos := n - 1 - t
			newIdx = (newIdx & ^(1 << pos)) | (bit << pos)
		}

		// update the state
		newState[newIdx] += state[i]
	}

	// update the qubit state
	qb.Update(vector.New(newState...))
}
