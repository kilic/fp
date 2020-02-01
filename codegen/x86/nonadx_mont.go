package x86

import (
	"fmt"
)

// Mongomerry reduction Q1 and Q3
// For lower sized representations Q1 only is enough.
func montQ13NoADX(rsize int, tape *tape, W *repr) *repr {
	stack := tape.stack

	// get carries and constants
	modulus := tape.lookupRepr("modulus")
	inp := tape.lookupLimb("inp")
	u := tape.lookupLimb("u")
	sCarry := tape.lookupLimb("short_carry")
	iCarry := tape.dx()       // internal carry
	lCarry := W.get().clone() // use first limb as long carry

	// size of an operand
	size := modulus.size

	// make first non zero element the head
	W.adjustIndex()
	span := rsize // bound of inner iteration
	var bound int
	var saveU = false
	if size < rsize {
		commentHeader("montgomery reduction")
		bound = size
		span = size
		assert(W.i == 0, "this case should be allowed only for q1")
	} else if W.i == 0 {
		commentHeader("montgomery reduction q1")
		// q1, with missing span
		bound = rsize
		saveU = true // u is required for the next part of reduction
	} else {
		commentHeader("montgomery reduction q3")
		// q3, remaining part of q1
		bound = size - rsize
		saveU = true // u is required for the next part of reduction
	}
	wOffset := W.i

	for i := 0; i < bound; i++ { // we go (span x span) in q1
		firstI, lastI := i == 0, i == bound-1
		modulus.updateIndex(0)
		W.updateIndex(i + wOffset)
		/////////////////////
		commentI(i + wOffset)
		W.commentState("W")
		commentU(i + wOffset) // u = Wi * inp
		/////////////////////
		// calculate u
		W.mul(_NO_ITER, inp, u, nil, _MUL_MOVE)
		sCarry.clear() // clear short carry
		commentSeperator()
		if saveU { // save u to a stack for the next part of reductions
			comment(fmt.Sprintf("save u%d", i+wOffset))
			u.move(tape.setLimbForKey(fmt.Sprintf("u%d", i+wOffset), stack.next()))
		}
		for j := 0; j < span; j++ {
			firstJ, lastJ := j == 0, j == span-1
			/////////////////////
			commentJ(j)
			W.commentCurrent("w")
			/////////////////////
			w := W.next() // w_(i+j)
			if firstJ {
				// (w_(i+j), scarry) := mi * ui
				modulus.next().mul(u, w, sCarry, _MUL_ADD)
				// make a register idle when it goes to zero
				// special case at i = 0, for w0 which is used as long carry, we don't free this register
				tape.freeIf(!firstI, w.delete())
				continue
			}
			if !lastJ {
				modulus.next().mul(u, w, nil, _MUL_ADD)
				iCarry.addCarry()
				w.addNoCarry(sCarry)
				sCarry.clear().adc(iCarry)
			} else { // lastJ
				// w_(i+j+1)
				w2 := W.next()
				if firstI {
					// (w_(i+j), icarry) := mi * ui
					modulus.next().mul(u, w, nil, _MUL_ADD)
					iCarry.addCarry()
					w.add(sCarry, _NO_CARRY)
					w2.comment("w", W.i-1)
					w2.add(iCarry, _CARRY)
				} else {
					modulus.next().mul(u, w, lCarry, _MUL_ADD)
					w.add(sCarry, _NO_CARRY)
					// where register rotation happens
					// if next wi is at memory
					// bring it to a register that should have
					// been zeroed before
					if w2.atMem() {
						r := tape.free(w2).next().assertAtReg(
							fmt.Sprintf("a register must have been idle, %d, %d", i, j))
						comment("move to idle register")
						w2.moveTo(r, _ASSIGN)
					}
					// add long carry
					w2.comment("w", W.i-1).adc(lCarry)
				}
				// make long carry
				lCarry.clearIf(!firstI).addCarry()
			}
			_, _ = firstJ, lastJ // fix: remove declaration if not necesaary
		}
		_, _ = firstI, lastI // fix: remove declaration if not necessary
	}
	return W
}

