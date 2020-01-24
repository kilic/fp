package fp

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"testing"
)

var fuz int = 1

var targetNumberOfLimb int = -1

var from = 2
var to = 8

func TestMain(m *testing.M) {
	_fuz := flag.Int("fuzz", 1, "# of iters")
	nol := flag.Int("nol", 0, "backend bit size")
	flag.Parse()
	fuz = *_fuz
	if *nol > 0 {
		targetNumberOfLimb = *nol
		if !(targetNumberOfLimb >= from && targetNumberOfLimb <= to) {
			panic(fmt.Sprintf("limb size %d not supported", targetNumberOfLimb))
		}
		from = targetNumberOfLimb
		to = targetNumberOfLimb
	}
	m.Run()
}

func randField(limbSize int) *field {
	byteSize := limbSize * 8
	pbig, err := rand.Prime(rand.Reader, 8*byteSize-1)
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

func debugBytes(a ...[]byte) {
	for _, b := range a {
		for i := (len(b) / 8) - 1; i > -1; i-- {
			fmt.Printf("0x%16.16x,\n", b[i*8:i*8+8])
		}
		fmt.Println()
	}
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
	var limbSize int
	if targetNumberOfLimb > 0 {
		limbSize = targetNumberOfLimb
	} else {
		return
	}
	field := randField(limbSize)
	if field.limbSize != limbSize {
		t.Fatalf("bad field construction")
	}
	bitSize := limbSize * 64
	a := field.randFieldElement(rand.Reader)
	b := field.randFieldElement(rand.Reader)
	c := field.newFieldElement()
	t.Run(fmt.Sprintf("%d_add", bitSize), func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.add(c, a, b)
		}
	})
	t.Run(fmt.Sprintf("%d_double", bitSize), func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.double(c, a)
		}
	})
	t.Run(fmt.Sprintf("%d_sub", bitSize), func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.sub(c, a, b)
		}
	})
	t.Run(fmt.Sprintf("%d_mul", bitSize), func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.mul(c, a, b)
		}
	})
	t.Run(fmt.Sprintf("%d_cmp", bitSize), func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			field.cmp(a, b)
		}
	})
}

func TestShift(t *testing.T) {
	two := big.NewInt(2)
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d_shift", limbSize*64), func(t *testing.T) {
			field := randField(limbSize)
			a := field.randFieldElement(rand.Reader)
			bi := field.toBigNoTransform(a)
			da := field.newFieldElement()
			field.copy(da, a)
			field.div_two(da)
			dbi := new(big.Int).Div(bi, two)
			dbi_2 := field.toBigNoTransform(da)
			if dbi.Cmp(dbi_2) != 0 {
				t.Fatalf("bad div 2 operation")
			}
			ma := field.newFieldElement()
			field.copy(ma, a)
			field.mul_two(ma)
			mbi := new(big.Int).Mul(bi, two)
			mbi_2 := field.toBigNoTransform(ma)
			if mbi.Cmp(mbi_2) != 0 {
				t.Fatalf("bad mul 2 operation")
			}
		})
	}
}

func TestCompare(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d_compare", limbSize*64), func(t *testing.T) {
			field := randField(limbSize)
			if field.cmp(field.r, field.r) != 0 {
				t.Fatalf("r == r (cmp)")
			}
			if !field.equal(field.r, field.r) {
				t.Fatalf("r == r (equal)")
			}
			if field.equal(field.p, field.r) {
				t.Fatalf("p != r")
			}
			if field.equal(field.r, field.zero) {
				t.Fatalf("r != 0")
			}
			if !field.equal(field.zero, field.zero) {
				t.Fatalf("0 == 0")
			}
			if field.cmp(field.p, field.r) != 1 {
				t.Fatalf("p > r")
			}
			if field.cmp(field.r, field.p) != -1 {
				t.Fatalf("r < p")
			}
			if is_even(field.p) {
				t.Fatalf("p is not even")
			}
		})
	}
}

func TestCopy(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d_copy", limbSize*64), func(t *testing.T) {
			field := randField(limbSize)
			a := field.randFieldElement(rand.Reader)
			b := field.newFieldElement()
			field.copy(b, a)
			if !field.equal(a, b) {
				t.Fatalf("copy operation fails")
			}
		})
	}
}

