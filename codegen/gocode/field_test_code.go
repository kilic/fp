package gocode

const fieldTestFixedModulus = `
import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"testing"
)

var fuz int

func TestMain(m *testing.M) {
	_fuz := flag.Int("fuzz", 1, "# of iters")
	flag.Parse()
	fuz = *_fuz
	m.Run()
}

func resolveLimbSize(bitSize int) int {
	size := (bitSize / 64)
	if bitSize%64 != 0 {
		size += 1
	}
	return size
}

func randBytes(max *big.Int) []byte {
	return padBytes(randBig(max).Bytes(), resolveLimbSize(max.BitLen())*8)
}

func randBig(max *big.Int) *big.Int {
	bi, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	return bi
}

func padBytes(in []byte, size int) []byte {
	out := make([]byte, size)
	copy(out[size-len(in):], in)
	return out
}

func BenchmarkField(t *testing.B) {
	in_a := randBytes(pbig)
	in_b := randBytes(pbig)
	a, _ := newFieldElementFromBytes(in_a)
	b, _ := newFieldElementFromBytes(in_b)
	c := newFieldElement()
	t.Run("Add", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			add(c, a, b)
		}
	})
	t.Run("Double", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			double(c, a)
		}
	})
	t.Run("Sub", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			sub(c, a, b)
		}
	})
	t.Run("Mul", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			mul(c, a, b)
		}
	})
	t.Run("Compare", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			equal(a, b)
		}
	})
}

func TestField(t *testing.T) {
	t.Run(fmt.Sprintf("A/%d: Serialization", limbSize*64), func(t *testing.T) {
		b1 := make([]byte, byteSize)
		b1[len(b1)-1] = byte(1)
		b2 := toBytes(r)
		if !bytes.Equal(b1, b2) {
			t.Fatalf("Bad serialization\n")
		}
		for i := 0; i < fuz; i++ {
			b1 := randBytes(pbig)
			a, _ := newFieldElementFromBytes(b1)
			b2 := toBytes(a)
			if !bytes.Equal(b1, b2) {
				t.Fatalf("Bad serialization")
			}
		}
	})
	t.Run(fmt.Sprintf("B0/%d: Addition cross test", limbSize*64), func(t *testing.T) {
		for i := 0; i < fuz; i++ {

			in_a := randBytes(pbig)
			in_b := randBytes(pbig)
			a, _ := newFieldElementFromBytes(in_a)
			b, _ := newFieldElementFromBytes(in_b)
			c := newFieldElement()
			big_a := new(big.Int).SetBytes(in_a)
			big_b := new(big.Int).SetBytes(in_b)
			big_c := new(big.Int)
			add(c, a, b)
			out_1 := toBytes(c)
			out_2 := padBytes(big_c.Add(big_a, big_b).Mod(big_c, pbig).Bytes(), byteSize)
			if !bytes.Equal(out_1, out_2) {
				t.Fatalf("Bad Addition: Cross test against big Int")
			}
			double(c, a)
			out_1 = toBytes(c)
			out_2 = padBytes(big_c.Add(big_a, big_a).Mod(big_c, pbig).Bytes(), byteSize)
			if !bytes.Equal(out_1, out_2) {
				t.Fatalf("Bad Doubling: Cross test against big Int")
			}
			sub(c, a, b)
			out_1 = toBytes(c)
			out_2 = padBytes(big_c.Sub(big_a, big_b).Mod(big_c, pbig).Bytes(), byteSize)
			if !bytes.Equal(out_1, out_2) {
				t.Fatalf("Bad Subtraction: Cross test against big Int")
			}
			neg(c, a)
			out_1 = toBytes(c)
			out_2 = padBytes(big_c.Neg(big_a).Mod(big_c, pbig).Bytes(), byteSize)
			if !bytes.Equal(out_1, out_2) {
				t.Fatalf("Bad Negation: Cross test against big Int")
			}
		}
	})
	t.Run(fmt.Sprintf("B0/%d: Additive properties", limbSize*64), func(t *testing.T) {
		for i := 0; i < fuz; i++ {

			in_a := randBytes(pbig)
			in_b := randBytes(pbig)
			a, _ := newFieldElementFromBytes(in_a)
			b, _ := newFieldElementFromBytes(in_b)
			c_1 := newFieldElement()
			c_2 := newFieldElement()
			add(c_1, a, zero)
			if !equal(c_1, a) {
				t.Fatalf("Bad Addition: Add by zero")
			}
			sub(c_1, a, zero)
			if !equal(c_1, a) {
				t.Fatalf("Bad Subtraction: Subtract by zero")
			}
			double(c_1, zero)
			if !equal(c_1, zero) {
				t.Fatalf("Bad Doubling: Doubling zero")
			}
			neg(c_1, zero)
			if !equal(c_1, zero) {
				t.Fatalf("Bad Negation: Negate zero")
			}
			sub(c_1, zero, a)
			neg(c_2, a)
			if !equal(c_1, c_2) {
				t.Fatalf("Bad Negation")
			}
			double(c_1, a)
			add(c_2, a, a)
			if !equal(c_1, c_2) {
				t.Fatalf("Bad Doubling")
			}
			add(c_1, a, b)
			add(c_2, b, a)
			if !equal(c_1, c_2) {
				t.Fatalf("Bad Addition: Commutativity")
			}
			sub(c_1, a, b)
			sub(c_2, b, a)
			neg(c_2, c_2)
			if !equal(c_1, c_2) {
				t.Fatalf("Bad Subtraction: Commutativity")
			}
			c_x, _ := newFieldElementFromBytes(randBytes(pbig))
			add(c_1, a, b)
			add(c_1, c_1, c_x)
			add(c_2, a, c_x)
			add(c_2, c_2, b)
			if !equal(c_1, c_2) {
				t.Fatalf("Bad Addition: Associativity")
			}
			sub(c_1, a, b)
			sub(c_1, c_1, c_x)
			sub(c_2, a, c_x)
			sub(c_2, c_2, b)
			if !equal(c_1, c_2) {
				t.Fatalf("Bad Addition: Associativity")
			}
		}
	})
	t.Run(fmt.Sprintf("C0/%d: Multiplication cross test", limbSize*64), func(t *testing.T) {
		for i := 0; i < fuz; i++ {
			in_a := randBytes(pbig)
			in_b := randBytes(pbig)
			a, _ := newFieldElementFromBytes(in_a)
			b, _ := newFieldElementFromBytes(in_b)
			c := newFieldElement()
			mul(c, a, b)
			out_1 := toBytes(c)
			big_a := new(big.Int).SetBytes(in_a)
			big_b := new(big.Int).SetBytes(in_b)
			big_c := new(big.Int)
			out_2 := padBytes(big_c.Mul(big_a, big_b).Mod(big_c, pbig).Bytes(), byteSize)
			if !bytes.Equal(out_1, out_2) {
				t.Fatalf("Bad Multiplication: Cross test against big.Int")
			}
		}
	})
	t.Run(fmt.Sprintf("C1/%d: Multiplication properties", limbSize*64), func(t *testing.T) {
		for i := 0; i < fuz; i++ {
			in_a := randBytes(pbig)
			in_b := randBytes(pbig)
			a, _ := newFieldElementFromBytes(in_a)
			b, _ := newFieldElementFromBytes(in_b)
			c_1 := newFieldElement()
			c_2 := newFieldElement()
			mul(c_1, a, zero)
			if !equal(c_1, zero) {
				t.Fatalf("Bad Multiplication: Mul by zero")
			}
			mul(c_1, a, one)
			if !equal(c_1, a) {
				t.Fatalf("Bad Multiplication: Mul by one")
			}
			mul(c_1, a, b)
			mul(c_2, b, a)
			if !equal(c_1, c_2) {
				t.Fatalf("Bad Multiplication: Commutativity")
			}
			c_x, _ := newFieldElementFromBytes(randBytes(pbig))
			mul(c_1, a, b)
			mul(c_1, c_1, c_x)
			mul(c_2, c_x, b)
			mul(c_2, c_2, a)
			if !equal(c_1, c_2) {
				t.Fatalf("Bad Multiplication: Associativity")
			}
		}
	})
}
`

