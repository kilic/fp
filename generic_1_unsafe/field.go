package fp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"unsafe"
)

// fieldElement is a pointer that addresses
// any field element in any limb size
type fieldElement = unsafe.Pointer

type field struct {
	limbSize int
	p        fieldElement
	inp      uint64
	one      fieldElement
	_one     fieldElement
	zero     fieldElement
	r        fieldElement
	r2       fieldElement
	pbig     *big.Int
	rbig     *big.Int
	equal    func(a, b fieldElement) bool
	cmp      func(a, b fieldElement) int8
	cpy      func(dst, stc fieldElement)
	_mul     func(c, a, b, p fieldElement, inp uint64)
	_add     func(c, a, b, p fieldElement)
	_double  func(c, a, p fieldElement)
	_sub     func(c, a, b, p fieldElement)
	_neg     func(c, a, p fieldElement)
	addn     func(a, b fieldElement) uint64
	subn     func(a, b fieldElement) uint64
	div_two  func(a fieldElement)
	mul_two  func(a fieldElement)
}

func newField(p []byte) *field {
	f := new(field)
	f.pbig = new(big.Int).SetBytes(p)
	f.p, f.limbSize = newFieldElementFromBytes(p)
	R := new(big.Int)
	R.SetBit(R, f.byteSize()*8, 1).Mod(R, f.pbig)
	R2 := new(big.Int)
	R2.Mul(R, R).Mod(R2, f.pbig)
	inpT := new(big.Int).ModInverse(new(big.Int).Neg(f.pbig), new(big.Int).SetBit(new(big.Int), 64, 1))
	f.r = newFieldElementFromBigUnchecked(f.limbSize, R)
	f.rbig = R
	f.one = newFieldElementFromBigUnchecked(f.limbSize, R)
	f.r2 = newFieldElementFromBigUnchecked(f.limbSize, R2)
	f._one = newFieldElementFromBigUnchecked(f.limbSize, big.NewInt(1))
	f.zero = newFieldElementFromBigUnchecked(f.limbSize, new(big.Int))
	if inpT == nil {
		return nil
	}
	f.inp = inpT.Uint64()
	switch f.limbSize {
	case 4:
		f.equal = eq4
		f.cpy = cpy4
		f.cmp = cmp4
		f.addn = addn4
		f.subn = subn4
		f._mul = mul4
		f._add = add4
		f._sub = sub4
		f._double = double4
		f._neg = _neg4
		f.div_two = div_two_4
		f.mul_two = mul_two_4
	case 5:
		f.equal = eq5
		f.cpy = cpy5
		f.cmp = cmp5
		f.addn = addn5
		f.subn = subn5
		f._mul = mul5
		f._add = add5
		f._sub = sub5
		f._double = double5
		f._neg = _neg5
		f.div_two = div_two_5
		f.mul_two = mul_two_5
	case 6:
		f.equal = eq6
		f.cpy = cpy6
		f.cmp = cmp6
		f.addn = addn6
		f.subn = subn6
		f._mul = mul6
		f._add = add6
		f._sub = sub6
		f._double = double6
		f._neg = _neg6
		f.div_two = div_two_6
		f.mul_two = mul_two_6
	case 7:
		f.equal = eq7
		f.cpy = cpy7
		f.cmp = cmp7
		f.addn = addn7
		f.subn = subn7
		f._mul = mul7
		f._add = add7
		f._sub = sub7
		f._double = double7
		f._neg = _neg7
		f.div_two = div_two_7
		f.mul_two = mul_two_7
	case 8:
		f.equal = eq8
		f.cpy = cpy8
		f.cmp = cmp8
		f.addn = addn8
		f.subn = subn8
		f._mul = mul8
		f._add = add8
		f._sub = sub8
		f._double = double8
		f._neg = _neg8
		f.div_two = div_two_8
		f.mul_two = mul_two_8
	default:
		panic("not implemented")
	}
	return f
}

func (f *field) toMont(c, a fieldElement) {
	f._mul(c, a, f.r2, f.p, f.inp)
}

func (f *field) fromMont(c, a fieldElement) {
	f._mul(c, a, f._one, f.p, f.inp)
}

func (f *field) add(c, a, b fieldElement) {
	f._add(c, a, b, f.p)
}

func (f *field) double(c, a fieldElement) {
	f._double(c, a, f.p)
}

func (f *field) sub(c, a, b fieldElement) {
	f._sub(c, a, b, f.p)
}

