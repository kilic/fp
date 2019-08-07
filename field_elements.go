package fp

import (
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"math/bits"
)

func (fe *Fe256) Bytes() []byte {
	out := make([]byte, 32)
	var a int
	for i := 0; i < 4; i++ {
		a = 32 - i*8
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

func (fe *Fe256) FromBytes(in []byte) *Fe256 {
	size := 32
	l := len(in)
	if l >= size {
		l = size
	}
	padded := make([]byte, size)
	copy(padded[size-l:], in[:])
	var a int
	for i := 0; i < 4; i++ {
		a = size - i*8
		fe[i] = uint64(padded[a-1]) | uint64(padded[a-2])<<8 |
			uint64(padded[a-3])<<16 | uint64(padded[a-4])<<24 |
			uint64(padded[a-5])<<32 | uint64(padded[a-6])<<40 |
			uint64(padded[a-7])<<48 | uint64(padded[a-8])<<56
	}
	return fe
}

func (fe *Fe256) SetBig(a *big.Int) *Fe256 {
	return fe.FromBytes(a.Bytes())
}

func (fe *Fe256) SetUint(a uint64) *Fe256 {
	fe[0] = a
	fe[1] = 0
	fe[2] = 0
	fe[3] = 0
	return fe
}

func (fe *Fe256) SetString(s string) (*Fe256, error) {
	if s[:2] == "0x" {
		s = s[2:]
	}
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return fe.FromBytes(bytes), nil
}

func (fe *Fe256) Set(fe2 *Fe256) *Fe256 {
	fe[0] = fe2[0]
	fe[1] = fe2[1]
	fe[2] = fe2[2]
	fe[3] = fe2[3]
	return fe
}

func (fe *Fe256) Big() *big.Int {
	return new(big.Int).SetBytes(fe.Bytes())
}

func (fe Fe256) String() (s string) {
	for i := 3; i >= 0; i-- {
		s = fmt.Sprintf("%s%16.16x", s, fe[i])
	}
	return "0x" + s
}

func (fe *Fe256) IsOdd() bool {
	var mask uint64 = 1
	return fe[0]&mask != 0
}

func (fe *Fe256) IsEven() bool {
	var mask uint64 = 1
	return fe[0]&mask == 0
}

func (fe *Fe256) IsZero() bool {
	return 0 == fe[0] && 0 == fe[1] && 0 == fe[2] && 0 == fe[3]
}

func (fe *Fe256) IsOne() bool {
	return 1 == fe[0] && 0 == fe[1] && 0 == fe[2] && 0 == fe[3]
}

func (fe *Fe256) Cmp(fe2 *Fe256) int64 {
	if fe[3] > fe2[3] {
		return 1
	} else if fe[3] < fe2[3] {
		return -1
	}
	if fe[2] > fe2[2] {
		return 1
	} else if fe[2] < fe2[2] {
		return -1
	}
	if fe[1] > fe2[1] {
		return 1
	} else if fe[1] < fe2[1] {
		return -1
	}
	if fe[0] > fe2[0] {
		return 1
	} else if fe[0] < fe2[0] {
		return -1
	}
	return 0
}

func (fe *Fe256) Equals(fe2 *Fe256) bool {
	return fe2[0] == fe[0] && fe2[1] == fe[1] && fe2[2] == fe[2] && fe2[3] == fe[3]
}

func (fe *Fe256) div2(e uint64) {
	fe[0] = fe[0]>>1 | fe[1]<<63
	fe[1] = fe[1]>>1 | fe[2]<<63
	fe[2] = fe[2]>>1 | fe[3]<<63
	fe[3] = fe[3]>>1 | e<<63
}

func (fe *Fe256) mul2() uint64 {
	e := fe[3] >> 63
	fe[3] = fe[3]<<1 | fe[2]>>63
	fe[2] = fe[2]<<1 | fe[1]>>63
	fe[1] = fe[1]<<1 | fe[0]>>63
	fe[0] = fe[0] << 1
	return e
}

func (fe *Fe256) bit(i int) bool {
	k := i >> 6
	i = i - k<<6
	b := (fe[k] >> uint(i)) & 1
	return b != 0
}

func (fe *Fe256) bitLen() int {
	for i := len(fe) - 1; i >= 0; i-- {
		if len := bits.Len64(fe[i]); len != 0 {
			return len + 64*i
		}
	}
	return 0
}

func (f *Fe256) rand(max *Fe256, r io.Reader) error {
	bitLen := bits.Len64(max[3]) + (4-1)*64
	k := (bitLen + 7) / 8
	b := uint(bitLen % 8)
	if b == 0 {
		b = 8
	}
	bytes := make([]byte, k)
	for {
		_, err := io.ReadFull(r, bytes)
		if err != nil {
			return err
		}
		bytes[0] &= uint8(int(1<<b) - 1)
		f.FromBytes(bytes)
		if f.Cmp(max) < 0 {
			break
		}
	}
	return nil
}

func (fe *Fe320) Bytes() []byte {
	out := make([]byte, 40)
	var a int
	for i := 0; i < 5; i++ {
		a = 40 - i*8
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

func (fe *Fe320) FromBytes(in []byte) *Fe320 {
	size := 40
	l := len(in)
	if l >= size {
		l = size
	}
	padded := make([]byte, size)
	copy(padded[size-l:], in[:])
	var a int
	for i := 0; i < 5; i++ {
		a = size - i*8
		fe[i] = uint64(padded[a-1]) | uint64(padded[a-2])<<8 |
			uint64(padded[a-3])<<16 | uint64(padded[a-4])<<24 |
			uint64(padded[a-5])<<32 | uint64(padded[a-6])<<40 |
			uint64(padded[a-7])<<48 | uint64(padded[a-8])<<56
	}
	return fe
}

func (fe *Fe320) SetBig(a *big.Int) *Fe320 {
	return fe.FromBytes(a.Bytes())
}

func (fe *Fe320) SetUint(a uint64) *Fe320 {
	fe[0] = a
	fe[1] = 0
	fe[2] = 0
	fe[3] = 0
	fe[4] = 0
	return fe
}

func (fe *Fe320) SetString(s string) (*Fe320, error) {
	if s[:2] == "0x" {
		s = s[2:]
	}
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return fe.FromBytes(bytes), nil
}

func (fe *Fe320) Set(fe2 *Fe320) *Fe320 {
	fe[0] = fe2[0]
	fe[1] = fe2[1]
	fe[2] = fe2[2]
	fe[3] = fe2[3]
	fe[4] = fe2[4]
	return fe
}

func (fe *Fe320) Big() *big.Int {
	return new(big.Int).SetBytes(fe.Bytes())
}

func (fe Fe320) String() (s string) {
	for i := 4; i >= 0; i-- {
		s = fmt.Sprintf("%s%16.16x", s, fe[i])
	}
	return "0x" + s
}

func (fe *Fe320) IsOdd() bool {
	var mask uint64 = 1
	return fe[0]&mask != 0
}

func (fe *Fe320) IsEven() bool {
	var mask uint64 = 1
	return fe[0]&mask == 0
}

func (fe *Fe320) IsZero() bool {
	return 0 == fe[0] && 0 == fe[1] && 0 == fe[2] && 0 == fe[3] && 0 == fe[4]
}

func (fe *Fe320) IsOne() bool {
	return 1 == fe[0] && 0 == fe[1] && 0 == fe[2] && 0 == fe[3] && 0 == fe[4]
}

func (fe *Fe320) Cmp(fe2 *Fe320) int64 {
	if fe[4] > fe2[4] {
		return 1
	} else if fe[4] < fe2[4] {
		return -1
	}
	if fe[3] > fe2[3] {
		return 1
	} else if fe[3] < fe2[3] {
		return -1
	}
	if fe[2] > fe2[2] {
		return 1
	} else if fe[2] < fe2[2] {
		return -1
	}
	if fe[1] > fe2[1] {
		return 1
	} else if fe[1] < fe2[1] {
		return -1
	}
	if fe[0] > fe2[0] {
		return 1
	} else if fe[0] < fe2[0] {
		return -1
	}
	return 0
}

func (fe *Fe320) Equals(fe2 *Fe320) bool {
	return fe2[0] == fe[0] && fe2[1] == fe[1] && fe2[2] == fe[2] && fe2[3] == fe[3] && fe2[4] == fe[4]
}

func (fe *Fe320) div2(e uint64) {
	fe[0] = fe[0]>>1 | fe[1]<<63
	fe[1] = fe[1]>>1 | fe[2]<<63
	fe[2] = fe[2]>>1 | fe[3]<<63
	fe[3] = fe[3]>>1 | fe[4]<<63
	fe[4] = fe[4]>>1 | e<<63
}

func (fe *Fe320) mul2() uint64 {
	e := fe[4] >> 63
	fe[4] = fe[4]<<1 | fe[3]>>63
	fe[3] = fe[3]<<1 | fe[2]>>63
	fe[2] = fe[2]<<1 | fe[1]>>63
	fe[1] = fe[1]<<1 | fe[0]>>63
	fe[0] = fe[0] << 1
	return e
}

func (fe *Fe320) bit(i int) bool {
	k := i >> 6
	i = i - k<<6
	b := (fe[k] >> uint(i)) & 1
	return b != 0
}

func (fe *Fe320) bitLen() int {
	for i := len(fe) - 1; i >= 0; i-- {
		if len := bits.Len64(fe[i]); len != 0 {
			return len + 64*i
		}
	}
	return 0
}

func (f *Fe320) rand(max *Fe320, r io.Reader) error {
	bitLen := bits.Len64(max[4]) + (5-1)*64
	k := (bitLen + 7) / 8
	b := uint(bitLen % 8)
	if b == 0 {
		b = 8
	}
	bytes := make([]byte, k)
	for {
		_, err := io.ReadFull(r, bytes)
		if err != nil {
			return err
		}
		bytes[0] &= uint8(int(1<<b) - 1)
		f.FromBytes(bytes)
		if f.Cmp(max) < 0 {
			break
		}
	}
	return nil
}

func (fe *Fe384) Bytes() []byte {
	out := make([]byte, 48)
	var a int
	for i := 0; i < 6; i++ {
		a = 48 - i*8
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

func (fe *Fe384) FromBytes(in []byte) *Fe384 {
	size := 48
	l := len(in)
	if l >= size {
		l = size
	}
	padded := make([]byte, size)
	copy(padded[size-l:], in[:])
	var a int
	for i := 0; i < 6; i++ {
		a = size - i*8
		fe[i] = uint64(padded[a-1]) | uint64(padded[a-2])<<8 |
			uint64(padded[a-3])<<16 | uint64(padded[a-4])<<24 |
			uint64(padded[a-5])<<32 | uint64(padded[a-6])<<40 |
			uint64(padded[a-7])<<48 | uint64(padded[a-8])<<56
	}
	return fe
}

func (fe *Fe384) SetBig(a *big.Int) *Fe384 {
	return fe.FromBytes(a.Bytes())
}

func (fe *Fe384) SetUint(a uint64) *Fe384 {
	fe[0] = a
	fe[1] = 0
	fe[2] = 0
	fe[3] = 0
	fe[4] = 0
	fe[5] = 0
	return fe
}

func (fe *Fe384) SetString(s string) (*Fe384, error) {
	if s[:2] == "0x" {
		s = s[2:]
	}
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return fe.FromBytes(bytes), nil
}

func (fe *Fe384) Set(fe2 *Fe384) *Fe384 {
	fe[0] = fe2[0]
	fe[1] = fe2[1]
	fe[2] = fe2[2]
	fe[3] = fe2[3]
	fe[4] = fe2[4]
	fe[5] = fe2[5]
	return fe
}

func (fe *Fe384) Big() *big.Int {
	return new(big.Int).SetBytes(fe.Bytes())
}

func (fe Fe384) String() (s string) {
	for i := 5; i >= 0; i-- {
		s = fmt.Sprintf("%s%16.16x", s, fe[i])
	}
	return "0x" + s
}

func (fe *Fe384) IsOdd() bool {
	var mask uint64 = 1
	return fe[0]&mask != 0
}

func (fe *Fe384) IsEven() bool {
	var mask uint64 = 1
	return fe[0]&mask == 0
}

func (fe *Fe384) IsZero() bool {
	return 0 == fe[0] && 0 == fe[1] && 0 == fe[2] && 0 == fe[3] && 0 == fe[4] && 0 == fe[5]
}

func (fe *Fe384) IsOne() bool {
	return 1 == fe[0] && 0 == fe[1] && 0 == fe[2] && 0 == fe[3] && 0 == fe[4] && 0 == fe[5]
}

func (fe *Fe384) Cmp(fe2 *Fe384) int64 {
	if fe[5] > fe2[5] {
		return 1
	} else if fe[5] < fe2[5] {
		return -1
	}
	if fe[4] > fe2[4] {
		return 1
	} else if fe[4] < fe2[4] {
		return -1
	}
	if fe[3] > fe2[3] {
		return 1
	} else if fe[3] < fe2[3] {
		return -1
	}
	if fe[2] > fe2[2] {
		return 1
	} else if fe[2] < fe2[2] {
		return -1
	}
	if fe[1] > fe2[1] {
		return 1
	} else if fe[1] < fe2[1] {
		return -1
	}
	if fe[0] > fe2[0] {
		return 1
	} else if fe[0] < fe2[0] {
		return -1
	}
	return 0
}

func (fe *Fe384) Equals(fe2 *Fe384) bool {
	return fe2[0] == fe[0] && fe2[1] == fe[1] && fe2[2] == fe[2] && fe2[3] == fe[3] && fe2[4] == fe[4] && fe2[5] == fe[5]
}

func (fe *Fe384) div2(e uint64) {
	fe[0] = fe[0]>>1 | fe[1]<<63
	fe[1] = fe[1]>>1 | fe[2]<<63
	fe[2] = fe[2]>>1 | fe[3]<<63
	fe[3] = fe[3]>>1 | fe[4]<<63
	fe[4] = fe[4]>>1 | fe[5]<<63
	fe[5] = fe[5]>>1 | e<<63
}

func (fe *Fe384) mul2() uint64 {
	e := fe[5] >> 63
	fe[5] = fe[5]<<1 | fe[4]>>63
	fe[4] = fe[4]<<1 | fe[3]>>63
	fe[3] = fe[3]<<1 | fe[2]>>63
	fe[2] = fe[2]<<1 | fe[1]>>63
	fe[1] = fe[1]<<1 | fe[0]>>63
	fe[0] = fe[0] << 1
	return e
}

func (fe *Fe384) bit(i int) bool {
	k := i >> 6
	i = i - k<<6
	b := (fe[k] >> uint(i)) & 1
	return b != 0
}

func (fe *Fe384) bitLen() int {
	for i := len(fe) - 1; i >= 0; i-- {
		if len := bits.Len64(fe[i]); len != 0 {
			return len + 64*i
		}
	}
	return 0
}

func (f *Fe384) rand(max *Fe384, r io.Reader) error {
	bitLen := bits.Len64(max[5]) + (6-1)*64
	k := (bitLen + 7) / 8
	b := uint(bitLen % 8)
	if b == 0 {
		b = 8
	}
	bytes := make([]byte, k)
	for {
		_, err := io.ReadFull(r, bytes)
		if err != nil {
			return err
		}
		bytes[0] &= uint8(int(1<<b) - 1)
		f.FromBytes(bytes)
		if f.Cmp(max) < 0 {
			break
		}
	}
	return nil
}

func (fe *Fe448) Bytes() []byte {
	out := make([]byte, 56)
	var a int
	for i := 0; i < 7; i++ {
		a = 56 - i*8
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

func (fe *Fe448) FromBytes(in []byte) *Fe448 {
	size := 56
	l := len(in)
	if l >= size {
		l = size
	}
	padded := make([]byte, size)
	copy(padded[size-l:], in[:])
	var a int
	for i := 0; i < 7; i++ {
		a = size - i*8
		fe[i] = uint64(padded[a-1]) | uint64(padded[a-2])<<8 |
			uint64(padded[a-3])<<16 | uint64(padded[a-4])<<24 |
			uint64(padded[a-5])<<32 | uint64(padded[a-6])<<40 |
			uint64(padded[a-7])<<48 | uint64(padded[a-8])<<56
	}
	return fe
}

func (fe *Fe448) SetBig(a *big.Int) *Fe448 {
	return fe.FromBytes(a.Bytes())
}

func (fe *Fe448) SetUint(a uint64) *Fe448 {
	fe[0] = a
	fe[1] = 0
	fe[2] = 0
	fe[3] = 0
	fe[4] = 0
	fe[5] = 0
	fe[6] = 0
	return fe
}

func (fe *Fe448) SetString(s string) (*Fe448, error) {
	if s[:2] == "0x" {
		s = s[2:]
	}
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return fe.FromBytes(bytes), nil
}

func (fe *Fe448) Set(fe2 *Fe448) *Fe448 {
	fe[0] = fe2[0]
	fe[1] = fe2[1]
	fe[2] = fe2[2]
	fe[3] = fe2[3]
	fe[4] = fe2[4]
	fe[5] = fe2[5]
	fe[6] = fe2[6]
	return fe
}

func (fe *Fe448) Big() *big.Int {
	return new(big.Int).SetBytes(fe.Bytes())
}

func (fe Fe448) String() (s string) {
	for i := 6; i >= 0; i-- {
		s = fmt.Sprintf("%s%16.16x", s, fe[i])
	}
	return "0x" + s
}

func (fe *Fe448) IsOdd() bool {
	var mask uint64 = 1
	return fe[0]&mask != 0
}

func (fe *Fe448) IsEven() bool {
	var mask uint64 = 1
	return fe[0]&mask == 0
}

func (fe *Fe448) IsZero() bool {
	return 0 == fe[0] && 0 == fe[1] && 0 == fe[2] && 0 == fe[3] && 0 == fe[4] && 0 == fe[5] && 0 == fe[6]
}

func (fe *Fe448) IsOne() bool {
	return 1 == fe[0] && 0 == fe[1] && 0 == fe[2] && 0 == fe[3] && 0 == fe[4] && 0 == fe[5] && 0 == fe[6]
}

func (fe *Fe448) Cmp(fe2 *Fe448) int64 {
	if fe[6] > fe2[6] {
		return 1
	} else if fe[6] < fe2[6] {
		return -1
	}
	if fe[5] > fe2[5] {
		return 1
	} else if fe[5] < fe2[5] {
		return -1
	}
	if fe[4] > fe2[4] {
		return 1
	} else if fe[4] < fe2[4] {
		return -1
	}
	if fe[3] > fe2[3] {
		return 1
	} else if fe[3] < fe2[3] {
		return -1
	}
	if fe[2] > fe2[2] {
		return 1
	} else if fe[2] < fe2[2] {
		return -1
	}
	if fe[1] > fe2[1] {
		return 1
	} else if fe[1] < fe2[1] {
		return -1
	}
	if fe[0] > fe2[0] {
		return 1
	} else if fe[0] < fe2[0] {
		return -1
	}
	return 0
}

func (fe *Fe448) Equals(fe2 *Fe448) bool {
	return fe2[0] == fe[0] && fe2[1] == fe[1] && fe2[2] == fe[2] && fe2[3] == fe[3] && fe2[4] == fe[4] && fe2[5] == fe[5] && fe2[6] == fe[6]
}

func (fe *Fe448) div2(e uint64) {
	fe[0] = fe[0]>>1 | fe[1]<<63
	fe[1] = fe[1]>>1 | fe[2]<<63
	fe[2] = fe[2]>>1 | fe[3]<<63
	fe[3] = fe[3]>>1 | fe[4]<<63
	fe[4] = fe[4]>>1 | fe[5]<<63
	fe[5] = fe[5]>>1 | fe[6]<<63
	fe[6] = fe[6]>>1 | e<<63
}

func (fe *Fe448) mul2() uint64 {
	e := fe[6] >> 63
	fe[6] = fe[6]<<1 | fe[5]>>63
	fe[5] = fe[5]<<1 | fe[4]>>63
	fe[4] = fe[4]<<1 | fe[3]>>63
	fe[3] = fe[3]<<1 | fe[2]>>63
	fe[2] = fe[2]<<1 | fe[1]>>63
	fe[1] = fe[1]<<1 | fe[0]>>63
	fe[0] = fe[0] << 1
	return e
}

func (fe *Fe448) bit(i int) bool {
	k := i >> 6
	i = i - k<<6
	b := (fe[k] >> uint(i)) & 1
	return b != 0
}

func (fe *Fe448) bitLen() int {
	for i := len(fe) - 1; i >= 0; i-- {
		if len := bits.Len64(fe[i]); len != 0 {
			return len + 64*i
		}
	}
	return 0
}

func (f *Fe448) rand(max *Fe448, r io.Reader) error {
	bitLen := bits.Len64(max[6]) + (7-1)*64
	k := (bitLen + 7) / 8
	b := uint(bitLen % 8)
	if b == 0 {
		b = 8
	}
	bytes := make([]byte, k)
	for {
		_, err := io.ReadFull(r, bytes)
		if err != nil {
			return err
		}
		bytes[0] &= uint8(int(1<<b) - 1)
		f.FromBytes(bytes)
		if f.Cmp(max) < 0 {
			break
		}
	}
	return nil
}

func (fe *Fe512) Bytes() []byte {
	out := make([]byte, 64)
	var a int
	for i := 0; i < 8; i++ {
		a = 64 - i*8
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

func (fe *Fe512) FromBytes(in []byte) *Fe512 {
	size := 64
	l := len(in)
	if l >= size {
		l = size
	}
	padded := make([]byte, size)
	copy(padded[size-l:], in[:])
	var a int
	for i := 0; i < 8; i++ {
		a = size - i*8
		fe[i] = uint64(padded[a-1]) | uint64(padded[a-2])<<8 |
			uint64(padded[a-3])<<16 | uint64(padded[a-4])<<24 |
			uint64(padded[a-5])<<32 | uint64(padded[a-6])<<40 |
			uint64(padded[a-7])<<48 | uint64(padded[a-8])<<56
	}
	return fe
}

func (fe *Fe512) SetBig(a *big.Int) *Fe512 {
	return fe.FromBytes(a.Bytes())
}

func (fe *Fe512) SetUint(a uint64) *Fe512 {
	fe[0] = a
	fe[1] = 0
	fe[2] = 0
	fe[3] = 0
	fe[4] = 0
	fe[5] = 0
	fe[6] = 0
	fe[7] = 0
	return fe
}

func (fe *Fe512) SetString(s string) (*Fe512, error) {
	if s[:2] == "0x" {
		s = s[2:]
	}
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return fe.FromBytes(bytes), nil
}

func (fe *Fe512) Set(fe2 *Fe512) *Fe512 {
	fe[0] = fe2[0]
	fe[1] = fe2[1]
	fe[2] = fe2[2]
	fe[3] = fe2[3]
	fe[4] = fe2[4]
	fe[5] = fe2[5]
	fe[6] = fe2[6]
	fe[7] = fe2[7]
	return fe
}

func (fe *Fe512) Big() *big.Int {
	return new(big.Int).SetBytes(fe.Bytes())
}

func (fe Fe512) String() (s string) {
	for i := 7; i >= 0; i-- {
		s = fmt.Sprintf("%s%16.16x", s, fe[i])
	}
	return "0x" + s
}

func (fe *Fe512) IsOdd() bool {
	var mask uint64 = 1
	return fe[0]&mask != 0
}

func (fe *Fe512) IsEven() bool {
	var mask uint64 = 1
	return fe[0]&mask == 0
}

func (fe *Fe512) IsZero() bool {
	return 0 == fe[0] && 0 == fe[1] && 0 == fe[2] && 0 == fe[3] && 0 == fe[4] && 0 == fe[5] && 0 == fe[6] && 0 == fe[7]
}

func (fe *Fe512) IsOne() bool {
	return 1 == fe[0] && 0 == fe[1] && 0 == fe[2] && 0 == fe[3] && 0 == fe[4] && 0 == fe[5] && 0 == fe[6] && 0 == fe[7]
}

func (fe *Fe512) Cmp(fe2 *Fe512) int64 {
	if fe[7] > fe2[7] {
		return 1
	} else if fe[7] < fe2[7] {
		return -1
	}
	if fe[6] > fe2[6] {
		return 1
	} else if fe[6] < fe2[6] {
		return -1
	}
	if fe[5] > fe2[5] {
		return 1
	} else if fe[5] < fe2[5] {
		return -1
	}
	if fe[4] > fe2[4] {
		return 1
	} else if fe[4] < fe2[4] {
		return -1
	}
	if fe[3] > fe2[3] {
		return 1
	} else if fe[3] < fe2[3] {
		return -1
	}
	if fe[2] > fe2[2] {
		return 1
	} else if fe[2] < fe2[2] {
		return -1
	}
	if fe[1] > fe2[1] {
		return 1
	} else if fe[1] < fe2[1] {
		return -1
	}
	if fe[0] > fe2[0] {
		return 1
	} else if fe[0] < fe2[0] {
		return -1
	}
	return 0
}

func (fe *Fe512) Equals(fe2 *Fe512) bool {
	return fe2[0] == fe[0] && fe2[1] == fe[1] && fe2[2] == fe[2] && fe2[3] == fe[3] && fe2[4] == fe[4] && fe2[5] == fe[5] && fe2[6] == fe[6] && fe2[7] == fe[7]
}

func (fe *Fe512) div2(e uint64) {
	fe[0] = fe[0]>>1 | fe[1]<<63
	fe[1] = fe[1]>>1 | fe[2]<<63
	fe[2] = fe[2]>>1 | fe[3]<<63
	fe[3] = fe[3]>>1 | fe[4]<<63
	fe[4] = fe[4]>>1 | fe[5]<<63
	fe[5] = fe[5]>>1 | fe[6]<<63
	fe[6] = fe[6]>>1 | fe[7]<<63
	fe[7] = fe[7]>>1 | e<<63
}

func (fe *Fe512) mul2() uint64 {
	e := fe[7] >> 63
	fe[7] = fe[7]<<1 | fe[6]>>63
	fe[6] = fe[6]<<1 | fe[5]>>63
	fe[5] = fe[5]<<1 | fe[4]>>63
	fe[4] = fe[4]<<1 | fe[3]>>63
	fe[3] = fe[3]<<1 | fe[2]>>63
	fe[2] = fe[2]<<1 | fe[1]>>63
	fe[1] = fe[1]<<1 | fe[0]>>63
	fe[0] = fe[0] << 1
	return e
}

func (fe *Fe512) bit(i int) bool {
	k := i >> 6
	i = i - k<<6
	b := (fe[k] >> uint(i)) & 1
	return b != 0
}

func (fe *Fe512) bitLen() int {
	for i := len(fe) - 1; i >= 0; i-- {
		if len := bits.Len64(fe[i]); len != 0 {
			return len + 64*i
		}
	}
	return 0
}

func (f *Fe512) rand(max *Fe512, r io.Reader) error {
	bitLen := bits.Len64(max[7]) + (8-1)*64
	k := (bitLen + 7) / 8
	b := uint(bitLen % 8)
	if b == 0 {
		b = 8
	}
	bytes := make([]byte, k)
	for {
		_, err := io.ReadFull(r, bytes)
		if err != nil {
			return err
		}
		bytes[0] &= uint8(int(1<<b) - 1)
		f.FromBytes(bytes)
		if f.Cmp(max) < 0 {
			break
		}
	}
	return nil
}
