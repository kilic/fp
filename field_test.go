package fp

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"math/big"
	"testing"
)

var n int

func TestMain(m *testing.M) {
	iter := flag.Int("iter", 1000, "# of iters")
	flag.Parse()
	n = *iter
	m.Run()
}

func ceil64(len int) int {
	size := 1 + ((len - 1) / 64)
	if size < 5 {
		return 4
	}
	return size
}

func randTestField(bitlen int) field {
	var field field
	for true {
		p, err := rand.Prime(rand.Reader, bitlen)
		if err != nil {
			panic(err)
		}
		switch ceil64(bitlen) {
		case 4:
			return NewField256(p.Bytes())
		case 5:
			return NewField320(p.Bytes())
		case 6:
			return NewField384(p.Bytes())
		case 7:
			return NewField448(p.Bytes())
		case 8:
			return NewField512(p.Bytes())
		}
		if field != nil {
			break
		}
	}
	return nil
}

func TestField(t *testing.T) {
	// example: subtest single run
	// go test -run 'Element/256_Enc' -iter 1 -v
	for i := 4; i <= 8; i++ {
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
		})
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
		})
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
		})
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
		})
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
		})
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
		})
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
		})
	}
}

func BenchmarkField256(t *testing.B) {
	var a, b, c Fe256
	var field *Field256
	for true {
		p, err := rand.Prime(rand.Reader, 256)
		if err != nil {
			t.Fatal(err)
		}
		field = NewField256(p.Bytes())
		if field != nil {
			break
		}
	}
	field.RandElement(&a, rand.Reader)
	field.RandElement(&b, rand.Reader)
	t.Run("Addition", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Add(&c, &a, &b)
		}
	})
	t.Run("Subtraction", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Sub(&c, &a, &b)
		}
	})
	t.Run("Doubling", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Double(&c, &a)
		}
	})
	t.Run("Multiplication", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Mul(&c, &a, &b)
		}
	})
	// t.Run("Squaring", func(t *testing.B) {
	// t.ResetTimer()
	// for i := 0; i < t.N; i++ {
	// field.Square(&c, &a) }})
	t.Run("Inversion", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.InvMontUp(&c, &a)
		}
	})
	t.Run("Exponentiation", func(t *testing.B) {
		e := new(big.Int).SetBytes(field.P.Bytes())
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Exp(&c, &a, e)
		}
	})
}

func BenchmarkField320(t *testing.B) {
	var a, b, c Fe320
	var field *Field320
	for true {
		p, err := rand.Prime(rand.Reader, 320)
		if err != nil {
			t.Fatal(err)
		}
		field = NewField320(p.Bytes())
		if field != nil {
			break
		}
	}
	field.RandElement(&a, rand.Reader)
	field.RandElement(&b, rand.Reader)
	t.Run("Addition", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Add(&c, &a, &b)
		}
	})
	t.Run("Subtraction", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Sub(&c, &a, &b)
		}
	})
	t.Run("Doubling", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Double(&c, &a)
		}
	})
	t.Run("Multiplication", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Mul(&c, &a, &b)
		}
	})
	// t.Run("Squaring", func(t *testing.B) {
	// t.ResetTimer()
	// for i := 0; i < t.N; i++ {
	// field.Square(&c, &a) }})
	t.Run("Inversion", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.InvMontUp(&c, &a)
		}
	})
	t.Run("Exponentiation", func(t *testing.B) {
		e := new(big.Int).SetBytes(field.P.Bytes())
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Exp(&c, &a, e)
		}
	})
}

func BenchmarkField384(t *testing.B) {
	var a, b, c Fe384
	var field *Field384
	for true {
		p, err := rand.Prime(rand.Reader, 384)
		if err != nil {
			t.Fatal(err)
		}
		field = NewField384(p.Bytes())
		if field != nil {
			break
		}
	}
	field.RandElement(&a, rand.Reader)
	field.RandElement(&b, rand.Reader)
	t.Run("Addition", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Add(&c, &a, &b)
		}
	})
	t.Run("Subtraction", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Sub(&c, &a, &b)
		}
	})
	t.Run("Doubling", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Double(&c, &a)
		}
	})
	t.Run("Multiplication", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Mul(&c, &a, &b)
		}
	})
	// t.Run("Squaring", func(t *testing.B) {
	// t.ResetTimer()
	// for i := 0; i < t.N; i++ {
	// field.Square(&c, &a) }})
	t.Run("Inversion", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.InvMontUp(&c, &a)
		}
	})
	t.Run("Exponentiation", func(t *testing.B) {
		e := new(big.Int).SetBytes(field.P.Bytes())
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Exp(&c, &a, e)
		}
	})
}

