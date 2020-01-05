package gocode

import (
	"fmt"
	"math/big"
)

func padBytes(in []byte, size int) []byte {
	out := make([]byte, size)
	copy(out[size-len(in):], in)
	return out
}

func encodeBig(name string, limbSize int, b *big.Int, pointer bool) string {
	encoded := "var "
	if pointer {
		encoded += fmt.Sprintf("%s = &fieldElement{\n", name)
	} else {
		encoded += fmt.Sprintf("%s = fieldElement{\n", name)
	}
	byteSize := limbSize * 8
	bts := padBytes(b.Bytes(), byteSize)
	for i := 0; i < limbSize; i++ {
		encoded += fmt.Sprintf("0x%16.16x,\n", bts[byteSize-(i+1)*8:byteSize-i*8])
	}
	return encoded + "}\n\n"
}

func fieldImpl(limbSize int, modulus *big.Int) string {

	if modulus == nil {
		return fieldImplNonFixedModulus
	} else {
		byteSize := limbSize * 8
		R := new(big.Int)
		R.SetBit(R, byteSize*8, 1).Mod(R, modulus)
		R2 := new(big.Int)
		R2.Mul(R, R).Mod(R2, modulus)
		inpT := new(big.Int).ModInverse(new(big.Int).Neg(modulus), new(big.Int).SetBit(new(big.Int), 64, 1))
		if inpT == nil {
			panic("cannot inverse modulus")
		}
		// imports
		code := fieldImplFixedModulus0
		// inp
		code += fmt.Sprintf("var inp uint64 = %d\n\n", inpT.Uint64())
		// modulus
		code += encodeBig("modulus", limbSize, modulus, false)
		// zero
		code += encodeBig("zero", limbSize, big.NewInt(0), true)
		// one & r
		code += encodeBig("r", limbSize, R, true)
		//
		code += encodeBig("one", limbSize, R, true)
		// r^2
		code += encodeBig("r2", limbSize, R2, true)
		// actual one
		code += encodeBig("_one", limbSize, big.NewInt(1), true)
		// pbig
		code += fmt.Sprintf("var pbig, _ = new(big.Int).SetString(\"%s\", 10)\n\n", modulus.String())
		// rbig
		code += fmt.Sprintf("var rbig, _ = new(big.Int).SetString(\"%s\", 10)\n\n", R.String())
		// impl
		code += fieldImplFixedModulus1
		return code
	}
}

const fieldImplNonFixedModulus = `

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
)

type field struct {
	p    *fieldElement
	one  *fieldElement
	_one *fieldElement
	zero *fieldElement
	r    *fieldElement
	r2   *fieldElement
	pbig *big.Int
	rbig *big.Int
	inp  uint64
}

func newField(p []byte) (*field, error) {
	f := new(field)
	f.pbig = new(big.Int).SetBytes(p)
	var err error
	f.p, err = new(fieldElement).fromBytes(p)
	if err != nil {
		return nil, err
	}
	R := new(big.Int)
	R.SetBit(R, byteSize*8, 1).Mod(R, f.pbig)
	R2 := new(big.Int)
	R2.Mul(R, R).Mod(R2, f.pbig)
	inpT := new(big.Int).ModInverse(new(big.Int).Neg(f.pbig), new(big.Int).SetBit(new(big.Int), 64, 1))
	if inpT == nil {
		return nil, fmt.Errorf("modulus is not inversive")
	}
	f.inp = inpT.Uint64()
	f.rbig = new(big.Int).Set(R)
	f.r, err = new(fieldElement).fromBytes(padBytes(R.Bytes(), byteSize))
	if err != nil {
		return nil, err
	}
	f.one = f.r
	f.r2, err = new(fieldElement).fromBytes(padBytes(R2.Bytes(), byteSize))
	if err != nil {
		return nil, err
	}
	f._one, err = new(fieldElement).fromBytes(padBytes([]byte{0, 0, 0, 1}, byteSize))
	if err != nil {
		return nil, err
	}
	f.zero, err = new(fieldElement).fromBytes(padBytes([]byte{0, 0, 0, 0}, byteSize))
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (f *field) newFieldElement() *fieldElement {
	return &fieldElement{}
}

func (f *field) randFieldElement(r io.Reader) (*fieldElement, error) {
	bi, err := rand.Int(r, f.pbig)
	if err != nil {
		return nil, err
	}
	return f.newFieldElementFromBig(bi)
}

func (f *field) newFieldElementFromBytes(in []byte) (*fieldElement, error) {
	fe, err := new(fieldElement).fromBytes(in)
	if err != nil {
		return nil, err
	}
	if !f.isValid(fe) {
		return nil, fmt.Errorf("input is not valid")
	}
	f.toMont(fe, fe)
	return fe, nil
}

func (f *field) newFieldElementFromBig(in *big.Int) (*fieldElement, error) {
	fe, err := new(fieldElement).fromBig(in)
	if err != nil {
		return nil, err
	}
	f.toMont(fe, fe)
	return fe, nil
}

func (f *field) newFieldElementFromString(in string) (*fieldElement, error) {
	fe, err := new(fieldElement).fromString(in)
	if err != nil {
		return nil, err
	}
	f.toMont(fe, fe)
	return fe, nil
}

func (f *field) toBytes(fe *fieldElement) []byte {
	t := new(fieldElement)
	f.fromMont(t, fe)
	return t.toBytes()
}

func (f *field) toBig(fe *fieldElement) *big.Int {
	t := new(fieldElement)
	f.fromMont(t, fe)
	return t.toBig()
}

func (f *field) toString(fe *fieldElement) string {
	return hex.EncodeToString(f.toBytes(fe))
}

func (f *field) isValid(fe *fieldElement) bool {
	if fe.cmp(f.p) != -1 {
		return false
	}
	return true
}

func (f *field) isZero(fe *fieldElement) bool {
	return fe.equal(f.zero)
}

func (f *field) isOne(fe *fieldElement) bool {
	return fe.equal(f.one)
}

func (f *field) toMont(c, a *fieldElement) {
	mul(c, a, f.r2, f.p, f.inp)
}

func (f *field) fromMont(c, a *fieldElement) {
	mul(c, a, f._one, f.p, f.inp)
}

func (f *field) add(c, a, b *fieldElement) {
	add(c, a, b, f.p)
}

func (f *field) double(c, a *fieldElement) {
	double(c, a, f.p)
}

func (f *field) sub(c, a, b *fieldElement) {
	sub(c, a, b, f.p)
}

func (f *field) neg(c, a *fieldElement) {
	if a.equal(f.zero) {
		a.set(f.zero)
		return
	}
	_neg(c, a, f.p)
}

func (f *field) mul(c, a, b *fieldElement) {
	mul(c, a, b, f.p, f.inp)
}

func (f *field) exp(c, a *fieldElement, e *big.Int) {
	z := f.newFieldElement()
	z.set(f.r)
	for i := e.BitLen(); i >= 0; i-- {
		f.mul(z, z, z)
		if e.Bit(i) == 1 {
			f.mul(z, z, a)
		}
	}
	c.set(z)
}

func padBytes(in []byte, size int) []byte {
	out := make([]byte, size)
	if len(in) > size {
		panic("bad input for padding")
	}
	copy(out[size-len(in):], in)
	return out
}
`

