package x86

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

/* Warnings
last iter BX carry addition ok
*/

func mul816NoAdx(tape *tape, A, B, R, Stack *repr, bi, carry *limb, firstIt bool) *repr {
	sizeA := A.size
	sizeB := B.size
	stackSize := Stack.size
	// Schoolbook multiplication is applied
	// Lowest limbs (GPRs) are saved to stack after calculated.
	// Then those idle GPRs are used for higher limbs.
	for i := 0; i < sizeA; i++ {
		Commentf("| \n\n/*\ti = %d\t\t\t\t*/\n", i)
		Commentf("| b%d @ %s", i, bi.Asm())
		A.next(_ITER).moveTo(bi, _NO_ASSIGN)
		if i != 0 { // Carry is not used in first round
			carry.clear()
		}
		for j := 0; j < sizeB; j++ {
			Commentf("| a%d * b%d ", j, i)
			if stackSize > 0 {
				R.updateIndex(i + j - 1)
			} else {
				R.updateIndex(i + j)
			}
			Ra := R.next(_ITER)
			if i+j == 0 {
				Rb := R.next(_ITER)
				if stackSize > 0 {
					B.next(_ITER).mul(bi, Stack.next(_ITER), Rb, _MUL_MOVE)
				} else {
					B.next(_ITER).mul(bi, Ra, Rb, _MUL_MOVE)
				}
			} else {
				if i == sizeA-1 && j == sizeB-1 {
					B.next(_ITER).mul(bi, Ra, carry, _MUL_ADD)
				} else {
					Rb := R.next(_ITER)
					B.next(_ITER).mul(bi, Ra, Rb, _MUL_ADD)
				}
			}
			if i == 0 {
			} else {
				if j == 0 {
					Rc := R.next(_ITER)
					Rc.addCarry()
					carry.addCarry()
					if firstIt {
						if i < sizeA-1 {
							Ra.moveTo(Stack.next(_ITER), _NO_ASSIGN)
							Ra.clear()
						}
					} else {
						if i <= sizeB+2 {
							Ra.moveTo(Stack.next(_ITER), _NO_ASSIGN)
							Ra.clear()
						}
					}
				} else if j != 0 && j != sizeB-2 && j != sizeB-1 {
					Rc := R.next(_ITER)
					Rc.add(carry, _CARRY)
					carry.clear()
					carry.addCarry()
				} else if j == sizeB-2 {
					if i == sizeA-1 {
						carry.addCarry()
					} else {
						Rc := R.next(_ITER)
						Rc.add(carry, _CARRY)
					}
				}
			}
		}
	}
	comment("")
	var W *repr
	if firstIt {
		W = tape.newReprNoAlloc(sizeA + sizeB)
		Stack.updateIndex(0)
		comment("xxx")
		R.updateIndex(sizeA - 2)
		for i := 0; i < sizeA*2; i++ {
			W.next(_ITER).set(Stack.next(_ITER))
		}
		W.updateIndex(sizeA - 1)
		for i := 0; i < sizeB; i++ {
			w := W.next(_ITER)
			w.load(R.next(_ITER).s, nil)
		}
		W.next(_ITER).load(carry.s, nil)
	} else {
		size := sizeA + sizeB
		W = tape.newReprNoAlloc(size)
		Stack.updateIndex(0)
		R.updateIndex(sizeA - 1)
		for i := 0; i < size-R.size-1; i++ {
			W.next(_ITER).set(Stack.next(_ITER))
		}
		for i := 0; i < R.size; i++ {
			W.next(_ITER).set(R.next(_ITER))
		}
		W.next(_ITER).set(carry)
	}
	return W
}