func BenchmarkField448(t *testing.B) {
	var a, b, c Fe448
	var field *Field448
	for true {
		p, err := rand.Prime(rand.Reader, 448)
		if err != nil {
			t.Fatal(err)
		}
		field = NewField448(p.Bytes())
		if field != nil {
			break
		}
	}
	field.RandElement(&a, rand.Reader)
	field.RandElement(&b, rand.Reader)
	t.Run("Addition", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Add(&c, &a, &b)
		}
	})
	t.Run("Subtraction", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Sub(&c, &a, &b)
		}
	})
	t.Run("Doubling", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Double(&c, &a)
		}
	})
	t.Run("Multiplication", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Mul(&c, &a, &b)
		}
	})
	// t.Run("Squaring", func(t *testing.B) {
	// t.ResetTimer()
	// for i := 0; i < t.N; i++ {
	// field.Square(&c, &a) }})
	t.Run("Inversion", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.InvMontUp(&c, &a)
		}
	})
	t.Run("Exponentiation", func(t *testing.B) {
		e := new(big.Int).SetBytes(field.P.Bytes())
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Exp(&c, &a, e)
		}
	})
}

func BenchmarkField512(t *testing.B) {
	var a, b, c Fe512
	var field *Field512
	for true {
		p, err := rand.Prime(rand.Reader, 512)
		if err != nil {
			t.Fatal(err)
		}
		field = NewField512(p.Bytes())
		if field != nil {
			break
		}
	}
	field.RandElement(&a, rand.Reader)
	field.RandElement(&b, rand.Reader)
	t.Run("Addition", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Add(&c, &a, &b)
		}
	})
	t.Run("Subtraction", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Sub(&c, &a, &b)
		}
	})
	t.Run("Doubling", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Double(&c, &a)
		}
	})
	t.Run("Multiplication", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Mul(&c, &a, &b)
		}
	})
	// t.Run("Squaring", func(t *testing.B) {
	// t.ResetTimer()
	// for i := 0; i < t.N; i++ {
	// field.Square(&c, &a) }})
	t.Run("Inversion", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.InvMontUp(&c, &a)
		}
	})
	t.Run("Exponentiation", func(t *testing.B) {
		e := new(big.Int).SetBytes(field.P.Bytes())
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Exp(&c, &a, e)
		}
	})
}

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
	inveea(inv, fe fieldElement)
}
type fieldElement interface {
	String() string
	Bytes() []byte
	Big() *big.Int
	setBig(b *big.Int) fieldElement
	setString(s string) (fieldElement, error)
	fromBytes(in []byte) fieldElement
	equals(other fieldElement) bool
	limb(i int) uint64
}

