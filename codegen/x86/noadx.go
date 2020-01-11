package x86

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func genMontMulNoAdx(size int, fixedmod bool, single bool) {
	/*
	   ("func mul%d(c *[%d]uint64, a, b *Fe%d)\n\n", i, i*2, i*64)
	*/
	if size < 4 {
		panic("not implemented")
	} else if size >= 4 || size < 9 {
		genMontMul48NoAdx(size, fixedmod, single)
	} else {
		panic("not implemented")
	}
}

func mul48NoAdx(tape *tape, A, B, R, Stack *repr, bi, carry *limb) *repr {
	size := A.size
	stackSize := Stack.size
	// Schoolbook multiplication is applied
	// Lowest limbs (GPRs) are saved to stack after calculated.
	// Then those idle GPRs are used for higher limbs.
	for i := 0; i < size; i++ {
		Commentf("| \n\n/*\ti = %d\t\t\t\t*/\n", i)
		Commentf("| b%d @ %s", i, bi.Asm())
		B.next(_ITER).moveTo(bi, _NO_ASSIGN)
		if i != 0 { // Carry is not used in first round
			carry.clear()
		}
		for j := 0; j < size; j++ {
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
					A.next(_ITER).mul(bi, Stack.next(_ITER), Rb, _MUL_MOVE)
				} else {
					A.next(_ITER).mul(bi, Ra, Rb, _MUL_MOVE)
				}
			} else {
				if i == size-1 && j == size-1 { // Very last multiplication
					A.next(_ITER).mul(bi, Ra, carry, _MUL_ADD)
				} else {
					Rb := R.next(_ITER)
					A.next(_ITER).mul(bi, Ra, Rb, _MUL_ADD)
				}
			}
			if i == 0 {
				// No carry operation is involved in first round
			} else {
				if j == 0 {
					Rc := R.next(_ITER)
					Rc.addCarry()
					carry.addCarry()
					if i < 2*size-9 {
						Ra.moveTo(Stack.next(_ITER), _NO_ASSIGN)
						Ra.clear()
					}
				} else if j != 0 && j != size-2 && j != size-1 {
					Rc := R.next(_ITER)
					Rc.add(carry, _CARRY)
					carry.clear()
					carry.addCarry()
				} else if j == size-2 {
					if i == size-1 {
						carry.addCarry()
					} else {
						Rc := R.next(_ITER)
						Rc.add(carry, _CARRY)
					}
				}
			}
		}
	}
	// W is 2n sized output
	W := tape.newReprNoAlloc(size * 2)
	if stackSize < 1 {
		R.updateIndex(0)
	}
	// Limbs at stack are lowest ones
	for i := 0; i < stackSize; i++ {
		W.next(_ITER).set(Stack.next(_ITER))
	}
	if stackSize < 1 {
		for i := 0; i < R.size-1; i++ {
			W.next(_ITER).set(R.next(_ITER))
		}
	} else {
		for i := 0; i < R.size; i++ {
			W.next(_ITER).set(R.next(_ITER))
		}
	}
	W.next(_ITER).set(carry)
	return W
}

