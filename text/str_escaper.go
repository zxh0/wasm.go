package text

import "strconv"

func escape(s string) []byte {
	n := len(s)
	data := make([]byte, 0, n)

	for i := 0; i < n; i++ {
		if s[i] != '\\' {
			data = append(data, s[i])
		} else {
			switch s[i+1] {
			case 't':
				data = append(data, '\t')
			case 'n':
				data = append(data, '\n')
			case 'r':
				data = append(data, '\r')
			case '"':
				data = append(data, '"')
			case '\\':
				data = append(data, '\\')
			case 'u':
				panic("TODO")
			default:
				k, _ := strconv.ParseUint(s[i+1:i+3], 16, 32)
				data = append(data, byte(k))
				i += 2
			}
		}
	}

	// TODO
	return data
}
