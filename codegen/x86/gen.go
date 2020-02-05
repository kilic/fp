package x86

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/mmcloughlin/avo/build"
)

const RSize int = 9

var logs = false

type bitFlags []int

func (flag *bitFlags) String() string {
	return "bit size flag"
}

func (flag *bitFlags) Set(value string) error {
	i, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}
	*flag = append(*flag, i)
	return nil
}

func GenX86All(output string) error {
	// a hack for avo output
	file := filepath.Join(output, "x86_arithmetic.s")
	if err := flag.Set("out", file); err != nil {
		return err
	}
	if err := os.MkdirAll(output, os.ModePerm); err != nil {
		return err
	}
	fixedmod, single, archTag := false, false, true
	for i := 1; i < 17; i++ {
		limbSize := i
		generateCopy(limbSize, single)
		generateEq(limbSize, single)
		generateCmp(limbSize, single)
		generateAdd(limbSize, fixedmod, single)
		generateAddNoCar(limbSize, single)
		generateDouble(limbSize, fixedmod, single)
		generateSub(limbSize, fixedmod, single)
		generateSubNoCar(limbSize, single)
		generateNeg(limbSize, fixedmod, single)
		generateMul2(limbSize, single)
		generateDiv2(limbSize, single)
		if limbSize != 1 {
			genMontMulADX(limbSize, fixedmod, single)
			genMontMulNoADX(limbSize, fixedmod, single, archTag)
		}
	}
	Generate()
	appendSingleLimbMultiplicationCode(file)
	appendIsEvenCode(file)
	pretty(file)
	return nil
}

func GenX86(output string, bitSize int, arch string, fixedmod bool, single bool) error {
	// a hack for avo output
	file := filepath.Join(output, "arithmetic.s")
	if err := flag.Set("out", file); err != nil {
		return err
	}
	// Package("github.com/kilic/fp/" + output)
	limbSize := bitSize / 64
	if bitSize%64 != 0 {
		return fmt.Errorf(fmt.Sprintf("bad bit size, %d\n", bitSize))
	}
	if limbSize < 2 || limbSize > 16 {
		return fmt.Errorf("limb size %d not implemented\n", limbSize)
	}
	generateCopy(limbSize, single)
	generateEq(limbSize, single)
	generateCmp(limbSize, single)
	generateAdd(limbSize, fixedmod, single)
	generateAddNoCar(limbSize, single)
	generateDouble(limbSize, fixedmod, single)
	generateSub(limbSize, fixedmod, single)
	generateSubNoCar(limbSize, single)
	generateNeg(limbSize, fixedmod, single)
	switch arch {
	case "ADX":
		genMontMulADX(limbSize, fixedmod, single)
	default:
		genMontMulNoADX(limbSize, fixedmod, single, false)
	}
	Generate()
	pretty(file)
	return nil
}

func GenDebugTest(limbs int, fixedmod bool, noadx bool, _logs bool) {
	logs = _logs
	file := "debug/mul.s"
	mkdirDebug()
	if err := flag.Set("out", file); err != nil {
		panic(err)
	}
	if noadx {
		genMontMulNoADX(limbs, fixedmod, true, false)
	} else {
		genMontMulADX(limbs, fixedmod, true)
	}
	Generate()
	pretty(file)
	generateTestCode(limbs, fixedmod)
}

func appendIsEvenCode(filename string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.WriteString(isEvenCode); err != nil {
		panic(err)
	}
}

func appendSingleLimbMultiplicationCode(filename string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.WriteString(singleLimbMultiplicationCode); err != nil {
		panic(err)
	}
	if _, err = f.WriteString(singleLimbMultiplicationNonAdxBmi2Code); err != nil {
		panic(err)
	}
}

func pretty(filename string) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln(err)
	}
	lines := strings.Split(string(input), "\n")
	for i, _ := range lines {
		lines[i] = strings.Replace(lines[i], "0x0000000000000000", "0x00", -1)
		lines[i] = strings.Replace(lines[i], "0x0000000000", "0x00", -1)
		lines[i] = strings.Replace(lines[i], "0x00000000", "0x00", -1)
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filename, []byte(output), 0600)
	if err != nil {
		log.Fatalln(err)
	}
}

func generateTestCode(limbs int, fixedmod bool) {
	outDir := filepath.Clean("./debug")
	writeToFile(declerationCode(limbs, fixedmod), filepath.Join(outDir, "decl.go"))
	writeToFile(testCode(fixedmod), filepath.Join(outDir, "mul_test.go"))
}

func writeToFile(content string, out string) {
	if err := ioutil.WriteFile(out, []byte(content), 0600); err != nil {
		panic(err)
	}
}

