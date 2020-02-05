package fp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"unsafe"

	"golang.org/x/sys/cpu"
)

// fieldElement is a pointer that addresses
// any field element in any limb size
type fieldElement = unsafe.Pointer

var nonADXBMI2 = !(cpu.X86.HasADX && cpu.X86.HasBMI2) || forceNonADXBMI2()

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
	copy     func(dst, stc fieldElement)
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

func newField(p []byte) (*field, error) {
	var err error
	f := new(field)
	f.pbig = new(big.Int).SetBytes(p)
	f.p, f.limbSize, err = newFieldElementFromBytes(p)
	if err != nil {
		return nil, err
	}
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
		return nil, fmt.Errorf("field is not applicable\n%s", hex.EncodeToString(p))
	}
	f.inp = inpT.Uint64()
	switch f.limbSize {
	case 1:
		f.equal = eq1
		f.copy = cpy1
		f.cmp = cmp1
		f.addn = addn1
		f.subn = subn1
		f._add = add1
		f._sub = sub1
		f._double = double1
		f._neg = _neg1
		f.div_two = div_two_1
		f.mul_two = mul_two_1
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_1
		} else {
			f._mul = mul1
		}
	case 2:
		f.equal = eq2
		f.copy = cpy2
		f.cmp = cmp2
		f.addn = addn2
		f.subn = subn2
		f._add = add2
		f._sub = sub2
		f._double = double2
		f._neg = _neg2
		f.div_two = div_two_2
		f.mul_two = mul_two_2
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_2
		} else {
			f._mul = mul2
		}
	case 3:
		f.equal = eq3
		f.copy = cpy3
		f.cmp = cmp3
		f.addn = addn3
		f.subn = subn3
		f._add = add3
		f._sub = sub3
		f._double = double3
		f._neg = _neg3
		f.div_two = div_two_3
		f.mul_two = mul_two_3
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_3
		} else {
			f._mul = mul3
		}
	case 4:
		f.equal = eq4
		f.copy = cpy4
		f.cmp = cmp4
		f.addn = addn4
		f.subn = subn4
		f._add = add4
		f._sub = sub4
		f._double = double4
		f._neg = _neg4
		f.div_two = div_two_4
		f.mul_two = mul_two_4
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_4
		} else {
			f._mul = mul4
		}
	case 5:
		f.equal = eq5
		f.copy = cpy5
		f.cmp = cmp5
		f.addn = addn5
		f.subn = subn5
		f._add = add5
		f._sub = sub5
		f._double = double5
		f._neg = _neg5
		f.div_two = div_two_5
		f.mul_two = mul_two_5
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_5
		} else {
			f._mul = mul5
		}
	case 6:
		f.equal = eq6
		f.copy = cpy6
		f.cmp = cmp6
		f.addn = addn6
		f.subn = subn6
		f._add = add6
		f._sub = sub6
		f._double = double6
		f._neg = _neg6
		f.div_two = div_two_6
		f.mul_two = mul_two_6
		f._mul = mul6
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_6
		} else {
			f._mul = mul6
		}
	case 7:
		f.equal = eq7
		f.copy = cpy7
		f.cmp = cmp7
		f.addn = addn7
		f.subn = subn7
		f._add = add7
		f._sub = sub7
		f._double = double7
		f._neg = _neg7
		f.div_two = div_two_7
		f.mul_two = mul_two_7
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_7
		} else {
			f._mul = mul7
		}
	case 8:
		f.equal = eq8
		f.copy = cpy8
		f.cmp = cmp8
		f.addn = addn8
		f.subn = subn8
		f._add = add8
		f._sub = sub8
		f._double = double8
		f._neg = _neg8
		f.div_two = div_two_8
		f.mul_two = mul_two_8
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_8
		} else {
			f._mul = mul8
		}
	case 9:
		f.equal = eq9
		f.copy = cpy9
		f.cmp = cmp9
		f.addn = addn9
		f.subn = subn9
		f._add = add9
		f._sub = sub9
		f._double = double9
		f._neg = _neg9
		f.div_two = div_two_9
		f.mul_two = mul_two_9
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_9
		} else {
			f._mul = mul9
		}
	case 10:
		f.equal = eq10
		f.copy = cpy10
		f.cmp = cmp10
		f.addn = addn10
		f.subn = subn10
		f._add = add10
		f._sub = sub10
		f._double = double10
		f._neg = _neg10
		f.div_two = div_two_10
		f.mul_two = mul_two_10
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_10
		} else {
			f._mul = mul10
		}
	case 11:
		f.equal = eq11
		f.copy = cpy11
		f.cmp = cmp11
		f.addn = addn11
		f.subn = subn11
		f._add = add11
		f._sub = sub11
		f._double = double11
		f._neg = _neg11
		f.div_two = div_two_11
		f.mul_two = mul_two_11
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_11
		} else {
			f._mul = mul11
		}
	case 12:
		f.equal = eq12
		f.copy = cpy12
		f.cmp = cmp12
		f.addn = addn12
		f.subn = subn12
		f._add = add12
		f._sub = sub12
		f._double = double12
		f._neg = _neg12
		f.div_two = div_two_12
		f.mul_two = mul_two_12
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_12
		} else {
			f._mul = mul12
		}
	case 13:
		f.equal = eq13
		f.copy = cpy13
		f.cmp = cmp13
		f.addn = addn13
		f.subn = subn13
		f._add = add13
		f._sub = sub13
		f._double = double13
		f._neg = _neg13
		f.div_two = div_two_13
		f.mul_two = mul_two_13
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_13
		} else {
			f._mul = mul13
		}
	case 14:
		f.equal = eq14
		f.copy = cpy14
		f.cmp = cmp14
		f.addn = addn14
		f.subn = subn14
		f._add = add14
		f._sub = sub14
		f._double = double14
		f._neg = _neg14
		f.div_two = div_two_14
		f.mul_two = mul_two_14
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_14
		} else {
			f._mul = mul14
		}
	case 15:
		f.equal = eq15
		f.copy = cpy15
		f.cmp = cmp15
		f.addn = addn15
		f.subn = subn15
		f._add = add15
		f._sub = sub15
		f._double = double15
		f._neg = _neg15
		f.div_two = div_two_15
		f.mul_two = mul_two_15
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_15
		} else {
			f._mul = mul15
		}
	case 16:
		f.equal = eq16
		f.copy = cpy16
		f.cmp = cmp16
		f.addn = addn16
		f.subn = subn16
		f._add = add16
		f._sub = sub16
		f._double = double16
		f._neg = _neg16
		f.div_two = div_two_16
		f.mul_two = mul_two_16
		if nonADXBMI2 {
			f._mul = mul_no_adx_bmi2_16
		} else {
			f._mul = mul16
		}
	default:
		return nil, fmt.Errorf("limb size %d is not implemented", f.limbSize)
	}
	return f, nil
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
		f.copy(a, f.zero)
		return
	}
	f._neg(c, a, f.p)
}

