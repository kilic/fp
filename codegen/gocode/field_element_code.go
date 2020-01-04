package gocode

import (
	"fmt"
)

func fieldElementImpl(limbSize int) string {

	code0 := `
import (
	"fmt"
)
`

	code1 := fmt.Sprintf(`
const byteSize = %d
const limbSize = %d

type fieldElement [%d]uint64
`, limbSize*8, limbSize, limbSize)

	code2 := `
func (fe fieldElement) Set(fe2 *fieldElement) {
	fe[0] = fe2[0]
	fe[1] = fe2[1]
	fe[2] = fe2[2]
	fe[3] = fe2[3]
}

func (fe *fieldElement) fromBytes(in []byte) (*fieldElement, error) {
	if len(in) != byteSize {
		return nil, fmt.Errorf("bad input len")
	}
	var a int
	var size = byteSize
	for i := 0; i < limbSize; i++ {
		a = size - i*8
		fe[i] = uint64(in[a-1]) | uint64(in[a-2])<<8 |
			uint64(in[a-3])<<16 | uint64(in[a-4])<<24 |
			uint64(in[a-5])<<32 | uint64(in[a-6])<<40 |
			uint64(in[a-7])<<48 | uint64(in[a-8])<<56
	}
	return fe, nil
}

func (fe *fieldElement) toBytes() []byte {
	out := make([]byte, byteSize)
	var a int
	for i := 0; i < limbSize; i++ {
		a = byteSize - i*8
		out[a-1] = byte(fe[i])
		out[a-2] = byte(fe[i] >> 8)
		out[a-3] = byte(fe[i] >> 16)
		out[a-4] = byte(fe[i] >> 24)
		out[a-5] = byte(fe[i] >> 32)
		out[a-6] = byte(fe[i] >> 40)
		out[a-7] = byte(fe[i] >> 48)
		out[a-8] = byte(fe[i] >> 56)
	}
	return out
}`
	return code0 + code1 + code2
}
