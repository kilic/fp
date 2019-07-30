package fp

import (
	"crypto/rand"
	"flag"
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
func TestFieldElement256(t *testing.T) {

	t.Run("Encoding & Decoding", func(t *testing.T) {
		var field *Field256
		for i := 0; i < n; i++ {
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
		}
		t.Run("1", func(t *testing.T) {
			bytes := []byte{
				0}
			if !new(Fe256).Unmarshal(bytes).Equals(&Fe256{0}) {
				t.Errorf("bad encoding\n")
			}
		})
		t.Run("2", func(t *testing.T) {
			bytes := []byte{
				254, 253}
			if new(Fe256).Unmarshal(bytes).Equals(&Fe256{0xfe, 0xfd}) {
				t.Errorf("bad encoding\n")
			}
		})
		t.Run("3", func(t *testing.T) {
			var a, b Fe256
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			bytes := make([]byte, 4*8)
			a.Marshal(bytes[:])
			b.Unmarshal(bytes[:])
			if !a.Equals(&b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
		t.Run("4", func(t *testing.T) {
			var a Fe256
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			b, er1 := new(Fe256).SetString(a.String())
			if er1 != nil {
				t.Fatal(er1)
			}
			if !a.Equals(b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
		t.Run("5", func(t *testing.T) {
			var a Fe256
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			b := new(Fe256).SetBig(a.Big())
			if er1 != nil {
				t.Fatal(er1)
			}
			if !a.Equals(b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
	})

	t.Run("Addition", func(t *testing.T) {
		var a, b, c, u, v Fe256
		zero := new(Fe256).SetUint(0)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Add(&u, &a, &b)
			field.Add(&u, &u, &c)
			field.Add(&v, &b, &c)
			field.Add(&v, &v, &a)
			if !u.Equals(&v) {
				t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)
			}
			field.Add(&u, &a, &b)
			field.Add(&v, &b, &a)
			if !u.Equals(&v) {
				t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv:%s\n", a, b, u, v)
			}
			field.Add(&u, &a, zero)
			if !u.Equals(&a) {
				t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)
			}
			field.Neg(&u, &a)
			field.Add(&u, &u, &a)
			if !u.Equals(zero) {
				t.Fatalf("Bad Negation\na:%s", a.String())
			}
		}
	})

	t.Run("Subtraction", func(t *testing.T) {
		var a, b, c, u, v Fe256
		zero := new(Fe256).SetUint(0)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Sub(&u, &a, &c)
			field.Sub(&u, &u, &b)
			field.Sub(&v, &a, &b)
			field.Sub(&v, &v, &c)
			if !u.Equals(&v) {
				t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)
			}
			field.Sub(&u, &a, zero)
			if !u.Equals(&a) {
				t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)
			}
			field.Sub(&u, &a, &b)
			field.Sub(&v, &b, &a)
			field.Add(&u, &u, &v)
			if !u.Equals(zero) {
				t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv: %s", a, b, u, v)
			}
			field.Sub(&u, &a, &b)
			field.Sub(&v, &b, &a)
			field.Neg(&v, &v)
			if !u.Equals(&u) {
				t.Fatalf("Bad Negation\na:%s", a.String())
			}
		}
	})

	t.Run("Doubling", func(t *testing.T) {
		var a, u, v Fe256
		for i := 0; i < n; i++ {
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
			err := field.RandElement(&a, rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			field.Double(&u, &a)
			field.Add(&v, &a, &a)
			if !u.Equals(&v) {
				t.Fatalf("Bad doubling\na: %s\nu: %s\nv: %s\n", a, u, v)
			}
		}
	})

	t.Run("Montgomerry", func(t *testing.T) {
		var a, b, c, u, v, w Fe256
		zero := new(Fe256).SetUint(0)
		one := new(Fe256).SetUint(1)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Mont(&u, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Demont(&u, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mont(&u, one)
			if !u.Equals(field.r1) {
				t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Demont(&u, field.r1)
			if !u.Equals(one) {
				t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mul(&u, &a, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad zero element\na: %s\nu: %s\np: %s\n", a, u, field.P)
			}
			field.Mul(&u, &a, one)
			field.Mul(&u, &u, field.r2)
			if !u.Equals(&a) {
				t.Fatalf("Multiplication identity does not hold, expected to equal itself\nu: %s\np: %s\n", u, field.P)
			}
			field.Mul(&u, field.r2, one)
			if !u.Equals(field.r1) {
				t.Fatalf("Multiplication identity does not hold, expected to equal r1\nu: %s\np: %s\n", u, field.P)
			}
			field.Mul(&u, &a, &b)
			field.Mul(&u, &u, &c)
			field.Mul(&v, &b, &c)
			field.Mul(&v, &v, &a)
			if !u.Equals(&v) {
				t.Fatalf("Multiplicative associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.P)
			}
			field.Add(&u, &a, &b)
			field.Mul(&u, &c, &u)
			field.Mul(&w, &a, &c)
			field.Mul(&v, &b, &c)
			field.Add(&v, &v, &w)
			if !u.Equals(&v) {
				t.Fatalf("Distributivity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.P)
			}
		}
	})

	t.Run("Exponentiation", func(t *testing.T) {
		var a, u, v Fe256
		bytes := make([]byte, 4*8)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			field.Exp(&u, &a, big.NewInt(0))
			if !u.Equals(field.r1) {
				t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			field.Exp(&u, &a, big.NewInt(1))
			if !u.Equals(&a) {
				t.Fatalf("Bad exponentiation, expected to equal a\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			field.Mul(&u, &a, &a)
			field.Mul(&u, &u, &u)
			field.Mul(&u, &u, &u)
			field.Exp(&v, &a, big.NewInt(8))
			if !u.Equals(&v) {
				t.Fatalf("Bad exponentiation\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			p := new(big.Int).SetBytes(field.P.Marshal(bytes))
			field.Exp(&u, &a, p)
			if !u.Equals(&a) {
				t.Fatalf("Bad exponentiation, expected to equal itself\nu: %s\na: %s\np: %s\n", u, a, field.P)
			}
			field.Exp(&u, &a, p.Sub(p, big.NewInt(1)))
			if !u.Equals(field.r1) {
				t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\na: %s\nr1: %s\np: %s\n", u, a, field.r1, field.P)
			}
		}
	})

	t.Run("Inversion", func(t *testing.T) {
		var u, a, v Fe256
		one := new(Fe256).SetUint(1)
		bytes := make([]byte, 4*8)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			field.InvMontUp(&u, &a)
			field.Mul(&u, &u, &a)
			if !u.Equals(field.r1) {
				t.Fatalf("Bad inversion, expected to equal r1\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mont(&u, &a)
			field.InvMontDown(&v, &u)
			field.Mul(&v, &v, &u)
			if !v.Equals(one) {
				t.Fatalf("Bad inversion, expected to equal 1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			p := new(big.Int).SetBytes(field.P.Marshal(bytes))
			field.Exp(&u, &a, p.Sub(p, big.NewInt(2)))
			field.InvMontUp(&v, &a)
			if !v.Equals(&u) {
				t.Fatalf("Bad inversion")
			}
			field.InvEEA(&u, &a)
			field.Mul(&u, &u, &a)
			field.Mul(&u, &u, field.r2)
			if !u.Equals(one) {
				t.Fatalf("Bad inversion")
			}
		}
	})

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
	er1 := field.RandElement(&a, rand.Reader)
	er2 := field.RandElement(&b, rand.Reader)
	if er1 != nil || er2 != nil {
		t.Fatal(er1, er2)
	}
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
		bytes := make([]byte, 4*8)
		e := new(big.Int).SetBytes(field.P.Marshal(bytes))
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Exp(&c, &a, e)
		}
	})
}

func TestFieldElement320(t *testing.T) {

	t.Run("Encoding & Decoding", func(t *testing.T) {
		var field *Field320
		for i := 0; i < n; i++ {
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
		}
		t.Run("1", func(t *testing.T) {
			bytes := []byte{
				0}
			if !new(Fe320).Unmarshal(bytes).Equals(&Fe320{0}) {
				t.Errorf("bad encoding\n")
			}
		})
		t.Run("2", func(t *testing.T) {
			bytes := []byte{
				254, 253}
			if new(Fe320).Unmarshal(bytes).Equals(&Fe320{0xfe, 0xfd}) {
				t.Errorf("bad encoding\n")
			}
		})
		t.Run("3", func(t *testing.T) {
			var a, b Fe320
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			bytes := make([]byte, 5*8)
			a.Marshal(bytes[:])
			b.Unmarshal(bytes[:])
			if !a.Equals(&b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
		t.Run("4", func(t *testing.T) {
			var a Fe320
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			b, er1 := new(Fe320).SetString(a.String())
			if er1 != nil {
				t.Fatal(er1)
			}
			if !a.Equals(b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
		t.Run("5", func(t *testing.T) {
			var a Fe320
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			b := new(Fe320).SetBig(a.Big())
			if er1 != nil {
				t.Fatal(er1)
			}
			if !a.Equals(b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
	})

	t.Run("Addition", func(t *testing.T) {
		var a, b, c, u, v Fe320
		zero := new(Fe320).SetUint(0)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Add(&u, &a, &b)
			field.Add(&u, &u, &c)
			field.Add(&v, &b, &c)
			field.Add(&v, &v, &a)
			if !u.Equals(&v) {
				t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)
			}
			field.Add(&u, &a, &b)
			field.Add(&v, &b, &a)
			if !u.Equals(&v) {
				t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv:%s\n", a, b, u, v)
			}
			field.Add(&u, &a, zero)
			if !u.Equals(&a) {
				t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)
			}
			field.Neg(&u, &a)
			field.Add(&u, &u, &a)
			if !u.Equals(zero) {
				t.Fatalf("Bad Negation\na:%s", a.String())
			}
		}
	})

	t.Run("Subtraction", func(t *testing.T) {
		var a, b, c, u, v Fe320
		zero := new(Fe320).SetUint(0)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Sub(&u, &a, &c)
			field.Sub(&u, &u, &b)
			field.Sub(&v, &a, &b)
			field.Sub(&v, &v, &c)
			if !u.Equals(&v) {
				t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)
			}
			field.Sub(&u, &a, zero)
			if !u.Equals(&a) {
				t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)
			}
			field.Sub(&u, &a, &b)
			field.Sub(&v, &b, &a)
			field.Add(&u, &u, &v)
			if !u.Equals(zero) {
				t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv: %s", a, b, u, v)
			}
			field.Sub(&u, &a, &b)
			field.Sub(&v, &b, &a)
			field.Neg(&v, &v)
			if !u.Equals(&u) {
				t.Fatalf("Bad Negation\na:%s", a.String())
			}
		}
	})

	t.Run("Doubling", func(t *testing.T) {
		var a, u, v Fe320
		for i := 0; i < n; i++ {
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
			err := field.RandElement(&a, rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			field.Double(&u, &a)
			field.Add(&v, &a, &a)
			if !u.Equals(&v) {
				t.Fatalf("Bad doubling\na: %s\nu: %s\nv: %s\n", a, u, v)
			}
		}
	})

	t.Run("Montgomerry", func(t *testing.T) {
		var a, b, c, u, v, w Fe320
		zero := new(Fe320).SetUint(0)
		one := new(Fe320).SetUint(1)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Mont(&u, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Demont(&u, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mont(&u, one)
			if !u.Equals(field.r1) {
				t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Demont(&u, field.r1)
			if !u.Equals(one) {
				t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mul(&u, &a, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad zero element\na: %s\nu: %s\np: %s\n", a, u, field.P)
			}
			field.Mul(&u, &a, one)
			field.Mul(&u, &u, field.r2)
			if !u.Equals(&a) {
				t.Fatalf("Multiplication identity does not hold, expected to equal itself\nu: %s\np: %s\n", u, field.P)
			}
			field.Mul(&u, field.r2, one)
			if !u.Equals(field.r1) {
				t.Fatalf("Multiplication identity does not hold, expected to equal r1\nu: %s\np: %s\n", u, field.P)
			}
			field.Mul(&u, &a, &b)
			field.Mul(&u, &u, &c)
			field.Mul(&v, &b, &c)
			field.Mul(&v, &v, &a)
			if !u.Equals(&v) {
				t.Fatalf("Multiplicative associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.P)
			}
			field.Add(&u, &a, &b)
			field.Mul(&u, &c, &u)
			field.Mul(&w, &a, &c)
			field.Mul(&v, &b, &c)
			field.Add(&v, &v, &w)
			if !u.Equals(&v) {
				t.Fatalf("Distributivity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.P)
			}
		}
	})

	t.Run("Exponentiation", func(t *testing.T) {
		var a, u, v Fe320
		bytes := make([]byte, 5*8)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			field.Exp(&u, &a, big.NewInt(0))
			if !u.Equals(field.r1) {
				t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			field.Exp(&u, &a, big.NewInt(1))
			if !u.Equals(&a) {
				t.Fatalf("Bad exponentiation, expected to equal a\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			field.Mul(&u, &a, &a)
			field.Mul(&u, &u, &u)
			field.Mul(&u, &u, &u)
			field.Exp(&v, &a, big.NewInt(8))
			if !u.Equals(&v) {
				t.Fatalf("Bad exponentiation\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			p := new(big.Int).SetBytes(field.P.Marshal(bytes))
			field.Exp(&u, &a, p)
			if !u.Equals(&a) {
				t.Fatalf("Bad exponentiation, expected to equal itself\nu: %s\na: %s\np: %s\n", u, a, field.P)
			}
			field.Exp(&u, &a, p.Sub(p, big.NewInt(1)))
			if !u.Equals(field.r1) {
				t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\na: %s\nr1: %s\np: %s\n", u, a, field.r1, field.P)
			}
		}
	})

	t.Run("Inversion", func(t *testing.T) {
		var u, a, v Fe320
		one := new(Fe320).SetUint(1)
		bytes := make([]byte, 5*8)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			field.InvMontUp(&u, &a)
			field.Mul(&u, &u, &a)
			if !u.Equals(field.r1) {
				t.Fatalf("Bad inversion, expected to equal r1\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mont(&u, &a)
			field.InvMontDown(&v, &u)
			field.Mul(&v, &v, &u)
			if !v.Equals(one) {
				t.Fatalf("Bad inversion, expected to equal 1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			p := new(big.Int).SetBytes(field.P.Marshal(bytes))
			field.Exp(&u, &a, p.Sub(p, big.NewInt(2)))
			field.InvMontUp(&v, &a)
			if !v.Equals(&u) {
				t.Fatalf("Bad inversion")
			}
			field.InvEEA(&u, &a)
			field.Mul(&u, &u, &a)
			field.Mul(&u, &u, field.r2)
			if !u.Equals(one) {
				t.Fatalf("Bad inversion")
			}
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
	er1 := field.RandElement(&a, rand.Reader)
	er2 := field.RandElement(&b, rand.Reader)
	if er1 != nil || er2 != nil {
		t.Fatal(er1, er2)
	}
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
		bytes := make([]byte, 5*8)
		e := new(big.Int).SetBytes(field.P.Marshal(bytes))
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Exp(&c, &a, e)
		}
	})
}

func TestFieldElement384(t *testing.T) {

	t.Run("Encoding & Decoding", func(t *testing.T) {
		var field *Field384
		for i := 0; i < n; i++ {
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
		}
		t.Run("1", func(t *testing.T) {
			bytes := []byte{
				0}
			if !new(Fe384).Unmarshal(bytes).Equals(&Fe384{0}) {
				t.Errorf("bad encoding\n")
			}
		})
		t.Run("2", func(t *testing.T) {
			bytes := []byte{
				254, 253}
			if new(Fe384).Unmarshal(bytes).Equals(&Fe384{0xfe, 0xfd}) {
				t.Errorf("bad encoding\n")
			}
		})
		t.Run("3", func(t *testing.T) {
			var a, b Fe384
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			bytes := make([]byte, 6*8)
			a.Marshal(bytes[:])
			b.Unmarshal(bytes[:])
			if !a.Equals(&b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
		t.Run("4", func(t *testing.T) {
			var a Fe384
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			b, er1 := new(Fe384).SetString(a.String())
			if er1 != nil {
				t.Fatal(er1)
			}
			if !a.Equals(b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
		t.Run("5", func(t *testing.T) {
			var a Fe384
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			b := new(Fe384).SetBig(a.Big())
			if er1 != nil {
				t.Fatal(er1)
			}
			if !a.Equals(b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
	})

	t.Run("Addition", func(t *testing.T) {
		var a, b, c, u, v Fe384
		zero := new(Fe384).SetUint(0)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Add(&u, &a, &b)
			field.Add(&u, &u, &c)
			field.Add(&v, &b, &c)
			field.Add(&v, &v, &a)
			if !u.Equals(&v) {
				t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)
			}
			field.Add(&u, &a, &b)
			field.Add(&v, &b, &a)
			if !u.Equals(&v) {
				t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv:%s\n", a, b, u, v)
			}
			field.Add(&u, &a, zero)
			if !u.Equals(&a) {
				t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)
			}
			field.Neg(&u, &a)
			field.Add(&u, &u, &a)
			if !u.Equals(zero) {
				t.Fatalf("Bad Negation\na:%s", a.String())
			}
		}
	})

	t.Run("Subtraction", func(t *testing.T) {
		var a, b, c, u, v Fe384
		zero := new(Fe384).SetUint(0)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Sub(&u, &a, &c)
			field.Sub(&u, &u, &b)
			field.Sub(&v, &a, &b)
			field.Sub(&v, &v, &c)
			if !u.Equals(&v) {
				t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)
			}
			field.Sub(&u, &a, zero)
			if !u.Equals(&a) {
				t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)
			}
			field.Sub(&u, &a, &b)
			field.Sub(&v, &b, &a)
			field.Add(&u, &u, &v)
			if !u.Equals(zero) {
				t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv: %s", a, b, u, v)
			}
			field.Sub(&u, &a, &b)
			field.Sub(&v, &b, &a)
			field.Neg(&v, &v)
			if !u.Equals(&u) {
				t.Fatalf("Bad Negation\na:%s", a.String())
			}
		}
	})

	t.Run("Doubling", func(t *testing.T) {
		var a, u, v Fe384
		for i := 0; i < n; i++ {
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
			err := field.RandElement(&a, rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			field.Double(&u, &a)
			field.Add(&v, &a, &a)
			if !u.Equals(&v) {
				t.Fatalf("Bad doubling\na: %s\nu: %s\nv: %s\n", a, u, v)
			}
		}
	})

	t.Run("Montgomerry", func(t *testing.T) {
		var a, b, c, u, v, w Fe384
		zero := new(Fe384).SetUint(0)
		one := new(Fe384).SetUint(1)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Mont(&u, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Demont(&u, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mont(&u, one)
			if !u.Equals(field.r1) {
				t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Demont(&u, field.r1)
			if !u.Equals(one) {
				t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mul(&u, &a, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad zero element\na: %s\nu: %s\np: %s\n", a, u, field.P)
			}
			field.Mul(&u, &a, one)
			field.Mul(&u, &u, field.r2)
			if !u.Equals(&a) {
				t.Fatalf("Multiplication identity does not hold, expected to equal itself\nu: %s\np: %s\n", u, field.P)
			}
			field.Mul(&u, field.r2, one)
			if !u.Equals(field.r1) {
				t.Fatalf("Multiplication identity does not hold, expected to equal r1\nu: %s\np: %s\n", u, field.P)
			}
			field.Mul(&u, &a, &b)
			field.Mul(&u, &u, &c)
			field.Mul(&v, &b, &c)
			field.Mul(&v, &v, &a)
			if !u.Equals(&v) {
				t.Fatalf("Multiplicative associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.P)
			}
			field.Add(&u, &a, &b)
			field.Mul(&u, &c, &u)
			field.Mul(&w, &a, &c)
			field.Mul(&v, &b, &c)
			field.Add(&v, &v, &w)
			if !u.Equals(&v) {
				t.Fatalf("Distributivity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.P)
			}
		}
	})

	t.Run("Exponentiation", func(t *testing.T) {
		var a, u, v Fe384
		bytes := make([]byte, 6*8)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			field.Exp(&u, &a, big.NewInt(0))
			if !u.Equals(field.r1) {
				t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			field.Exp(&u, &a, big.NewInt(1))
			if !u.Equals(&a) {
				t.Fatalf("Bad exponentiation, expected to equal a\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			field.Mul(&u, &a, &a)
			field.Mul(&u, &u, &u)
			field.Mul(&u, &u, &u)
			field.Exp(&v, &a, big.NewInt(8))
			if !u.Equals(&v) {
				t.Fatalf("Bad exponentiation\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			p := new(big.Int).SetBytes(field.P.Marshal(bytes))
			field.Exp(&u, &a, p)
			if !u.Equals(&a) {
				t.Fatalf("Bad exponentiation, expected to equal itself\nu: %s\na: %s\np: %s\n", u, a, field.P)
			}
			field.Exp(&u, &a, p.Sub(p, big.NewInt(1)))
			if !u.Equals(field.r1) {
				t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\na: %s\nr1: %s\np: %s\n", u, a, field.r1, field.P)
			}
		}
	})

	t.Run("Inversion", func(t *testing.T) {
		var u, a, v Fe384
		one := new(Fe384).SetUint(1)
		bytes := make([]byte, 6*8)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			field.InvMontUp(&u, &a)
			field.Mul(&u, &u, &a)
			if !u.Equals(field.r1) {
				t.Fatalf("Bad inversion, expected to equal r1\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mont(&u, &a)
			field.InvMontDown(&v, &u)
			field.Mul(&v, &v, &u)
			if !v.Equals(one) {
				t.Fatalf("Bad inversion, expected to equal 1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			p := new(big.Int).SetBytes(field.P.Marshal(bytes))
			field.Exp(&u, &a, p.Sub(p, big.NewInt(2)))
			field.InvMontUp(&v, &a)
			if !v.Equals(&u) {
				t.Fatalf("Bad inversion")
			}
			field.InvEEA(&u, &a)
			field.Mul(&u, &u, &a)
			field.Mul(&u, &u, field.r2)
			if !u.Equals(one) {
				t.Fatalf("Bad inversion")
			}
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
	er1 := field.RandElement(&a, rand.Reader)
	er2 := field.RandElement(&b, rand.Reader)
	if er1 != nil || er2 != nil {
		t.Fatal(er1, er2)
	}
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
		bytes := make([]byte, 6*8)
		e := new(big.Int).SetBytes(field.P.Marshal(bytes))
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Exp(&c, &a, e)
		}
	})
}

func TestFieldElement448(t *testing.T) {

	t.Run("Encoding & Decoding", func(t *testing.T) {
		var field *Field448
		for i := 0; i < n; i++ {
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
		}
		t.Run("1", func(t *testing.T) {
			bytes := []byte{
				0}
			if !new(Fe448).Unmarshal(bytes).Equals(&Fe448{0}) {
				t.Errorf("bad encoding\n")
			}
		})
		t.Run("2", func(t *testing.T) {
			bytes := []byte{
				254, 253}
			if new(Fe448).Unmarshal(bytes).Equals(&Fe448{0xfe, 0xfd}) {
				t.Errorf("bad encoding\n")
			}
		})
		t.Run("3", func(t *testing.T) {
			var a, b Fe448
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			bytes := make([]byte, 7*8)
			a.Marshal(bytes[:])
			b.Unmarshal(bytes[:])
			if !a.Equals(&b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
		t.Run("4", func(t *testing.T) {
			var a Fe448
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			b, er1 := new(Fe448).SetString(a.String())
			if er1 != nil {
				t.Fatal(er1)
			}
			if !a.Equals(b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
		t.Run("5", func(t *testing.T) {
			var a Fe448
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			b := new(Fe448).SetBig(a.Big())
			if er1 != nil {
				t.Fatal(er1)
			}
			if !a.Equals(b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
	})

	t.Run("Addition", func(t *testing.T) {
		var a, b, c, u, v Fe448
		zero := new(Fe448).SetUint(0)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Add(&u, &a, &b)
			field.Add(&u, &u, &c)
			field.Add(&v, &b, &c)
			field.Add(&v, &v, &a)
			if !u.Equals(&v) {
				t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)
			}
			field.Add(&u, &a, &b)
			field.Add(&v, &b, &a)
			if !u.Equals(&v) {
				t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv:%s\n", a, b, u, v)
			}
			field.Add(&u, &a, zero)
			if !u.Equals(&a) {
				t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)
			}
			field.Neg(&u, &a)
			field.Add(&u, &u, &a)
			if !u.Equals(zero) {
				t.Fatalf("Bad Negation\na:%s", a.String())
			}
		}
	})

	t.Run("Subtraction", func(t *testing.T) {
		var a, b, c, u, v Fe448
		zero := new(Fe448).SetUint(0)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Sub(&u, &a, &c)
			field.Sub(&u, &u, &b)
			field.Sub(&v, &a, &b)
			field.Sub(&v, &v, &c)
			if !u.Equals(&v) {
				t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)
			}
			field.Sub(&u, &a, zero)
			if !u.Equals(&a) {
				t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)
			}
			field.Sub(&u, &a, &b)
			field.Sub(&v, &b, &a)
			field.Add(&u, &u, &v)
			if !u.Equals(zero) {
				t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv: %s", a, b, u, v)
			}
			field.Sub(&u, &a, &b)
			field.Sub(&v, &b, &a)
			field.Neg(&v, &v)
			if !u.Equals(&u) {
				t.Fatalf("Bad Negation\na:%s", a.String())
			}
		}
	})

	t.Run("Doubling", func(t *testing.T) {
		var a, u, v Fe448
		for i := 0; i < n; i++ {
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
			err := field.RandElement(&a, rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			field.Double(&u, &a)
			field.Add(&v, &a, &a)
			if !u.Equals(&v) {
				t.Fatalf("Bad doubling\na: %s\nu: %s\nv: %s\n", a, u, v)
			}
		}
	})

	t.Run("Montgomerry", func(t *testing.T) {
		var a, b, c, u, v, w Fe448
		zero := new(Fe448).SetUint(0)
		one := new(Fe448).SetUint(1)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Mont(&u, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Demont(&u, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mont(&u, one)
			if !u.Equals(field.r1) {
				t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Demont(&u, field.r1)
			if !u.Equals(one) {
				t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mul(&u, &a, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad zero element\na: %s\nu: %s\np: %s\n", a, u, field.P)
			}
			field.Mul(&u, &a, one)
			field.Mul(&u, &u, field.r2)
			if !u.Equals(&a) {
				t.Fatalf("Multiplication identity does not hold, expected to equal itself\nu: %s\np: %s\n", u, field.P)
			}
			field.Mul(&u, field.r2, one)
			if !u.Equals(field.r1) {
				t.Fatalf("Multiplication identity does not hold, expected to equal r1\nu: %s\np: %s\n", u, field.P)
			}
			field.Mul(&u, &a, &b)
			field.Mul(&u, &u, &c)
			field.Mul(&v, &b, &c)
			field.Mul(&v, &v, &a)
			if !u.Equals(&v) {
				t.Fatalf("Multiplicative associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.P)
			}
			field.Add(&u, &a, &b)
			field.Mul(&u, &c, &u)
			field.Mul(&w, &a, &c)
			field.Mul(&v, &b, &c)
			field.Add(&v, &v, &w)
			if !u.Equals(&v) {
				t.Fatalf("Distributivity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.P)
			}
		}
	})

	t.Run("Exponentiation", func(t *testing.T) {
		var a, u, v Fe448
		bytes := make([]byte, 7*8)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			field.Exp(&u, &a, big.NewInt(0))
			if !u.Equals(field.r1) {
				t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			field.Exp(&u, &a, big.NewInt(1))
			if !u.Equals(&a) {
				t.Fatalf("Bad exponentiation, expected to equal a\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			field.Mul(&u, &a, &a)
			field.Mul(&u, &u, &u)
			field.Mul(&u, &u, &u)
			field.Exp(&v, &a, big.NewInt(8))
			if !u.Equals(&v) {
				t.Fatalf("Bad exponentiation\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			p := new(big.Int).SetBytes(field.P.Marshal(bytes))
			field.Exp(&u, &a, p)
			if !u.Equals(&a) {
				t.Fatalf("Bad exponentiation, expected to equal itself\nu: %s\na: %s\np: %s\n", u, a, field.P)
			}
			field.Exp(&u, &a, p.Sub(p, big.NewInt(1)))
			if !u.Equals(field.r1) {
				t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\na: %s\nr1: %s\np: %s\n", u, a, field.r1, field.P)
			}
		}
	})

	t.Run("Inversion", func(t *testing.T) {
		var u, a, v Fe448
		one := new(Fe448).SetUint(1)
		bytes := make([]byte, 7*8)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			field.InvMontUp(&u, &a)
			field.Mul(&u, &u, &a)
			if !u.Equals(field.r1) {
				t.Fatalf("Bad inversion, expected to equal r1\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mont(&u, &a)
			field.InvMontDown(&v, &u)
			field.Mul(&v, &v, &u)
			if !v.Equals(one) {
				t.Fatalf("Bad inversion, expected to equal 1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			p := new(big.Int).SetBytes(field.P.Marshal(bytes))
			field.Exp(&u, &a, p.Sub(p, big.NewInt(2)))
			field.InvMontUp(&v, &a)
			if !v.Equals(&u) {
				t.Fatalf("Bad inversion")
			}
			field.InvEEA(&u, &a)
			field.Mul(&u, &u, &a)
			field.Mul(&u, &u, field.r2)
			if !u.Equals(one) {
				t.Fatalf("Bad inversion")
			}
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
	er1 := field.RandElement(&a, rand.Reader)
	er2 := field.RandElement(&b, rand.Reader)
	if er1 != nil || er2 != nil {
		t.Fatal(er1, er2)
	}
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
		bytes := make([]byte, 7*8)
		e := new(big.Int).SetBytes(field.P.Marshal(bytes))
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Exp(&c, &a, e)
		}
	})
}

func TestFieldElement512(t *testing.T) {

	t.Run("Encoding & Decoding", func(t *testing.T) {
		var field *Field512
		for i := 0; i < n; i++ {
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
		}
		t.Run("1", func(t *testing.T) {
			bytes := []byte{
				0}
			if !new(Fe512).Unmarshal(bytes).Equals(&Fe512{0}) {
				t.Errorf("bad encoding\n")
			}
		})
		t.Run("2", func(t *testing.T) {
			bytes := []byte{
				254, 253}
			if new(Fe512).Unmarshal(bytes).Equals(&Fe512{0xfe, 0xfd}) {
				t.Errorf("bad encoding\n")
			}
		})
		t.Run("3", func(t *testing.T) {
			var a, b Fe512
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			bytes := make([]byte, 8*8)
			a.Marshal(bytes[:])
			b.Unmarshal(bytes[:])
			if !a.Equals(&b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
		t.Run("4", func(t *testing.T) {
			var a Fe512
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			b, er1 := new(Fe512).SetString(a.String())
			if er1 != nil {
				t.Fatal(er1)
			}
			if !a.Equals(b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
		t.Run("5", func(t *testing.T) {
			var a Fe512
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			b := new(Fe512).SetBig(a.Big())
			if er1 != nil {
				t.Fatal(er1)
			}
			if !a.Equals(b) {
				t.Errorf("bad encoding or decoding\n")
			}
		})
	})

	t.Run("Addition", func(t *testing.T) {
		var a, b, c, u, v Fe512
		zero := new(Fe512).SetUint(0)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Add(&u, &a, &b)
			field.Add(&u, &u, &c)
			field.Add(&v, &b, &c)
			field.Add(&v, &v, &a)
			if !u.Equals(&v) {
				t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)
			}
			field.Add(&u, &a, &b)
			field.Add(&v, &b, &a)
			if !u.Equals(&v) {
				t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv:%s\n", a, b, u, v)
			}
			field.Add(&u, &a, zero)
			if !u.Equals(&a) {
				t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)
			}
			field.Neg(&u, &a)
			field.Add(&u, &u, &a)
			if !u.Equals(zero) {
				t.Fatalf("Bad Negation\na:%s", a.String())
			}
		}
	})

	t.Run("Subtraction", func(t *testing.T) {
		var a, b, c, u, v Fe512
		zero := new(Fe512).SetUint(0)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Sub(&u, &a, &c)
			field.Sub(&u, &u, &b)
			field.Sub(&v, &a, &b)
			field.Sub(&v, &v, &c)
			if !u.Equals(&v) {
				t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)
			}
			field.Sub(&u, &a, zero)
			if !u.Equals(&a) {
				t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)
			}
			field.Sub(&u, &a, &b)
			field.Sub(&v, &b, &a)
			field.Add(&u, &u, &v)
			if !u.Equals(zero) {
				t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv: %s", a, b, u, v)
			}
			field.Sub(&u, &a, &b)
			field.Sub(&v, &b, &a)
			field.Neg(&v, &v)
			if !u.Equals(&u) {
				t.Fatalf("Bad Negation\na:%s", a.String())
			}
		}
	})

	t.Run("Doubling", func(t *testing.T) {
		var a, u, v Fe512
		for i := 0; i < n; i++ {
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
			err := field.RandElement(&a, rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			field.Double(&u, &a)
			field.Add(&v, &a, &a)
			if !u.Equals(&v) {
				t.Fatalf("Bad doubling\na: %s\nu: %s\nv: %s\n", a, u, v)
			}
		}
	})

	t.Run("Montgomerry", func(t *testing.T) {
		var a, b, c, u, v, w Fe512
		zero := new(Fe512).SetUint(0)
		one := new(Fe512).SetUint(1)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			er2 := field.RandElement(&b, rand.Reader)
			er3 := field.RandElement(&c, rand.Reader)
			if er1 != nil || er2 != nil || er3 != nil {
				t.Fatal(er1, er2, er3)
			}
			field.Mont(&u, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Demont(&u, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mont(&u, one)
			if !u.Equals(field.r1) {
				t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Demont(&u, field.r1)
			if !u.Equals(one) {
				t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mul(&u, &a, zero)
			if !u.Equals(zero) {
				t.Fatalf("Bad zero element\na: %s\nu: %s\np: %s\n", a, u, field.P)
			}
			field.Mul(&u, &a, one)
			field.Mul(&u, &u, field.r2)
			if !u.Equals(&a) {
				t.Fatalf("Multiplication identity does not hold, expected to equal itself\nu: %s\np: %s\n", u, field.P)
			}
			field.Mul(&u, field.r2, one)
			if !u.Equals(field.r1) {
				t.Fatalf("Multiplication identity does not hold, expected to equal r1\nu: %s\np: %s\n", u, field.P)
			}
			field.Mul(&u, &a, &b)
			field.Mul(&u, &u, &c)
			field.Mul(&v, &b, &c)
			field.Mul(&v, &v, &a)
			if !u.Equals(&v) {
				t.Fatalf("Multiplicative associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.P)
			}
			field.Add(&u, &a, &b)
			field.Mul(&u, &c, &u)
			field.Mul(&w, &a, &c)
			field.Mul(&v, &b, &c)
			field.Add(&v, &v, &w)
			if !u.Equals(&v) {
				t.Fatalf("Distributivity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.P)
			}
		}
	})

	t.Run("Exponentiation", func(t *testing.T) {
		var a, u, v Fe512
		bytes := make([]byte, 8*8)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			field.Exp(&u, &a, big.NewInt(0))
			if !u.Equals(field.r1) {
				t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			field.Exp(&u, &a, big.NewInt(1))
			if !u.Equals(&a) {
				t.Fatalf("Bad exponentiation, expected to equal a\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			field.Mul(&u, &a, &a)
			field.Mul(&u, &u, &u)
			field.Mul(&u, &u, &u)
			field.Exp(&v, &a, big.NewInt(8))
			if !u.Equals(&v) {
				t.Fatalf("Bad exponentiation\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			p := new(big.Int).SetBytes(field.P.Marshal(bytes))
			field.Exp(&u, &a, p)
			if !u.Equals(&a) {
				t.Fatalf("Bad exponentiation, expected to equal itself\nu: %s\na: %s\np: %s\n", u, a, field.P)
			}
			field.Exp(&u, &a, p.Sub(p, big.NewInt(1)))
			if !u.Equals(field.r1) {
				t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\na: %s\nr1: %s\np: %s\n", u, a, field.r1, field.P)
			}
		}
	})

	t.Run("Inversion", func(t *testing.T) {
		var u, a, v Fe512
		one := new(Fe512).SetUint(1)
		bytes := make([]byte, 8*8)
		for i := 0; i < n; i++ {
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
			er1 := field.RandElement(&a, rand.Reader)
			if er1 != nil {
				t.Fatal(er1)
			}
			field.InvMontUp(&u, &a)
			field.Mul(&u, &u, &a)
			if !u.Equals(field.r1) {
				t.Fatalf("Bad inversion, expected to equal r1\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P)
			}
			field.Mont(&u, &a)
			field.InvMontDown(&v, &u)
			field.Mul(&v, &v, &u)
			if !v.Equals(one) {
				t.Fatalf("Bad inversion, expected to equal 1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)
			}
			p := new(big.Int).SetBytes(field.P.Marshal(bytes))
			field.Exp(&u, &a, p.Sub(p, big.NewInt(2)))
			field.InvMontUp(&v, &a)
			if !v.Equals(&u) {
				t.Fatalf("Bad inversion")
			}
			field.InvEEA(&u, &a)
			field.Mul(&u, &u, &a)
			field.Mul(&u, &u, field.r2)
			if !u.Equals(one) {
				t.Fatalf("Bad inversion")
			}
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
	er1 := field.RandElement(&a, rand.Reader)
	er2 := field.RandElement(&b, rand.Reader)
	if er1 != nil || er2 != nil {
		t.Fatal(er1, er2)
	}
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
		bytes := make([]byte, 8*8)
		e := new(big.Int).SetBytes(field.P.Marshal(bytes))
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field.Exp(&c, &a, e)
		}
	})
}