func (f *field) mul(c, a, b fieldElement) {
	f._mul(c, a, b, f.p, f.inp)
}

func (f *field) square(c, a fieldElement) {
	f._mul(c, a, a, f.p, f.inp)
}

func (f *field) exp(c, a fieldElement, e *big.Int) {
	z := f.newFieldElement()
	f.copy(z, f.r)
	for i := e.BitLen(); i >= 0; i-- {
		f.mul(z, z, z)
		if e.Bit(i) == 1 {
			f.mul(z, z, a)
		}
	}
	f.copy(c, z)
}

func (f *field) isOne(fe fieldElement) bool {
	return f.equal(fe, f.one)
}

func (f *field) isZero(fe fieldElement) bool {
	return f.equal(fe, f.zero)
}

func (f *field) isValid(fe []byte) bool {
	feBig := new(big.Int).SetBytes(fe)
	if feBig.Cmp(f.pbig) != -1 {
		return false
	}
	return true
}

func (f *field) newFieldElement() fieldElement {
	fe, err := newFieldElement(f.limbSize)
	if err != nil {
		// panic("this is unexpected")
	}
	return fe
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
	fe, limbSize, err := newFieldElementFromBytes(in)
	if err != nil {
		return nil, err
	}
	if limbSize != f.limbSize {
		// panic("this is unexpected")
	}
	return fe, nil
}

func (f *field) newFieldElementFromBytes(in []byte) (fieldElement, error) {
	if len(in) != f.byteSize() {
		return nil, fmt.Errorf("bad input size")
	}
	if !f.isValid(in) {
		return nil, fmt.Errorf("input is a larger number than modulus")
	}
	fe, limbSize, err := newFieldElementFromBytes(in)
	if err != nil {
		return nil, err
	}
	if limbSize != f.limbSize {
		// panic("this is unexpected")
	}
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
	fe, limbSize, err := newFieldElementFromBytes(padBytes(in, f.byteSize()))
	if err != nil {
		return nil, err
	}
	if limbSize != f.limbSize {
		// panic("this is unexpected")
	}
	f.toMont(fe, fe)
	return fe, nil
}