func (f *field) neg(c, a fieldElement) {
	if f.equal(a, f.zero) {
		f.cpy(a, f.zero)
		return
	}
	f._neg(c, a, f.p)
}

func (f *field) mul(c, a, b fieldElement) {
	f._mul(c, a, b, f.p, f.inp)
}

func (f *field) exp(c, a fieldElement, e *big.Int) {
	z := f.newFieldElement()
	f.cpy(z, f.r)
	for i := e.BitLen(); i >= 0; i-- {
		f.mul(z, z, z)
		if e.Bit(i) == 1 {
			f.mul(z, z, a)
		}
	}
	f.cpy(c, z)
}

func (f *field) isValid(fe []byte) bool {
	feBig := new(big.Int).SetBytes(fe)
	if feBig.Cmp(f.pbig) != -1 {
		return false
	}
	return true
}

func (f *field) newFieldElement() fieldElement {
	return newFieldElement(f.limbSize)
}

func (f *field) randFieldElement(r io.Reader) fieldElement {
	bi, err := rand.Int(r, f.pbig)
	if err != nil {
		panic(err)
	}
	return newFieldElementFromBigUnchecked(f.limbSize, bi)
}

func (f *field) newFieldElementFromBytesNoTransform(in []byte) (fieldElement, error) {
	if len(in) != f.byteSize() {
		return nil, fmt.Errorf("bad input size")
	}
	fe, _ := newFieldElementFromBytes(in)
	return fe, nil
}

func (f *field) newFieldElementFromBytes(in []byte) (fieldElement, error) {
	if len(in) != f.byteSize() {
		return nil, fmt.Errorf("bad input size")
	}
	if !f.isValid(in) {
		return nil, fmt.Errorf("input is a larger number than modulus")
	}
	fe, _ := newFieldElementFromBytes(in)
	// if limbSize != _limbSize { // panic("") // is not expected // }
	f.toMont(fe, fe)
	return fe, nil
}

func (f *field) newFieldElementFromString(hexStr string) (fieldElement, error) {
	str := hexStr
	if len(str) > 1 && str[:2] == "0x" {
		str = hexStr[:2]
	}
	in, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	if !f.isValid(in) {
		return nil, fmt.Errorf("input is a larger number than modulus")
	}
	if len(in) > f.byteSize() {
		return nil, fmt.Errorf("bad input size")
	}
	fe, _ := newFieldElementFromBytes(padBytes(in, f.byteSize()))
	f.toMont(fe, fe)
	return fe, nil
}

func (f *field) newFieldElementFromBig(a *big.Int) (fieldElement, error) {
	in := a.Bytes()
	if !f.isValid(in) {
		return nil, fmt.Errorf("input is a larger number than modulus")
	}
	if len(in) > f.byteSize() {
		return nil, fmt.Errorf("bad input size")
	}
	fe, _ := newFieldElementFromBytes(padBytes(in, f.byteSize()))
	f.toMont(fe, fe)
	return fe, nil
}

func (f *field) toBytes(in fieldElement) []byte {
	t := f.newFieldElement()
	f.fromMont(t, in)
	return f.toBytesNoTransform(t)
}

func (f *field) toBytesNoTransform(in fieldElement) []byte {
	switch f.limbSize {
	case 4:
		return toBytes((*[4]uint64)(in)[:])
	case 5:
		return toBytes((*[5]uint64)(in)[:])
	case 6:
		return toBytes((*[6]uint64)(in)[:])
	case 7:
		return toBytes((*[7]uint64)(in)[:])
	case 8:
		return toBytes((*[8]uint64)(in)[:])
	default:
		panic("not implemented")
	}
}

func (f *field) toBig(in fieldElement) *big.Int {
	return new(big.Int).SetBytes(f.toBytes(in))
}

func (f *field) toBigNoTransform(in fieldElement) *big.Int {
	return new(big.Int).SetBytes(f.toBytesNoTransform(in))
}

func (f *field) toString(in fieldElement) string {
	return hex.EncodeToString(f.toBytes(in))
}

func (f *field) toStringNoTransform(in fieldElement) string {
	return hex.EncodeToString(f.toBytesNoTransform(in))
}

func (f *field) byteSize() int {
	return f.limbSize * 8
}

