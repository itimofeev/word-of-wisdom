package pow

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSolveChallenge(t *testing.T) {
	difficulty := uint(4)
	p := New(10)
	data := []byte("hello, there1")
	nonce := p.SolveChallenge(data, difficulty)

	require.True(t, p.ValidateSolution(data, nonce, difficulty))

	hash := p.generateHash(data, nonce)
	fmt.Println(hash)
}
