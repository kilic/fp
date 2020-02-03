package x86

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func mul(size int) {
	funcName := "mul"
	TEXT(funcName, NOSPLIT, fmt.Sprintf("func(c *[%d]uint64, a, b *[%d]uint64)", size*2, size))
	commentHeader("inputs")
	tape := newTape(_NO_SWAP, ax.s, bx.s, dx.s)
	A := tape.newReprAtParam(size, "a", tape.di(), 0)
	B := tape.newReprAtParam(size, "b", tape.si(), 0)
	R := tape.newReprAllocGPRs(RSize)
	R.debug("R")
	assert(size < RSize*2+1, "only upto two partial multiplications is allowed")
	var W *repr
	if size > RSize {
		Wr := partialMulADX(tape, A, B, R).commentState("W right").debug("W right")
		tape.moveToStack(Wr).commentState("W right at stack").debug("W right at stack")
		Wl := partialMulADX(tape, A, B, R).commentState("W left").debug("W left")
		Wr.commentState("W right")
		Wl.setSwap(tape.ax())
		W = combinePartialResults(tape, Wr, Wl).commentState("W combined").debug("W combined")
	} else {
		W = partialMulADX(tape, A, B, R).commentState("W").debug("W")
	}
	C := tape.newReprAtParam(W.size, "c", tape.ax(), 0)
	W.setSwap(dx)
	W.updateIndex(0)
	for i := 0; i < W.size; i++ {
		W.next().move(C.next())
	}
	tape.ret()
}

func montMul(size int, fixedmod bool) {
	mulRSize := RSize
	funcName := "mmul"
	modulusName := "·modulus"
	if fixedmod {
		TEXT(funcName, NOSPLIT, fmt.Sprintf("func(c, a, b *[%d]uint64)", size))
	} else {
		TEXT(funcName, NOSPLIT, fmt.Sprintf("func(c, a, b, p *[%d]uint64, inp uint64)", size))
	}
	commentHeader("inputs")
	tape := newTape(_NO_SWAP, ax.s, bx.s, dx.s)
	A := tape.newReprAtParam(size, "a", tape.di(), 0)
	B := tape.newReprAtParam(size, "b", tape.si(), 0)

	// fix:
	// R := tape.newReprAllocGPRs(RSize)
	// R.debug("R")
	r := tape.newReprAllocRemainingGPRs()
	R := r.slice(0, mulRSize)
	R.debug("R")

	assert(size < RSize*2+1, "only upto two partial multiplications is allowed")
	var W *repr
	if size > RSize {
		Wr := partialMulADX(tape, A, B, R).commentState("W right").debug("W right")
		tape.moveToStack(Wr).commentState("W right at stack").debug("W right at stack")
		Wl := partialMulADX(tape, A, B, R).commentState("W left").debug("W left")
		Wr.commentState("W right")
		Wl.setSwap(tape.ax())
		W = combinePartialResults(tape, Wr, Wl).commentState("W combined").debug("W combined")
	} else {
		W = partialMulADX(tape, A, B, R).commentState("W").debug("W")
	}
	var modulus *repr
	var inp *limb
	if fixedmod {
		inp = newLimb(NewDataAddr(Symbol{Name: fmt.Sprintf("·inp")}, 0))
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: modulusName}, 0), 0)
	} else {
		inp = newLimb(NewParamAddr("inp", 32))
	}
	var montRsize int
	// var lCarry, hi *limb
	var hi *limb
	if size < 6 {
		montRsize = mulRSize
		tape.free(B.base.clone())
		if !fixedmod {
			p := tape.next().assertAtReg("there should be an idle register")
			comment("fetch modulus")
			modulus = tape.newReprAtParam(size, "p", p, 0)
		}
		hi = tape.bx()
	} else {
		tape.free(B.base, tape.bx())
		if fixedmod {
			montRsize = mulRSize + 1
			transitionMulToMont2(tape, W, 1)
		} else {
			montRsize = mulRSize
			transitionMulToMont2(tape, W, 2)
			comment("fetch modulus")
			p := tape.next().assertAtReg("should be spared at transition, p")
			modulus = tape.newReprAtParam(size, "p", p, 0)
		}
		hi = tape.next().assertAtReg("should be spared at transition, p")
	}
	W.commentState("W ready to mont").debug("ready to mont")
	// lCarry = W.at(0).clone()
	tape.setLimbForKey("inp", inp)
	tape.setLimbForKey("hi", hi)
	tape.setReprForKey("modulus", modulus)
	var lastBit *limb
	if montRsize >= size {
		montQ13(montRsize, tape, W).commentState("W montgomery reduction ends").debug("W montgomery reduction ends")
		tape.debug()
		tape.free(hi, ax)
		lastBit = tape.lookupLimb("long_long_carry").assertAtReg("must exist")
	}
	T, Red := modularReduction(tape, W, lastBit)
	T.debug("T")
	Red.debug("RED")
	out(tape, T, Red, lastBit)
	tape.ret()
}

func modularReduction(tape *tape, W *repr, lastBit *limb) (*repr, *repr) {
	commentHeader("modular reduction")
	modulus := tape.lookupRepr("modulus")
	modulus.updateIndex(0)
	size := W.size / 2
	Red := W.slice(size, size*2)
	var T *repr
	swap := tape.dx()
	// use swap register for last limb
	T = tape.newReprAlloc(Red.size - 1).extend(swap)
	// T = tape.newReprAlloc(Red.size)
	T.setSwap(swap)
	modulus.updateIndex(0)
	for i := 0; i < Red.size; i++ {
		T.next().loadSubSafe(
			Red.next(),
			modulus.next(),
			i != 0,
		)
	}
	SBBQ(U32(0), lastBit.s)
	return T, Red
}

func out(tape *tape, T, Red *repr, r *limb) {
	commentHeader("out")
	C := tape.newReprAtParam(Red.size, "c", r, 0)
	for i := 0; i < Red.size; i++ {
		T.next().moveIfNotCFAux(
			Red.next(), C.next())
	}
}

func transitionMulToMont2(tape *tape, W *repr, spare int) {
	hasAux := false
	ws := W.stackSlice()
	wr := W.registerSlice().reverse()
	for i := 0; i < ws.size; i++ {
		o := tape.next()
		if o.atReg() {
			comment("A")
			s := ws.next()
			tape.free(s)
			s.moveAssign(o)
			hasAux = true
		} else {
			comment("B")
			assert(hasAux, "transition should be done with auxilarry registers")
			r := wr.next()
			t := r.clone()
			r.moveAssign(o)
			s := ws.next()
			tape.free(s)
			s.moveAssign(t)
		}
	}
	spared := []*limb{}
	for i := 0; i < spare; i++ {
		comment("C")
		r := wr.next()
		s := tape.next()
		s.assertAtMem("sparing to stack")
		spared = append(spared, r.clone())
		r.moveAssign(s)
	}
	tape.free(spared...)
}