func (fe *Fe256) setBig(b *big.Int) fieldElement {
	return fe.SetBig(b)
}
func (fe *Fe256) setString(s string) (fieldElement, error) {
	return fe.SetString(s)
}
func (fe *Fe256) fromBytes(in []byte) fieldElement {
	return fe.FromBytes(in)
}
func (fe *Fe256) equals(other fieldElement) bool {
	for i := 0; i < len(fe); i++ {
		if fe[i] != other.limb(i) {
			return false
		}
	}
	return true
}
func (fe *Fe256) limb(i int) uint64 {
	return fe[i]
}
func (f *Field256) add(c, a, b fieldElement) {
	f.Add(c.(*Fe256), a.(*Fe256), b.(*Fe256))
}
func (f *Field256) double(c, a fieldElement) {
	f.Double(c.(*Fe256), a.(*Fe256))
}
func (f *Field256) sub(c, a, b fieldElement) {
	f.Sub(c.(*Fe256), a.(*Fe256), b.(*Fe256))
}
func (f *Field256) neg(c, a fieldElement) {
	f.Neg(c.(*Fe256), a.(*Fe256))
}
func (f *Field256) square(c, a fieldElement) {
	f.Square(c.(*Fe256), a.(*Fe256))
}
func (f *Field256) mul(c, a, b fieldElement) {
	f.Mul(c.(*Fe256), a.(*Fe256), b.(*Fe256))
}
func (f *Field256) exp(c, a fieldElement, e *big.Int) {
	f.Exp(c.(*Fe256), a.(*Fe256), e)
}
func (f *Field256) mont(c, a fieldElement) {
	f.Mont(c.(*Fe256), a.(*Fe256))
}
func (f *Field256) demont(c, a fieldElement) {
	f.Demont(c.(*Fe256), a.(*Fe256))
}
func (f *Field256) one() fieldElement {
	return new(Fe256).Set(f.r1)
}
func (f *Field256) two() fieldElement {
	return new(Fe256).Set(f.r2)
}
func (f *Field256) p() fieldElement {
	return new(Fe256).Set(f.P)
}
func (f *Field256) newElement() fieldElement {
	return &Fe256{}
}
func (f *Field256) randElement(fe fieldElement, r io.Reader) fieldElement {
	_, err := f.RandElement(fe.(*Fe256), r)
	if err != nil {
		panic(err)
	}
	return fe
}
func (f *Field256) invmu(inv, fe fieldElement) {
	f.InvMontUp(inv.(*Fe256), fe.(*Fe256))
}
func (f *Field256) invmd(inv, fe fieldElement) {
	f.InvMontDown(inv.(*Fe256), fe.(*Fe256))
}
func (f *Field256) inveea(inv, fe fieldElement) {
	f.InvEEA(inv.(*Fe256), fe.(*Fe256))
}

func (fe *Fe320) setBig(b *big.Int) fieldElement {
	return fe.SetBig(b)
}
func (fe *Fe320) setString(s string) (fieldElement, error) {
	return fe.SetString(s)
}
func (fe *Fe320) fromBytes(in []byte) fieldElement {
	return fe.FromBytes(in)
}
func (fe *Fe320) equals(other fieldElement) bool {
	for i := 0; i < len(fe); i++ {
		if fe[i] != other.limb(i) {
			return false
		}
	}
	return true
}
func (fe *Fe320) limb(i int) uint64 {
	return fe[i]
}
func (f *Field320) add(c, a, b fieldElement) {
	f.Add(c.(*Fe320), a.(*Fe320), b.(*Fe320))
}
func (f *Field320) double(c, a fieldElement) {
	f.Double(c.(*Fe320), a.(*Fe320))
}
func (f *Field320) sub(c, a, b fieldElement) {
	f.Sub(c.(*Fe320), a.(*Fe320), b.(*Fe320))
}
func (f *Field320) neg(c, a fieldElement) {
	f.Neg(c.(*Fe320), a.(*Fe320))
}
func (f *Field320) square(c, a fieldElement) {
	f.Square(c.(*Fe320), a.(*Fe320))
}
func (f *Field320) mul(c, a, b fieldElement) {
	f.Mul(c.(*Fe320), a.(*Fe320), b.(*Fe320))
}
func (f *Field320) exp(c, a fieldElement, e *big.Int) {
	f.Exp(c.(*Fe320), a.(*Fe320), e)
}
func (f *Field320) mont(c, a fieldElement) {
	f.Mont(c.(*Fe320), a.(*Fe320))
}
func (f *Field320) demont(c, a fieldElement) {
	f.Demont(c.(*Fe320), a.(*Fe320))
}
func (f *Field320) one() fieldElement {
	return new(Fe320).Set(f.r1)
}
func (f *Field320) two() fieldElement {
	return new(Fe320).Set(f.r2)
}
func (f *Field320) p() fieldElement {
	return new(Fe320).Set(f.P)
}
func (f *Field320) newElement() fieldElement {
	return &Fe320{}
}
func (f *Field320) randElement(fe fieldElement, r io.Reader) fieldElement {
	_, err := f.RandElement(fe.(*Fe320), r)
	if err != nil {
		panic(err)
	}
	return fe
}
func (f *Field320) invmu(inv, fe fieldElement) {
	f.InvMontUp(inv.(*Fe320), fe.(*Fe320))
}
func (f *Field320) invmd(inv, fe fieldElement) {
	f.InvMontDown(inv.(*Fe320), fe.(*Fe320))
}
func (f *Field320) inveea(inv, fe fieldElement) {
	f.InvEEA(inv.(*Fe320), fe.(*Fe320))
}

