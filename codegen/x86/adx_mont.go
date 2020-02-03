package x86

import "fmt"

func montQ13(rsize int, tape *tape, W *repr) *repr {
	stack := tape.stack
	// fetch constants
	modulus := tape.lookupRepr("modulus")
	inp := tape.lookupLimb("inp")
	hi := tape.lookupLimb("hi")
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
	fullCap := rsize == bound
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
			dx.move(tape.setLimbForKey(fmt.Sprintf("u%d", i+wOffset), stack.next()))
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
			modulus.next().mulx(ax, hi)
			w1.adoxq(ax)

			tmp := false
			if lastJ {
				if w2.atMem() {
					tape.free(w2)
					W.commentPrevious("w")
					if fullCap {
						comment("move to temp register")
						w2.moveAssign(tape.ax())
						tmp = true
					} else {
						comment("move to an idle register")
						w2.moveAssign(tape.next().assertAtReg("there should be an idle register"))
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
					w2.moveAssign(tape.next().assertAtReg("there should be an idle register"))
				}
				lCarry1.adcxq(lCarry1)
				if !lastI {
					comment("clear flags")
					ax.xorself()
				} else {
					// !!!! ????
					// lCarry1.adoxq(lCarry1)
				}
			}
			if firstJ {
				if !firstI {
					lCarry0 = lCarry1.clone()
				} else {
					lCarry0 = w1.clone()
				}
				lCarry1 = w1.clone()
				w1.delete()
				// make a register idle when it goes to zero
				// special case at i = 0, for w0 which is used as long carry, we don't free this register
				// tape.freeIf(!firstI, w1.delete())
				// tape.freeIf(!lastI, w1.delete())
			}
			_, _ = firstJ, lastJ // fix: remove declaration if not necesaary
		}
		if lastI {
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
			w2.assertAtMem("very last element expected to be at memory")
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
			// w2.assertAtMem("w2, span - 2")
			// w2.moveAssign(tape.next().assertAtReg("at fisrt j a register went to zero"))
			// comment("add aggregated carry")
			// w2.adcxq(hi)
			// tape.free(llCarry)
			w2.assertAtMem("w2, span - 2")
			llCarry.adcxq(hi)
		} else {
			assert(lastJ, "should be last iter")
			w1.assertAtMem("w1, last j")
			w2.assertAtMem("w2, last j")
			r := tape.next().assertAtReg("just above long long carry is freed")
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
			// w2.moveAssign(tape.next().assertAtReg("just above long long carry is freed"))
			// w2.adcxq(gi)
		}
		_, _ = firstJ, lastJ
	}
	tape.free(ax)
	return W
}

// func montQ13(rsize int, tape *tape, W *repr) *repr {
// 	stack := tape.stack
// 	// fetch constants
// 	modulus := tape.lookupRepr("modulus")
// 	inp := tape.lookupLimb("inp")
// 	hi := tape.lookupLimb("hi")
// 	// 'dx' is one of multiplication operand of mulx
// 	// thus we will store 'u' at 'dx'
// 	// 'ax' will be used to get lower limb of mulx result
// 	dx, ax := tape.dx(), tape.ax()
// 	// each iteration creates a zero which is actuallly the lowest limb
// 	// we use zero to clear flags by xoring itself
// 	zero := ax.clone()
// 	var lCarry *limb

// 	size := modulus.size
// 	W.adjustIndex()
// 	// calculate bounds
// 	span := rsize
// 	var bound int
// 	var saveU = false
// 	if size < rsize+1 {
// 		commentHeader("montgomery reduction")
// 		bound = size
// 		span = size
// 		assert(W.i == 0, "this case should be allowed only for q1")
// 	} else if W.i == 0 {
// 		commentHeader("montgomery reduction q1")
// 		// q1, with missing span
// 		bound = rsize
// 		saveU = true // u is required for the next part of reduction
// 	} else {
// 		commentHeader("montgomery reduction q3")
// 		// q3, remaining part of q1
// 		bound = size - rsize
// 		saveU = true // u is required for the next part of reduction
// 	}
// 	wOffset := W.i
// 	comment("clear flags")
// 	zero.xorself()
// 	for i := 0; i < bound; i++ {
// 		firstI, lastI := i == 0, i == bound-1
// 		modulus.updateIndex(0)
// 		W.updateIndex(i + wOffset)
// 		/////////////////////
// 		commentI(i + wOffset)
// 		W.commentState("W")
// 		commentU(i + wOffset) // u = Wi * inp
// 		/////////////////////
// 		W.get().move(dx)
// 		// 'u' is stored at dx, 'hi' is just a placeholder here
// 		inp.mulx(dx, hi)
// 		if saveU {
// 			comment(fmt.Sprintf("save u%d", i+wOffset))
// 			dx.move(tape.setLimbForKey(fmt.Sprintf("u%d", i+wOffset), stack.next()))
// 		}
// 		commentSeperator()
// 		for j := 0; j < span; j++ {
// 			firstJ, lastJ := j == 0, j == span-1
// 			W.updateIndex(i + j + wOffset)
// 			/////////////////////
// 			commentJ(j)
// 			W.commentCurrent("w")
// 			/////////////////////
// 			w1, w2 := W.next(), W.next()
// 			modulus.next().mulx(ax, hi)
// 			w1.adoxq(ax)

// 			if w2.atMem() {
// 				assert(lastJ, "this should occur in last j")
// 				comment("move to an idle register")
// 				tape.free(w2)
// 				w2.moveAssign(tape.next().assertAtReg("there should be an idle register"))
// 			}

// 			w2.adcxq(hi)
// 			if lastJ {
// 				w2.adoxq(lCarry)
// 				lCarry.clearIf(!firstI).adcxq(lCarry)
// 				if !lastI {
// 					comment("clear flags")
// 					ax.xorself()
// 				} else {
// 					lCarry.adoxq(lCarry)
// 				}
// 			}
// 			if firstJ {
// 				if firstI {
// 					lCarry = w1.clone()
// 				} else {
// 					zero.set(w1.clone())
// 				}
// 				// make a register idle when it goes to zero
// 				// special case at i = 0, for w0 which is used as long carry, we don't free this register
// 				tape.freeIf(!firstI, w1.delete())
// 			}
// 			_, _ = firstJ, lastJ // fix: remove declaration if not necesaary
// 		}
// 		if lastI {
// 			tape.setLimbForKey("long_long_carry", lCarry)
// 			tape.alloc(lCarry)
// 		}
// 		_, _ = firstI, lastI // fix: remove declaration if not necessary
// 	}
// 	return W
// }
