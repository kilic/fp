package x86

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func montMulNoADX(size int, fixedmod bool) {
	funcName := "mul"
	modulusName := "·modulus"
	if fixedmod {
		TEXT(funcName, NOSPLIT, fmt.Sprintf("func(c, a, b *[%d]uint64)", size))
	} else {
		TEXT(funcName, NOSPLIT, fmt.Sprintf("func(c, a, b, p *[%d]uint64, inp uint64)", size))
	}
	commentHeader("inputs")
	tape := newTape(_NO_SWAP, ax.s, dx.s)
	A := tape.newReprAtParam(size, "a", tape.di(), 0)
	B := tape.newReprAtParam(size, "b", tape.si(), 0)
	ai := tape.newLimb()
	carry := tape.newLimb()

	mulRSize := RSize - 1
	if size < 5 {
		mulRSize = size*2 - 1
	}
	R := tape.newReprAllocGPRs(mulRSize).debug("R")

	assert(size < RSize*2+1, "only upto two partial multiplications is allowed")
	var W *repr
	if size > RSize-1 {
		// for larger integers, multiplication is done in two parts
		// result of these parts are combined afterwards
		// calculate part 1
		Wr := partialMulNoADX(tape, A, B, R, ai, carry).commentState("W part 1 multiplication").debug("W part 1 multiplication")
		// move intermediate resut to stack
		tape.moveToStack(Wr).commentState("W part 1 moved to stack").debug("W part 1 moved to stack")
		// calculate part2
		Wl := partialMulNoADX(tape, A, B, R, ai, carry).commentState("W part 2 multiplication").debug("W part 2 multiplication")
		Wr.commentState("W part 1").setSwap(ax)
		// combine results
		W = combinePartialResults(tape, Wr, Wl).commentState("W combined").debug("W combined")
	} else {
		W = partialMulNoADX(tape, A, B, R, ai, carry).commentState("W").debug("mul end")
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
	var lCarry, sCarry, u *limb
	tape.free(A.base.clone(), B.base.clone(), ai)
	if fixedmod {
		transitionMulToMont2(tape, W, 2)
		sCarry = tape.next().assertAtReg()
		u = tape.next().assertAtReg()
		montRsize = mulRSize + 1
	} else {
		transitionMulToMont2(tape, W, 3)
		sCarry = tape.next().assertAtReg()
		u = tape.next().assertAtReg()
		p := tape.next().assertAtReg()
		comment("fetch modulus")
		modulus = tape.newReprAtParam(size, "p", p, 0)
		montRsize = mulRSize
	}
	lCarry = W.at(0).clone()
	tape.setLimbForKey("u", u)
	tape.setLimbForKey("short_carry", sCarry)
	tape.setLimbForKey("inp", inp)
	tape.setReprForKey("modulus", modulus)
	var lastBit *limb
	if montRsize >= size { // the case where only q1 part is enough
		montQ13NoADX(montRsize, tape, W).commentState("W montgomerry reduction ends").debug("W montgomery reduction ends")
		lastBit = lCarry.clone()
	} else {
		montQ13NoADX(montRsize, tape, W).commentState("W q1").debug("W q1")
		q2SpecialCase := (montRsize+1 == size)
		if q2SpecialCase {
			montQ2SpecialCaseNoADX(montRsize, tape, W, sCarry, lCarry).commentState("W q2").debug("W q2")
			// notice that long and short carry are switched
			montQ3SpecialCaseNoADX(montRsize, tape, W, lCarry, sCarry).commentState("W q3").debug("W q3")
			lastBit = lCarry.clone()
		} else {
			// long long carry from q1
			llCarry := tape.next().assertAtMem()
			comment("save the carry from q1")
			comment(fmt.Sprintf("should be added to w%d", W.i))
			lCarry.move(llCarry)
			tape.setLimbForKey("long_long_carry", llCarry)
			// Q2
			montQ2NoADX(montRsize, tape, W, lCarry).commentState("q2").debug("q2")
			// long long carry from q2
			comment("save the carry from q2")
			comment(fmt.Sprintf("should be added to w%d", W.i))
			lCarry.move(llCarry)
			// swapping to fit to q3 part
			spare := transitionQ2toQ3NoADX(W, lCarry)
			W.adjustIndex().commentState("W q2 q3 transition").debug("W q2 q3 transition")
			// u has been used in q2, need new one
			u.set(spare)
			tape.setLimbForKey("u", u)
			// Q3
			// long carry in q3 will the first non zero element of aggregated result
			lCarry.set(W.adjustIndex().get())
			montQ13NoADX(montRsize, tape, W).commentState("W q3").debug("W q3")
			// we must aggregate carries from q2 and q3
			// these two carries will be added to same limb
			comment("aggregate carries from q2 & q3")
			comment(fmt.Sprintf("should be added to w%d", W.i))
			llCarry.adc(lCarry)
			// Q4
			montQ4NoADX(montRsize, tape, W, lCarry).commentState("W q4").debug("W q4")
			lastBit = lCarry
		}
	}

	commentHeader("modular reduction")
	Red := W.slice(size, size*2)
	var T *repr
	T = tape.newReprAlloc(Red.size)
	T.setSwap(dx)
	modulus.updateIndex(0)
	for i := 0; i < Red.size; i++ {
		T.next().loadSubSafe(
			Red.next(),
			modulus.next(),
			i != 0,
		)
	}
	SBBQ(U32(0), lastBit.s)
	commentHeader("out")
	C := tape.newReprAtParam(Red.size, "c", lastBit, 0)
	for i := 0; i < Red.size; i++ {
		T.next().moveIfNotCFAux(
			Red.next(), C.next())
	}
	tape.ret()
	_ = lastBit
}

func transitionQ2toQ3NoADX(W *repr, aux *limb) *limb {
	W.adjustIndex().next()
	Ws := W.slice(W.i, W.size)
	countStack := func() (i int) {
		ws := Ws.clone()
		for ws.next().atMem() {
			i++
		}
		return i
	}
	bound := countStack()
	if bound == 0 {
		return aux
	}
	commentHeader("q2 q3 transition swap")
	R := Ws.registerSlice().reverse()
	var spare *limb
	for i := 0; i < bound; i++ {
		s := Ws.next()
		assert(s.atMem(), "expected to be at stack")
		t := s.clone()
		if i == 0 {
			spare = aux.clone()
		}
		s.moveAssign(spare)
		r := R.next()
		spare = r.clone()
		r.moveAssign(t)
	}
	return spare
}

func transitionMulToMont(tape *tape, W *repr, aux []*limb, spare int) []*limb {
	// this is kind of pesky too :(
	commentHeader("swap")
	W.updateIndex(0)
	var stackSize int
	for i := 0; i < W.size; i++ {
		if W.next().atReg() {
			stackSize = i
			break
		}
	}
	spared := []*limb{}
	W.updateIndex(0)
	regSize := W.size - stackSize
	if regSize != W.size {
		regs := W.slice(stackSize, W.size)
		regs.previous()
		// ******
		limit := len(aux) + regSize - spare
		if stackSize < limit {
			limit = stackSize
		}
		d := limit + spare
		// ******
		for i := 0; i < d; i++ {
			if i < len(aux) { // A
				fmt.Println("xxx")
				Comment("A")
				w := W.next()
				tape.free(w.clone())
				w.moveTo(aux[i], _ASSIGN)
			} else if i < limit { // B
				fmt.Println("yyy")
				Comment("B")
				r := regs.get().clone()
				s := tape.stack.next()
				regs.previous().moveTo(s, _ASSIGN)
				w := W.next()
				tape.free(w.clone())
				w.moveTo(r, _ASSIGN)
			} else { // C
				fmt.Println("zzz")
				Comment("C")
				r := regs.previous()
				spared = append(spared, r.clone())
				s := tape.stack.next()
				r.moveTo(s, _ASSIGN)
			}
		}
	}
	W.updateIndex(0)
	return spared
}

func combinePartialResults(tape *tape, Wr, Wl *repr) *repr {
	size := Wr.size
	W := tape.newReprNoAlloc(size)
	Wr.updateIndex(0)
	Wl.setSwap(tape.ax()).updateIndex(0)
	car := false
	for i := 0; i < size; i++ {
		wr, wl, w := Wr.next(), Wl.next(), W.next()
		if wl.isEmpty() {
			w.set(wr.clone())
			continue
		}
		if !wr.isEmpty() {
			wl.add(wr, car)
			car = true
			tape.free(wr)
		} else {
			wl.addCarry()
		}
		w.set(wl)
	}
	return W
}
