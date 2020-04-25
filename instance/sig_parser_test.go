package instance

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSigParser(t *testing.T) {
	testSigParser(t, "(i32,f64)->(f32,i64)")
	testSigParser(t, "(i32)->(f32,i64)")
	testSigParser(t, "()->(f32)")
	testSigParser(t, "(i32)->()")
}

func testSigParser(t *testing.T, nameAndSig string) {
	name, sig := parseNameAndSig(nameAndSig)
	require.Equal(t, nameAndSig, name+sig.GetSignature())
}