// Mongomerry reduction Q2
func montQ2NoADX(rsize int, tape *tape, W *repr, lCarry *limb) *repr {
	commentHeader("montgomerry reduction q2")

	// get carries and constants
	modulus := tape.lookupRepr("modulus")
	llCarry := tape.lookupLimb("long_long_carry")
	sCarry := tape.lookupLimb("short_carry")
	iCarry := tape.dx() // internal carry

	// size of an operand
	size := modulus.size
	// bounds
	bound, span := rsize, size-rsize
	// high limbs of modulus will be used in q2
	modulusOffset := rsize

	// in first swap we will use an idle register
	// which has been used for 'u' value at q1
	firstSwap := true
	// bounds for swap operations
	// we will tolerate stack limb by 1
	spaceRequired := span + bound + 1
	stackRequired := spaceRequired - rsize - 3
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
		sCarry.clear() // clear short carry
		commentSeperator()
		for j := 0; j < span; j++ {
			firstJ, lastJ := j == 0, j == span-1
			/////////////////////
			commentJ(modulusOffset + j)
			W.commentCurrent("w")
			/////////////////////
			w := W.next()
			if firstJ {
				modulus.next().mul(u, w, sCarry, _MUL_ADD)
				// if required move to stack due to rotation of gp registers
				needStack := !firstI && (stackRequired > 0)
				if needStack {
					m := tape.stack.next()
					tape.free(w.clone())
					w.moveAssign(m)
					stackRequired -= 1
				}
				continue
			}
			if !lastJ {
				modulus.next().mul(u, w, nil, _MUL_ADD)
				iCarry.addCarry()
				w.addNoCarry(sCarry)
				sCarry.clear().adc(iCarry)
			} else {
				// w_(i+j+1)
				w2 := W.next()
				if firstI {
					modulus.next().mul(u, w, nil, _MUL_ADD)
					if span == rsize {
						comment("bring the carry from q1")
						iCarry.adc(llCarry)
					} else {
						iCarry.addCarry()
					}
					w.add(sCarry, _NO_CARRY)
					if w2.atMem() {
						comment("move to an idle register")
						r := tape.lookupLimb("u")
						w2.moveAssign(r)
						firstSwap = false
					}
					w2.comment("w", W.i-1).adc(iCarry)
					// :)
				} else {
					modulus.next().mul(u, w, lCarry, _MUL_ADD)
					w.add(sCarry, _NO_CARRY)
					if w2.atMem() {
						if firstSwap {
							// use 'u' from q1
							comment("move to an idle register")
							r := tape.lookupLimb("u")
							w2.moveAssign(r)
							firstSwap = false
						} else {
							if lastI {
								comment("tolarete this limb to stay in stack")
							} else {
								comment("move to an idle register")
								r := tape.next()
								assert(r.atReg(), "register expected")
								w2.moveAssign(r)
							}
						}
					}
					// add long carry
					k := (W.i - 1 + W.size) % W.size // mod
					w2.comment("w", k).adc(lCarry)
				}
				// make long carry
				if i == rsize-span-1 {
					// this is the point where we should inlude long-long-carry from q1
					comment("bring the carry from q1")
					llCarry.move(lCarry)
					lCarry.addCarry()
				} else {
					lCarry.clear().addCarry()
				}
			}
			_, _ = firstJ, lastJ // fix: remove if not necesaary
		}
		tape.free(u)
		_, _ = firstI, lastI // fix: remove if not necessary
	}
	return W
}