func mont48NoAdx(tape *tape, W *repr, inp Op, modulus *repr, u, sCarry, lCarry *limb, fixedModulus bool) {
	// Check for expected size of double-precision input number
	if W.size%2 != 0 || W.size > 16 {
		panic("")
	}
	size := W.size / 2
	iCarry := newLimb(RDX, _NO_SWAP)
	var k int
	if W.at(0).atMem() || W.at(0).Asm() == lCarry.Asm() {
		k = 1
	}
	for i := 0; i < size; i++ {
		Commentf("| \n\n/*\ti = %d\t\t\t\t*/\n", i)
		W.updateIndex(i)
		W.mul(_NO_ITER, inp, u, nil, _MUL_MOVE)
		sCarry.clear()
		Commentf("|")
		firstI := i == 0
		for j := 0; j < size; j++ {
			lastJ := j == size-1
			if j == 0 {
				modulus.next(_ITER).mul(u, W.next(_ITER), sCarry, _MUL_ADD)
			} else {
				W.commentCurrent("w")
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
							Commentf("| w%d @ %s", W.i, W.at(k))
							Comment("| move to emptied register")
							w2.moveTo(W.at(k), _ASSIGN)
							k++
						} else {
							W.commentCurrent("w")
						}
						W.next(_ITER).add(iCarry, _CARRY)
						lCarry.clear()
						lCarry.addCarry()
					} else {
						modulus.next(_ITER).mul(u, w, lCarry, _MUL_ADD)
						w.add(sCarry, _NO_CARRY)
						// where rotation happens
						w2 := W.get()
						if w2.atMem() {
							Commentf("| w%d @ %s", W.i, W.at(k))
							Comment("| move to emptied register")
							w2.moveTo(W.at(k), _ASSIGN)
							k++
						} else {
							W.commentCurrent("w")
						}
						w2.add(lCarry, _CARRY)
						lCarry.clear()
						lCarry.addCarry()
					}
				}
			}
		}
	}
	comment("reduction")
	C_red := W.slice(size, size*2)
	tape.freeAll()
	tape.reserveGp(C_red.ops()...)
	tape.reserveGp(lCarry.s)
	if !fixedModulus {
		tape.reserveGp(modulus.base)
	}
	tape.swap = tape.newLimb().s.(GPPhysical)
	T := tape.newReprAlloc(C_red.size)
	modulus.updateIndex(0)
	for i := 0; i < C_red.size; i++ {
		T.next(_ITER).loadSubSafe(
			*C_red.next(_ITER),
			*modulus.next(_ITER),
			i != 0,
		)
	}
	SBBQ(U32(0), lCarry.s)
	Commentf("|")
	C := tape.newReprAtParam(C_red.size, "c", lCarry.s.(Register))
	for i := 0; i < C_red.size; i++ {
		T.next(_ITER).moveIfNotCFAux(
			*C_red.next(_ITER),
			*C.next(_ITER))
	}
}

