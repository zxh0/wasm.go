package text

import (
	"math"
	"strconv"
	"strings"
)

func parseU32(s string) uint32 {
	base := 10
	s = strings.ReplaceAll(s, "_", "")
	if strings.Index(s, "0x") >= 0 {
		base = 16
		s = strings.Replace(s, "0x", "", 1)
	}

	i, err := strconv.ParseUint(s, base, 32)
	if err != nil {
		panic(err) // TODO
	}
	return uint32(i)
}

func parseI32(s string) int32 {
	return int32(parseInt(s, 32))
}

func parseI64(s string) int64 {
	return parseInt(s, 64)
}

func parseInt(s string, bitSize int) int64 {
	var i int64
	var err error
	base := 10

	s = strings.ReplaceAll(s, "_", "")
	if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	if strings.Index(s, "0x") >= 0 {
		s = strings.Replace(s, "0x", "", 1)
		base = 16
	}

	if strings.HasPrefix(s, "-") {
		i, err = strconv.ParseInt(s, base, bitSize)
	} else {
		var u uint64
		u, err = strconv.ParseUint(s, base, bitSize)
		i = int64(u)
	}

	if err != nil {
		panic(err) // TODO
	}
	return i
}

func parseF32(s string) float32 {
	if strings.Index(s, "nan") >= 0 {
		return parseNaN32(s)
	}
	return float32(parseFloat(s, 32))
}

func parseF64(s string) float64 {
	if strings.Index(s, "nan") >= 0 {
		return parseNaN64(s)
	}
	return parseFloat(s, 64)
}

func parseFloat(s string, bitSize int) float64 {
	s = strings.ReplaceAll(s, "_", "")
	if strings.Index(s, "0x") >= 0 &&
		strings.IndexByte(s, 'p') < 0 &&
		strings.IndexByte(s, 'P') < 0 {

		s += "p0"
	}

	f, err := strconv.ParseFloat(s, bitSize)
	if err != nil {
		panic(err) // TODO
	}
	return f
}
func parseNaN32(s string) float32 {
	s = strings.ReplaceAll(s, "_", "")
	f := float32(math.NaN())
	if s[0] == '-' {
		f = -f
		s = s[1:] // remove sign
	} else if s[0] == '+' {
		s = s[1:] // remove sign
	}
	if strings.HasPrefix(s, "nan:0x") {
		payload, err := strconv.ParseUint(s[6:], 16, 32)
		if err != nil {
			panic(err)
		}
		bits := math.Float32bits(f) & 0xFFBFFFFF
		f = math.Float32frombits(bits | uint32(payload))
	}
	return f
}
func parseNaN64(s string) float64 {
	s = strings.ReplaceAll(s, "_", "")
	f := math.NaN()
	if s[0] == '-' {
		f = -f
		s = s[1:] // remove sign
	} else if s[0] == '+' {
		s = s[1:] // remove sign
	}
	if strings.HasPrefix(s, "nan:0x") {
		payload, err := strconv.ParseUint(s[6:], 16, 64)
		if err != nil {
			panic(err)
		}
		bits := math.Float64bits(f) & 0xFFF7FFFFFFFFFFFE
		f = math.Float64frombits(bits | payload)
	} else {
		bits := math.Float64bits(f) & 0xFFFFFFFFFFFFFFFE
		f = math.Float64frombits(bits)
	}
	return f
}