// Mongomerry reduction Q4
func montQ4NoADX(rsize int, tape *tape, W *repr, lCarry *limb) *repr {
	commentHeader("montgomerry reduction q4")

	// get carries and constants
	modulus := tape.lookupRepr("modulus")
	llCarry := tape.lookupLimb("long_long_carry")
	sCarry := tape.lookupLimb("short_carry")
	idle := tape.lookupLimb("u") // no need u anymore, then use in rotation
	iCarry := tape.dx()          // internal carry

	// size of an operand
	size := modulus.size
	// calculate bounds
	bound, span := size-rsize, size-rsize
	modulusOffset := rsize
	wOffset := 2 * rsize
	for i := 0; i < span; i++ {
		firstI, lastI := i == 0, i == bound-1
		W.updateIndex(i + wOffset)
		modulus.updateIndex(modulusOffset)
		/////////////////////
		commentI(i)
		W.commentState("W")
		/////////////////////
		u := tape.lookupLimb(fmt.Sprintf("u%d", i+modulusOffset))
		sCarry.clear() // clear short carry
		commentSeperator()
		for j := 0; j < span; j++ {
			firstJ, lastJ := j == 0, j == span-1
			/////////////////////
			commentJ(modulusOffset + j)
			W.commentCurrent("w")
			/////////////////////
			w := W.next() // w_(i+j)
			if firstJ {
				modulus.next().mul(u, w, sCarry, _MUL_ADD)
				if !firstI && !lastI {
					tape.moveAssign(w)
				}
				continue
			}
			if !lastJ {
				modulus.next().mul(u, w, nil, _MUL_ADD)
				iCarry.addCarry()
				w.addNoCarry(sCarry)
				sCarry.clear().adc(iCarry)
			} else {
				w2 := W.next().assertAtMem("expected at mem w2")
				if firstI {
					modulus.next().mul(u, w, nil, _MUL_ADD)
					iCarry.adc(llCarry)
					w.add(sCarry, _NO_CARRY)
					w2.moveAssign(idle)
					w2.comment("w", W.i-1)
					w2.add(iCarry, _CARRY)
				} else {
					modulus.next().mul(u, w, lCarry, _MUL_ADD)
					w.add(sCarry, _NO_CARRY)
					if lastI {
						comment("very last limb goes to short carry register")
						tape.free(w2)
						w2.moveAssign(sCarry)
					} else {
						r := tape.next()
						assert(r.atReg(), "register expected")
						tape.free(w2)
						w2.moveAssign(r)
					}
					w2.comment("w", W.i-1)
					w2.add(lCarry, _CARRY)
				}
				lCarry.clear().addCarry()
			}
			_, _ = firstJ, lastJ // fix: remove if not necesaary
		}
		_, _ = firstI, lastI // fix: remove if not necessary
	}
	return W
}

// Mongomery reduction Q2
// Handles the special case where rsize + 1 = size
func montQ2SpecialCaseNoADX(rsize int, tape *tape, W *repr, sCarry, lCarry *limb) *repr {
	commentHeader("montgomerry reduction q2")

	// get carries and constants
	modulus := tape.lookupRepr("modulus")
	iCarry := tape.dx() // internal carry

	// calculate bounds
	size := modulus.size
	offset, bound := size-1, size-1
	W.updateIndex(offset)

	// process where j = size - 1 only
	sCarry.clear()
	for i := 0; i < bound; i++ {
		firstI, lastI := i == 0, i == size-1
		W.updateIndex(offset + i)
		modulus.updateIndex(-1)
		/////////////////////
		commentI(i)
		W.commentCurrent("w")
		/////////////////////
		u := tape.lookupLimb(fmt.Sprintf("u%d", i))
		w1, w2 := W.next(), W.next()
		if firstI {
			modulus.next().mul(u, w1, w2, _MUL_ADD)
			sCarry.addCarry()
			continue
		}
		modulus.next().mul(u, w1, nil, _MUL_ADD)
		iCarry.adc(sCarry)
		sCarry.clear()
		w2.add(iCarry, _NO_CARRY)
		// may not work
		if i != size-3 {
			sCarry.addCarry()
		} else {
			comment("carry from q1")
			sCarry.adc(lCarry)
		}
		tape.free(u)
		_, _ = firstI, lastI
	}
	return W
}