func (fe *Fe384) setBig(b *big.Int) fieldElement {
	return fe.SetBig(b)
}
func (fe *Fe384) setString(s string) (fieldElement, error) {
	return fe.SetString(s)
}
func (fe *Fe384) fromBytes(in []byte) fieldElement {
	return fe.FromBytes(in)
}
func (fe *Fe384) equals(other fieldElement) bool {
	for i := 0; i < len(fe); i++ {
		if fe[i] != other.limb(i) {
			return false
		}
	}
	return true
}
func (fe *Fe384) limb(i int) uint64 {
	return fe[i]
}
func (f *Field384) add(c, a, b fieldElement) {
	f.Add(c.(*Fe384), a.(*Fe384), b.(*Fe384))
}
func (f *Field384) double(c, a fieldElement) {
	f.Double(c.(*Fe384), a.(*Fe384))
}
func (f *Field384) sub(c, a, b fieldElement) {
	f.Sub(c.(*Fe384), a.(*Fe384), b.(*Fe384))
}
func (f *Field384) neg(c, a fieldElement) {
	f.Neg(c.(*Fe384), a.(*Fe384))
}
func (f *Field384) square(c, a fieldElement) {
	f.Square(c.(*Fe384), a.(*Fe384))
}
func (f *Field384) mul(c, a, b fieldElement) {
	f.Mul(c.(*Fe384), a.(*Fe384), b.(*Fe384))
}
func (f *Field384) exp(c, a fieldElement, e *big.Int) {
	f.Exp(c.(*Fe384), a.(*Fe384), e)
}
func (f *Field384) mont(c, a fieldElement) {
	f.Mont(c.(*Fe384), a.(*Fe384))
}
func (f *Field384) demont(c, a fieldElement) {
	f.Demont(c.(*Fe384), a.(*Fe384))
}
func (f *Field384) one() fieldElement {
	return new(Fe384).Set(f.r1)
}
func (f *Field384) two() fieldElement {
	return new(Fe384).Set(f.r2)
}
func (f *Field384) p() fieldElement {
	return new(Fe384).Set(f.P)
}
func (f *Field384) newElement() fieldElement {
	return &Fe384{}
}
func (f *Field384) randElement(fe fieldElement, r io.Reader) fieldElement {
	_, err := f.RandElement(fe.(*Fe384), r)
	if err != nil {
		panic(err)
	}
	return fe
}
func (f *Field384) invmu(inv, fe fieldElement) {
	f.InvMontUp(inv.(*Fe384), fe.(*Fe384))
}
func (f *Field384) invmd(inv, fe fieldElement) {
	f.InvMontDown(inv.(*Fe384), fe.(*Fe384))
}
func (f *Field384) inveea(inv, fe fieldElement) {
	f.InvEEA(inv.(*Fe384), fe.(*Fe384))
}

