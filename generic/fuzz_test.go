package fp

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
)

func TestMultiplicationFuzzAgainstBigInt(t *testing.T) {
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
					fmt.Printf("p big\n%#x\n", field.pbig)
					fmt.Printf("r big\n%#x\n", field.rbig)
					fmt.Printf("r\n%#x\n", field.toBytesNoTransform(field.r))
					fmt.Printf("aR\n%#x\n", field.toBytesNoTransform(a))
					fmt.Printf("bR\n%#x\n", field.toBytesNoTransform(b))
					fmt.Printf("cR\n%#x\n", field.toBytesNoTransform(c))
					t.Fatal("i", i)
				}
			}
		})
	}
}

func TestExponentiationFuzz(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d_exponention", limbSize*64), func(t *testing.T) {
			for i := 0; i < fuz; i++ {
				field := randField(limbSize)
				a := field.randFieldElement(rand.Reader)
				u := field.newFieldElement()
				p := new(big.Int).SetBytes(field.pbig.Bytes())
				field.exp(u, a, p)
				if !field.equal(u, a) {
					fmt.Printf("p big\n%#x\n", field.pbig)
					fmt.Printf("r big\n%#x\n", field.rbig)
					fmt.Printf("aR\n%#x\n", field.toBytesNoTransform(a))
					t.Fatal("i", i)
				}
			}
		})
	}
}
