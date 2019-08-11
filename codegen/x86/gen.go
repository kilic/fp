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
		generateSquare(i, globalModulus)
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
	tape := newTape(RBX, RAX)
	A := tape.newReprAtParam(size, "a", RDI)
	B := tape.newReprAtParam(size, "b", RSI)
	C_sum := tape.newReprAlloc(size)
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
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0))
	} else {
		modulus = tape.newReprAtParam(size, "p", B.base)
	}
	C_red := tape.newReprAlloc(size)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).loadSubSafe(*C_sum.next(_ITER), *modulus.next(_ITER), i != 0)
	}
	SBBQ(Imm(0), RAX)
	Commentf("|")
	C := tape.newReprAtParam(size, "c", RDI)
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
	tape := newTape(RBX, RAX)
	A := tape.newReprAtParam(size, "a", RDI)
	B := tape.newReprAtParam(size, "b", RSI)
	C_sum := tape.newReprAlloc(size)
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
	tape := newTape(RBX, RAX)
	if !globMod {
		tape.reserveGp(RSI)
	}
	A := tape.newReprAtParam(size, "a", RDI)
	C_sum := tape.newReprAlloc(size)
	XORQ(RAX, RAX)
	for i := 0; i < size; i++ {
		C_sum.next(_ITER).loadDouble(*A.next(_ITER), i != 0)
	}
	ADCQ(Imm(0), RAX)
	Commentf("|")
	var modulus *repr
	if globMod {
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0))
	} else {
		modulus = tape.newReprAtParam(size, "p", RSI)
	}
	C_red := tape.newReprAlloc(size)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).loadSubSafe(*C_sum.next(_ITER), *modulus.next(_ITER), i != 0)
	}
	SBBQ(Imm(0), RAX)
	Commentf("|")
	C := tape.newReprAtParam(size, "c", RDI)
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
	tape := newTape(RBX, RAX)
	A := tape.newReprAtParam(size, "a", RDI)
	B := tape.newReprAtParam(size, "b", RSI)
	C_sub := tape.newReprAlloc(size)
	zero := tape.newReprNoAlloc(size)
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
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0))
	} else {
		modulus = tape.newReprAtParam(size, "p", B.base)
	}
	C_mod := tape.newReprAlloc(size)
	for i := 0; i < size; i++ {
		zero.next(_ITER).moveIfNotCFAux(*modulus.next(_ITER), *C_mod.next(_ITER))
	}
	Commentf("|")
	C := tape.newReprAtParam(size, "c", RDI)
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
	tape := newTape(RBX, RAX)
	A := tape.newReprAtParam(size, "a", RDI)
	B := tape.newReprAtParam(size, "b", RSI)
	C_sum := tape.newReprAlloc(size)
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
	tape := newTape(RBX, RAX)
	A := tape.newReprAtParam(size, "a", RDI)
	if !globMod {
		// reserve in advace for modulus
		tape.reserveGp(RSI)
	}
	C_sub := tape.newReprAlloc(size)
	Commentf("|")
	var modulus *repr
	if globMod {
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0))
	} else {
		modulus = tape.newReprAtParam(size, "p", RSI)
	}
	for i := 0; i < size; i++ {
		C_sub.next(_ITER).loadSub(*modulus.next(_ITER), *A.next(_ITER), i != 0)
	}
	Commentf("|")
	C := tape.newReprAtParam(size, "c", RDI)
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
	u := RCX
	tape := newTape(RBX, RAX, RDX, u)
	A := tape.newReprAtParam(size, "a", RDI)
	B := tape.newReprAtParam(size, "b", RSI)
	w := tape.newReprAlloc(size * 2)
	mul(w, A, B, u)
	Commentf("|")
	w.updateIndex(0)
	C := tape.newReprAtParam(2*size, "c", A.base)
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
	u := RCX
	tape := newTape(RBX, RAX, RDX, u)
	carrySet := newGpSet(tape.reserveGp(R14, R13, R15)...)
	w := tape.newReprAtParam(2*size, "w", RDI)
	var modulus *repr
	var inp Mem
	if globMod {
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0))
		inp = NewDataAddr(Symbol{Name: fmt.Sprintf("·inp%d", size)}, 0)
	} else {
		modulus = tape.newReprAtParam(size, "p", RSI)
		inp = NewParamAddr("inp", 24)
	}
	rotation := tape.newReprAlloc(size + 1).load(w)
	C_mont, longCarry := mont(tape, carrySet, inp, modulus, w, rotation, u)
	Commentf("| Reduce by modulus")
	tape.free(carrySet.slice()...)
	tape.free(RAX, RDX, u)
	tape.reserveGp(longCarry)
	C_red := tape.newReprAlloc(size)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).loadSubSafe(*C_mont.next(_ITER), *modulus.next(_ITER), i != 0)
	}
	SBBQ(Imm(0), longCarry)
	Commentf("| Compare & Return")
	C := tape.newReprAtParam(2*size, "c", w.base)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).moveIfNotCFAux(*C_mont.next(_ITER), *C.next(_ITER))
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
	u := RCX
	tape := newTape(RBX, RAX, RDX, u)
	Commentf("| Multiplication")
	A := tape.newReprAtParam(size, "a", RDI)
	B := tape.newReprAtParam(size, "b", RSI)
	w := tape.newReprAlloc(size * 2)
	Commentf("|")
	mul(w, A, B, u)
	Commentf("|")
	w.updateIndex(0)
	Commentf("| Montgomerry Reduction")
	var longCarry Op
	if tape.sizeFreeGp() > 0 {
		longCarry = tape.next(_ALLOC)
	} else {
		longCarry = tape.donate(w)
	}
	var modulus *repr
	var inp Mem
	if globMod {
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0))
		inp = NewDataAddr(Symbol{Name: fmt.Sprintf("·inp%d", size)}, 0)
	} else {
		inp = NewParamAddr("inp", 32)
		if tape.sizeFreeGp() > 0 {
			r := tape.next(_ALLOC).(GPPhysical)
			modulus = tape.newReprAtParam(size, "p", r)
		} else {
			modulus = tape.newReprAtParam(size, "p", tape.donate(w).(Register))
		}
	}
	carrySet := newGpSet(A.base, B.base, longCarry)
	// tape.reserveGp(carries...)  Alreadry reserved in multiplication phase
	rotation := w.slice(0, (w.size/2)+1)
	C_mont, _ := mont(tape, carrySet, inp, modulus, w, rotation, u)
	Commentf("| Reduce by modulus")
	tape.free(RAX, RDX, u)
	// tape.free(carrySet.slice()...)
	tape.reserveGp(B.base)
	C_red := tape.newReprAlloc(size)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).loadSubSafe(*C_mont.next(_ITER), *modulus.next(_ITER), i != 0)
	}
	SBBQ(Imm(0), longCarry)
	Commentf("| Compare & Return")
	C := tape.newReprAtParam(size, "c", A.base)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).moveIfNotCFAux(*C_mont.next(_ITER), *C.next(_ITER))
	}
	tape.ret()
	RET()
}