const fieldTestNonFixedModulus = `
import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"testing"
)

var fuz int

func TestMain(m *testing.M) {
	_fuz := flag.Int("fuzz", 1, "# of iters")
	flag.Parse()
	fuz = *_fuz
	m.Run()
}

func randField() *field {
	pbig, err := rand.Prime(rand.Reader, byteSize*8)
	if err != nil {
		panic(err)
	}
	rawpbytes := pbig.Bytes()
	pbytes := make([]byte, byteSize)
	copy(pbytes[byteSize-len(rawpbytes):], pbig.Bytes())
	field, err := newField(pbytes)
	if err != nil {
		panic(err)
	}
	return field
}

func resolveLimbSize(bitSize int) int {
	size := (bitSize / 64)
	if bitSize%64 != 0 {
		size += 1
	}
	return size
}

func randBytes(max *big.Int) []byte {
	return padBytes(randBig(max).Bytes(), resolveLimbSize(max.BitLen())*8)
}

func randBig(max *big.Int) *big.Int {
	bi, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	return bi
}

func BenchmarkField(t *testing.B) {
	field := randField()
	in_a := randBytes(field.pbig)
	in_b := randBytes(field.pbig)
	a, _ := field.newFieldElementFromBytes(in_a)
	b, _ := field.newFieldElementFromBytes(in_b)
	c := field.newFieldElement()
	t.Run("Add", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.add(c, a, b)
		}
	})
	t.Run("Double", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.double(c, a)
		}
	})
	t.Run("Sub", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.sub(c, a, b)
		}
	})
	t.Run("Mul", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.mul(c, a, b)
		}
	})
	t.Run("Compare", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.equal(a, b)
		}
	})
}

func TestField(t *testing.T) {
	_ = randBytes
	t.Run(fmt.Sprintf("A/%d: Serialization", limbSize*64), func(t *testing.T) {
		field := randField()
		bts := make([]byte, byteSize)
		bts[len(bts)-1] = byte(1)
		_bts := field.toBytes(field.r)
		if !bytes.Equal(bts, _bts) {
			t.Fatalf("Bad serialization\n")
		}
		for i := 0; i < fuz; i++ {
			field := randField()
			bts := randBytes(field.pbig)
			a, _ := field.newFieldElementFromBytes(bts)
			_bts = field.toBytes(a)
			if !bytes.Equal(bts, _bts) {
				t.Fatalf("Bad serialization")
			}
		}
	})

	t.Run(fmt.Sprintf("B0/%d: Addition cross test", limbSize*64), func(t *testing.T) {
		for i := 0; i < fuz; i++ {
			field := randField()
			in_a := randBytes(field.pbig)
			in_b := randBytes(field.pbig)
			a, _ := field.newFieldElementFromBytes(in_a)
			b, _ := field.newFieldElementFromBytes(in_b)
			c := field.newFieldElement()
			big_a := new(big.Int).SetBytes(in_a)
			big_b := new(big.Int).SetBytes(in_b)
			big_c := new(big.Int)
			field.add(c, a, b)
			out_1 := field.toBytes(c)
			out_2 := padBytes(big_c.Add(big_a, big_b).Mod(big_c, field.pbig).Bytes(), byteSize)
			if !bytes.Equal(out_1, out_2) {
				t.Fatalf("Bad Addition: Cross test against big Int")
			}
			field.double(c, a)
			out_1 = field.toBytes(c)
			out_2 = padBytes(big_c.Add(big_a, big_a).Mod(big_c, field.pbig).Bytes(), byteSize)
			if !bytes.Equal(out_1, out_2) {
				t.Fatalf("Bad Doubling: Cross test against big Int")
			}
			field.sub(c, a, b)
			out_1 = field.toBytes(c)
			out_2 = padBytes(big_c.Sub(big_a, big_b).Mod(big_c, field.pbig).Bytes(), byteSize)
			if !bytes.Equal(out_1, out_2) {
				t.Fatalf("Bad Subtraction: Cross test against big Int")
			}
			field.neg(c, a)
			out_1 = field.toBytes(c)
			out_2 = padBytes(big_c.Neg(big_a).Mod(big_c, field.pbig).Bytes(), byteSize)
			if !bytes.Equal(out_1, out_2) {
				t.Fatalf("Bad Negation: Cross test against big Int")
			}
		}
	})
	t.Run(fmt.Sprintf("B0/%d: Additive properties", limbSize*64), func(t *testing.T) {
		for i := 0; i < fuz; i++ {
			field := randField()
			in_a := randBytes(field.pbig)
			in_b := randBytes(field.pbig)
			a, _ := field.newFieldElementFromBytes(in_a)
			b, _ := field.newFieldElementFromBytes(in_b)
			c_1 := field.newFieldElement()
			c_2 := field.newFieldElement()
			field.add(c_1, a, field.zero)
			if !field.equal(c_1, a) {
				t.Fatalf("Bad Addition: Add by zero")
			}
			field.sub(c_1, a, field.zero)
			if !field.equal(c_1, a) {
				t.Fatalf("Bad Subtraction: Subtract by zero")
			}
			field.double(c_1, field.zero)
			if !field.equal(c_1, field.zero) {
				t.Fatalf("Bad Doubling: Doubling zero")
			}
			field.neg(c_1, field.zero)
			if !field.equal(c_1, field.zero) {
				t.Fatalf("Bad Negation: Negate zero")
			}
			field.sub(c_1, field.zero, a)
			field.neg(c_2, a)
			if !field.equal(c_1, c_2) {
				t.Fatalf("Bad Negation")
			}
			field.double(c_1, a)
			field.add(c_2, a, a)
			if !field.equal(c_1, c_2) {
				t.Fatalf("Bad Doubling")
			}
			field.add(c_1, a, b)
			field.add(c_2, b, a)
			if !field.equal(c_1, c_2) {
				t.Fatalf("Bad Addition: Commutativity")
			}
			field.sub(c_1, a, b)
			field.sub(c_2, b, a)
			field.neg(c_2, c_2)
			if !field.equal(c_1, c_2) {
				t.Fatalf("Bad Subtraction: Commutativity")
			}
			c_x, _ := field.newFieldElementFromBytes(randBytes(field.pbig))
			field.add(c_1, a, b)
			field.add(c_1, c_1, c_x)
			field.add(c_2, a, c_x)
			field.add(c_2, c_2, b)
			if !field.equal(c_1, c_2) {
				t.Fatalf("Bad Addition: Associativity")
			}
			field.sub(c_1, a, b)
			field.sub(c_1, c_1, c_x)
			field.sub(c_2, a, c_x)
			field.sub(c_2, c_2, b)
			if !field.equal(c_1, c_2) {
				t.Fatalf("Bad Addition: Associativity")
			}
		}
	})
	t.Run(fmt.Sprintf("C0/%d: Multiplication cross test", limbSize*64), func(t *testing.T) {
		for i := 0; i < fuz; i++ {
			field := randField()
			in_a := randBytes(field.pbig)
			in_b := randBytes(field.pbig)
			a, _ := field.newFieldElementFromBytes(in_a)
			b, _ := field.newFieldElementFromBytes(in_b)
			c := field.newFieldElement()
			field.mul(c, a, b)
			big_a := new(big.Int).SetBytes(in_a)
			big_b := new(big.Int).SetBytes(in_b)
			big_c := new(big.Int)
			out_1 := field.toBytes(c)
			out_2 := padBytes(big_c.Mul(big_a, big_b).Mod(big_c, field.pbig).Bytes(), byteSize)
			if !bytes.Equal(out_1, out_2) {
				t.Fatalf("Bad Multiplication: Cross test against big.Int")
			}
		}
	})
	t.Run(fmt.Sprintf("C1/%d: Multiplication properties", limbSize*64), func(t *testing.T) {
		for i := 0; i < fuz; i++ {
			field := randField()
			in_a := randBytes(field.pbig)
			in_b := randBytes(field.pbig)
			a, _ := field.newFieldElementFromBytes(in_a)
			b, _ := field.newFieldElementFromBytes(in_b)
			c_1 := field.newFieldElement()
			c_2 := field.newFieldElement()
			field.mul(c_1, a, field.zero)
			if !field.equal(c_1, field.zero) {
				t.Fatalf("Bad Multiplication: Mul by zero")
			}
			field.mul(c_1, a, field.one)
			if !field.equal(c_1, a) {
				t.Fatalf("Bad Multiplication: Mul by one")
			}
			field.mul(c_1, a, b)
			field.mul(c_2, b, a)
			if !field.equal(c_1, c_2) {
				t.Fatalf("Bad Multiplication: Commutativity")
			}
			c_x, _ := field.newFieldElementFromBytes(randBytes(field.pbig))
			field.mul(c_1, a, b)
			field.mul(c_1, c_1, c_x)
			field.mul(c_2, c_x, b)
			field.mul(c_2, c_2, a)
			if !field.equal(c_1, c_2) {
				t.Fatalf("Bad Multiplication: Associativity")
			}
		}
	})
}

`