func declerationCode(limbs int, fixedmod bool) string {

	if fixedmod {
		return fmt.Sprintf(`package multest

const s = %d

type fl [s*2]uint64
type fe [s]uint64

//go:noescape
func mul(c, a, b *fe)
		`, limbs)
	}
	return fmt.Sprintf(`package multest

const s = %d

type fl [s*2]uint64
type fe [s]uint64

func mul(c, a, b, p *fe, inp uint64)
`, limbs)
}

func testCode(fixedmod bool) string {
	if fixedmod {
		return testCodeBase + _testCodeFixedmod
	}
	return testCodeBase + _testCode
}

const _testCodeFixedmod = `
func Benchmark(b *testing.B) {
	newField()
	a1 := randFe()
	a2 := randFe()
	var c fe
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mul(&c, a1, a2)
	}
}

func TestOne(t *testing.T) {
	for i := 0; i < fuz; i++ {
		newField()
		var c fe
		mul(&c, r, one)
		z := feToBig(c[:])
		expect := new(big.Int).SetUint64(1)
		if z.Cmp(expect) != 0 {
			debugFe(r, "r")
			debugFe(&c, "c")
			debugFe(&modulus, "p")
			t.Fatal("")
		}
	}
}

func TestSoftBox(t *testing.T) {
	for i := 0; i < fuz; i++ {
		newField()
		a := randFe()
		var c fe
		mul(&c, a, r)
		x := feToBig(a[:])
		y := feToBig(c[:])
		if x.Cmp(y) != 0 {
			fmt.Println(i)
			debugFe(a, "a")
			debugFe(&c, "c")
			debugFe(r, "r")
			debugFe(&modulus, "p")
			t.Fatal("")
		}
	}
}

func TestHardBox(t *testing.T) {
	for i := 0; i < fuz; i++ {
		newField()
		a, b := randFe(), randFe()
		var c fe
		mul(&c, a, b)
		ri, cw, ch := new(big.Int), new(big.Int), feToBig(c[:])
		ri.ModInverse(rbig, p)
		cw.Mul(
			feToBig(a[:]),
			feToBig(b[:]),
		).Mul(cw, ri).Mod(cw, p)
		if ch.Cmp(cw) != 0 {
			fmt.Println(i)
			debugFe(a, "a")
			debugFe(b, "b")
			fmt.Printf("ch = %#x\n", ch)
			fmt.Printf("cw = %#x\n", cw)
			debugFe(r, "r")
			debugFe(&modulus, "p")
			t.Fatal("")
		}
	}
}

func TestFF(t *testing.T) {
	for i := 0; i < fuz; i++ {
		newField()
		l := modulus[s-1]>>1 | 0xfffffffffffffff
		a, b := &fe{}, &fe{}
		for i := 0; i < s-1; i++ {
			b[i], a[i] = 0xffffffffffffffff, 0xffffffffffffffff
		}
		a[s-1], b[s-1] = l, l
		var c fe
		mul(&c, a, b)
		ri, cw, ch := new(big.Int), new(big.Int), feToBig(c[:])
		ri.ModInverse(rbig, p)
		cw.Mul(
			feToBig(a[:]),
			feToBig(b[:]),
		).Mul(cw, ri).Mod(cw, p)
		if ch.Cmp(cw) != 0 {
			fmt.Println(i)
			debugFe(a, "a")
			fmt.Printf("ch = %#x\n", ch)
			fmt.Printf("cw = %#x\n", cw)
			debugFe(r, "r")
			debugFe(&modulus, "p")
			t.Fatal("")
		}
	}
}
`