// Mongomery reduction Q3
// Handles the special case where rsize + 1 = size
func montQ3SpecialCaseNoADX(rsize int, tape *tape, W *repr, sCarry, lCarry *limb) *repr {
	commentHeader("montgomery reduction q3")

	// get carries and constants
	modulus := tape.lookupRepr("modulus")
	u := tape.lookupLimb("u")
	inp := tape.lookupLimb("inp")
	iCarry := tape.dx() // internal carry

	// calculate bounds
	size := modulus.size
	offset, bound := size-1, size

	// process where i = size - 1 only
	W.updateIndex(offset)
	modulus.updateIndex(0)
	commentI(size - 1)
	commentU(size - 1) // u = Wi * inp
	W.mul(_NO_ITER, inp, u, nil, _MUL_MOVE)
	sCarry.clear() // clear short carry
	for j := 0; j < bound; j++ {
		firstJ, lastJ := j == 0, j == size-1
		W.updateIndex(offset + j)
		commentJ(j)
		W.commentCurrent("w")
		w := W.next()
		if firstJ {
			modulus.next().mul(u, w, sCarry, _MUL_ADD)
			tape.free(w.delete())
			continue
		}
		if !lastJ {
			modulus.next().mul(u, w, nil, _MUL_ADD)
			iCarry.addCarry()
			w.addNoCarry(sCarry)
			sCarry.clear().adc(iCarry)
		} else {
			assert(w.atMem(), "expected to be at stack")
			r := tape.next().assertAtReg(
				fmt.Sprintf("a register must have been idle"))
			comment("move to idle register")
			w.moveTo(r, _ASSIGN)
			w2 := W.next()
			modulus.next().mul(u, w, lCarry, _MUL_ADD)
			w.add(sCarry, _NO_CARRY)
			w2.comment("w", W.size-1)
			lCarry.adc(w2)
			w2.set(lCarry)
			comment("care the last bit")
			sCarry.clear().addCarry()
		}
		_, _ = firstJ, lastJ
	}
	return W
}

// func montQ3NoADX(rsize int, tape *tape, W *repr, inp *limb, modulus *repr, u, sCarry *limb) {
// 	commentHeader("montgomery reduction q3")
// 	stack := tape.stack
// 	size := modulus.size
// 	iCarry := tape.dx()       // internal carry
// 	lCarry := W.get().clone() // use first limb as long carry

// 	span := rsize // inner iteration bound
// 	bound := size - rsize
// 	offset := W.adjustIndex().i // start from non zero elements

// 	for i := 0; i < bound; i++ { // (size - rsize) x (rsize) in q3
// 		firstI, lastI := i == 0, i == bound-1
// 		W.updateIndex(offset + i)
// 		modulus.updateIndex(0)
// 		commentI(i + offset)
// 		W.commentState("W")
// 		commentU(i + offset) // u = Wi * inp
// 		// calculate u
// 		W.mul(_NO_ITER, inp, u, nil, _MUL_MOVE)
// 		sCarry.clear() // clear short carry
// 		// save u to a stack for the next part of reductions
// 		comment(fmt.Sprintf("save u%d", i+offset))
// 		u.move(tape.setLimbForKey(fmt.Sprintf("u%d", i+offset), stack.next()))
// 		commentHeader("")
// 		for j := 0; j < span; j++ {
// 			firstJ, lastJ := j == 0, j == span-1
// 			commentJ(j)
// 			W.commentCurrent("w")
// 			w := W.next() // w_(i+j)
// 			if firstJ {
// 				modulus.next().mul(u, w, sCarry, _MUL_ADD)
// 				tape.freeIf(!firstI, w.delete())
// 				continue
// 			}
// 			if !lastJ {
// 				modulus.next().mul(u, w, nil, _MUL_ADD)
// 				iCarry.addCarry()
// 				w.addNoCarry(sCarry)
// 				sCarry.clear().adc(iCarry)
// 			} else {
// 				w2 := W.next()
// 				if firstI {
// 					modulus.next().mul(u, w, nil, _MUL_ADD)
// 					iCarry.addCarry()
// 					w.add(sCarry, _NO_CARRY)
// 					w2.comment("w", W.i-1)
// 					w2.add(iCarry, _CARRY)
// 				} else {
// 					modulus.next().mul(u, w, lCarry, _MUL_ADD)
// 					w.add(sCarry, _NO_CARRY)
// 					// rotation
// 					if w2.atMem() {
// 						r := tape.free(w2).next().assertAtReg(
// 							fmt.Sprintf("a register must have been idle, %d, %d", i, j))
// 						comment("move to idle register")
// 						w2.moveTo(r, _ASSIGN)
// 					}
// 					w2.comment("w", W.i-1)
// 					w2.add(lCarry, _CARRY)
// 				}
// 				lCarry.clearIf(!firstI).addCarry()
// 			}
// 			_, _ = firstJ, lastJ // fix: remove declaration if not necessary
// 		}
// 		_, _ = firstI, lastI // fix: remove declaration if not necessary
// 	}
// }