func toBytes(fe []uint64) []byte {
	size := len(fe)
	byteSize := size * 8
	out := make([]byte, byteSize)
	var a int
	for i := 0; i < size; i++ {
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

// newFieldElement returns pointer of an uint64 array.
// limbSize is calculated according to size of input slice
func newFieldElementFromBytes(in []byte) (fieldElement, int) {
	byteSize := len(in)
	limbSize := byteSize / 8
	if byteSize%8 != 0 {
		panic("bad input byte size")
	}
	a := newFieldElement(limbSize)
	var data []uint64
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	sh.Data = uintptr(a)
	sh.Len, sh.Cap = limbSize, limbSize
	limbSliceFromBytes(data[:], in)
	return a, limbSize
}

func newFieldElement(limbSize int) fieldElement {
	switch limbSize {
	case 4:
		return unsafe.Pointer(&[4]uint64{})
	case 5:
		return unsafe.Pointer(&[5]uint64{})
	case 6:
		return unsafe.Pointer(&[6]uint64{})
	case 7:
		return unsafe.Pointer(&[7]uint64{})
	case 8:
		return unsafe.Pointer(&[8]uint64{})
	default:
		panic("not implemented")
	}
}

func newFieldElementFromBigUnchecked(limbSize int, bi *big.Int) fieldElement {
	in := bi.Bytes()
	byteSize := limbSize * 8
	fe, _ := newFieldElementFromBytes(padBytes(in, byteSize))
	return fe
}

func limbSliceFromBytes(out []uint64, in []byte) {
	var byteSize = len(in)
	var limbSize = len(out)
	if limbSize*8 != byteSize {
		panic("non ... input output sizes")
	}
	var a int
	for i := 0; i < limbSize; i++ {
		a = byteSize - i*8
		out[i] = uint64(in[a-1]) | uint64(in[a-2])<<8 |
			uint64(in[a-3])<<16 | uint64(in[a-4])<<24 |
			uint64(in[a-5])<<32 | uint64(in[a-6])<<40 |
			uint64(in[a-7])<<48 | uint64(in[a-8])<<56
	}
}

func padBytes(in []byte, size int) []byte {
	out := make([]byte, size)
	if len(in) > size {
		panic("bad input for padding")
	}
	copy(out[size-len(in):], in)
	return out
}

func (f *field) inverse(inv, e fieldElement) {
	u, v, s, r := f.newFieldElement(),
		f.newFieldElement(),
		f.newFieldElement(),
		f.newFieldElement()
	zero := f.newFieldElement()
	f.cpy(u, f.p)
	f.cpy(v, e)
	f.cpy(s, f._one)
	var k int
	var found = false
	byteSize := f.byteSize()
	bitSize := byteSize * 8
	// Phase 1
	for i := 0; i < bitSize*2; i++ {
		if f.equal(v, zero) {
			found = true
			break
		}
		if is_even(u) {
			f.div_two(u)
			f.mul_two(s)
		} else if is_even(v) {
			f.div_two(v)
			f.mul_two(r)
		} else if f.cmp(u, v) == 1 {
			f.subn(u, v)
			f.div_two(u)
			f.addn(r, s)
			f.mul_two(s)
		} else {
			f.subn(v, u)
			f.div_two(v)
			f.addn(s, r)
			f.mul_two(r)
		}
		k += 1
	}
	if !found {
		f.cpy(inv, zero)
		return
	}
	if k < bitSize {
		/*
			THIS IS UNEXPECTED
		*/
		f.cpy(inv, zero)
		return
	}

	if f.cmp(r, f.p) != -1 {
		f.subn(r, f.p)
	}
	f.cpy(u, f.p)
	f.subn(u, r)

	// Phase 2
	for i := k; i < bitSize*2; i++ {
		f.double(u, u)
	}
	f.cpy(inv, u)

}

// func fieldElementToBytes2(out []byte, in fieldElement) error {
// 	byteSize := len(out)
// 	limbSize := byteSize / 8
// 	if byteSize%8 != 0 && limbSize < 1 {
// 		return fmt.Errorf("bad output allocation")
// 	}
// 	var data []byte
// 	sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
// 	sh.Data = uintptr(in)
// 	sh.Len, sh.Cap = byteSize, byteSize
// 	for i := 0; i < limbSize; i++ {
// 		l := i * 8
// 		binary.BigEndian.PutUint64(
// 			out[byteSize-l-8:byteSize-l],
// 			binary.LittleEndian.Uint64(data[l:l+8]),
// 		)
// 	}
// 	return nil
// }