/* Warnings
last iter BX carry addition

*/
func mont816NoAdxI0(tape *tape, W *repr, gpSize int, inp Op, modulus *repr, u, sCarry *limb, fixedModulus bool) *limb {
	stack := tape.stack
	size := modulus.size
	iCarry := newLimb(RDX, _NO_SWAP)
	lCarry := W.at(0).clone()
	WR := tape.newReprNoAlloc(size + gpSize + 1)
	var k int = 1
	// for i := 0; i < size*2; i++ {
	modulusOffset := 0
	for i := 0; i < size; i++ {
		Commentf("| \n\n/*\ti = %d\t\t\t\t*/\n", i)
		W.updateIndex(i)
		modulus.updateIndex(modulusOffset)
		W.mul(_NO_ITER, inp, u, nil, _MUL_MOVE)
		sCarry.clear()
		Commentf("|")
		firstI := i == 0
		idle := newLimbEmpty(nil)
		for j := 0; j < gpSize; j++ {
			lastJ := j == gpSize-1
			W.commentCurrent("w")
			if j == 0 {
				w := W.next(_ITER)
				idle.set(w)
				modulus.next(_ITER).mul(u, w, sCarry, _MUL_ADD)
				wr := WR.next(_ITER)
				if i >= gpSize {
					s := stack.next(_ALLOC)
					// s := Stack.next(_ITER)
					w.moveTo(s, _ASSIGN)
					wr.set(s.clone())
				} else {
					w.delete()
				}
			} else {
				w := W.next(_ITER)
				if !lastJ {
					modulus.next(_ITER).mul(u, w, nil, _MUL_ADD)
					iCarry.addCarry()
					w.add(sCarry, _NO_CARRY)
					sCarry.clear()
					sCarry.add(iCarry, _CARRY)
				} else {
					if firstI {
						modulus.next(_ITER).mul(u, w, nil, _MUL_ADD)
						iCarry.addCarry()
						w.add(sCarry, _NO_CARRY)
						// W.commentCurrent("w")
						w2 := W.get()
						if w2.atMem() {
							Commentf("| w%d @ %s", W.i, idle)
							Comment("| move to idle register")
							stack.free(w2)
							w2.moveTo(idle.clone(), _ASSIGN)
							k++
						} else {
							W.commentCurrent("w")
						}
						W.next(_ITER).add(iCarry, _CARRY)
						lCarry.clear().addCarry()
					} else {
						modulus.next(_ITER).mul(u, w, lCarry, _MUL_ADD)
						w.add(sCarry, _NO_CARRY)
						// where rotation happens
						w2 := W.get()
						if w2.atMem() {
							Commentf("| w%d @ %s", W.i, idle)
							Comment("| move to idle register")
							stack.free(w2)
							w2.moveTo(idle.clone(), _ASSIGN)
							k++
						} else {
							W.commentCurrent("w")
						}
						w2.add(lCarry, _CARRY)
						lCarry.clear().addCarry()
					}
				}
			}
		}
	}
	dCarry := stack.next(_ALLOC)
	lCarry.moveTo(dCarry, _NO_ASSIGN)
	return dCarry
}

func mont816NoAdxI1(tape *tape, W *repr, gpSize int, inp Op, modulus *repr, u, sCarry *limb, fixedModulus bool) *limb {
	stack := tape.stack
	size := modulus.size
	span := size - gpSize
	iCarry := newLimb(RDX, _NO_SWAP)
	lCarry := W.at(0).clone()
	WR := tape.newReprNoAlloc(size + gpSize + 1)
	var k int = 1
	// for i := 0; i < size*2; i++ {
	modulusOffset := gpSize
	for i := 0; i < size; i++ {
		Commentf("| \n\n/*\ti = %d\t\t\t\t*/\n", i)
		W.updateIndex(i)
		modulus.updateIndex(modulusOffset)
		W.mul(_NO_ITER, inp, u, nil, _MUL_MOVE)
		sCarry.clear()
		Commentf("|")
		firstI := i == 0
		idle := newLimbEmpty(nil)
		for j := 0; j < span; j++ {
			lastJ := j == span-1
			W.commentCurrent("w")
			if j == 0 {
				w := W.next(_ITER)
				idle.set(w)
				modulus.next(_ITER).mul(u, w, sCarry, _MUL_ADD)
				wr := WR.next(_ITER)
				if i >= gpSize {
					s := stack.next(_ALLOC)
					w.moveTo(s, _ASSIGN)
					wr.set(s.clone())
				} else {
					w.delete()
				}
			} else {
				w := W.next(_ITER)
				if !lastJ {
					modulus.next(_ITER).mul(u, w, nil, _MUL_ADD)
					iCarry.addCarry()
					w.add(sCarry, _NO_CARRY)
					sCarry.clear()
					sCarry.add(iCarry, _CARRY)
				} else {
					if firstI {
						modulus.next(_ITER).mul(u, w, nil, _MUL_ADD)
						iCarry.addCarry()
						w.add(sCarry, _NO_CARRY)
						w2 := W.get()
						if w2.atMem() {
							Commentf("| w%d @ %s", W.i, idle)
							Comment("| move to idle register")
							stack.free(w2)
							w2.moveTo(idle.clone(), _ASSIGN)
							k++
						} else {
							W.commentCurrent("w")
						}
						W.next(_ITER).add(iCarry, _CARRY)
						lCarry.clear().addCarry()
					} else {
						modulus.next(_ITER).mul(u, w, lCarry, _MUL_ADD)
						w.add(sCarry, _NO_CARRY)
						// where rotation happens
						w2 := W.get()
						if w2.atMem() {
							Commentf("| w%d @ %s", W.i, idle)
							Comment("| move to idle register")
							stack.free(w2)
							w2.moveTo(idle.clone(), _ASSIGN)
							k++
						} else {
							W.commentCurrent("w")
						}
						w2.add(lCarry, _CARRY)
						lCarry.clear().addCarry()
					}
				}
			}
		}
	}
	dCarry := stack.next(_ALLOC)
	lCarry.moveTo(dCarry, _NO_ASSIGN)
	return dCarry
}

