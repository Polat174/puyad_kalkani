package toolchain_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestRapidVeTestifyBaglantisi(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		require.NotNil(t, t)
	})
}
