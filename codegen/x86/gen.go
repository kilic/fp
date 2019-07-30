// +build ignore

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
	flag.Parse()
	GenerateX86(*from, *to)
}

func GenerateX86(from, to int) {
	Package("github.com/kilic/fp/codegen/generated")
	for i := from; i <= to; i++ {
		generateAdd(i)
		generateAddNoCar(i)
		generateSub(i)
		generateSubNoCar(i)
		generateDouble(i)
		generateNeg(i)
		generateMul(i)
		generateSquare(i)
		generateMont(i)
		generateMontMul(i)
		generateMontSquare(i)
	}
	Generate()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func generateAdd(size int) {
	modulus := newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0), RBX)
	Implement(fmt.Sprintf("add%d", size))
	reservedGps := []Op{RAX, RBX, RDI, RSI}
	tape := newTape(reservedGps...)
	C_sum := tape.newReprAlloc(size, RBX)
	C_red := tape.newReprAlloc(size, RBX)
	Commentf("|")
	A := newReprAtMemory(size, Mem{Base: Load(Param("a"), RDI)}, RBX)
	B := newReprAtMemory(size, Mem{Base: Load(Param("b"), RSI)}, RBX)
	XORQ(RAX, RAX)
	Commentf("|")
	for i := 0; i < size; i++ {
		C_sum.next(_ITER).loadAdd(*A.next(_ITER), *B.next(_ITER), i != 0)
	}
	ADCQ(Imm(0), RAX)
	Commentf("|")
	for i := 0; i < size; i++ {
		// improvement : swap register could be used
		C_red.next(_ITER).loadSubSafe(*C_sum.next(_ITER), *modulus.next(_ITER), i != 0)
	}
	SBBQ(Imm(0), RAX)
	Commentf("|")
	C := newReprAtMemory(size, Mem{Base: Load(Param("c"), RDI)}, RBX)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).moveIfNotCFAux(*C_sum.next(_ITER), *C.next(_ITER))
	}
	tape.ret()
	RET()
}

func generateAddNoCar(size int) {
	Implement(fmt.Sprintf("addn%d", size))
	reservedGps := []Op{RAX, RBX, RDI, RSI}
	tape := newTape(reservedGps...)
	C_sum := tape.newReprAlloc(size, RBX)
	Commentf("|")
	A := newReprAtMemory(size, Mem{Base: Load(Param("a"), RDI)}, RBX)
	B := newReprAtMemory(size, Mem{Base: Load(Param("b"), RSI)}, RBX)
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

func generateDouble(size int) {
	modulus := newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0), RBX)
	Implement(fmt.Sprintf("double%d", size))
	reservedGps := []Op{RAX, RBX, RDI}
	tape := newTape(reservedGps...)
	C_sum := tape.newReprAlloc(size, RBX)
	C_red := tape.newReprAlloc(size, RBX)
	Commentf("|")
	A := newReprAtMemory(size, Mem{Base: Load(Param("a"), RDI)}, RBX)
	XORQ(RAX, RAX)
	for i := 0; i < size; i++ {
		C_sum.next(_ITER).loadDouble(*A.next(_ITER), i != 0)
	}
	ADCQ(Imm(0), RAX)
	Commentf("|")
	for i := 0; i < size; i++ {
		C_red.next(_ITER).loadSubSafe(*C_sum.next(_ITER), *modulus.next(_ITER), i != 0)
	}
	SBBQ(Imm(0), RAX)
	Commentf("|")
	C := newReprAtMemory(size, Mem{Base: Load(Param("c"), RDI)}, RBX)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).moveIfNotCFAux(*C_sum.next(_ITER), *C.next(_ITER))
	}
	tape.ret()
	RET()
}

func generateSub(size int) {
	modulus := newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0), RBX)
	Implement(fmt.Sprintf("sub%d", size))
	reservedGps := []Op{RAX, RBX, RDI, RSI}
	tape := newTape(reservedGps...)
	C_sub := tape.newReprAlloc(size, RBX)
	C_mod := tape.newReprAlloc(size, RBX)
	zero := tape.newReprNoAlloc(size, RBX)
	for i := 0; i < size; i++ {
		zero.next(_ITER).set(RAX)
	}
	Commentf("|")
	A := newReprAtMemory(size, Mem{Base: Load(Param("a"), RDI)}, RBX)
	B := newReprAtMemory(size, Mem{Base: Load(Param("b"), RSI)}, RBX)
	XORQ(RAX, RAX)
	for i := 0; i < size; i++ {
		C_sub.next(_ITER).loadSub(*A.next(_ITER), *B.next(_ITER), i != 0)
	}
	Commentf("|")
	for i := 0; i < size; i++ {
		zero.next(_ITER).moveIfNotCFAux(*modulus.next(_ITER), *C_mod.next(_ITER))
	}
	Commentf("|")
	C := newReprAtMemory(size, Mem{Base: Load(Param("c"), RDI)}, RBX)
	for i := 0; i < size; i++ {
		C.next(_ITER).loadAdd(*C_sub.next(_ITER), *C_mod.next(_ITER), i != 0)
	}
	tape.ret()
	RET()
}

