package fp

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
)

func TestFuzzMultiplicationAgainstBigInt(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d", limbSize*64), func(t *testing.T) {
			for i := 0; i < fuz; i++ {
				field := randField(limbSize)
				for j := 0; j < fieldLifetime; j++ {
					if field.limbSize != limbSize {
						fmt.Println("---")
						fmt.Println(i)
						fmt.Println("bad field construction")
						field.debug()
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
						fmt.Println("---")
						fmt.Println(i)
						field.debug()
						fmt.Printf("aR\n%#x\n", field.toBytesNoTransform(a))
						fmt.Printf("bR\n%#x\n", field.toBytesNoTransform(b))
						fmt.Printf("cR\n%#x\n", field.toBytesNoTransform(c))
						fmt.Printf("c\n%#x\n", out_2)
					}
				}
			}
		})
	}
}

func TestFuzzExponentiation(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d", limbSize*64), func(t *testing.T) {
			for i := 0; i < fuz; i++ {
				field := randField(limbSize)
				for j := 0; j < fieldLifetime; j++ {

					a := field.randFieldElement(rand.Reader)
					u := field.newFieldElement()
					p := new(big.Int).SetBytes(field.pbig.Bytes())
					field.exp(u, a, p)
					if !field.equal(u, a) {
						fmt.Println("---")
						fmt.Println(i)
						field.debug()
						fmt.Printf("aR\n%#x\n", field.toBytesNoTransform(a))
					}
				}
			}
		})
	}
}

func TestFuzzInversion(t *testing.T) {
	for limbSize := from; limbSize < to+1; limbSize++ {
		t.Run(fmt.Sprintf("%d", limbSize*64), func(t *testing.T) {
			for i := 0; i < fuz; i++ {
				field := randField(limbSize)
				for j := 0; j < fieldLifetime; j++ {
					u := field.newFieldElement()
					field.inverse(u, field.one)
					if !field.equal(u, field.one) {
						fmt.Println("---")
						fmt.Println("(1^-1) == 1)")
						fmt.Println(i)
						field.debug()
						fmt.Printf("u\n%#x\n", field.toBytesNoTransform(u))
					}
					a := field.randFieldElement(rand.Reader)
					inv := field.toBig(a)
					inv.ModInverse(inv, field.pbig)
					field.inverse(u, a)
					inv2 := field.toBig(u)
					if inv.Cmp(inv2) != 0 {
						fmt.Println("---")
						fmt.Println("cross against big int")
						fmt.Println(i)
						field.debug()
						fmt.Printf("a\n%#x\n", field.toBytesNoTransform(a))
						fmt.Printf("u\n%#x\n", field.toBytesNoTransform(u))
						fmt.Printf("u2\n%#x\n", inv2.Bytes())
					}
					a = field.randFieldElement(rand.Reader)
					v := field.randFieldElement(rand.Reader)
					field.inverse(v, a)
					field.mul(u, v, a)
					if !field.equal(u, field.r) {
						fmt.Println("---")
						fmt.Println("(r*a) * r*(a^-1) == r)")
						fmt.Println(i)
						field.debug()
						fmt.Printf("a\n%#x\n", field.toBytesNoTransform(a))
						fmt.Printf("v\n%#x\n", field.toBytesNoTransform(v))
						fmt.Printf("u\n%#x\n", field.toBytesNoTransform(u))
					}
					p := new(big.Int).Set(field.pbig)
					field.exp(u, a, p.Sub(p, big.NewInt(2)))
					field.inverse(v, a)
					if !field.equal(v, u) {
						fmt.Println("---")
						fmt.Println("a^(p-2) == a^-1")
						fmt.Println(i)
						field.debug()
						fmt.Printf("a\n%#x\n", field.toBytesNoTransform(a))
						fmt.Printf("u\n%#x\n", field.toBytesNoTransform(u))
						fmt.Printf("v\n%#x\n", field.toBytesNoTransform(v))
					}
				}
			}
		})
	}
}