func TestSerialization(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d_serialization", limbSize*64), func(t *testing.T) {
			field := randField(limbSize)
			if field.limbSize != limbSize {
				t.Fatalf("bad field construction\n")
			}
			// demont(r) == 1
			b0 := make([]byte, field.byteSize())
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
				field := randField(limbSize)
				if field.limbSize != limbSize {
					t.Fatalf("bad field construction")
				}
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
				if !field.equal(a0, a1) {
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
				if !field.equal(a0, a1) {
					t.Fatalf("bad serialization (big.Int)")
				}
			}
		})
	}
}

func TestAdditionCrossAgainstBigInt(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d_addition_cross", limbSize*64), func(t *testing.T) {
			for i := 0; i < fuz; i++ {
				field := randField(limbSize)
				if field.limbSize != limbSize {
					t.Fatalf("Bad field construction")
				}
				a := field.randFieldElement(rand.Reader)
				b := field.randFieldElement(rand.Reader)
				c := field.newFieldElement()
				big_a := field.toBig(a)
				big_b := field.toBig(b)
				big_c := new(big.Int)
				field.add(c, a, b)
				out_1 := field.toBytes(c)
				out_2 := padBytes(big_c.Add(big_a, big_b).Mod(big_c, field.pbig).Bytes(), field.byteSize())
				if !bytes.Equal(out_1, out_2) {
					t.Fatalf("cross test against big.Int is not satisfied A")
				}
				field.double(c, a)
				out_1 = field.toBytes(c)
				out_2 = padBytes(big_c.Add(big_a, big_a).Mod(big_c, field.pbig).Bytes(), field.byteSize())
				if !bytes.Equal(out_1, out_2) {
					t.Fatalf("cross test against big.Int is not satisfied B")
				}
				field.sub(c, a, b)
				out_1 = field.toBytes(c)
				out_2 = padBytes(big_c.Sub(big_a, big_b).Mod(big_c, field.pbig).Bytes(), field.byteSize())
				if !bytes.Equal(out_1, out_2) {
					t.Fatalf("cross test against big.Int is not satisfied C")
				}
				field.neg(c, a)
				out_1 = field.toBytes(c)
				out_2 = padBytes(big_c.Neg(big_a).Mod(big_c, field.pbig).Bytes(), field.byteSize())
				if !bytes.Equal(out_1, out_2) {
					t.Fatalf("cross test against big.Int is not satisfied D")
				}
			}
		})
	}
}

func TestAdditionProperties(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d_addition_properties", limbSize*64), func(t *testing.T) {
			for i := 0; i < fuz; i++ {
				field := randField(limbSize)
				if field.limbSize != limbSize {
					t.Fatalf("bad field construction")
				}
				a := field.randFieldElement(rand.Reader)
				b := field.randFieldElement(rand.Reader)
				c_1 := field.newFieldElement()
				c_2 := field.newFieldElement()
				field.add(c_1, a, field.zero)
				if !field.equal(c_1, a) {
					t.Fatalf("a + 0 == a")
				}
				field.sub(c_1, a, field.zero)
				if !field.equal(c_1, a) {
					t.Fatalf("a - 0 == a")
				}
				field.double(c_1, field.zero)
				if !field.equal(c_1, field.zero) {
					t.Fatalf("2 * 0 == 0")
				}
				field.neg(c_1, field.zero)
				if !field.equal(c_1, field.zero) {
					t.Fatalf("-0 == 0")
				}
				field.sub(c_1, field.zero, a)
				field.neg(c_2, a)
				if !field.equal(c_1, c_2) {
					t.Fatalf("0-a == -a")
				}
				field.double(c_1, a)
				field.add(c_2, a, a)
				if !field.equal(c_1, c_2) {
					t.Fatalf("2 * a == a + a")
				}
				field.add(c_1, a, b)
				field.add(c_2, b, a)
				if !field.equal(c_1, c_2) {
					t.Fatalf("a + b = b + a")
				}
				field.sub(c_1, a, b)
				field.sub(c_2, b, a)
				field.neg(c_2, c_2)
				if !field.equal(c_1, c_2) {
					t.Fatalf("a - b = - ( b - a )")
				}
				c_x := field.randFieldElement(rand.Reader)
				field.add(c_1, a, b)
				field.add(c_1, c_1, c_x)
				field.add(c_2, a, c_x)
				field.add(c_2, c_2, b)
				if !field.equal(c_1, c_2) {
					t.Fatalf("(a + b) + c == (a + c ) + b")
				}
				field.sub(c_1, a, b)
				field.sub(c_1, c_1, c_x)
				field.sub(c_2, a, c_x)
				field.sub(c_2, c_2, b)
				if !field.equal(c_1, c_2) {
					t.Fatalf("(a - b) - c == (a - c ) -b")
				}
			}
		})
	}
}

