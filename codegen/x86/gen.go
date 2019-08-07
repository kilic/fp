package main

import (
	"flag"
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func main() {
	from := flag.Int("from", 4, "")
	to := flag.Int("to", 16, "")
	globalModulus := flag.Bool("globmod", false, "")
	flag.Parse()
	GenerateX86(*from, *to, *globalModulus)
}

func GenerateX86(from, to int, globalModulus bool) {
	Package("github.com/kilic/fp/codegen/generated")
	for i := from; i <= to; i++ {
		generateAdd(i, globalModulus)
		generateAddNoCar(i)
		generateSub(i, globalModulus)
		generateSubNoCar(i)
		generateDouble(i, globalModulus)
		generateNeg(i, globalModulus)
		generateMul(i)
		generateMont(i, globalModulus)
		generateMontMul(i, globalModulus)
		generateSquare(i)
		generateMontSquare(i, globalModulus)
	}
	Generate()
}

func generateAdd(size int, globMod bool) {
	/*
		("func add%d(c, a, b *Fe%d)\n\n", i, i*64)
		("func add%d(c, a, b, n *Fe%d)\n\n", i, i*64)
	*/
	Implement(fmt.Sprintf("add%d", size))
	Commentf("|")
	reservedGps := []Op{RAX, RBX}
	tape := newTape(reservedGps...)
	A := tape.newReprAtParam(size, "a", RDI, RBX)
	B := tape.newReprAtParam(size, "b", RSI, RBX)
	C_sum := tape.newReprAlloc(size, RBX)
	XORQ(RAX, RAX)
	Commentf("|")
	for i := 0; i < size; i++ {
		C_sum.next(_ITER).loadAdd(*A.next(_ITER), *B.next(_ITER), i != 0)
	}
	ADCQ(Imm(0), RAX)
	Commentf("|")
	var modulus *repr
	if globMod {
		tape.free(B.base)
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0), RBX)
	} else {
		modulus = tape.newReprAtParam(size, "p", B.base, RBX)
	}
	C_red := tape.newReprAlloc(size, RBX)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).loadSubSafe(*C_sum.next(_ITER), *modulus.next(_ITER), i != 0)
	}
	SBBQ(Imm(0), RAX)
	Commentf("|")
	C := tape.newReprAtParam(size, "c", RDI, RBX)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).moveIfNotCFAux(*C_sum.next(_ITER), *C.next(_ITER))
	}
	tape.ret()
	RET()
}

func generateAddNoCar(size int) {
	/*
		("func addn%d(a, b *Fe%d) uint64\n\n", i, i*64)
	*/
	Implement(fmt.Sprintf("addn%d", size))
	Commentf("|")
	reservedGps := []Op{RAX, RBX}
	tape := newTape(reservedGps...)
	A := tape.newReprAtParam(size, "a", RDI, RBX)
	B := tape.newReprAtParam(size, "b", RSI, RBX)
	C_sum := tape.newReprAlloc(size, RBX)
	XORQ(RAX, RAX)
	Commentf("|")
	for i := 0; i < size; i++ {
		C_sum.next(_ITER).loadAdd(*A.next(_ITER), *B.next(_ITER), i != 0)
	}
	ADCQ(Imm(0), RAX)
	Commentf("|")
	for i := 0; i < size; i++ {
		C_sum.next(_ITER).moveTo(*A.next(_ITER), _NO_ASSIGN)
	}
	Store(RAX, ReturnIndex(0))
	tape.ret()
	RET()
}

func generateDouble(size int, globMod bool) {
	/*
		("func double%d(c, a *Fe%d)\n\n", i, i*64)
		("func double%d(c, a, p *Fe%d)\n\n", i, i*64)
	*/
	Implement(fmt.Sprintf("double%d", size))
	Commentf("|")
	reservedGps := []Op{RAX, RBX}
	tape := newTape(reservedGps...)
	if !globMod {
		tape.reserveGp(RSI)
	}
	A := tape.newReprAtParam(size, "a", RDI, RBX)
	C_sum := tape.newReprAlloc(size, RBX)
	XORQ(RAX, RAX)
	for i := 0; i < size; i++ {
		C_sum.next(_ITER).loadDouble(*A.next(_ITER), i != 0)
	}
	ADCQ(Imm(0), RAX)
	Commentf("|")
	var modulus *repr
	if globMod {
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0), RBX)
	} else {
		modulus = tape.newReprAtParam(size, "p", RSI, RBX)
	}
	C_red := tape.newReprAlloc(size, RBX)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).loadSubSafe(*C_sum.next(_ITER), *modulus.next(_ITER), i != 0)
	}
	SBBQ(Imm(0), RAX)
	Commentf("|")
	C := tape.newReprAtParam(size, "c", RDI, RBX)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).moveIfNotCFAux(*C_sum.next(_ITER), *C.next(_ITER))
	}
	tape.ret()
	RET()
}

