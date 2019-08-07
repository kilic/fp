package main

import (
	"fmt"
	"io/ioutil"
)

func GenerateFieldElementTests(out string, from, to int) {

	codeStr := pkg("fp")
	codeStr = imports(codeStr, []string{"math/big", "testing", "crypto/rand"})
	codeStr += `
var n int
func TestMain(m *testing.M) {
	iter := flag.Int("iter", 1000, "# of iters")
	flag.Parse()
	n = *iter
	m.Run()
}
`
	// Rand Test Suite Generator
	data := struct {
		LimbSizes []int
	}{make([]int, 1+(to-from))}
	for i := 0; i < len(data.LimbSizes); i++ {
		data.LimbSizes[i] = from + i
	}
	funcs := map[string]interface{}{
		"mul": mul,
	}
	if strTestRandField, err := generate("", []string{tmplTestRandField}, funcs, data); err != nil {
		panic(err)
	} else {
		codeStr += strTestRandField
	}
	// Test Code
	codeStr += `
func TestField(t *testing.T) {`
	codeStr += `
	// example: subtest single run 
	// go test -run 'Element/256_Enc' -iter 1 -v
`
	codeStr += fmt.Sprintf("for i := %d; i <= %d; i++ {", from, to)
	codeStr += strTestEncoding
	codeStr += strTestAddition
	codeStr += strTestDouble
	codeStr += strTestSubtraction
	codeStr += strTestMongomerry
	codeStr += strTestExp
	codeStr += strTestInv
	codeStr += `
	}`
	codeStr += `
}`
	// Benches
	for i := from; i <= to; i++ {
		data := struct {
			Field        string
			FieldElement string
			Bit          int
		}{
			FieldElement: fmt.Sprintf("Fe%d", 64*i),
			Field:        fmt.Sprintf("Field%d", 64*i),
			Bit:          64 * i,
		}
		if strBenches, err := generate("", []string{tmplBenches}, utilFuncs, data); err != nil {
			panic(err)
		} else {
			codeStr += strBenches
		}
	}
	// Interfaces
	codeStr += strTestInterfaces
	// Interface implementations
	for i := from; i <= to; i++ {
		data := struct {
			Field        string
			FieldElement string
		}{
			FieldElement: fmt.Sprintf("Fe%d", 64*i),
			Field:        fmt.Sprintf("Field%d", 64*i),
		}
		if interfaceImpls, err := generate("", []string{tmplTestInterfaceImpls}, nil, data); err != nil {
			panic(err)
		} else {
			codeStr += interfaceImpls
		}
	}
	// // Remove template related break lines
	// strings.ReplaceAll(codeStr, "\n\n", "\n")

	if err := ioutil.WriteFile(out, []byte(codeStr), 0600); err != nil {
		panic(err)
	}
}