// func montQ1NoADX(rsize int, tape *tape, W *repr, inp *limb, modulus *repr, u, sCarry *limb) {
// 	commentHeader("montgomerry reduction q1")
// 	stack := tape.stack
// 	W.updateIndex(0)
// 	size := modulus.size
// 	// iteration bound of q1 part
// 	bound := rsize
// 	if size < rsize {
// 		bound = size
// 	}
// 	// u is required for the next part of reduction
// 	// if q1 only is not enough for whole reduction
// 	saveU := size > rsize
// 	iCarry := tape.dx()          // internal carry
// 	lCarry := W.get().clone()    // use first limb as long carry
// 	for i := 0; i < bound; i++ { // we go (span x span) in q1
// 		firstI, lastI := i == 0, i == bound-1
// 		W.updateIndex(i)
// 		modulus.updateIndex(0)
// 		commentI(i)
// 		W.commentState("W")
// 		commentU(i) // u = Wi * inp
// 		// calculate u
// 		W.mul(_NO_ITER, inp, u, nil, _MUL_MOVE)
// 		sCarry.clear() // clear short carry
// 		commentHeader("")
// 		if saveU { // save u to a stack for the next part of reductions
// 			comment(fmt.Sprintf("save u%d", i))
// 			u.move(tape.setLimbForKey(fmt.Sprintf("u%d", i), stack.next()))
// 		}
// 		for j := 0; j < bound; j++ {
// 			firstJ, lastJ := j == 0, j == bound-1
// 			commentJ(j)
// 			W.commentCurrent("w")
// 			w := W.next() // w_(i+j)
// 			if firstJ {
// 				// (w_(i+j), scarry) := mi * ui
// 				modulus.next().mul(u, w, sCarry, _MUL_ADD)
// 				// make a register idle when it goes to zero
// 				// special case at i = 0, for w0 which is used as long carry, we don't free this register
// 				tape.freeIf(!firstI, w.delete())
// 				continue
// 			}
// 			if !lastJ {
// 				modulus.next().mul(u, w, nil, _MUL_ADD)
// 				iCarry.addCarry()
// 				w.addNoCarry(sCarry)
// 				sCarry.clear().adc(iCarry)
// 			} else { // lastJ
// 				// w_(i+j+1)
// 				w2 := W.next()
// 				if firstI {
// 					// (w_(i+j), icarry) := mi * ui
// 					modulus.next().mul(u, w, nil, _MUL_ADD)
// 					iCarry.addCarry()
// 					w.add(sCarry, _NO_CARRY)
// 					w2.comment("w", W.i-1)
// 					w2.add(iCarry, _CARRY)
// 				} else {
// 					modulus.next().mul(u, w, lCarry, _MUL_ADD)
// 					w.add(sCarry, _NO_CARRY)
// 					// where register rotation happens
// 					// if next wi is at memory
// 					// bring it to a register that should have
// 					// been zeroed before
// 					if w2.atMem() {
// 						r := tape.free(w2).next().assertAtReg(
// 							fmt.Sprintf("a register must have been idle, %d, %d", i, j))
// 						comment("move to idle register")
// 						w2.moveTo(r, _ASSIGN)
// 					}
// 					w2.comment("w", W.i-1)
// 					w2.add(lCarry, _CARRY)
// 				}
// 				lCarry.clearIf(!firstI).addCarry()
// 			}
// 			_, _ = firstJ, lastJ // fix: remove declaration if not necesaary
// 		}
// 		_, _ = firstI, lastI // fix: remove declaration if not necessary
// 	}
// }