func generateSquare(size int, globMod bool) {
	/*
		("func square%d(c *[%d]uint64, a *Fe%d)\n\n", i, i*2, i*64)
		("func square%d(c *[%d]uint64, a *Fe%d)\n\n", i, i*2, i*64)
	*/
	Implement(fmt.Sprintf("square%d", size))
	// TEXT("square", NOSPLIT, "func(w *[8]uint64, a *[4]uint64)")
	Commentf("|")
	u := R8
	tape := newTape(RBX, RAX, RDX, u)
	A := tape.newReprAtParam(size, "a", RDI)
	w := tape.newReprAlloc(size * 2)
	square(w, A, u)
	Commentf("|")
	w.updateIndex(0)
	W := tape.newReprAtParam(2*size, "c", A.base)
	for i := 0; i < 2*size; i++ {
		w.next(_ITER).moveTo(*W.next(_ITER), _NO_ASSIGN)
	}
	tape.ret()
	RET()
}

func generateMontSquare(size int, globMod bool) {
	/*
	   ("func montsquare%d(c, a *Fe%d)\n\n", i, i*64)
	   ("func montsquare%d(c, a, p *Fe%d, inp uint64)\n\n", i, i*64)
	*/
	Implement(fmt.Sprintf("montsquare%d", size))
	// TEXT("montsquare", NOSPLIT, "func(c *[4]uint64, a *[4]uint64)")
	u := R8
	tape := newTape(RBX, RAX, RDX, u)
	A := tape.newReprAtParam(size, "a", RDI)
	w := tape.newReprAlloc(size * 2)
	square(w, A, u)
	Commentf("|")
	w.updateIndex(0)
	Commentf("| Montgomerry Reduction")
	// Carries
	longCarry := A.base
	var c1, c2 Op
	if tape.sizeFreeGp() > 0 {
		c1 = tape.next(_ALLOC)
	} else {
		c1 = tape.donate(w)
	}
	if tape.sizeFreeGp() > 0 {
		c2 = tape.next(_ALLOC)
	} else {
		c2 = tape.donate(w)
	}
	carrySet := newGpSet(c1, c2, longCarry)
	// Modulus
	var modulus *repr
	var inp Mem
	if globMod {
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0))
		inp = NewDataAddr(Symbol{Name: fmt.Sprintf("·inp%d", size)}, 0)
	} else {
		inp = NewParamAddr("inp", 24)
		if tape.sizeFreeGp() > 0 {
			modulus = tape.newReprAtParam(size, "p", tape.next(_ALLOC).(Register))
		} else {
			modulus = tape.newReprAtParam(size, "p", tape.donate(w).(Register))
		}
	}
	rotation := w.slice(0, (w.size/2)+1)
	C_mont, _ := mont(tape, carrySet, inp, modulus, w, rotation, u)
	Commentf("| Compare & Return")
	tape.free(RAX, RDX, c1, c2, u)
	tape.free(carrySet.slice()...)
	tape.reserveGp(longCarry)
	C_red := tape.newReprAlloc(size)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).loadSubSafe(*C_mont.next(_ITER), *modulus.next(_ITER), i != 0)
	}
	SBBQ(Imm(0), longCarry)
	C := tape.newReprAtParam(size, "c", longCarry)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).moveIfNotCFAux(*C_mont.next(_ITER), *C.next(_ITER))
	}
	tape.ret()
	RET()
}