func genMontMul48NoAdx(size int, fixedmod bool, single bool) {
	funcName := "mul"
	modulusName := "·modulus"
	if !single {
		funcName = fmt.Sprintf("%s%d", funcName, size)
		modulusName = fmt.Sprintf("%s%d", modulusName, size)
	}
	if fixedmod {
		TEXT(funcName, NOSPLIT, fmt.Sprintf("func(c *[%d]uint64, a, b *[%d]uint64)", size*2, size))
	} else {
		TEXT(funcName, NOSPLIT, fmt.Sprintf("func(c *[%d]uint64, a, b, p *[%d]uint64, inp uint64)", size*2, size))
	}
	comment("inputs")
	tape := newTape(_NO_SWAP, mlo, mhi)
	A := tape.newReprAtParam(size, "a", RDI)
	B := tape.newReprAtParam(size, "b", RSI)

	// Expect all GPRs free,
	// expect, DI and SI allocated for inputs
	// AX, DX allocated for multiplication
	if tape.sizeFreeGp() != 10 {
		panic("")
	}

	// `bi` is allocd for limb of a second operand(input b)
	bi := tape.newLimb()
	// `carry` will be assigned to last limb of mul result
	carry := tape.newLimb()

	R := tape.newReprAllocRemainingGPRs()
	// Size of r must be equal 8.
	// Registers named R8 ... R15 are expected.
	if R.size != 8 {
		panic("")
	}

	// Allocate stack if limb size is larger than 4.
	// Otherwise we don't need stack space.
	// We mostly use (stackSize > 0) control to apply logic for (limbSize > 4)
	stackSize := 2*size - 9
	if stackSize < 0 {
		stackSize = 0
	}
	Stack := tape.allocStack(stackSize)

	// Do zero GPRs
	for i := 0; i < R.size; i++ {
		r := R.next(_ITER)
		if (i != 0 && i != 1) || (stackSize > 0 && i == 1) {
			r.clear()
		}
	}

	W := mul48NoAdx(tape, A, B, R, Stack, bi, carry)
	if W.size != 2*size {
		panic("")
	}

	var modulus *repr
	var inp Mem // fix: inp to limb type
	if fixedmod {
		inp = NewDataAddr(Symbol{Name: fmt.Sprintf("·inp")}, 0)
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: modulusName}, 0))
	} else {
		inp = NewParamAddr("inp", 32)
	}

	comment("swap")
	var lCarry, sCarry, u *limb
	switch size {
	case 4:
		if !fixedmod {
			modulus = tape.newReprAtParam(size, "p", R.at(R.size-1).s.(Register))
		}
		lCarry = newLimb(A.base, nil)
		sCarry = newLimb(B.base, nil)
		u = bi
	case 5:
		lCarry = newLimb(A.base, nil)
		W.updateIndex(0)
		W.next(_ITER).moveTo(lCarry, _ASSIGN)
		sCarry = newLimb(B.base, nil)
		u = newLimb(bi.s, nil)
		if !fixedmod {
			w := W.next(_ITER)
			t := newLimb(w.s, nil)
			w.moveTo(Stack.next(_ITER), _ASSIGN)
			modulus = tape.newReprAtParam(size, "p", t.s.(Register))
		}
	case 6:
		lCarry = newLimb(A.base, nil)
		Stack.updateIndex(0)
		W.next(_ITER).moveTo(lCarry, _ASSIGN)
		W.next(_ITER).moveTo(B.base, _ASSIGN)
		W.next(_ITER).moveTo(bi, _ASSIGN)
		if !fixedmod {
			W.updateIndex(9)
		} else {
			W.updateIndex(10)
		}
		w := W.next(_ITER)
		sCarry = newLimb(w.s, nil)
		w.moveTo(Stack.next(_ITER), _ASSIGN)
		//
		w = W.next(_ITER)
		u = newLimb(w.s, nil)
		w.moveTo(Stack.next(_ITER), _ASSIGN)
		//
		if !fixedmod {
			w := W.next(_ITER)
			t := newLimb(w.s, nil)
			w.moveTo(Stack.next(_ITER), _ASSIGN)
			modulus = tape.newReprAtParam(size, "p", t.s.(Register))
		}
	case 7:
		lCarry = newLimb(A.base, nil)
		Stack.updateIndex(0)
		W.next(_ITER).moveTo(lCarry, _ASSIGN)
		W.next(_ITER).moveTo(B.base, _ASSIGN)
		W.next(_ITER).moveTo(bi, _ASSIGN)
		w3 := W.next(_ITER)
		w4 := W.next(_ITER)
		if !fixedmod {
			W.updateIndex(9)
		} else {
			W.updateIndex(10)
		}
		w10 := W.next(_ITER)
		t := newLimb(w10.s, nil)
		w10.moveTo(Stack.next(_ITER), _ASSIGN)
		w3.moveTo(t, _ASSIGN)

		w11 := W.next(_ITER)
		t = newLimb(w11.s, nil)
		w11.moveTo(Stack.next(_ITER), _ASSIGN)
		w4.moveTo(t, _ASSIGN)

		w := W.next(_ITER)
		sCarry = newLimb(w.s, nil)
		w.moveTo(Stack.next(_ITER), _ASSIGN)

		w = W.next(_ITER)
		u = newLimb(w.s, nil)
		w.moveTo(Stack.next(_ITER), _ASSIGN)

		if !fixedmod {
			w := W.next(_ITER)
			t := newLimb(w.s, nil)
			w.moveTo(Stack.next(_ITER), _ASSIGN)
			modulus = tape.newReprAtParam(size, "p", t.s.(Register))
		}
	case 8:
		lCarry = newLimb(A.base, nil)
		Stack.updateIndex(0)
		W.next(_ITER).moveTo(lCarry, _ASSIGN)
		W.next(_ITER).moveTo(B.base, _ASSIGN)
		W.next(_ITER).moveTo(bi, _ASSIGN)
		w3 := W.next(_ITER)
		w4 := W.next(_ITER)
		w5 := W.next(_ITER)
		w6 := W.next(_ITER)
		if !fixedmod {
			W.updateIndex(9)
		} else {
			W.updateIndex(10)
		}
		w10 := W.next(_ITER)
		t := w10.clone()
		w10.moveTo(Stack.next(_ITER), _ASSIGN)
		w3.moveTo(t, _ASSIGN)

		w11 := W.next(_ITER)
		t = w11.clone()
		w11.moveTo(Stack.next(_ITER), _ASSIGN)
		w4.moveTo(t, _ASSIGN)

		w12 := W.next(_ITER)
		t = w12.clone()
		w12.moveTo(Stack.next(_ITER), _ASSIGN)
		w5.moveTo(t, _ASSIGN)

		w13 := W.next(_ITER)
		t = w13.clone()
		w13.moveTo(Stack.next(_ITER), _ASSIGN)
		w6.moveTo(t, _ASSIGN)

		w := W.next(_ITER)
		sCarry = newLimb(w.s, nil)
		w.moveTo(Stack.next(_ITER), _ASSIGN)

		w = W.next(_ITER)
		u = newLimb(w.s, nil)
		w.moveTo(Stack.next(_ITER), _ASSIGN)

		if !fixedmod {
			w := W.next(_ITER)
			t := newLimb(w.s, nil)
			w.moveTo(Stack.next(_ITER), _ASSIGN)
			modulus = tape.newReprAtParam(size, "p", t.s.(Register))
		}
	}
	W.updateIndex(0)
	mont48NoAdx(tape, W, inp, modulus, u, sCarry, lCarry, fixedmod)
	tape.ret()
	RET()
	comment("end")
}

