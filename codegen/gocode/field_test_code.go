package gocode

const fieldTestFixedModulus = `
import (
	"bytes"
	"crypto/rand"
	"flag"
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

func BenchmarkField(t *testing.B) {
	in_a := randBytes(pbig)
	in_b := randBytes(pbig)
	a, _ := newFieldElementFromBytes(in_a)
	b, _ := newFieldElementFromBytes(in_b)
	c := newFieldElement()
	t.Run("add", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			add(c, a, b)
		}
	})
	t.Run("double", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			double(c, a)
		}
	})
	t.Run("sub", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			sub(c, a, b)
		}
	})
	t.Run("mul", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			mul(c, a, b)
		}
	})
	t.Run("cmp", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			a.equal(b)
		}
	})
}

func TestCompare(t *testing.T) {
	if r.cmp(r) != 0 {
		t.Fatalf("r == r (cmp)")
	}
	if !r.equal(r) {
		t.Fatalf("r == r (equal)")
	}
	if r.equal(&modulus) {
		t.Fatalf("p != r")
	}
	if r.equal(zero) {
		t.Fatalf("r != 0")
	}
	if !zero.equal(zero) {
		t.Fatalf("0 == 0")
	}
	if modulus.cmp(r) != 1 {
		t.Fatalf("p > r")
	}
	if r.cmp(&modulus) != -1 {
		t.Fatalf("r < p")
	}
}

func TestSerialization(t *testing.T) {
	// demont(r) == 1
	b0 := make([]byte, byteSize)
	b0[len(b0)-1] = byte(1)
	b1 := toBytes(r)
	if !bytes.Equal(b0, b1) {
		t.Fatalf("demont(r) must be equal to 1\n")
	}
	// is a => modulus should not be valid
	_, err := newFieldElementFromBytes(pbig.Bytes())
	if err == nil {
		t.Fatalf("a number eq or larger than modulus must not be valid")
	}
	for i := 0; i < fuz; i++ {
		// bytes
		b0 := randBytes(pbig)
		a0, err := newFieldElementFromBytes(b0)
		if err != nil {
			t.Fatal(err)
		}
		b1 = toBytes(a0)
		if !bytes.Equal(b0, b1) {
			t.Fatalf("bad serialization (bytes)")
		}
		// string
		s := toString(a0)
		a1, err := newFieldElementFromString(s)
		if err != nil {
			t.Fatal(err)
		}
		if !a0.equal(a1) {
			t.Fatalf("bad serialization (str)")
		}
		// big int
		a0, err = newFieldElementFromBytes(b0)
		if err != nil {
			t.Fatal(err)
		}
		bi := toBig(a0)
		a1, err = newFieldElementFromBig(bi)
		if err != nil {
			t.Fatal(err)
		}
		if !a0.equal(a1) {
			t.Fatalf("bad serialization (big.Int)")
		}
	}
}

func TestAdditionCrossAgainstBigInt(t *testing.T) {
	for i := 0; i < fuz; i++ {
		a, _ := randFieldElement(rand.Reader)
		b, _ := randFieldElement(rand.Reader)
		c := newFieldElement()
		big_a := toBig(a)
		big_b := toBig(b)
		big_c := new(big.Int)
		add(c, a, b)
		out_1 := toBytes(c)
		out_2 := padBytes(big_c.Add(big_a, big_b).Mod(big_c, pbig).Bytes(), byteSize)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied A")
		}
		double(c, a)
		out_1 = toBytes(c)
		out_2 = padBytes(big_c.Add(big_a, big_a).Mod(big_c, pbig).Bytes(), byteSize)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied B")
		}
		sub(c, a, b)
		out_1 = toBytes(c)
		out_2 = padBytes(big_c.Sub(big_a, big_b).Mod(big_c, pbig).Bytes(), byteSize)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied C")
		}
		neg(c, a)
		out_1 = toBytes(c)
		out_2 = padBytes(big_c.Neg(big_a).Mod(big_c, pbig).Bytes(), byteSize)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied D")
		}
	}
}

func TestAdditionProperties(t *testing.T) {
	for i := 0; i < fuz; i++ {
		a, _ := randFieldElement(rand.Reader)
		b, _ := randFieldElement(rand.Reader)
		c_1 := newFieldElement()
		c_2 := newFieldElement()
		add(c_1, a, zero)
		if !c_1.equal(a) {
			t.Fatalf("a + 0 == a")
		}
		sub(c_1, a, zero)
		if !c_1.equal(a) {
			t.Fatalf("a - 0 == a")
		}
		double(c_1, zero)
		if !c_1.equal(zero) {
			t.Fatalf("2 * 0 == 0")
		}
		neg(c_1, zero)
		if !c_1.equal(zero) {
			t.Fatalf("-0 == 0")
		}
		sub(c_1, zero, a)
		neg(c_2, a)
		if !c_1.equal(c_2) {
			t.Fatalf("0 - a == -a")
		}
		double(c_1, a)
		add(c_2, a, a)
		if !c_1.equal(c_2) {
			t.Fatalf("2 * a == a + a")
		}
		add(c_1, a, b)
		add(c_2, b, a)
		if !c_1.equal(c_2) {
			t.Fatalf("a + b = b + a")
		}
		sub(c_1, a, b)
		sub(c_2, b, a)
		neg(c_2, c_2)
		if !c_1.equal(c_2) {
			t.Fatalf("a - b = - ( b - a )")
		}
		c_x, _ := randFieldElement(rand.Reader)
		add(c_1, a, b)
		add(c_1, c_1, c_x)
		add(c_2, a, c_x)
		add(c_2, c_2, b)
		if !c_1.equal(c_2) {
			t.Fatalf("(a + b) + c == (a + c ) + b")
		}
		sub(c_1, a, b)
		sub(c_1, c_1, c_x)
		sub(c_2, a, c_x)
		sub(c_2, c_2, b)
		if !c_1.equal(c_2) {
			t.Fatalf("(a - b) - c == (a - c ) -b")
		}
	}
}

func TestMultiplicationCrossAgainstBigInt(t *testing.T) {
	for i := 0; i < fuz; i++ {
		a, _ := randFieldElement(rand.Reader)
		b, _ := randFieldElement(rand.Reader)
		c := newFieldElement()
		big_a := toBig(a)
		big_b := toBig(b)
		big_c := new(big.Int)
		mul(c, a, b)
		out_1 := toBytes(c)
		out_2 := padBytes(big_c.Mul(big_a, big_b).Mod(big_c, pbig).Bytes(), byteSize)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied")
		}
	}
}

func TestMultiplicationProperties(t *testing.T) {
	for i := 0; i < fuz; i++ {
		a, _ := randFieldElement(rand.Reader)
		b, _ := randFieldElement(rand.Reader)
		c_1 := newFieldElement()
		c_2 := newFieldElement()
		mul(c_1, a, zero)
		if !c_1.equal(zero) {
			t.Fatalf("a * 0 == 0")
		}
		mul(c_1, a, one)
		if !c_1.equal(a) {
			t.Fatalf("a * 1 == a")
		}
		mul(c_1, a, b)
		mul(c_2, b, a)
		if !c_1.equal(c_2) {
			t.Fatalf("a * b == b * a")
		}
		c_x, _ := randFieldElement(rand.Reader)
		mul(c_1, a, b)
		mul(c_1, c_1, c_x)
		mul(c_2, c_x, b)
		mul(c_2, c_2, a)
		if !c_1.equal(c_2) {
			t.Fatalf("(a * b) * c == (a * c) * b")
		}
	}
}

func TestExponentiation(t *testing.T) {
	for i := 0; i < fuz; i++ {
		a, _ := randFieldElement(rand.Reader)
		u := newFieldElement()
		exp(u, a, big.NewInt(0))
		if !u.equal(one) {
			t.Fatalf("a^0 == 1")
		}
		exp(u, a, big.NewInt(1))
		if !u.equal(a) {
			t.Fatalf("a^1 == a")
		}
		v := newFieldElement()
		mul(u, a, a)
		mul(u, u, u)
		mul(u, u, u)
		exp(v, a, big.NewInt(8))
		if !u.equal(v) {
			t.Fatalf("((a^2)^2)^2 == a^8")
		}
		p := new(big.Int).SetBytes(pbig.Bytes())
		exp(u, a, p)
		if !u.equal(a) {
			t.Fatalf("a^p == a")
		}
		exp(u, a, p.Sub(p, big.NewInt(1)))
		if !u.equal(r) {
			t.Fatalf("a^(p-1) == 1")
		}
	}
}

`