func generateSub(size int, globMod bool) {
	/*
	   ("func sub%d(c, a, b *Fe%d)\n\n", i, i*64)
	   ("func sub%d(c, a, b, p *Fe%d)\n\n", i, i*64)
	*/
	Implement(fmt.Sprintf("sub%d", size))
	Commentf("|")
	reservedGps := []Op{RAX, RBX}
	tape := newTape(reservedGps...)
	A := tape.newReprAtParam(size, "a", RDI, RBX)
	B := tape.newReprAtParam(size, "b", RSI, RBX)
	C_sub := tape.newReprAlloc(size, RBX)
	zero := tape.newReprNoAlloc(size, RBX)
	for i := 0; i < size; i++ {
		zero.next(_ITER).set(RAX)
	}
	XORQ(RAX, RAX)
	for i := 0; i < size; i++ {
		C_sub.next(_ITER).loadSub(*A.next(_ITER), *B.next(_ITER), i != 0)
	}
	Commentf("|")
	var modulus *repr
	if globMod {
		tape.free(B.base)
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0), RBX)
	} else {
		modulus = tape.newReprAtParam(size, "p", B.base, RBX)
	}
	C_mod := tape.newReprAlloc(size, RBX)
	for i := 0; i < size; i++ {
		zero.next(_ITER).moveIfNotCFAux(*modulus.next(_ITER), *C_mod.next(_ITER))
	}
	Commentf("|")
	C := tape.newReprAtParam(size, "c", RDI, RBX)
	for i := 0; i < size; i++ {
		C.next(_ITER).loadAdd(*C_sub.next(_ITER), *C_mod.next(_ITER), i != 0)
	}
	tape.ret()
	RET()
}

func generateSubNoCar(size int) {
	/*
		("func subn%d(a, b *Fe%d) uint64\n\n", i, i*64)
	*/
	Implement(fmt.Sprintf("subn%d", size))
	Commentf("|")
	reservedGps := []Op{RAX, RBX}
	tape := newTape(reservedGps...)
	A := tape.newReprAtParam(size, "a", RDI, RBX)
	B := tape.newReprAtParam(size, "b", RSI, RBX)
	C_sum := tape.newReprAlloc(size, RBX)
	XORQ(RAX, RAX)
	Commentf("|")
	for i := 0; i < size; i++ {
		C_sum.next(_ITER).loadSub(*A.next(_ITER), *B.next(_ITER), i != 0)
	}
	ADCQ(Imm(0), RAX)
	Commentf("|")
	for i := 0; i < size; i++ {
		C_sum.next(_ITER).moveTo(*A.next(_ITER), _NO_ASSIGN)
	}
	Store(RAX, ReturnIndex(0))
	tape.ret()
	RET()
}

func generateNeg(size int, globMod bool) {
	/*
	   ("func neg%d(c, a *Fe%d)\n\n", i, i*64)
	   ("func neg%d(c, a, p *Fe%d)\n\n", i, i*64)
	*/
	Implement(fmt.Sprintf("neg%d", size))
	Commentf("|")
	reservedGps := []Op{RAX, RBX}
	tape := newTape(reservedGps...)
	A := tape.newReprAtParam(size, "a", RDI, RBX)
	if !globMod {
		// reserve in advace for modulus
		tape.reserveGp(RSI)
	}
	C_sub := tape.newReprAlloc(size, RBX)
	Commentf("|")
	var modulus *repr
	if globMod {
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0), RBX)
	} else {
		modulus = tape.newReprAtParam(size, "p", RSI, RBX)
	}
	for i := 0; i < size; i++ {
		C_sub.next(_ITER).loadSub(*modulus.next(_ITER), *A.next(_ITER), i != 0)
	}
	Commentf("|")
	C := tape.newReprAtParam(size, "c", RDI, RBX)
	for i := 0; i < size; i++ {
		C_sub.next(_ITER).moveTo(*C.next(_ITER), _NO_ASSIGN)
	}
	tape.ret()
	RET()
}