func (fe *Fe448) setBig(b *big.Int) fieldElement {
	return fe.SetBig(b)
}
func (fe *Fe448) setString(s string) (fieldElement, error) {
	return fe.SetString(s)
}
func (fe *Fe448) fromBytes(in []byte) fieldElement {
	return fe.FromBytes(in)
}
func (fe *Fe448) equals(other fieldElement) bool {
	for i := 0; i < len(fe); i++ {
		if fe[i] != other.limb(i) {
			return false
		}
	}
	return true
}
func (fe *Fe448) limb(i int) uint64 {
	return fe[i]
}
func (f *Field448) add(c, a, b fieldElement) {
	f.Add(c.(*Fe448), a.(*Fe448), b.(*Fe448))
}
func (f *Field448) double(c, a fieldElement) {
	f.Double(c.(*Fe448), a.(*Fe448))
}
func (f *Field448) sub(c, a, b fieldElement) {
	f.Sub(c.(*Fe448), a.(*Fe448), b.(*Fe448))
}
func (f *Field448) neg(c, a fieldElement) {
	f.Neg(c.(*Fe448), a.(*Fe448))
}
func (f *Field448) square(c, a fieldElement) {
	f.Square(c.(*Fe448), a.(*Fe448))
}
func (f *Field448) mul(c, a, b fieldElement) {
	f.Mul(c.(*Fe448), a.(*Fe448), b.(*Fe448))
}
func (f *Field448) exp(c, a fieldElement, e *big.Int) {
	f.Exp(c.(*Fe448), a.(*Fe448), e)
}
func (f *Field448) mont(c, a fieldElement) {
	f.Mont(c.(*Fe448), a.(*Fe448))
}
func (f *Field448) demont(c, a fieldElement) {
	f.Demont(c.(*Fe448), a.(*Fe448))
}
func (f *Field448) one() fieldElement {
	return new(Fe448).Set(f.r1)
}
func (f *Field448) two() fieldElement {
	return new(Fe448).Set(f.r2)
}
func (f *Field448) p() fieldElement {
	return new(Fe448).Set(f.P)
}
func (f *Field448) newElement() fieldElement {
	return &Fe448{}
}
func (f *Field448) randElement(fe fieldElement, r io.Reader) fieldElement {
	_, err := f.RandElement(fe.(*Fe448), r)
	if err != nil {
		panic(err)
	}
	return fe
}
func (f *Field448) invmu(inv, fe fieldElement) {
	f.InvMontUp(inv.(*Fe448), fe.(*Fe448))
}
func (f *Field448) invmd(inv, fe fieldElement) {
	f.InvMontDown(inv.(*Fe448), fe.(*Fe448))
}
func (f *Field448) inveea(inv, fe fieldElement) {
	f.InvEEA(inv.(*Fe448), fe.(*Fe448))
}

func (fe *Fe512) setBig(b *big.Int) fieldElement {
	return fe.SetBig(b)
}
func (fe *Fe512) setString(s string) (fieldElement, error) {
	return fe.SetString(s)
}
func (fe *Fe512) fromBytes(in []byte) fieldElement {
	return fe.FromBytes(in)
}
func (fe *Fe512) equals(other fieldElement) bool {
	for i := 0; i < len(fe); i++ {
		if fe[i] != other.limb(i) {
			return false
		}
	}
	return true
}
func (fe *Fe512) limb(i int) uint64 {
	return fe[i]
}
func (f *Field512) add(c, a, b fieldElement) {
	f.Add(c.(*Fe512), a.(*Fe512), b.(*Fe512))
}
func (f *Field512) double(c, a fieldElement) {
	f.Double(c.(*Fe512), a.(*Fe512))
}
func (f *Field512) sub(c, a, b fieldElement) {
	f.Sub(c.(*Fe512), a.(*Fe512), b.(*Fe512))
}
func (f *Field512) neg(c, a fieldElement) {
	f.Neg(c.(*Fe512), a.(*Fe512))
}
func (f *Field512) square(c, a fieldElement) {
	f.Square(c.(*Fe512), a.(*Fe512))
}
func (f *Field512) mul(c, a, b fieldElement) {
	f.Mul(c.(*Fe512), a.(*Fe512), b.(*Fe512))
}
func (f *Field512) exp(c, a fieldElement, e *big.Int) {
	f.Exp(c.(*Fe512), a.(*Fe512), e)
}
func (f *Field512) mont(c, a fieldElement) {
	f.Mont(c.(*Fe512), a.(*Fe512))
}
func (f *Field512) demont(c, a fieldElement) {
	f.Demont(c.(*Fe512), a.(*Fe512))
}
func (f *Field512) one() fieldElement {
	return new(Fe512).Set(f.r1)
}
func (f *Field512) two() fieldElement {
	return new(Fe512).Set(f.r2)
}
func (f *Field512) p() fieldElement {
	return new(Fe512).Set(f.P)
}
func (f *Field512) newElement() fieldElement {
	return &Fe512{}
}
func (f *Field512) randElement(fe fieldElement, r io.Reader) fieldElement {
	_, err := f.RandElement(fe.(*Fe512), r)
	if err != nil {
		panic(err)
	}
	return fe
}
func (f *Field512) invmu(inv, fe fieldElement) {
	f.InvMontUp(inv.(*Fe512), fe.(*Fe512))
}
func (f *Field512) invmd(inv, fe fieldElement) {
	f.InvMontDown(inv.(*Fe512), fe.(*Fe512))
}
func (f *Field512) inveea(inv, fe fieldElement) {
	f.InvEEA(inv.(*Fe512), fe.(*Fe512))
}