const fieldTestNonFixedModulus = `
import (
	"bytes"
	"crypto/rand"
	"flag"
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

func BenchmarkField(t *testing.B) {
	field := randField()
	in_a := randBytes(field.pbig)
	in_b := randBytes(field.pbig)
	a, _ := field.newFieldElementFromBytes(in_a)
	b, _ := field.newFieldElementFromBytes(in_b)
	c := field.newFieldElement()
	t.Run("add", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.add(c, a, b)
		}
	})
	t.Run("double", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.double(c, a)
		}
	})
	t.Run("sub", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.sub(c, a, b)
		}
	})
	t.Run("mul", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.mul(c, a, b)
		}
	})
	t.Run("cmp", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			a.equal(b)
		}
	})
}

func TestCompare(t *testing.T) {
	field := randField()
	r := field.r
	modulus := field.p
	zero := field.zero
	if r.cmp(r) != 0 {
		t.Fatalf("r == r (cmp)")
	}
	if !r.equal(r) {
		t.Fatalf("r == r (equal)")
	}
	if r.equal(modulus) {
		t.Fatalf("p != r")
	}
	if r.equal(zero) {
		t.Fatalf("r != 0")
	}
	if !zero.equal(zero) {
		t.Fatalf("0 == 0")
	}
	if modulus.cmp(r) != 1 {
		t.Fatalf("p > r")
	}
	if r.cmp(modulus) != -1 {
		t.Fatalf("r < p")
	}
}

func TestSerialization(t *testing.T) {
	field := randField()
	// demont(r) == 1
	b0 := make([]byte, byteSize)
	b0[len(b0)-1] = byte(1)
	b1 := field.toBytes(field.r)
	if !bytes.Equal(b0, b1) {
		t.Fatalf("demont(r) must be equal to 1\n")
	}
	// is a => modulus should not be valid
	_, err := field.newFieldElementFromBytes(field.pbig.Bytes())
	if err == nil {
		t.Fatalf("a number eq or larger than modulus must not be valid")
	}
	for i := 0; i < fuz; i++ {
		field := randField()
		// bytes
		b0 := randBytes(field.pbig)
		a0, err := field.newFieldElementFromBytes(b0)
		if err != nil {
			t.Fatal(err)
		}
		b1 = field.toBytes(a0)
		if !bytes.Equal(b0, b1) {
			t.Fatalf("bad serialization (bytes)")
		}
		// string
		s := field.toString(a0)
		a1, err := field.newFieldElementFromString(s)
		if err != nil {
			t.Fatal(err)
		}
		if !a0.equal(a1) {
			t.Fatalf("bad serialization (str)")
		}
		// big int
		a0, err = field.newFieldElementFromBytes(b0)
		if err != nil {
			t.Fatal(err)
		}
		bi := field.toBig(a0)
		a1, err = field.newFieldElementFromBig(bi)
		if err != nil {
			t.Fatal(err)
		}
		if !a0.equal(a1) {
			t.Fatalf("bad serialization (big.Int)")
		}
	}
}

func TestAdditionCrossAgainstBigInt(t *testing.T) {
	for i := 0; i < fuz; i++ {
		field := randField()
		a, _ := field.randFieldElement(rand.Reader)
		b, _ := field.randFieldElement(rand.Reader)
		c := field.newFieldElement()
		big_a := field.toBig(a)
		big_b := field.toBig(b)
		big_c := new(big.Int)
		field.add(c, a, b)
		out_1 := field.toBytes(c)
		out_2 := padBytes(big_c.Add(big_a, big_b).Mod(big_c, field.pbig).Bytes(), byteSize)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied A")
		}
		field.double(c, a)
		out_1 = field.toBytes(c)
		out_2 = padBytes(big_c.Add(big_a, big_a).Mod(big_c, field.pbig).Bytes(), byteSize)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied B")
		}
		field.sub(c, a, b)
		out_1 = field.toBytes(c)
		out_2 = padBytes(big_c.Sub(big_a, big_b).Mod(big_c, field.pbig).Bytes(), byteSize)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied C")
		}
		field.neg(c, a)
		out_1 = field.toBytes(c)
		out_2 = padBytes(big_c.Neg(big_a).Mod(big_c, field.pbig).Bytes(), byteSize)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied D")
		}
	}
}

func TestAdditionProperties(t *testing.T) {
	for i := 0; i < fuz; i++ {
		field := randField()
		zero := field.zero
		a, _ := field.randFieldElement(rand.Reader)
		b, _ := field.randFieldElement(rand.Reader)
		c_1 := field.newFieldElement()
		c_2 := field.newFieldElement()
		field.add(c_1, a, zero)
		if !c_1.equal(a) {
			t.Fatalf("a + 0 == a")
		}
		field.sub(c_1, a, zero)
		if !c_1.equal(a) {
			t.Fatalf("a - 0 == a")
		}
		field.double(c_1, zero)
		if !c_1.equal(zero) {
			t.Fatalf("2 * 0 == 0")
		}
		field.neg(c_1, zero)
		if !c_1.equal(zero) {
			t.Fatalf("-0 == 0")
		}
		field.sub(c_1, zero, a)
		field.neg(c_2, a)
		if !c_1.equal(c_2) {
			t.Fatalf("0 - a == -a")
		}
		field.double(c_1, a)
		field.add(c_2, a, a)
		if !c_1.equal(c_2) {
			t.Fatalf("2 * a == a + a")
		}
		field.add(c_1, a, b)
		field.add(c_2, b, a)
		if !c_1.equal(c_2) {
			t.Fatalf("a + b = b + a")
		}
		field.sub(c_1, a, b)
		field.sub(c_2, b, a)
		field.neg(c_2, c_2)
		if !c_1.equal(c_2) {
			t.Fatalf("a - b = - ( b - a )")
		}
		c_x, _ := field.randFieldElement(rand.Reader)
		field.add(c_1, a, b)
		field.add(c_1, c_1, c_x)
		field.add(c_2, a, c_x)
		field.add(c_2, c_2, b)
		if !c_1.equal(c_2) {
			t.Fatalf("(a + b) + c == (a + c ) + b")
		}
		field.sub(c_1, a, b)
		field.sub(c_1, c_1, c_x)
		field.sub(c_2, a, c_x)
		field.sub(c_2, c_2, b)
		if !c_1.equal(c_2) {
			t.Fatalf("(a - b) - c == (a - c ) -b")
		}
	}
}

func TestMultiplicationCrossAgainstBigInt(t *testing.T) {
	for i := 0; i < fuz; i++ {
		field := randField()
		a, _ := field.randFieldElement(rand.Reader)
		b, _ := field.randFieldElement(rand.Reader)
		c := field.newFieldElement()
		big_a := field.toBig(a)
		big_b := field.toBig(b)
		big_c := new(big.Int)
		field.mul(c, a, b)
		out_1 := field.toBytes(c)
		out_2 := padBytes(big_c.Mul(big_a, big_b).Mod(big_c, field.pbig).Bytes(), byteSize)
		if !bytes.Equal(out_1, out_2) {
			t.Fatalf("cross test against big.Int is not satisfied")
		}
	}
}

func TestMultiplicationProperties(t *testing.T) {
	for i := 0; i < fuz; i++ {
		field := randField()
		zero := field.zero
		one := field.one
		a, _ := field.randFieldElement(rand.Reader)
		b, _ := field.randFieldElement(rand.Reader)
		c_1 := field.newFieldElement()
		c_2 := field.newFieldElement()
		field.mul(c_1, a, zero)
		if !c_1.equal(zero) {
			t.Fatalf("a * 0 == 0")
		}
		field.mul(c_1, a, one)
		if !c_1.equal(a) {
			t.Fatalf("a * 1 == a")
		}
		field.mul(c_1, a, b)
		field.mul(c_2, b, a)
		if !c_1.equal(c_2) {
			t.Fatalf("a * b == b * a")
		}
		c_x, _ := field.randFieldElement(rand.Reader)
		field.mul(c_1, a, b)
		field.mul(c_1, c_1, c_x)
		field.mul(c_2, c_x, b)
		field.mul(c_2, c_2, a)
		if !c_1.equal(c_2) {
			t.Fatalf("(a * b) * c == (a * c) * b")
		}
	}
}

func TestExponentiation(t *testing.T) {
	for i := 0; i < fuz; i++ {
		field := randField()
		a, _ := field.randFieldElement(rand.Reader)
		u := field.newFieldElement()
		field.exp(u, a, big.NewInt(0))
		if !u.equal(field.one) {
			t.Fatalf("a^0 == 1")
		}
		field.exp(u, a, big.NewInt(1))
		if !u.equal(a) {
			t.Fatalf("a^1 == a")
		}
		v := field.newFieldElement()
		field.mul(u, a, a)
		field.mul(u, u, u)
		field.mul(u, u, u)
		field.exp(v, a, big.NewInt(8))
		if !u.equal(v) {
			t.Fatalf("((a^2)^2)^2 == a^8")
		}
		p := new(big.Int).SetBytes(field.pbig.Bytes())
		field.exp(u, a, p)
		if !u.equal(a) {
			t.Fatalf("a^p == a")
		}
		field.exp(u, a, p.Sub(p, big.NewInt(1)))
		if !u.equal(field.r) {
			t.Fatalf("a^(p-1) == 1")
		}
	}
}
`
