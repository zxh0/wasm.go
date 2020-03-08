package text

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// https://en.wikipedia.org/wiki/NaN
func TestNaN(t *testing.T) {
	require.Equal(t, uint32(0x7fc00000), math.Float32bits(float32(math.NaN())))
	require.Equal(t, uint32(0xffc00000), math.Float32bits(float32(-math.NaN())))
	require.Equal(t, uint32(0x7f800000), math.Float32bits(float32(math.Inf(1))))
	require.Equal(t, uint32(0xff800000), math.Float32bits(float32(math.Inf(-1))))
	require.Equal(t, uint64(0x7ff8000000000001), math.Float64bits(math.NaN()))
	require.Equal(t, uint64(0xfff8000000000001), math.Float64bits(-math.NaN()))
	require.Equal(t, uint64(0x7ff0000000000000), math.Float64bits(math.Inf(1)))
	require.Equal(t, uint64(0xfff0000000000000), math.Float64bits(math.Inf(-1)))
}

func TestParseFloat(t *testing.T) {
	f1 := parseF32("+0x1.00000100000000001p-50")
	f2 := parseF32("+0x1.000002p-50")
	require.Equal(t, f1, f2, fmt.Sprintf("%b != %b\n", f1, f2))
}
