package text

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompileErrors(t *testing.T) {
	files, err := ioutil.ReadDir("./testdata")
	require.NoError(t, err)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "err") {
			testCompileErr(t, "./testdata/"+f.Name())
		}
	}
}
func testCompileErr(t *testing.T, filename string) {
	data, err := ioutil.ReadFile(filename)
	require.NoError(t, err)

	testWAT := string(data)
	sep := ";;------------------------------;;"
	if strings.Index(testWAT, sep) < 0 {
		expectedErr := getExpectedErr(filename, testWAT)
		_, err := CompileModuleFile(filename)
		require.Error(t, err, filename)
		require.Equal(t, expectedErr, err.Error())
	} else {
		for _, wat := range strings.Split(testWAT, sep) {
			wat = strings.TrimSpace(wat)
			expectedErr := getExpectedErr("Obtained from string", wat)
			_, err := CompileModuleStr(wat)
			require.Error(t, err, filename+"\n"+expectedErr)
			require.Equal(t, expectedErr, err.Error())
		}
	}
}
func getExpectedErr(filename, wat string) string {
	start := strings.Index(wat, "(;;") + 3
	end := strings.Index(wat, ";;)")
	err := strings.TrimSpace(wat[start:end])
	return strings.ReplaceAll(err, "err.wat", filename)
}