// func reduceNoAdx(C_red *repr, modulus *repr) {
// 	comment("reduction")
// 	swap := RSI
// 	tape := newTape(swap, C_red.ops()...)
// 	tape.reserveGp(RDI)
// 	T := tape.newReprAlloc(C_red.size)
// 	modulus.updateIndex(0)
// 	for i := 0; i < C_red.size; i++ {
// 		T.next(_ITER).loadSubSafe(
// 			*C_red.next(_ITER),
// 			*modulus.next(_ITER),
// 			i != 0,
// 		)
// 	}
// 	C := tape.newReprAtParam(C_red.size, "c", RDI)
// 	for i := 0; i < C_red.size; i++ {
// 		T.next(_ITER).moveIfNotCFAux(*C_red.next(_ITER), *C.next(_ITER))
// 	}
// }

// func stitchMul(size int, gpsize int) {
// 	TEXT(fmt.Sprintf("mul%d", size), NOSPLIT, fmt.Sprintf("func(c *[%d]uint64, a, b *[%d]uint64)", size*2, size))
// 	Commentf("|")
// 	bi := RCX
// 	tape := newTape(_NO_SWAP, MLO, MHI, bi, RBX)
// 	A := tape.newReprAtParam(size, "a", RDI)
// 	B := tape.newReprAtParam(size, "b", RSI)
// 	R := tape.newReprAlloc(gpsize)
// 	carry := newLimb(RBX, nil)
// 	Commentf("|")
// 	for i := 0; i < R.size; i++ {
// 		r := R.next(_ITER)
// 		if i != 0 {
// 			r.clear()
// 		}
// 	}
// 	n := 2
// 	Stack := make([]*repr, n)
// 	Stack[0] = tape.allocStack(size + gpsize)
// 	Stack[1] = tape.allocStack(size) // todo fix
// 	var mostRecentlyMovedToStackIndex int
// 	for k := 0; k < n; k++ {
// 		span := gpsize
// 		if k == n-1 {
// 			span = size - gpsize
// 		}
// 		Commentf("| \n\n/*\tBlock %d\t\t\t\t*/\n", k)
// 		S := Stack[k]
// 		for i := 0; i < size; i++ {
// 			Commentf("| \n\n/*\ti = %d\t\t\t\t*/\n", i)
// 			Commentf("| b%d @ CX", i)
// 			if span == 1 {
// 				if i == 0 {
// 					A.next(_ITER).moveTo(bi, _NO_ASSIGN)
// 				}
// 			} else {
// 				B.next(_ITER).moveTo(bi, _NO_ASSIGN)
// 			}
// 			if i != 0 && span > 2 {
// 				carry.clear()
// 			}
// 			A.updateIndex(gpsize * k)
// 			for j := 0; j < span; j++ {
// 				Commentf("| a%d * b%d ", j, i)
// 				R.updateIndex(i + j - 1)
// 				indexOfRi := R.i
// 				Ra, Rb := R.next(_ITER), R.next(_ITER)
// 				if span == 1 {
// 					if i+j == 0 {
// 						B.next(_ITER).mul(bi, S.next(_ITER), Rb, _MUL_MOVE)
// 					} else {
// 						B.next(_ITER).mul(bi, Ra, Rb, _MUL_ADD)
// 					}
// 				} else {
// 					if i+j == 0 {
// 						A.next(_ITER).mul(bi, S.next(_ITER), Rb, _MUL_MOVE)
// 					} else {
// 						A.next(_ITER).mul(bi, Ra, Rb, _MUL_ADD)
// 					}
// 				}
// 				if i == 0 {
// 					// No carry operation is involved in first round
// 				} else {
// 					if j == 0 && !(i == size-1 && span < 2) {
// 						Rc := R.next(_ITER)
// 						Rc.addCarry()
// 						if span > 2 {
// 							carry.addCarry()
// 						}
// 						if i < span+size-gpsize || k == 0 {
// 							Ra.moveTo(S.next(_ITER), _NO_ASSIGN)
// 							Ra.clear()
// 							mostRecentlyMovedToStackIndex = indexOfRi
// 						}
// 					} else if j == span-2 {
// 						Rc := R.next(_ITER)
// 						Rc.add(carry, _CARRY)
// 					} else if j != span-1 {
// 						Rc := R.next(_ITER)
// 						Rc.add(carry, _CARRY)
// 						carry.clear()
// 						carry.addCarry()
// 					}
// 				}
// 			}
// 		}
// 		Commentf("| ")
// 		if k == 0 {
// 			R.updateIndex(mostRecentlyMovedToStackIndex + 1)
// 			for i := 0; i < gpsize; i++ {
// 				R.next(_ITER).moveTo(S.next(_ITER), _NO_ASSIGN)
// 			}
// 			Commentf("| ")
// 			for i := 0; i < gpsize; i++ {
// 				R.next(_ITER).clear()
// 			}
// 		}
// 		Commentf("| \n\n/*\tBlock %d end\t\t\t\t*/\n", k)
// 		Commentf("| ")
// 	}