const (
	strTestInterfaces = `
type field interface {
add(c, a, b fieldElement)
double(c, a fieldElement)
sub(c, a, b fieldElement)
neg(c, a fieldElement)
square(c, a fieldElement)
mul(c, a, b fieldElement)
exp(c, a fieldElement, e *big.Int)
mont(c, a fieldElement)
demont(c, a fieldElement)
randElement(fe fieldElement, r io.Reader) fieldElement
newElement() fieldElement
p() fieldElement
one() fieldElement
two() fieldElement
invmu(inv, fe fieldElement)
invmd(inv, fe fieldElement)
inveea(inv, fe fieldElement)}
type fieldElement interface {
String() string
Bytes() []byte
Big() *big.Int
setBig(b *big.Int) fieldElement
setString(s string) (fieldElement, error)
fromBytes(in []byte) fieldElement
equals(other fieldElement) bool
limb(i int) uint64}
	`

	tmplTestRandField = `
func ceil64(len int) int {
size := 1 + ((len - 1) / 64)
if size < 5 {
return 4 }
return size }
{{ $FIELDS := .LimbSizes }} 
func randTestField(bitlen int) field {
var field field
for true {
p, err := rand.Prime(rand.Reader, bitlen)
if err != nil {
panic(err) }
switch ceil64(bitlen) { 
{{- range $N_LIMB := $FIELDS }} 
case {{ $N_LIMB }}: 
return NewField{{ mul $N_LIMB 64 }}(p.Bytes())
{{- end }}}
if field != nil {
break }}
return nil }
`

	tmplTestInterfaceImpls = `
{{ $FE := .FieldElement }} {{ $FIELD := .Field }}
func (fe *{{ $FE }}) setBig(b *big.Int) fieldElement {
return fe.SetBig(b)}
func (fe *{{ $FE }}) setString(s string) (fieldElement, error) {
return fe.SetString(s)}
func (fe *{{ $FE }}) fromBytes(in []byte) fieldElement {
return fe.FromBytes(in)}
func (fe *{{ $FE }}) equals(other fieldElement) bool {
for i := 0; i < len(fe); i++ {
if fe[i] != other.limb(i) {
return false}}
return true}
func (fe *{{ $FE }}) limb(i int) uint64 {
return fe[i]}
func (f *{{ $FIELD }}) add(c, a, b fieldElement) {
f.Add(c.(*{{ $FE }}), a.(*{{ $FE }}), b.(*{{ $FE }}))}
func (f *{{ $FIELD }}) double(c, a fieldElement) {
f.Double(c.(*{{ $FE }}), a.(*{{ $FE }}))}
func (f *{{ $FIELD }}) sub(c, a, b fieldElement) {
f.Sub(c.(*{{ $FE }}), a.(*{{ $FE }}), b.(*{{ $FE }}))}
func (f *{{ $FIELD }}) neg(c, a fieldElement) {
f.Neg(c.(*{{ $FE }}), a.(*{{ $FE }}))}
func (f *{{ $FIELD }}) square(c, a fieldElement) {
f.Square(c.(*{{ $FE }}), a.(*{{ $FE }}))}
func (f *{{ $FIELD }}) mul(c, a, b fieldElement) {
f.Mul(c.(*{{ $FE }}), a.(*{{ $FE }}), b.(*{{ $FE }}))}
func (f *{{ $FIELD }}) exp(c, a fieldElement, e *big.Int) {
f.Exp(c.(*{{ $FE }}), a.(*{{ $FE }}), e)}
func (f *{{ $FIELD }}) mont(c, a fieldElement) {
f.Mont(c.(*{{ $FE }}), a.(*{{ $FE }}))}
func (f *{{ $FIELD }}) demont(c, a fieldElement) {
f.Demont(c.(*{{ $FE }}), a.(*{{ $FE }}))}
func (f *{{ $FIELD }}) one() fieldElement {
return new({{ $FE }}).Set(f.r1)}
func (f *{{ $FIELD }}) two() fieldElement {
return new({{ $FE }}).Set(f.r2)}
func (f *{{ $FIELD }}) p() fieldElement {
return new({{ $FE }}).Set(f.P)}
func (f *{{ $FIELD }}) newElement() fieldElement {
return &{{ $FE }}{}}
func (f *{{ $FIELD }}) randElement(fe fieldElement, r io.Reader) fieldElement {
_, err := f.RandElement(fe.(*{{ $FE }}), r)
if err != nil {
panic(err) }
return fe }
func (f *{{ $FIELD }}) invmu(inv, fe fieldElement) {
f.InvMontUp(inv.(*{{ $FE }}), fe.(*{{ $FE }}))}
func (f *{{ $FIELD }}) invmd(inv, fe fieldElement) {
f.InvMontDown(inv.(*{{ $FE }}), fe.(*{{ $FE }}))}
func (f *{{ $FIELD }}) inveea(inv, fe fieldElement) {
f.InvEEA(inv.(*{{ $FE }}), fe.(*{{ $FE }}))}
`
	strTestEncoding = `
t.Run(fmt.Sprintf("%d Encoding & Decoding", i*64), func(t *testing.T) {
	field := randTestField(i * 64)
	zero := field.newElement().fromBytes([]byte{0})
	t.Run("1", func(t *testing.T) {
		bytes := []byte{0}
		fe := field.newElement()
		fe.fromBytes(bytes)
		if !fe.equals(zero) {
			t.Errorf("bad encoding\n")
		}
	})
	t.Run("2", func(t *testing.T) {
		in := []byte{254, 253}
		fe := field.newElement()
		fe.fromBytes(in)
		if bytes.Equal(in, fe.Bytes()) {
			t.Errorf("bad encoding\n")
		}
	})
	t.Run("3", func(t *testing.T) {
		a := field.randElement(field.newElement(), rand.Reader)
		b := field.newElement()
		b.fromBytes(a.Bytes())
		if !a.equals(b) {
			t.Errorf("bad encoding or decoding\n")
		}
	})
	t.Run("4", func(t *testing.T) {
		a := field.randElement(field.newElement(), rand.Reader)
		b := field.newElement()
		if _, err := b.setString(a.String()); err != nil {
			t.Errorf("bad encoding or decoding\n")
		}
		if !a.equals(b) {
			t.Errorf("bad encoding or decoding\n")
		}
	})
	t.Run("5", func(t *testing.T) {
		a := field.randElement(field.newElement(), rand.Reader)
		b := field.newElement()
		b.setBig(a.Big())
		if !a.equals(b) {
			t.Errorf("bad encoding or decoding\n")
		}
	})
})`

	strTestAddition = `
t.Run(fmt.Sprintf("%d Addition", i*64), func(t *testing.T) {
	var a, b, c, u, v fieldElement
	for j := 0; j < n; j++ {
		field := randTestField(i * 64)
		zero := field.newElement().fromBytes([]byte{0})
		u = field.newElement()
		v = field.newElement()
		a = field.randElement(field.newElement(), rand.Reader)
		b = field.randElement(field.newElement(), rand.Reader)
		c = field.randElement(field.newElement(), rand.Reader)
		field.add(u, a, b)
		field.add(u, u, c)
		field.add(v, b, c)
		field.add(v, v, a)
		if !u.equals(v) {
			t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)
		}
		field.add(u, a, b)
		field.add(v, b, a)
		if !u.equals(v) {
			t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv:%s\n", a, b, u, v)
		}
		field.add(u, a, zero)
		if !u.equals(a) {
			t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)
		}
		field.neg(u, a)
		field.add(u, u, a)
		if !u.equals(zero) {
			t.Fatalf("Bad Negation\na:%s", a.String())
		}
	}
})`

	strTestSubtraction = `
t.Run(fmt.Sprintf("%d Subtraction", i*64), func(t *testing.T) {
	var a, b, c, u, v fieldElement
	for j := 0; j < n; j++ {
		field := randTestField(i * 64)
		zero := field.newElement().fromBytes([]byte{0})
		u = field.newElement()
		v = field.newElement()
		a = field.randElement(field.newElement(), rand.Reader)
		b = field.randElement(field.newElement(), rand.Reader)
		c = field.randElement(field.newElement(), rand.Reader)
		field.sub(u, a, c)
		field.sub(u, u, b)
		field.sub(v, a, b)
		field.sub(v, v, c)
		if !u.equals(v) {
			t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)
		}
		field.sub(u, a, zero)
		if !u.equals(a) {
			t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)
		}
		field.sub(u, a, b)
		field.sub(v, b, a)
		field.add(u, u, v)
		if !u.equals(zero) {
			t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv: %s", a, b, u, v)
		}
		field.sub(u, a, b)
		field.sub(v, b, a)
		field.neg(v, v)
		if !u.equals(u) {
			t.Fatalf("Bad Negation\na:%s", a.String())
		}
	}
})`

	strTestDouble = `
t.Run(fmt.Sprintf("%d Doubling", i*64), func(t *testing.T) {
	var a, u, v fieldElement
	for j := 0; j < n; j++ {
		field := randTestField(i * 64)
		u = field.newElement()
		v = field.newElement()
		a = field.randElement(field.newElement(), rand.Reader)
		field.double(u, a)
		field.add(v, a, a)
		if !u.equals(v) {
			t.Fatalf("Bad doubling\na: %s\nu: %s\nv: %s\n", a, u, v)
		}
	}
})`

	strTestMongomerry = `
t.Run(fmt.Sprintf("%d Montgomerry", i*64), func(t *testing.T) {
	var a, b, c, u, v, w fieldElement
	for j := 0; j < n; j++ {
		field := randTestField(i * 64)
		zero := field.newElement().fromBytes([]byte{0})
		one := field.newElement().fromBytes([]byte{1})
		u = field.newElement()
		v = field.newElement()
		w = field.newElement()
		a = field.randElement(field.newElement(), rand.Reader)
		b = field.randElement(field.newElement(), rand.Reader)
		c = field.randElement(field.newElement(), rand.Reader)
		field.mont(u, zero)
		if !u.equals(zero) {
			t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.one(), field.p())
		}
		field.demont(u, zero)
		if !u.equals(zero) {
			t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.one(), field.p())
		}
		field.mont(u, one)
		if !u.equals(field.one()) {
			t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.one(), field.p())
		}
		field.demont(u, field.one())
		if !u.equals(one) {
			t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.one(), field.p())
		}
		field.mul(u, a, zero)
		if !u.equals(zero) {
			t.Fatalf("Bad zero element\na: %s\nu: %s\np: %s\n", a, u, field.p())
		}
		field.mul(u, a, one)
		field.mul(u, u, field.two())
		if !u.equals(a) {
			t.Fatalf("Multiplication identity does not hold, expected to equal itself\nu: %s\np: %s\n", u, field.p())
		}
		field.mul(u, field.two(), one)
		if !u.equals(field.one()) {
			t.Fatalf("Multiplication identity does not hold, expected to equal r1\nu: %s\np: %s\n", u, field.p())
		}
		field.mul(u, a, b)
		field.mul(u, u, c)
		field.mul(v, b, c)
		field.mul(v, v, a)
		if !u.equals(v) {
			t.Fatalf("Multiplicative associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.p())
		}
		field.add(u, a, b)
		field.mul(u, c, u)
		field.mul(w, a, c)
		field.mul(v, b, c)
		field.add(v, v, w)
		if !u.equals(v) {
			t.Fatalf("Distributivity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.p())
		}
	}
})`

	strTestExp = `
t.Run(fmt.Sprintf("%d Exponentiation", i*64), func(t *testing.T) {
	var a, u, v fieldElement
	for j := 0; j < n; j++ {
		field := randTestField(i * 64)
		u = field.newElement()
		v = field.newElement()
		a = field.randElement(field.newElement(), rand.Reader)
		field.exp(u, a, big.NewInt(0))
		if !u.equals(field.one()) {
			t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.p())
		}
		field.exp(u, a, big.NewInt(1))
		if !u.equals(a) {
			t.Fatalf("Bad exponentiation, expected to equal a\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.p())
		}
		field.mul(u, a, a)
		field.mul(u, u, u)
		field.mul(u, u, u)
		field.exp(v, a, big.NewInt(8))
		if !u.equals(v) {
			t.Fatalf("Bad exponentiation\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.p())
		}
		p := new(big.Int).SetBytes(field.p().Bytes())
		field.exp(u, a, p)
		if !u.equals(a) {
			t.Fatalf("Bad exponentiation, expected to equal itself\nu: %s\na: %s\np: %s\n", u, a, field.p())
		}
		field.exp(u, a, p.Sub(p, big.NewInt(1)))
		if !u.equals(field.one()) {
			t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\na: %s\nr1: %s\np: %s\n", u, a, field.one(), field.p())
		}
	}
})`

	strTestInv = `
t.Run(fmt.Sprintf("%d Inversion", i*64), func(t *testing.T) {
	var a, u, v fieldElement
	for j := 0; j < n; j++ {
		field := randTestField(i * 64)
		one := field.newElement().fromBytes([]byte{1})
		u = field.newElement()
		v = field.newElement()
		a = field.randElement(field.newElement(), rand.Reader)
		field.invmu(u, a)
		field.mul(u, u, a)
		if !u.equals(field.one()) {
			t.Fatalf("Bad inversion, expected to equal r1\nu: %s\nr1: %s\np: %s\n", u, field.one(), field.p())
		}
		field.mont(u, a)
		field.invmd(v, u)
		field.mul(v, v, u)
		if !v.equals(one) {
			t.Fatalf("Bad inversion, expected to equal 1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.p())
		}
		p := new(big.Int).SetBytes(field.p().Bytes())
		field.exp(u, a, p.Sub(p, big.NewInt(2)))
		field.invmu(v, a)
		if !v.equals(u) {
			t.Fatalf("Bad inversion")
		}
		field.inveea(u, a)
		field.mul(u, u, a)
		field.mul(u, u, field.two())
		if !u.equals(one) {
			t.Fatalf("Bad inversion")
		}
	}
})`

	tmplBenches = `
{{ $FE := .FieldElement }} {{ $FIELD := .Field }} {{ $N_BIT := .Bit }}
func Benchmark{{ $FIELD }}(t *testing.B) {
var a, b, c {{ $FE }}
var field *{{ $FIELD }}
for true {
p, err := rand.Prime(rand.Reader, {{ $N_BIT }})
if err != nil {
t.Fatal(err) }
field = New{{ $FIELD }}(p.Bytes())
if field != nil {
break }}
field.RandElement(&a, rand.Reader)
field.RandElement(&b, rand.Reader)
t.Run("Addition", func(t *testing.B) {
t.ResetTimer()
for i := 0; i < t.N; i++ {
field.Add(&c, &a, &b) }})
t.Run("Subtraction", func(t *testing.B) {
t.ResetTimer()
for i := 0; i < t.N; i++ {
field.Sub(&c, &a, &b) }})
t.Run("Doubling", func(t *testing.B) {
t.ResetTimer()
for i := 0; i < t.N; i++ {
field.Double(&c, &a) }})
t.Run("Multiplication", func(t *testing.B) {
t.ResetTimer()
for i := 0; i < t.N; i++ {
field.Mul(&c, &a, &b) }})
// t.Run("Squaring", func(t *testing.B) {
// t.ResetTimer()
// for i := 0; i < t.N; i++ {
// field.Square(&c, &a) }})
t.Run("Inversion", func(t *testing.B) {
t.ResetTimer()
for i := 0; i < t.N; i++ {
field.InvMontUp(&c, &a) }})
t.Run("Exponentiation", func(t *testing.B) {
e := new(big.Int).SetBytes(field.P.Bytes())
t.ResetTimer()
for i := 0; i < t.N; i++ {
field.Exp(&c, &a, e) }})}`
)