func square(w, A *repr, ai Op) {
	lo, hi := RAX, RDX
	size := A.size
	carryAdded := make(map[*limb]bool)
	numberAdded := make(map[*limb]bool)
	for i := 0; i < 2*size; i++ {
		z := w.next(_ITER)
		if z.atReg() {
			if i > 1 {
				XORQ(z.s, z.s)
			}
			continue
		}
		MOVQ(U32(0), z.s)
	}
	for i := 0; i < size; i++ {
		first, last := (i == 0), (i == size-1)
		w.updateIndex(2 * i)
		A.updateIndex(i)
		Commentf("| a%d", i)
		Commentf("| w%d @ %s", 2*i, w.next(_NO_ITER).s.Asm())
		A.next(_NO_ITER).moveTo(ai, _ASSIGN)
		wa := w.next(_ITER)
		wb := w.next(_ITER)
		if first {
			A.mul(_ITER, ai, wa.s, wb.s, _MUL_MOVE)
		} else {
			A.mul(_ITER, ai, wa.s, wb.s, _MUL_ADD)
		}
		wc := w.next(_ITER)
		wd := w.next(_ITER)
		if numberAdded[wb] {
			carryAdded[wc] = true
			wc.addCarry()
			if numberAdded[wc] {
				carryAdded[wd] = true
				wd.addCarry()
			}
		}
		numberAdded[wa] = true
		numberAdded[wb] = true
		for j := i + 1; j < size; j++ {
			w.updateIndex(i + j)
			Commentf("| w%d @ %s", i+j, w.next(_NO_ITER).s.Asm())
			w.updateIndex(i + j + 2)
			A.mul(_ITER, ai, lo, hi, _MUL_ADD)
			carryAdded[w.next(_NO_ITER)] = true
			w.next(_NO_ITER).addCarry()
			w.updateIndex(i + j)
			wa := w.next(_ITER)
			wb := w.next(_ITER)
			wc := w.next(_ITER)
			wa.add(lo, _NO_CARRY)
			wb.add(hi, _CARRY)
			if numberAdded[wb] || carryAdded[wb] {
				carryAdded[wc] = true
				wc.addCarry()
			}
			numberAdded[wa] = true
			numberAdded[wb] = true
		}
		_ = last
	}
}

