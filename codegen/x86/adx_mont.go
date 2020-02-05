package x86

import "fmt"

func montQ13(rsize int, tape *tape, W *repr) *repr {
	stack := tape.stack
	// fetch constants
	modulus := tape.lookupRepr("modulus")
	inp := tape.lookupLimb("inp")
	hi := tape.lookupLimb("hi")
	lo := tape.ax()
	// 'dx' is one of multiplication operand of mulx
	// thus we will store 'u' at 'dx'
	// 'ax' will be used to get lower limb of mulx result
	dx, ax := tape.dx(), tape.ax()
	// each iteration creates a zero which is actuallly the lowest limb
	// we use zero to clear flags by xoring itself
	zero := ax.clone()
	var lCarry1 *limb
	var lCarry0 *limb

	size := modulus.size
	W.adjustIndex()
	// calculate bounds
	span := rsize
	var bound int
	var saveU = false
	if size < rsize+1 {
		commentHeader("montgomery reduction")
		bound = size
		span = size
		assert(W.i == 0, "this case should be allowed only for q1")
	} else if W.i == 0 {
		commentHeader("montgomery reduction q1")
		// q1, with missing span
		bound = rsize
		saveU = true // u is required for the next part of reduction
		// notice that a this is full capacity iteration
	} else {
		commentHeader("montgomery reduction q3")
		// q3, remaining part of q1
		bound = size - rsize
		saveU = true // u is required for the next part of reduction
	}
	// when span (inner iteration) size equals
	// to available register size we apply diffrent
	// long carry propagation trick
	fullCap := rsize == span
	wOffset := W.i
	comment("clear flags")
	zero.xorself()
	for i := 0; i < bound; i++ {
		firstI, lastI := i == 0, i == bound-1
		modulus.updateIndex(0)
		W.updateIndex(i + wOffset)
		/////////////////////
		commentI(i + wOffset)
		W.commentState("W")
		commentU(i + wOffset) // u = Wi * inp
		/////////////////////
		W.get().move(dx)
		// 'u' is stored at dx, 'hi' is just a placeholder here
		inp.mulx(dx, hi)
		if saveU {
			comment(fmt.Sprintf("save u%d", i+wOffset))
			s := stack.next()
			dx.move(tape.setLimbForKey(fmt.Sprintf("u%d", i+wOffset), s))
		}
		commentSeperator()
		for j := 0; j < span; j++ {
			firstJ, lastJ := j == 0, j == span-1
			W.updateIndex(i + j + wOffset)
			/////////////////////
			commentJ(j)
			W.commentCurrent("w")
			/////////////////////
			w1, w2 := W.next(), W.next()
			modulus.next().mulx(lo, hi)
			w1.adoxq(lo)

			tmp := false
			if lastJ {
				if w2.atMem() {
					W.commentPrevious("w")
					if fullCap {
						comment("move to temp register")
						tape.moveAssign(w2, lo)
						tmp = true
					} else {
						comment("move to an idle register")
						tape.moveAssignNext(w2).assertAtReg()
					}
				}
			}
			w2.adcxq(hi)
			if lastJ {
				w2.adoxq(lCarry0)
				tape.free(lCarry0)
				if tmp {
					comment("move to an idle register")
					W.commentPrevious("w")
					w2.moveAssign(tape.next().assertAtReg())
				}
				lCarry1.adcxq(lCarry1)
				if !lastI {
					comment("clear flags")
					lo.xorself()
				} else {
					// !!!! ????
					// lCarry1.adoxq(lCarry1)
				}
			}
			if firstJ {
				// swap and switch carries
				// use zero (w1) as next long carry
				if !firstI {
					lCarry0 = lCarry1.clone()
				} else {
					lCarry0 = w1.clone()
				}
				lCarry1 = w1.clone()
				w1.delete()
			}
			_, _ = firstJ, lastJ // fix: remove declaration if not necesaary
		}
		if lastI {
			// save this for carry for
			// last bit at modular reduction
			// or further quarters of mont reduction
			tape.setLimbForKey("long_long_carry", lCarry1)
			tape.alloc(lCarry1)
		}
		_, _ = firstI, lastI // fix: remove declaration if not necessary
	}
	return W
}