func generateSubNoCar(size int) {
	Implement(fmt.Sprintf("subn%d", size))
	reservedGps := []Op{RAX, RBX, RDI, RSI}
	tape := newTape(reservedGps...)
	C_sum := tape.newReprAlloc(size, RBX)
	Commentf("|")
	A := newReprAtMemory(size, Mem{Base: Load(Param("a"), RDI)}, RBX)
	B := newReprAtMemory(size, Mem{Base: Load(Param("b"), RSI)}, RBX)
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

func generateNeg(size int) {
	modulus := newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0), RBX)
	Implement(fmt.Sprintf("neg%d", size))
	reservedGps := []Op{RAX, RBX, RDI}
	tape := newTape(reservedGps...)
	C_sub := tape.newReprAlloc(size, RBX)
	Commentf("|")
	A := newReprAtMemory(size, Mem{Base: Load(Param("a"), RDI)}, RBX)
	Commentf("|")
	for i := 0; i < size; i++ {
		C_sub.next(_ITER).loadSub(*modulus.next(_ITER), *A.next(_ITER), i != 0)
	}
	Commentf("|")
	C := newReprAtMemory(size, Mem{Base: Load(Param("c"), RDI)}, RBX)
	for i := 0; i < size; i++ {
		C_sub.next(_ITER).moveTo(*C.next(_ITER), _NO_ASSIGN)
	}
	tape.ret()
	RET()
}

func generateMul(size int) {
	Implement(fmt.Sprintf("mul%d", size))
	Commentf("|")
	reservedGps := []Op{RAX, RBX, RCX, RDX, RDI, RSI}
	tape := newTape(reservedGps...)

	w := tape.newReprAlloc(size*2, RBX)
	A := newReprAtMemory(size, Mem{Base: Load(Param("a"), RDI)}, RBX)
	B := newReprAtMemory(size, Mem{Base: Load(Param("b"), RSI)}, RBX)
	Commentf("|")
	mul(w, A, B)
	Commentf("|")
	w.updateIndex(0)
	C := newReprAtMemory(2*size, Mem{Base: Load(Param("c"), RDI)}, RBX)
	for i := 0; i < 2*size; i++ {
		w.next(_ITER).moveTo(*C.next(_ITER), _NO_ASSIGN)
	}
	tape.ret()
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

func generateMont(size int) {

	modulus := newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0), nil)
	inp := NewDataAddr(Symbol{Name: fmt.Sprintf("·inp%d", size)}, 0)
	Implement(fmt.Sprintf("mont%d", size))
	W := newReprAtMemory(size*2, Mem{Base: Load(Param("w"), RDI)}, nil)
	reservedGps := []Op{RAX, RDX, RCX, RBX, RDI, R14, R13, R15}
	tape := newTape(reservedGps...)
	carries := []Op{R14, R13, R15}
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
	C := newReprAtMemory(2*size, Mem{Base: Load(Param("c"), RDI)}, RBX)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).moveIfNotCF(*C_mont.next(_NO_ITER))
		C_mont.next(_ITER).moveTo(*C.next(_ITER), _ASSING)
	}
	tape.ret()
	RET()
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

func generateMontMul(size int) {
	modulus := newReprAtMemory(size, NewDataAddr(Symbol{Name: fmt.Sprintf("·modulus%d", size)}, 0), nil)
	inp := NewDataAddr(Symbol{Name: fmt.Sprintf("·inp%d", size)}, 0)
	Implement(fmt.Sprintf("montmul%d", size))
	Commentf("|")
	Commentf("| Multiplication")
	reservedGps := []Op{RAX, RBX, RCX, RDX, RDI, RSI}
	tape := newTape(reservedGps...)
	w := tape.newReprAlloc(size*2, RBX)
	A := newReprAtMemory(size, Mem{Base: Load(Param("a"), RDI)}, RBX)
	B := newReprAtMemory(size, Mem{Base: Load(Param("b"), RSI)}, RBX)
	Commentf("|")
	mul(w, A, B)
	Commentf("|")
	w.updateIndex(0)

	Commentf("| Montgomerry Reduction")
	// RDI and RSI was referencing to field element inputs.
	// Inputs are already processed, and related registers is
	// going to be allocated for short carries in montgomerry operation
	// in which we exactly need two of them.
	// We also need a register to oparate the long carry
	// that is used in montgomery limb iterations.
	// At this point all cpu registers are probably full with multiplication result.
	// Therefore the last general purpose that is allocated for multiplication result
	// is going to be reserved for long carry.
	// And new stack position is to be allocated in return of the exchange.
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
	// todo : consider notation
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
	C := newReprAtMemory(2*size, Mem{Base: Load(Param("c"), RDI)}, RBX)
	for i := 0; i < size; i++ {
		C_red.next(_ITER).moveIfNotCF(*C_mont.next(_NO_ITER))
		C_mont.next(_ITER).moveTo(*C.next(_ITER), _ASSING)
	}
	tape.ret()
	RET()
}

func generateSquare(size int) {
	Implement(fmt.Sprintf("square%d", size))
	// todo
	RET()
}

func generateMontSquare(size int) {
	Implement(fmt.Sprintf("montsquare%d", size))
	// todo
	RET()
}
