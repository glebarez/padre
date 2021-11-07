package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandomSlice(t *testing.T) {
	var randoms = make([][]byte, 0, 10)

	// generate some random slices
	for i := 0; i < 10; i++ {
		newRandom := RandomSlice(13)
		// check uniqness
		require.NotContains(t, randoms, newRandom)
		randoms = append(randoms, newRandom)
	}

}