// 	C := tape.newReprAtParam(size*2, "c", RDI)

// 	Commentf("| c[0, l] = u1[0, l]")
// 	u1 := Stack[0]
// 	u2 := Stack[1].updateIndex(0)
// 	for i := 0; i < gpsize; i++ {
// 		u1.next(_ITER).moveTo(C.next(_ITER).s, _NO_ASSIGN)
// 	}
// 	R.updateIndex(mostRecentlyMovedToStackIndex + 1)
// 	Commentf("| c[l, l+n] = u1[l, l+n] + u2[0, n]")
// 	for i := 0; i < size; i++ {
// 		if i < 2*(size-gpsize) {
// 			u1.next(_ITER).moveTo(RCX, _NO_ASSIGN)
// 			if i == 0 {
// 				u2.next(_ITER).addTo(RCX, _NO_CARRY)
// 			} else {
// 				u2.next(_ITER).addTo(RCX, _CARRY)
// 			}
// 			C.next(_ITER).load(RCX, nil)
// 		} else {
// 			r := R.next(_ITER)
// 			u1.next(_ITER).add(r, _CARRY)
// 			r.moveTo(C.next(_ITER), _NO_ASSIGN)
// 		}
// 	}

// 	Commentf("| c[l+n, 2n] = u2[n, l+n]")
// 	for i := 0; i < size-gpsize; i++ {
// 		if i == 0 {
// 			R.next(_ITER).addCarry().moveTo(C.next(_ITER), _NO_ASSIGN)
// 		} else {
// 			R.next(_ITER).moveTo(C.next(_ITER), _NO_ASSIGN)
// 		}
// 	}
// 	tape.ret()
// 	RET()
// }