// func mont816NoAdxdepr(tape *tape, W, Stack *repr, gpSize int, inp Op, modulus *repr, u, sCarry *limb, fixedModulus bool) *repr {
// 	size := modulus.size
// 	iCarry := newLimb(RDX, _NO_SWAP)
// 	lCarry := W.at(0).clone()
// 	WR := tape.newReprNoAlloc(size + gpSize + 1)
// 	var k int = 1
// 	// for i := 0; i < size*2; i++ {
// 	modulusOffset := 0
// 	for i := 0; i < size; i++ {
// 		Commentf("| \n\n/*\ti = %d\t\t\t\t*/\n", i)
// 		W.updateIndex(i)
// 		modulus.updateIndex(modulusOffset)
// 		W.mul(_NO_ITER, inp, u, nil, _MUL_MOVE)
// 		sCarry.clear()
// 		Commentf("|")
// 		firstI := i == 0
// 		idle := newLimbEmpty(nil)
// 		for j := 0; j < gpSize; j++ {
// 			lastJ := j == gpSize-1
// 			W.commentCurrent("w")
// 			if j == 0 {
// 				w := W.next(_ITER)
// 				idle.set(w)
// 				modulus.next(_ITER).mul(u, w, sCarry, _MUL_ADD)
// 				wr := WR.next(_ITER)
// 				if i >= gpSize {
// 					s := Stack.next(_ITER)
// 					w.moveTo(s, _ASSIGN)
// 					wr.set(s.clone())
// 				} else {
// 					w.delete()
// 				}
// 			} else {
// 				w := W.next(_ITER)
// 				if !lastJ {
// 					modulus.next(_ITER).mul(u, w, nil, _MUL_ADD)
// 					iCarry.addCarry()
// 					w.add(sCarry, _NO_CARRY)
// 					sCarry.clear()
// 					sCarry.add(iCarry, _CARRY)
// 				} else {
// 					if firstI {
// 						modulus.next(_ITER).mul(u, w, nil, _MUL_ADD)
// 						iCarry.addCarry()
// 						w.add(sCarry, _NO_CARRY)
// 						// W.commentCurrent("w")
// 						w2 := W.get()
// 						if w2.atMem() {
// 							Commentf("| w%d @ %s", W.i, W.at(k))
// 							Comment("| move to idle register")
// 							// w2.moveTo(W.at(k), _ASSIGN)
// 							w2.moveTo(idle.clone(), _ASSIGN)
// 							k++
// 						} else {
// 							W.commentCurrent("w")
// 						}
// 						W.next(_ITER).add(iCarry, _CARRY)
// 						lCarry.clear().addCarry()
// 					} else {
// 						modulus.next(_ITER).mul(u, w, lCarry, _MUL_ADD)
// 						w.add(sCarry, _NO_CARRY)
// 						// where rotation happens
// 						w2 := W.get()
// 						if w2.atMem() {
// 							Commentf("| w%d @ %s", W.i, W.at(k))
// 							Comment("| move to idle register")
// 							// w2.moveTo(W.at(k), _ASSIGN)
// 							w2.moveTo(idle.clone(), _ASSIGN)
// 							k++
// 						} else {
// 							W.commentCurrent("w")
// 						}
// 						w2.add(lCarry, _CARRY)
// 						lCarry.clear().addCarry()
// 					}
// 				}
// 			}
// 		}
// 	}

// 	return WR.updateIndex(gpSize)
// }
