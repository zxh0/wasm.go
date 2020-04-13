package instance

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func f1()         {}
func f2() int32   { return 0 }
func f3() int64   { return 0 }
func f4() float32 { return 0 }
func f5() float64 { return 0 }

func f6(a int32, b int64, c float32, d float64) {}

func add(a, b int32) int32 {
	return a + b
}

func TestNativeFuncCall(t *testing.T) {
	nf, err := wrapNativeFunc(add)
	require.NoError(t, err)
	results, err := nf.Call(int32(1), int32(2))
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, int32(3), results[0])
}

func TestNativeFuncType(t *testing.T) {
	require.Equal(t, "()->()", getNativeFuncSig(f1))
	require.Equal(t, "()->(i32)", getNativeFuncSig(f2))
	require.Equal(t, "()->(i64)", getNativeFuncSig(f3))
	require.Equal(t, "()->(f32)", getNativeFuncSig(f4))
	require.Equal(t, "()->(f64)", getNativeFuncSig(f5))
	require.Equal(t, "(i32,i64,f32,f64)->()", getNativeFuncSig(f6))
}

func getNativeFuncSig(nf interface{}) string {
	ft, err := getNativeFuncType(nf)
	if err != nil {
		panic(err)
	}
	return ft.GetSignature()
}