func generateMul(size int) {
	/*
	   ("func mul%d(c *[%d]uint64, a, b *Fe%d)\n\n", i, i*2, i*64)
	*/

	Implement(fmt.Sprintf("mul%d", size))
	Commentf("|")
	reservedGps := []Op{RAX, RBX, RCX, RDX}
	tape := newTape(reservedGps...)
	A := tape.newReprAtParam(size, "a", RDI, RBX)
	B := tape.newReprAtParam(size, "b", RSI, RBX)
	w := tape.newReprAlloc(size*2, RBX)
	Commentf("|")
	mul(w, A, B)
	Commentf("|")
	w.updateIndex(0)
	C := tape.newReprAtParam(size, "c", RDI, RBX)
	for i := 0; i < 2*size; i++ {
		w.next(_ITER).moveTo(*C.next(_ITER), _NO_ASSIGN)
	}
	tape.ret()
	RET()
}

func generateMont(size int, globMod bool) {
	/*
		("func mont%d(c *Fe%d, w *[%d]uint64)\n\n", i, i*64, i*2, i*64)
		("func mont%d(c *Fe%d, w *[%d]uint64, p *Fe%d,inp uint64)\n\n", i, i*64, i*2, i*64)
	*/
	Implement(fmt.Sprintf("mont%d", size))
	reservedGps := []Op{RAX, RDX, RCX, RBX}
	tape := newTape(reservedGps...)
	carries := []Op{R14, R13, R15}
	tape.reserveGp(carries...)
	W := tape.newReprAtParam(2*size, "w", RDI, RBX)
	var modulus *repr
	var inp Mem
	if globMod {
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0), RBX)
		inp = NewDataAddr(Symbol{Name: fmt.Sprintf("·inp%d", size)}, 0)
	} else {
		tape.reserveGp(RSI)
		modulus = tape.newReprAtParam(size, "p", RSI, RBX)
		inp = NewParamAddr("inp", 24)
	}
	rotation := tape.newReprAlloc(size+1, RBX)
	rotation.load(W)
	C_mont := mont(tape, carries, inp, modulus, rotation, W)
	Commentf("| Reduce by modulus")
	tape.free(carries[:2]...)
	tape.free(RAX, RDX, RCX)
	C_red := tape.newReprAlloc(size, RBX)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).loadSubSafe(*C_mont.next(_ITER), *modulus.next(_ITER), i != 0)
	}
	SBBQ(Imm(0), carries[2])
	Commentf("| Compare & Return")
	C := tape.newReprAtParam(2*size, "c", RDI, RBX)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).moveIfNotCF(*C_mont.next(_NO_ITER))
		C_mont.next(_ITER).moveTo(*C.next(_ITER), _ASSING)
	}
	tape.ret()
	RET()
}

func generateMontMul(size int, globMod bool) {
	/*
	 ("func montmul%d(c, a, b *Fe%d)\n\n", i, i*64)
	 ("func montmul%d(c, a, b *Fe%d, p *Fe%d,inp uint64)\n\n", i, i*64, i*64)
	*/
	Implement(fmt.Sprintf("montmul%d", size))
	Commentf("|")
	reservedGps := []Op{RAX, RBX, RCX, RDX}
	tape := newTape(reservedGps...)
	Commentf("| Multiplication")
	A := tape.newReprAtParam(size, "a", RDI, RBX)
	B := tape.newReprAtParam(size, "b", RSI, RBX)
	w := tape.newReprAlloc(size*2, RBX)
	Commentf("|")
	mul(w, A, B)
	Commentf("|")
	w.updateIndex(0)
	Commentf("| Montgomerry Reduction")
	var longCarry Op
	if tape.sizeFreeGp() > 0 {
		longCarry = tape.next(_ALLOC)
	} else {
		for j := w.size - 1; ; j-- {
			if w.limbs[j].atReg() {
				longCarry = w.limbs[j].s
				w.limbs[j].moveTo(tape.next(_ALLOC), _ASSING)
				break
			}
		}
	}
	var modulus *repr
	var inp Mem
	if globMod {
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0), RBX)
		inp = NewDataAddr(Symbol{Name: fmt.Sprintf("·inp%d", size)}, 0)
	} else {
		inp = NewParamAddr("inp", 32)
		if tape.sizeFreeGp() > 0 {
			r := tape.next(_ALLOC).(Register)
			modulus = tape.newReprAtParam(size, "p", r, RBX)
		} else {
			// multiplication resut ([8]uint)
			// donates a general purpose register
			// for modulus address
			for j := w.size - 1; ; j-- {
				if w.limbs[j].atReg() {
					r := w.limbs[j].s.(Register)
					w.limbs[j].moveTo(tape.next(_ALLOC), _ASSING)
					modulus = tape.newReprAtParam(size, "p", r, RBX)
					break
				}
			}
		}
	}

	carries := []Op{RDI, RSI, longCarry}
	// tape.reserveGp(carries...) // Alreadry reserved in multiplication phase
	rotation := w.slice(0, size+1)
	C_mont := mont(tape, carries, inp, modulus, rotation, w)
	Commentf("| Reduce by modulus")
	tape.free(RAX, RDX, RCX, RSI)
	C_red := tape.newReprAlloc(size, RBX)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).loadSubSafe(*C_mont.next(_ITER), *modulus.next(_ITER), i != 0)
	}
	SBBQ(Imm(0), longCarry)
	Commentf("| Compare & Return")
	C := tape.newReprAtParam(size, "c", RDI, RBX)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).moveIfNotCF(*C_mont.next(_NO_ITER))
		C_mont.next(_ITER).moveTo(*C.next(_ITER), _ASSING)
	}
	tape.ret()
	RET()
}