func mul(w, a, b *repr, bi Op) {
	Commentf("|")
	size := a.size
	for i := 0; i < 2*size; i++ {
		z := w.next(_ITER)
		if z.atReg() {
			if i > 1 {
				XORQ(z.s, z.s)
			}
			continue
		}
		MOVQ(U32(0), z.s)
	}
	w.updateIndex(0)
	for i := 0; i < size; i++ {
		Commentf("|")
		Commentf("| b%d", i)
		b.next(_ITER).moveTo(bi, _NO_ASSIGN)
		for j := 0; j < size; j++ {
			Commentf("| a%d * b%d", j, i)
			w.updateIndex(i + j)
			wa := w.next(_ITER)
			wb := w.next(_ITER)
			wc := w.next(_ITER)
			wd := w.next(_ITER)
			addOneCarry := (i+j+2 < 2*size && i > 0)
			addTwoCarry := (i+j+3 < 2*size)
			if addOneCarry {
				if addTwoCarry {
					Commentf("| (w%d, w%d, w%d, w%d) @ (%s, %s, %s, %s)", i+j, i+j+1, i+j+2, i+j+3, wa.s.Asm(), wb.s.Asm(), wc.s.Asm(), wd.s.Asm())
				} else {
					Commentf("| (w%d, w%d, w%d) @ (%s, %s, %s)", i+j, i+j+1, i+j+2, wa.s.Asm(), wb.s.Asm(), wc.s.Asm())
				}
			} else {
				Commentf("| (w%d, w%d) @ (%s, %s)", i+j, i+j+1, wa.s.Asm(), wb.s.Asm())
			}
			if i+j == 0 {
				a.next(_ITER).mul(bi, wa.s, wb.s, _MUL_MOVE)
				continue
			}
			a.next(_ITER).mul(bi, wa.s, wb.s, _MUL_ADD)
			if addOneCarry {
				wc.addCarry()
				if addTwoCarry {
					wd.addCarry()
				}
			}
		}
	}
}

func mont(tape *tape, carries *gpSet, inp Op, modulus, w, rotation *repr, u Op) (*repr, Op) {
	size := w.size / 2
	carries.freeAll()
	shortCarries := []Op{carries.next(_ALLOC), carries.next(_ALLOC)}
	longCarry := carries.next(_ALLOC)
	C_mont := tape.newReprNoAlloc(size)
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
		rotation.updateIndex(i)
		Commentf("| (u @ %s) = (w%d @ %s) * inp", u.Asm(), i, rotation.next(_NO_ITER).Asm())
		rotation.mul(_NO_ITER, inp, u, nil, _MUL_MOVE)
		if i == last {
			tape.free(rotation.limbs[i])
		}
		var carry1, carry2 Op
		for j := 0; j < size; j++ {
			Commentf("| w%d @ %s", i+j, rotation.next(_NO_ITER).Asm())
			carry1, carry2 = shortCarries[j%2], shortCarries[(j+1)%2]
			XORQ(carry1, carry1)
			e := rotation.next(_ITER)
			modulus.mul(_ITER, u, e.s, carry1, _MUL_ADD)
			if j > 0 {
				e.add(carry2, _NO_CARRY)
				ADCQ(Imm(0), carry1)
			}
		}
		j := i + size
		Commentf("| w%d @ %s", j, rotation.next(_NO_ITER).Asm())
		// e := rotation.next(_ITER)
		C_mont.set(i, rotation.next(_NO_ITER).s)
		if i == 0 {
			rotation.next(_ITER).add(carry1, _NO_CARRY)
		} else {
			ADDQ(carry1, longCarry)
			rotation.next(_ITER).add(longCarry, _CARRY)
		}
		MOVQ(U64(0), longCarry)
		ADCQ(Imm(0), longCarry)
	}
	return C_mont, longCarry
}