func montQ2SpecialCase(rsize int, tape *tape, W *repr, llCarry *limb) *repr {
	commentHeader("montgomerry reduction q2")
	modulus := tape.lookupRepr("modulus")

	size := modulus.size
	offset, bound := size-1, size-1
	W.updateIndex(offset)
	modulus.updateIndex(-1)
	modulus.next().move(tape.dx())
	ax.xorself()
	hi := tape.lookupLimb("hi")
	// process where j = size - 1
	for i := 0; i < bound; i++ {
		firstI, lastI := i == 0, i == bound-1
		W.updateIndex(offset + i)
		modulus.updateIndex(-1)
		/////////////////////
		commentI(i)
		W.commentCurrent("w")
		/////////////////////
		u := tape.lookupLimb(fmt.Sprintf("u%d", i))
		w1, w2 := W.next(), W.next()
		u.mulx(ax, hi)
		w1.adoxq(ax)
		if lastI {
			// very last element expected to be at memory
			w2.assertAtMem()
			comment("aggregate carries")
			comment(fmt.Sprintf("%v + %v should be added to w%d @ %v", llCarry, hi, W.i-1, w2))
			comment("notice that aggregated value can be at most (2^64 - 1)")
			llCarry.adcxq(hi)
			llCarry.adoxq(ax.clear())
		} else {
			w2.adcxq(hi)
		}
		_, _ = firstI, lastI
	}
	return W
}

func montQ3SpecialCase(rsize int, tape *tape, W *repr, llCarry *limb) *repr {
	commentHeader("montgomerry reduction q3 & q4")
	modulus := tape.lookupRepr("modulus")
	inp := tape.lookupLimb("inp")
	hi := tape.lookupLimb("hi")

	size := modulus.size
	offset, span := size-1, size

	W.updateIndex(offset)
	modulus.updateIndex(0)
	commentI(size - 1)
	commentU(size - 1) // u = Wi * inp
	ax.xorself()

	W.get().move(dx)
	// 'u' is stored at dx, 'hi' is just a placeholder here
	inp.mulx(dx, hi)
	for j := 0; j < span; j++ {
		firstJ, lastJ := j == 0, j == span-1
		W.updateIndex(offset + j)
		/////////////////////
		commentJ(j)
		W.commentCurrent("w")
		/////////////////////
		w1, w2 := W.next(), W.next()
		modulus.next().mulx(ax, hi)

		if !lastJ {
			w1.adoxq(ax)
		}

		if firstJ {
			tape.free(w1.delete())
		}

		if j < span-2 {
			w2.adcxq(hi)
		} else if j == span-2 {
			w2.assertAtMem()
			llCarry.adcxq(hi)
		} else {
			assert(lastJ, "should be last iter")
			w1.assertAtMem()
			w2.assertAtMem()
			// just above register for long long carry is freed
			r := tape.next().assertAtReg()
			llCarry.adoxq(ax)
			r.adcxq(hi)

			comment("the last bit")
			hi.clear()
			r.adoxq(hi)

			llCarry.addNoCarry(w1)
			r.adc(w2)
			hi.addCarry()
			tape.free(w1, w2)

			w1.set(llCarry)
			w2.set(r)
		}
		_, _ = firstJ, lastJ
	}
	tape.free(ax)
	return W
}

