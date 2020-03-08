package binary

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReads(t *testing.T) {
	reader := WasmReader{data: []byte{
		0x01,
		0x02, 0x03, 0x04, 0x05,
		0x00, 0x00, 0xc0, 0x3f,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf8, 0x3f,
		0xE5, 0x8E, 0x26, // https://en.wikipedia.org/wiki/LEB128#Unsigned_LEB128
		0xC0, 0xBB, 0x78, // https://en.wikipedia.org/wiki/LEB128#Signed_LEB128
		0xC0, 0xBB, 0x78,
		0x03, 0x01, 0x02, 0x03,
		0x03, 0x66, 0x6f, 0x6f,
	}}
	require.Equal(t, byte(0x01), discardError(reader.readByte()))
	require.Equal(t, uint32(0x05040302), discardError(reader.readU32()))
	require.Equal(t, float32(1.5), discardError(reader.readF32()))
	require.Equal(t, 1.5, discardError(reader.readF64()))
	require.Equal(t, uint32(624485), discardError(reader.readVarU32()))
	require.Equal(t, int32(-123456), discardError(reader.readVarS32()))
	require.Equal(t, int64(-123456), discardError(reader.readVarS64()))
	require.Equal(t, []byte{0x01, 0x02, 0x03}, discardError(reader.readBytes()))
	require.Equal(t, "foo", discardError(reader.readName()))
	require.Equal(t, 0, reader.remaining())
}

func discardError(x interface{}, err error) interface{} {
	if err != nil {
		return err
	}
	return x
}