func (f *field) newFieldElementFromBig(a *big.Int) (fieldElement, error) {
	in := a.Bytes()
	if !f.isValid(in) {
		return nil, fmt.Errorf("input is a larger number than modulus")
	}
	if len(in) > f.byteSize() {
		return nil, fmt.Errorf("bad input size %d", len(in))
	}
	fe, limbSize, err := newFieldElementFromBytes(padBytes(in, f.byteSize()))
	if err != nil {
		return nil, err
	}
	if limbSize != f.limbSize {
		// panic("this is unexpected")
	}
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
	case 1:
		return toBytes((*[1]uint64)(in)[:])
	case 2:
		return toBytes((*[2]uint64)(in)[:])
	case 3:
		return toBytes((*[3]uint64)(in)[:])
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
	case 9:
		return toBytes((*[9]uint64)(in)[:])
	case 10:
		return toBytes((*[10]uint64)(in)[:])
	case 11:
		return toBytes((*[11]uint64)(in)[:])
	case 12:
		return toBytes((*[12]uint64)(in)[:])
	case 13:
		return toBytes((*[13]uint64)(in)[:])
	case 14:
		return toBytes((*[14]uint64)(in)[:])
	case 15:
		return toBytes((*[15]uint64)(in)[:])
	case 16:
		return toBytes((*[16]uint64)(in)[:])
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
func newFieldElementFromBytes(in []byte) (fieldElement, int, error) {
	byteSize := len(in)
	limbSize := byteSize / 8
	if byteSize%8 != 0 {
		return nil, 0, fmt.Errorf("bad input byte size %d", byteSize)
	}
	a, err := newFieldElement(limbSize)
	if err != nil {
		return nil, 0, err
	}
	var data []uint64
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	sh.Data = uintptr(a)
	sh.Len, sh.Cap = limbSize, limbSize
	if err := limbSliceFromBytes(data[:], in); err != nil {
		// panic("this is unexpected")
	}
	return a, limbSize, nil
}

func newFieldElement(limbSize int) (fieldElement, error) {
	switch limbSize {
	case 1:
		return unsafe.Pointer(&[1]uint64{}), nil
	case 2:
		return unsafe.Pointer(&[2]uint64{}), nil
	case 3:
		return unsafe.Pointer(&[3]uint64{}), nil
	case 4:
		return unsafe.Pointer(&[4]uint64{}), nil
	case 5:
		return unsafe.Pointer(&[5]uint64{}), nil
	case 6:
		return unsafe.Pointer(&[6]uint64{}), nil
	case 7:
		return unsafe.Pointer(&[7]uint64{}), nil
	case 8:
		return unsafe.Pointer(&[8]uint64{}), nil
	case 9:
		return unsafe.Pointer(&[9]uint64{}), nil
	case 10:
		return unsafe.Pointer(&[10]uint64{}), nil
	case 11:
		return unsafe.Pointer(&[11]uint64{}), nil
	case 12:
		return unsafe.Pointer(&[12]uint64{}), nil
	case 13:
		return unsafe.Pointer(&[13]uint64{}), nil
	case 14:
		return unsafe.Pointer(&[14]uint64{}), nil
	case 15:
		return unsafe.Pointer(&[15]uint64{}), nil
	case 16:
		return unsafe.Pointer(&[16]uint64{}), nil
	default:
		return nil, fmt.Errorf("limb size %d is not implemented", limbSize)
	}
}

func newFieldElementFromBigUnchecked(limbSize int, bi *big.Int) fieldElement {
	in := bi.Bytes()
	byteSize := limbSize * 8
	fe, _, _ := newFieldElementFromBytes(padBytes(in, byteSize))
	return fe
}

func limbSliceFromBytes(out []uint64, in []byte) error {
	var byteSize = len(in)
	var limbSize = len(out)
	if limbSize*8 != byteSize {
		return fmt.Errorf("(byteSize != limbSize * 8), %d, %d", byteSize, limbSize)
	}
	var a int
	for i := 0; i < limbSize; i++ {
		a = byteSize - i*8
		out[i] = uint64(in[a-1]) | uint64(in[a-2])<<8 |
			uint64(in[a-3])<<16 | uint64(in[a-4])<<24 |
			uint64(in[a-5])<<32 | uint64(in[a-6])<<40 |
			uint64(in[a-7])<<48 | uint64(in[a-8])<<56
	}
	return nil
}

func padBytes(in []byte, size int) []byte {
	out := make([]byte, size)
	if len(in) > size {
		panic("bad input for padding")
	}
	copy(out[size-len(in):], in)
	return out
}

func (f *field) inverse(inv, e fieldElement) bool {
	u, v, s, r := f.newFieldElement(),
		f.newFieldElement(),
		f.newFieldElement(),
		f.newFieldElement()
	zero := f.newFieldElement()
	f.copy(u, f.p)
	f.copy(v, e)
	f.copy(s, f._one)
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
		f.copy(inv, zero)
		return false
	}
	if k < bitSize {
		/*
			this is unexpected
		*/
		f.copy(inv, zero)
		return false
	}

	if f.cmp(r, f.p) != -1 {
		f.subn(r, f.p)
	}
	f.copy(u, f.p)
	f.subn(u, r)

	// Phase 2
	for i := k; i < bitSize*2; i++ {
		f.double(u, u)
	}
	f.copy(inv, u)
	return true
}