const fieldImplFixedModulus0 = `
import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"fmt"
	"math/big"
)
`
const fieldImplFixedModulus1 = `

func randFieldElement(r io.Reader) (*fieldElement, error) {
	bi, err := rand.Int(r, pbig)
	if err != nil {
		return nil, err
	}
	return newFieldElementFromBig(bi)
}

func newFieldElement() *fieldElement {
	return &fieldElement{}
}

func newFieldElementFromBytes(in []byte) (*fieldElement, error) {
	fe, err := new(fieldElement).fromBytes(in)
	if err != nil {
		return nil, err
	}
	if !isValid(fe) {
		return nil, fmt.Errorf("input is not valid")
	}
	toMont(fe, fe)
	return fe, nil
}

func newFieldElementFromBig(in *big.Int) (*fieldElement, error) {
	fe, err := new(fieldElement).fromBig(in)
	if err != nil {
		return nil, err
	}
	toMont(fe, fe)
	return fe, nil
}

func newFieldElementFromString(in string) (*fieldElement, error) {
	fe, err := new(fieldElement).fromString(in)
	if err != nil {
		return nil, err
	}
	toMont(fe, fe)
	return fe, nil
}

func toBytes(fe *fieldElement) []byte {
	t := new(fieldElement)
	fromMont(t, fe)
	return t.toBytes()
}

func toBig(fe *fieldElement) *big.Int {
	t := new(fieldElement)
	fromMont(t, fe)
	return t.toBig()
}

func toString(fe *fieldElement) string {
	return hex.EncodeToString(toBytes(fe))
}

func isValid(fe *fieldElement) bool {
	if fe.cmp(&modulus) != -1 {
		return false
	}
	return true
}

func isZero(fe *fieldElement) bool {
	return fe.equal(zero)
}

func isOne(fe *fieldElement) bool {
	return fe.equal(one)
}

func toMont(c, a *fieldElement) {
	mul(c, a, r2)
}

func fromMont(c, a *fieldElement) {
	mul(c, a, _one)
}

func neg(c, a *fieldElement) {
	if a.equal(zero) {
		a.set(zero)
		return
	}
	_neg(c, a)
}

func exp(c, a *fieldElement, e *big.Int) {
	z := newFieldElement()
	z.set(r)
	for i := e.BitLen(); i >= 0; i-- {
		mul(z, z, z)
		if e.Bit(i) == 1 {
			mul(z, z, a)
		}
	}
	c.set(z)
}

func padBytes(in []byte, size int) []byte {
	out := make([]byte, size)
	if len(in) > size {
		panic("bad input for padding")
	}
	copy(out[size-len(in):], in)
	return out
}
`