const _testCode = `
func Benchmark(b *testing.B) {
	newField()
	a1 := randFe()
	a2 := randFe()
	var c fe
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mul(&c, a1, a2, &modulus, inp)
	}
}

func TestOne(t *testing.T) {
	for i := 0; i < fuz; i++ {
		newField()
		var c fe
		mul(&c, r, one, &modulus, inp)
		z := feToBig(c[:])
		expect := new(big.Int).SetUint64(1)
		if z.Cmp(expect) != 0 {
			debugFe(r, "r")
			debugFe(&c, "c")
			debugFe(&modulus, "p")
			t.Fatal("")
		}
	}
}

func TestSoftBox(t *testing.T) {
	for i := 0; i < fuz; i++ {
		newField()
		a := randFe()
		var c fe
		mul(&c, a, r, &modulus, inp)
		x := feToBig(a[:])
		y := feToBig(c[:])
		if x.Cmp(y) != 0 {
			fmt.Println(i)
			debugFe(a, "a")
			debugFe(&c, "c")
			debugFe(r, "r")
			debugFe(&modulus, "p")
			t.Fatal("")
		}
	}
}

func TestHardBox(t *testing.T) {
	for i := 0; i < fuz; i++ {
		newField()
		a, b := randFe(), randFe()
		var c fe
		mul(&c, a, b, &modulus, inp)
		ri, cw, ch := new(big.Int), new(big.Int), feToBig(c[:])
		ri.ModInverse(rbig, p)
		cw.Mul(
			feToBig(a[:]),
			feToBig(b[:]),
		).Mul(cw, ri).Mod(cw, p)
		if ch.Cmp(cw) != 0 {
			fmt.Println(i)
			debugFe(a, "a")
			debugFe(b, "b")
			fmt.Printf("ch = %#x\n", ch)
			fmt.Printf("cw = %#x\n", cw)
			debugFe(r, "r")
			debugFe(&modulus, "p")
			t.Fatal("")
		}
	}
}

func TestFF(t *testing.T) {
	for i := 0; i < fuz; i++ {
		newField()
		l := modulus[s-1]>>1 | 0xfffffffffffffff
		a, b := &fe{}, &fe{}
		for i := 0; i < s-1; i++ {
			b[i], a[i] = 0xffffffffffffffff, 0xffffffffffffffff
		}
		a[s-1], b[s-1] = l, l
		var c fe
		mul(&c, a, b, &modulus, inp)
		ri, cw, ch := new(big.Int), new(big.Int), feToBig(c[:])
		ri.ModInverse(rbig, p)
		cw.Mul(
			feToBig(a[:]),
			feToBig(b[:]),
		).Mul(cw, ri).Mod(cw, p)
		if ch.Cmp(cw) != 0 {
			fmt.Println(i)
			debugFe(a, "a")
			fmt.Printf("ch = %#x\n", ch)
			fmt.Printf("cw = %#x\n", cw)
			debugFe(r, "r")
			debugFe(&modulus, "p")
			t.Fatal("")
		}
	}
}
`

const testCodeBase = `package multest

import (
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"testing"
)

var fuz int

func TestMain(m *testing.M) {
	_fuz := flag.Int("fuzz", 50, "# of iters")
	flag.Parse()
	fuz = *_fuz
	m.Run()
}

func feToBig(a []uint64) *big.Int {
	r := new(big.Int)
	for i := 0; i < len(a); i++ {
		b := new(big.Int)
		b.SetUint64(a[i]).Lsh(b, 64*uint(i))
		r.Or(r, b)
	}
	return r
}

func randFe() *fe {
	r, _ := rand.Int(rand.Reader, p)
	return fromBytes(r.Bytes())
}

func fromBytes(_in []byte) *fe {
	in := padBytes(_in, s*8)
	out := &fe{}
	var a int
	for i := 0; i < s; i++ {
		a = s*8 - i*8
		out[i] = uint64(in[a-1]) | uint64(in[a-2])<<8 |
			uint64(in[a-3])<<16 | uint64(in[a-4])<<24 |
			uint64(in[a-5])<<32 | uint64(in[a-6])<<40 |
			uint64(in[a-7])<<48 | uint64(in[a-8])<<56
	}
	return out
}

func padBytes(in []byte, size int) []byte {
	out := make([]byte, size)
	if len(in) > size {
		panic("bad input for padding")
	}
	copy(out[size-len(in):], in)
	return out
}

var modulus fe
var p *big.Int
var r *fe
var rbig *big.Int
var r2 *fe
var one *fe
var inp uint64

func newField() {
	byteSize := s * 8
	p, _ = rand.Prime(rand.Reader, byteSize*8)
	modulus = *fromBytes(p.Bytes())
	R := new(big.Int)
	R.SetBit(R, byteSize*8, 1).Mod(R, p)
	rbig = R
	R2 := new(big.Int)
	R2.Mul(R, R).Mod(R2, p)
	inpT := new(big.Int).ModInverse(new(big.Int).Neg(p), new(big.Int).SetBit(new(big.Int), 64, 1))
	r = fromBytes(R.Bytes())
	r2 = fromBytes(R2.Bytes())
	one = fromBytes(big.NewInt(1).Bytes())
	if inpT == nil {
		panic("cannot construct field")
	}
	inp = inpT.Uint64()
}

func debugFl(a *fl, desc string) {
	var str string
	for i := 0; i < s*2; i++ {
		s := fmt.Sprintf("%16.16x", a[i])
		fmt.Println(s)
		str = s + str
	}
	str = "0x" + str
	fmt.Println(desc, "=", str)
}

func debugFe(a *fe, desc string) {
	str := "0x"
	for i := s; i > 0; i-- {
		str += fmt.Sprintf("%16.16x", a[i-1])
	}
	fmt.Println(desc, "=", str)
}
`

func mkdirDebug() {
	output := filepath.Clean("./debug")
	s, err := os.Stat(output)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(output, os.ModePerm); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		if !s.IsDir() {
			panic("output should be directory")
		}
	}
}
