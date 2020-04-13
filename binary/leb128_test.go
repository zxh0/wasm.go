package binary

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVarUint(t *testing.T) {
	for i := 0; i < 100; i++ {
		val := rand.Uint64()
		data := encodeVarUint(val, 64)
		val2, n := decodeVarUint(data, 64)
		require.Equal(t, val, val2)
		require.Equal(t, len(data), n)
	}
	for i := 0; i < 100; i++ {
		val := rand.Uint32()
		data := encodeVarUint(uint64(val), 32)
		val2, n := decodeVarUint(data, 32)
		require.Equal(t, val, uint32(val2))
		require.Equal(t, len(data), n)
	}
}
func TestVarInt(t *testing.T) {
	for i := 0; i < 100; i++ {
		val := int64(rand.Uint64())
		data := encodeVarInt(val, 64)
		val2, n := decodeVarInt(data, 64)
		require.Equal(t, val, val2)
		require.Equal(t, len(data), n)
	}
	for i := 0; i < 100; i++ {
		val := int32(rand.Uint32())
		data := encodeVarInt(int64(val), 32)
		val2, n := decodeVarInt(data, 32)
		require.Equal(t, val, int32(val2))
		require.Equal(t, len(data), n)
	}
}

func TestDecodeVarUint(t *testing.T) {
	data := []byte{
		0b1_0111111,
		0b1_0011111,
		0b1_0001111,
		0b1_0000111,
		0b1_0000011,
		0b0_0000001}
	testDecodeVarUint32(t, data[5:], 0b0000001, 1)
	testDecodeVarUint32(t, data[4:], 0b1_0000011, 2)
	testDecodeVarUint32(t, data[3:], 0b1_0000011_0000111, 3)
	testDecodeVarUint32(t, data[2:], 0b1_0000011_0000111_0001111, 4)
	testDecodeVarUint32(t, data[1:], 0b1_0000011_0000111_0001111_0011111, 5)
	//testDecodeVarUint32(t, data[0:], 0, 0)
}

func TestDecodeVarInt(t *testing.T) {
	testDecodeVarInt32(t, []byte{0xC0, 0xBB, 0x78}, int32(-123456), 3)
	testDecodeVarInt32(t, []byte{0x7F}, int32(-1), 1)
	testDecodeVarInt32(t, []byte{0x7E}, int32(-2), 1)
	testDecodeVarInt32(t, []byte{0x7D}, int32(-3), 1)
	testDecodeVarInt32(t, []byte{0x7C}, int32(-4), 1)
	testDecodeVarInt32(t, []byte{0x40}, int32(-64), 1)
}

func testDecodeVarUint32(t *testing.T, data []byte, n uint32, w int) {
	_n, _w := decodeVarUint(data, 32)
	require.Equal(t, n, uint32(_n))
	require.Equal(t, w, _w)
}
func testDecodeVarInt32(t *testing.T, data []byte, n int32, w int) {
	_n, _w := decodeVarInt(data, 32)
	require.Equal(t, n, int32(_n))
	require.Equal(t, w, _w)
}