func TestMultiplicationCrossAgainstBigInt(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d_multiplication_cross", limbSize*64), func(t *testing.T) {
			for i := 0; i < fuz; i++ {
				field := randField(limbSize)
				if field.limbSize != limbSize {
					t.Fatalf("bad field construction")
				}
				a := field.randFieldElement(rand.Reader)
				b := field.randFieldElement(rand.Reader)
				c := field.newFieldElement()
				big_a := field.toBig(a)
				big_b := field.toBig(b)
				big_c := new(big.Int)
				field.mul(c, a, b)
				out_1 := field.toBytes(c)
				out_2 := padBytes(big_c.Mul(big_a, big_b).Mod(big_c, field.pbig).Bytes(), field.byteSize())
				if !bytes.Equal(out_1, out_2) {
					t.Fatalf("cross test against big.Int is not satisfied")
				}
			}
		})
	}
}

func TestMultiplicationProperties(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d_multiplication_properties", limbSize*64), func(t *testing.T) {
			for i := 0; i < fuz; i++ {
				field := randField(limbSize)
				if field.limbSize != limbSize {
					t.Fatalf("bad field construction")
				}
				a := field.randFieldElement(rand.Reader)
				b := field.randFieldElement(rand.Reader)
				c_1 := field.newFieldElement()
				c_2 := field.newFieldElement()
				field.mul(c_1, a, field.zero)
				if !field.equal(c_1, field.zero) {
					t.Fatalf("a * 0 == 0")
				}
				field.mul(c_1, a, field.one)
				if !field.equal(c_1, a) {
					t.Fatalf("a * 1 == a")
				}
				field.mul(c_1, a, b)
				field.mul(c_2, b, a)
				if !field.equal(c_1, c_2) {
					t.Fatalf("a * b == b * a")
				}
				c_x := field.randFieldElement(rand.Reader)
				field.mul(c_1, a, b)
				field.mul(c_1, c_1, c_x)
				field.mul(c_2, c_x, b)
				field.mul(c_2, c_2, a)
				if !field.equal(c_1, c_2) {
					t.Fatalf("(a * b) * c == (a * c) * b")
				}
			}
		})
	}
}

func TestExponentiation(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d_exponention", limbSize*64), func(t *testing.T) {
			for i := 0; i < fuz; i++ {
				field := randField(limbSize)
				if field.limbSize != limbSize {
					t.Fatalf("bad field construction")
				}
				a := field.randFieldElement(rand.Reader)
				u := field.newFieldElement()
				field.exp(u, a, big.NewInt(0))
				if !field.equal(u, field.one) {
					t.Fatalf("a^0 == 1")
				}
				field.exp(u, a, big.NewInt(1))
				if !field.equal(u, a) {
					t.Fatalf("a^1 == a")
				}
				v := field.newFieldElement()
				field.mul(u, a, a)
				field.mul(u, u, u)
				field.mul(u, u, u)
				field.exp(v, a, big.NewInt(8))
				if !field.equal(u, v) {
					t.Fatalf("((a^2)^2)^2 == a^8")
				}
				p := new(big.Int).SetBytes(field.pbig.Bytes())
				field.exp(u, a, p)
				if !field.equal(u, a) {
					t.Fatalf("a^p == a")
				}
				field.exp(u, a, p.Sub(p, big.NewInt(1)))
				if !field.equal(u, field.r) {
					t.Fatalf("a^(p-1) == 1")
				}
			}
		})
	}
}

func TestInversion(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d_inversion", limbSize*64), func(t *testing.T) {
			for i := 0; i < fuz; i++ {
				field := randField(limbSize)
				u := field.newFieldElement()
				field.inverse(u, field.zero)
				if !field.equal(u, field.zero) {
					t.Fatalf("(0^-1) == 0)")
				}
				field.inverse(u, field.one)
				if !field.equal(u, field.one) {
					t.Fatalf("(1^-1) == 1)")
				}
				a := field.randFieldElement(rand.Reader)
				field.inverse(u, a)
				field.mul(u, u, a)
				if !field.equal(u, field.r) {
					t.Fatalf("(r*a) * r*(a^-1) == r)")
				}
				v := field.newFieldElement()
				p := new(big.Int).Set(field.pbig)
				field.exp(u, a, p.Sub(p, big.NewInt(2)))
				field.inverse(v, a)
				if !field.equal(v, u) {
					t.Fatalf("a^(p-2) == a^-1")
				}
			}
		})
	}
}
