package binary

// https://en.wikipedia.org/wiki/LEB128#Encode_unsigned_integer
func encodeVarUint(val uint64, size int) []byte {
	buf := make([]byte, 0, 2)
	for {
		b := val & 0x7f
		val >>= 7
		if val != 0 {
			b |= 0x80
		}
		buf = append(buf, byte(b))
		if val == 0 {
			break
		}
	}
	return buf
}

// https://en.wikipedia.org/wiki/LEB128#Decode_unsigned_integer
func decodeVarUint(data []byte, size int) (uint64, int) {
	result := uint64(0)
	for i, b := range data {
		if i > size/7 {
			break
		}
		result |= (uint64(b) & 0x7f) << (i * 7)
		if b&0x80 == 0 {
			return result, i + 1
		}
	}
	return 0, 0
}

// https://en.wikipedia.org/wiki/LEB128#Encode_signed_integer
func encodeVarInt(val int64, size int) []byte {
	buf := make([]byte, 0, 2)
	more := true
	for more {
		b := val & 0x7f
		val >>= 7
		if (val == 0 && (0x40&b == 0)) || (val == -1 && (0x40&b != 0)) {
			more = false
		} else {
			b |= 0x80
		}
		buf = append(buf, byte(b))
	}
	return buf
}

// https://en.wikipedia.org/wiki/LEB128#Decode_signed_integer
func decodeVarInt(data []byte, size int) (int64, int) {
	result := int64(0)
	for i, b := range data {
		if i > size/7 {
			break
		}
		result |= (int64(b) & 0x7f) << (i * 7)
		if b&0x80 == 0 {
			if (i*7 < size) && (b&0x40 != 0) {
				result = result | (-1 << ((i + 1) * 7))
			}
			return result, i + 1
		}
	}
	return 0, 0
}
