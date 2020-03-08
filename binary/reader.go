package binary

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type WasmReader struct {
	data []byte
}

func (reader *WasmReader) remaining() int {
	return len(reader.data)
}

func (reader *WasmReader) nextByte() byte {
	if len(reader.data) < 1 {
		return 0
	}
	return reader.data[0]
}

func (reader *WasmReader) readByte() (byte, error) {
	if len(reader.data) < 1 {
		return 0, io.EOF
	}
	b := reader.data[0]
	reader.data = reader.data[1:]
	return b, nil
}

func (reader *WasmReader) readBytes() ([]byte, error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	if len(reader.data) < int(n) {
		return nil, fmt.Errorf("insufficient bytes")
	}
	bytes := reader.data[:n]
	reader.data = reader.data[n:]
	return bytes, nil
}

func (reader *WasmReader) readName() (string, error) {
	bytes, err := reader.readBytes()
	return string(bytes), err
}

func (reader *WasmReader) readU32() (uint32, error) {
	if len(reader.data) < 4 {
		return 0, io.EOF
	}
	n := binary.LittleEndian.Uint32(reader.data)
	reader.data = reader.data[4:]
	return n, nil
}

func (reader *WasmReader) readF32() (float32, error) {
	if len(reader.data) < 4 {
		return 0, io.EOF
	}
	n := binary.LittleEndian.Uint32(reader.data)
	reader.data = reader.data[4:]
	return math.Float32frombits(n), nil
}

func (reader *WasmReader) readF64() (float64, error) {
	if len(reader.data) < 8 {
		return 0, io.EOF
	}
	n := binary.LittleEndian.Uint64(reader.data)
	reader.data = reader.data[8:]
	return math.Float64frombits(n), nil
}

func (reader *WasmReader) readVarU32() (uint32, error) {
	n, w := readVarUint(reader.data, 32)
	if w <= 0 {
		return 0, fmt.Errorf("LEB128 error")
	}
	reader.data = reader.data[w:]
	return uint32(n), nil
}

func (reader *WasmReader) readVarS32() (int32, error) {
	n, w := readVarInt(reader.data, 32)
	if w <= 0 {
		return 0, fmt.Errorf("LEB128 error")
	}
	reader.data = reader.data[w:]
	return int32(n), nil
}

func (reader *WasmReader) readVarS64() (int64, error) {
	n, w := readVarInt(reader.data, 64)
	if w <= 0 {
		return 0, fmt.Errorf("LEB128 error")
	}
	reader.data = reader.data[w:]
	return n, nil
}

// io.Reader
func (reader *WasmReader) Read(p []byte) (n int, err error) {
	if len(reader.data) < len(p) {
		return 0, io.EOF
	}

	n = copy(p, reader.data)
	reader.data = reader.data[n:]
	return
}
