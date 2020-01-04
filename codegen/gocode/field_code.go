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


func (f *field) toMont(c, a *fieldElement) {
	mul(c, a, f.r2, f.p, f.inp)
}

func (f *field) fromMont(c, a *fieldElement) {
	mul(c, a, f._one, f.p, f.inp)
}

func (f *field) randFieldElement(r io.Reader) (*fieldElement, error) {
	bi, err := rand.Int(r, f.pbig)
	if err != nil {
		panic(err)
	}
	return f.newFieldElementFromBig(bi)
}

func (f *field) equal(a1, a2 *fieldElement) bool {
	for i := 0; i < limbSize; i++ {
		if a1[i] != a2[i] {
			return false
		}
	}
	return true
}

func (f *field) toBytes(fe *fieldElement) []byte {
	t := new(fieldElement)
	f.fromMont(t, fe)
	return t.toBytes()
}

func (f *field) toBytesNoTransform(fe *fieldElement) []byte {
	return fe.toBytes()
}

func (f *field) toBig(fe *fieldElement) *big.Int {
	t := new(fieldElement)
	f.fromMont(t, fe)
	return new(big.Int).SetBytes(f.toBytes(t))
}

func (f *field) toBigNoTransform(fe *fieldElement) *big.Int {
	return new(big.Int).SetBytes(f.toBytes(fe))
}

func (f *field) toString(fe *fieldElement) string {
	return hex.EncodeToString(f.toBytes(fe))
}

func (f *field) newFieldElementFromBytes(in []byte) (*fieldElement, error) {
	fe, err := new(fieldElement).fromBytes(in)
	if err != nil {
		return nil, err
	}
	f.toMont(fe, fe)
	return fe, nil
}

func (f *field) newFieldElementFromBig(bi *big.Int) (*fieldElement, error) {
	bts := bi.Bytes()
	in := make([]byte, byteSize)
	copy(in[len(in)-len(bts):], bts[:])
	return f.newFieldElementFromBytes(in)
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
	if f.equal(a, f.zero) {
		a.Set(f.zero)
		return
	}
	_neg(c, a, f.p)
}

func (f *field) mul(c, a, b *fieldElement) {
	mul(c, a, b, f.p, f.inp)
}

func padBytes(in []byte, size int) []byte {
	out := make([]byte, size)
	copy(out[size-len(in):], in)
	return out
}
`

const fieldImplFixedModulus0 = `
import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"math/big"
)
`
const fieldImplFixedModulus1 = `

func newFieldElement() *fieldElement {
	return &fieldElement{}
}

func toMont(c, a *fieldElement) {
	mul(c, a, r2)
}

func fromMont(c, a *fieldElement) {
	mul(c, a, _one)
}

func equal(a1, a2 *fieldElement) bool {
	for i := 0; i < limbSize; i++ {
		if a1[i] != a2[i] {
			return false
		}
	}
	return true
}

func randFieldElement(r io.Reader) (*fieldElement, error) {
	bi, err := rand.Int(r, pbig)
	if err != nil {
		panic(err)
	}
	return newFieldElementFromBig(bi)
}

func toBytes(fe *fieldElement) []byte {
	t := new(fieldElement)
	fromMont(t, fe)
	return t.toBytes()
}

func toBytesNoTransform(fe *fieldElement) []byte {
	return fe.toBytes()
}

func toBig(fe *fieldElement) *big.Int {
	t := new(fieldElement)
	fromMont(t, fe)
	return new(big.Int).SetBytes(toBytes(t))
}

func toBigNoTransform(fe *fieldElement) *big.Int {
	return new(big.Int).SetBytes(toBytes(fe))
}

func toString(fe *fieldElement) string {
	return hex.EncodeToString(toBytes(fe))
}

func newFieldElementFromBytes(in []byte) (*fieldElement, error) {
	fe, err := new(fieldElement).fromBytes(in)
	if err != nil {
		return nil, err
	}
	toMont(fe, fe)
	return fe, nil
}

func newFieldElementFromBig(bi *big.Int) (*fieldElement, error) {
	bts := bi.Bytes()
	in := make([]byte, byteSize)
	copy(in[len(in)-len(bts):], bts[:])
	return newFieldElementFromBytes(in)
}


func neg(c, a *fieldElement) {
	if equal(a, zero) {
		a.Set(zero)
		return
	}
	_neg(c, a)
}
`
