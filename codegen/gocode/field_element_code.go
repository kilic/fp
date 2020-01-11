package gocode

import (
	"fmt"
)

func fieldElementImpl(limbSize int) string {

	code0 := `
import (
	"fmt"
	"math/big"
	"encoding/hex"
)
`

	code1 := fmt.Sprintf(`
const byteSize = %d
const limbSize = %d

type fieldElement [%d]uint64
`, limbSize*8, limbSize, limbSize)

	code2 := `
func (fe *fieldElement) set(fe2 *fieldElement) *fieldElement {
	for i := 0; i < limbSize; i++ {
		fe[i] = fe2[i]
	}
	return fe
}

func (fe *fieldElement) cmp(fe2 *fieldElement) int8 {
	for i := limbSize-1; i > -1; i-- {
		if fe[i] > fe2[i] {
			return 1
		} else if fe[i] < fe2[i] {
			return -1
		}
	}
	return 0
}

func (fe *fieldElement) equal(fe2 *fieldElement) bool {
	for i := 0; i < limbSize; i++ {
		if fe[i] != fe2[i] {
			return false
		}
	}
	return true
}

func (fe *fieldElement) fromBytes(in []byte) (*fieldElement, error) {
	if len(in) != byteSize {
		return nil, fmt.Errorf("bad input size")
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
}

func (fe *fieldElement) fromBig(b *big.Int) (*fieldElement, error) {
	in := padBytes(b.Bytes(), byteSize)
	return fe.fromBytes(in)
}

func (fe *fieldElement) toBig() *big.Int {
	return new(big.Int).SetBytes(fe.toBytes())
}

func (fe *fieldElement) fromString(str string) (*fieldElement, error) {
	in := str
	if len(in) > 2 && in[:2] == "0x" {
		in = in[2:]
	}
	data, err := hex.DecodeString(in)
	if err != nil {
		return nil, err
	}
	return fe.fromBytes(padBytes(data, byteSize))
}

func (fe *fieldElement) toString() string {
	return 	hex.EncodeToString(fe.toBytes())
}
`
	return code0 + code1 + code2
}