func generateSquare(size int) {
	/*
		("func square%d(c *[%d]uint64, a *Fe%d)\n\n", i, i*2, i*64)
		("func square%d(c *[%d]uint64, a *Fe%d)\n\n", i, i*2, i*64)
	*/
	Implement(fmt.Sprintf("square%d", size))
	// todo
	RET()
}

func generateMontSquare(size int, globMod bool) {
	/*
	   ("func montsquare%d(c, a *Fe%d)\n\n", i, i*64)
	   ("func montsquare%d(c, a, p *Fe%d, inp uint64)\n\n", i, i*64)
	*/
	Implement(fmt.Sprintf("montsquare%d", size))
	// todo
	RET()
}

func mul(w, a, b *repr) {
	Commentf("|")
	size := a.size
	for i := 0; i < 2*size; i++ {
		z := w.next(_ITER)
		if z.atReg() {
			XORQ(z.s, z.s)
			continue
		}
		MOVQ(U32(0), z.s)
	}
	w.updateIndex(0)
	for i := 0; i < size; i++ {
		Commentf("|")
		Commentf("| b%d", i)
		b.next(_ITER).moveTo(RCX, _NO_ASSIGN)
		for j := 0; j < size; j++ {
			Commentf("| a%d * b%d", j, i)
			w.updateIndex(i + j)
			wa := w.next(_ITER)
			wb := w.next(_ITER)
			wc := w.next(_ITER)
			wd := w.next(_ITER)
			if i+j == 0 {
				a.next(_ITER).mul(RCX, wa.s, wb.s, _MUL_MOVE)
				continue
			}
			a.next(_ITER).mul(RCX, wa.s, wb.s, _MUL_ADD)
			if i+j+2 < 2*size && i > 0 {
				wc.addCarry()
				if i+j+3 < 2*size {
					wd.addCarry()
				}
			}
		}
	}
}

func mont(tape *tape, carries []Op, inp Op, modulus, rotation, w *repr) *repr {
	Commentf("|")
	size := w.size / 2
	shortCarries := carries[:2]
	longCarry := carries[2]
	C_mont := tape.newReprNoAlloc(size, RBX)
	for i, last := 0, size-1; i < size; i++ {
		Commentf("|")
		if i != 0 {
			next := w.next(_ITER)
			if next.atMem() {
				rotation.limbs[i-1].load(*next, nil)
			} else {
				rotation.limbs[i-1].set(*next)
			}
		}
		Commentf("| u @ CX = w%d * inp", i)
		rotation.updateIndex(i)
		rotation.mul(_NO_ITER, inp, RCX, nil, _MUL_MOVE)
		if i == last {
			tape.free(rotation.limbs[i])
		}
		var carry1, carry2 Register
		for j := 0; j < size; j++ {
			Commentf("| w%d", i+j)
			carry1 = shortCarries[j%2].(GPPhysical)
			carry2 = shortCarries[(j+1)%2].(GPPhysical)
			XORQ(carry1, carry1)
			e := rotation.next(_ITER)
			modulus.mul(_ITER, RCX, e.s, carry1, _MUL_ADD)
			if j > 0 {
				e.add(carry2, _NO_CARRY)
				ADCQ(Imm(0), carry1)
			}
		}
		j := i + size
		Commentf("| w%d", j)
		e := rotation.next(_ITER)
		C_mont.set(i, *e)
		if i == 0 {
			e.add(carry1, _NO_CARRY)
		} else {
			ADDQ(carry1, longCarry)
			e.add(longCarry, _CARRY)
		}
		MOVQ(U64(0), longCarry)
		ADCQ(Imm(0), longCarry)
	}
	return C_mont
}