func montQ2(rsize int, tape *tape, W *repr, lCarry *limb) *repr {
	commentHeader("montgomerry reduction q2")

	modulus := tape.lookupRepr("modulus")
	hi := tape.lookupLimb("hi")
	llCarry := tape.lookupLimb("long_long_carry").assertAtMem()

	dx, ax := tape.dx(), tape.ax()
	size := modulus.size

	// bounds of upper and inner iteration
	bound, span := rsize, size-rsize
	assert(span < rsize, "q2 is not implemented for full capacity spanning")

	// high limbs of modulus will be used in q2
	modulusOffset := rsize

	spaceRequired := span + bound
	stackRequired := spaceRequired - bound
	comment("clear flags")
	lCarry.xorself()
	for i := 0; i < bound; i++ {
		firstI, lastI := i == 0, i == bound-1
		W.updateIndex(i + modulusOffset)
		modulus.updateIndex(modulusOffset)
		/////////////////////
		commentI(i)
		W.commentState("W")
		/////////////////////
		// fetch 'u' calculated at q1
		u := tape.lookupLimb(fmt.Sprintf("u%d", i))
		comment(fmt.Sprintf("u%d @ %v", i, u))
		u.move(dx)
		// free 'u' from memory
		tape.free(u)
		commentSeperator()
		for j := 0; j < span; j++ {
			firstJ, lastJ := j == 0, j == span-1
			W.updateIndex(i + j + modulusOffset)
			/////////////////////
			commentJ(modulusOffset + j)
			W.commentCurrent("w")
			/////////////////////
			w1, w2 := W.next(), W.next()
			modulus.next().mulx(ax, hi)
			w1.adoxq(ax)
			if firstJ {
				needStack := !firstI && stackRequired > 0
				if needStack {
					m := tape.stack.next()
					tape.free(w1.clone())
					w1.moveAssign(m)
					stackRequired -= 1
				}
			}
			if !lastJ {
				w2.adcxq(hi)
			} else {
				if w2.atMem() {
					W.commentPrevious("w")
					comment("move to an idle register")
					tape.moveAssignNext(w2).assertAtReg()
					W.commentPrevious("w")
				}
				w2.adcxq(hi).adoxq(lCarry)
				if i == rsize-span-1 {
					// at this point we bring the carry from q1
					comment("bring the carry from q1")
					llCarry.move(lCarry)
					lCarry.addCarry()
				} else {
					lCarry.clear().adcxq(lCarry)
				}
			}
			_, _ = firstJ, lastJ // fix: remove if not necesaary
		}
		if !lastI {
			comment("clear flags")
			ax.xorself()
		}
		_, _ = firstI, lastI // fix: remove if not necessary
	}
	return W
}

func montQ4(rsize int, tape *tape, W *repr, lCarry *limb) *repr {
	commentHeader("montgomerry reduction q4")

	modulus := tape.lookupRepr("modulus")
	hi := tape.lookupLimb("hi")
	lo := tape.ax()
	dx, ax := tape.dx(), tape.ax()
	size := modulus.size

	bound, span := size-rsize, size-rsize
	modulusOffset := rsize
	wOffset := 2 * rsize
	comment("clear flags")
	ax.xorself()
	for i := 0; i < bound; i++ {
		firstI, lastI := i == 0, i == bound-1
		W.updateIndex(i + wOffset)
		modulus.updateIndex(modulusOffset)
		/////////////////////
		commentI(i)
		W.commentState("W")
		/////////////////////
		// fetch 'u' calculated at q1
		u := tape.lookupLimb(fmt.Sprintf("u%d", i+modulusOffset))
		comment(fmt.Sprintf("u%d @ %v", i, u))
		u.move(dx)
		// free 'u' from memory
		tape.free(u)
		commentSeperator()
		for j := 0; j < span; j++ {
			firstJ, lastJ := j == 0, j == span-1
			W.updateIndex(i + j + wOffset)
			/////////////////////
			commentJ(modulusOffset + j)
			W.commentCurrent("w")
			/////////////////////
			w1, w2 := W.next(), W.next()
			modulus.next().mulx(lo, hi)
			w1.adoxq(lo)
			if firstJ {
				w2.adcxq(hi)
				if !lastI {
					// no free register is expected here
					tape.moveAssignNext(w1).assertAtMem()
				}
				continue
			}
			if !lastJ {
				w2.adcxq(hi)
			} else {
				W.commentPrevious("w")
				w2.assertAtMem()
				comment("move to an idle register")
				var r *limb
				if !lastI {
					r = tape.next().assertAtReg()
				} else {
					r = tape.ax().clone()
				}
				tape.moveAssign(w2, r)
				w2.adcxq(hi)
				if firstI {
					comment("bring carry from q2 & q3")
				}
				W.commentPrevious("w")
				w2.adoxq(lCarry)
				lCarry.clear().adcxq(lCarry)
			}

			_, _ = firstJ, lastJ // fix: remove if not necesaary
		}
		_, _ = firstI, lastI // fix: remove if not necessary
	}
	tape.free(hi)
	return W
}
