package x86

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func genMontMulAdx(size int, fixedmod bool, single bool) {
	/*
	   ("func mul%d(c *[%d]uint64, a, b *Fe%d)\n\n", i, i*2, i*64)
	*/
	if size < 2 {
		panic("not implemented")
	} else if size >= 2 || size < 9 {
		genMontMul48Adx(size, fixedmod, single)
	} else {
		panic("not implemented")
	}
}

func mulx(a0, a1, a2 *limb) {
	MULXQ(a0.s, a1.s, a2.s)
}

func adcxq(a0, a1 *limb) {
	ADCXQ(a0.s, a1.s)
}

func adoxq(a0, a1 *limb) {
	ADOXQ(a0.s, a1.s)
}

func mul48Adx(tape *tape, A, B, R, Stack *repr) *repr {
	// sonda SI
	bi := newLimb(RDX, nil)
	ax := newLimb(RAX, nil)
	bx := newLimb(RBX, nil)
	XORQ(RAX, RAX)
	size := A.size
	stackSize := Stack.size
	for i := 0; i < size; i++ {
		Comment("|")
		Comment("|")
		B.next(_ITER).moveTo(bi, _NO_ASSIGN)
		if i != 0 {
			if i != size-1 {
				r := R.next(_ITER).s
				XORQ(r, r)
			} else {
				XORQ(B.base, B.base)
			}
		}
		for j := 0; j < size; j++ {
			Comment("|")
			if stackSize > 0 {
				R.updateIndex(i + j - 1)
			} else {
				R.updateIndex(i + j)
			}
			// todo xor
			r1 := R.next(_ITER)
			var r2 *limb
			if i+j != 2*(size-1) {
				r2 = R.next(_ITER)
			} else {
				r2 = new(limb).set(B.base)
			}
			ai := A.next(_ITER)
			if i == 0 {
				mulx(ai, ax, r2)
				if j == 0 {
					if stackSize > 0 {
						s := Stack.next(_ITER).s
						MOVQ(RAX, s)
					} else {
						MOVQ(RAX, r1.s)
					}
				} else {
					adcxq(ax, r1)
				}
				if j == size-1 {
					r2.addCarry()
				}
			} else {
				mulx(ai, ax, bx)
				adoxq(ax, r1)
				if j == size-1 {
					if i != size-1 {
						adoxq(r2, r2)
						adcxq(bx, r2)
					} else {
						adoxq(bx, r2)
						r2.addCarry()
					}
				} else {
					adcxq(bx, r2)
				}
				if j == 0 && i < 2*size-10 {
					r1.moveTo(Stack.next(_ITER), _NO_ASSIGN)
				}
			}
		}
	}
	W := tape.newReprNoAlloc(size * 2)
	for i := 0; i < stackSize; i++ {
		W.next(_ITER).set(Stack.next(_ITER))
	}
	if stackSize < 1 {
		R.updateIndex(0)
		for i := 0; i < size*2-1; i++ {
			W.next(_ITER).set(R.next(_ITER))
		}
	} else {
		for i := 0; i < R.size; i++ {
			W.next(_ITER).set(R.next(_ITER))
		}
	}

	W.next(_ITER).set(B.base)
	W.updateIndex(0)
	return W
}

func mont48Adx(tape *tape, W *repr, inp Op, modulus *repr, hi, c *limb, fixedModulus bool) {
	if W.size%2 != 0 || W.size > 16 {
		panic("")
	}
	size := W.size / 2
	dx := newLimb(RDX, nil)
	ax := newLimb(RAX, nil)
	W.updateIndex(0)
	zero := new(limb)
	for i := 0; i < size; i++ {
		Comment("|")
		Comment("|")
		if i == 0 {
			c.xorself()
		}
		W.updateIndex(i)
		W.get().moveTo(dx, _NO_ASSIGN)
		MULXQ(inp, RDX, hi.s)
		for j := 0; j < size; j++ {
			Comment("|")
			W.updateIndex(i + j)
			w1 := W.next(_ITER)
			w2 := W.next(_ITER)
			mulx(modulus.next(_ITER), ax, hi)
			adoxq(ax, w1)
			adcxq(hi, w2)
			if j == size-1 {
				adoxq(c, w2)
				if i != 0 {
					c.clear()
				}
				adcxq(c, c)
				if i != size-1 {
					zero.xorself()
				} else {
					adoxq(zero, c)
				}
			}
			if j == 0 {
				zero.set(w1)
			}
		}

		if i != size-1 {
			w := W.next(_ITER)
			if w.atMem() {
				w.moveTo(zero, _ASSIGN)
			}
		}
	}
	comment("reduction")
	C_red := W.slice(size, size*2)
	tape.freeAll()
	tape.reserveGp(C_red.ops()...)
	tape.reserveGp(c.s)
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
	SBBQ(U32(0), c.s)
	Commentf("|")
	C := tape.newReprAtParam(C_red.size, "c", c.s.(Register), 0)
	for i := 0; i < C_red.size; i++ {
		T.next(_ITER).moveIfNotCFAux(
			*C_red.next(_ITER),
			*C.next(_ITER))
	}
}

func genMontMul48Adx(size int, fixedmod bool, single bool) {
	funcName := "mul"
	modulusName := "·modulus"
	if !single {
		funcName = fmt.Sprintf("%s%d", funcName, size)
		modulusName = fmt.Sprintf("%s%d", modulusName, size)
	}
	if fixedmod {
		TEXT(funcName, NOSPLIT, fmt.Sprintf("func(c *[%d]uint64, a, b *[%d]uint64)", size, size))
	} else {
		TEXT(funcName, NOSPLIT, fmt.Sprintf("func(c *[%d]uint64, a, b, p *[%d]uint64, inp uint64)", size, size))
	}
	comment("inputs")
	tape := newTape(_NO_SWAP, RDX, RAX, RBX)
	A := tape.newReprAtParam(size, "a", RDI, 0)
	B := tape.newReprAtParam(size, "b", RSI, 0)
	if tape.sizeFreeGp() != 9 {
		panic("")
	}

	R := tape.newReprAllocRemainingGPRs()
	if R.size != 9 {
		panic("")
	}

	var r *repr
	switch size {
	case 2:
		r = R.slice(0, 3)
	case 3:
		r = R.slice(0, 5)
	case 4:
		r = R.slice(0, 7)
	default:
		r = R
	}
	rLast := R.last()

	stackSize := 2*size - 10
	if stackSize < 0 {
		stackSize = 0
	}
	Stack := tape.allocStack(stackSize)

	W := mul48Adx(tape, A, B, r, Stack)
	if W.size != 2*size {
		panic("")
	}
	Comment("|")
	var modulus *repr
	var inp Mem // fix: inp to limb type
	var lCarry, hi *limb
	if fixedmod {
		inp = NewDataAddr(Symbol{Name: fmt.Sprintf("·inp")}, 0)
		modulus = tape.newReprAtMemory(size, NewDataAddr(Symbol{Name: modulusName}, 0), 0)
	} else {
		inp = NewParamAddr("inp", 32)
	}
	W.updateIndex(0)
	switch size {
	case 2:
		if !fixedmod {
			modulus = tape.newReprAtParam(size, "p", rLast.asRegister(), 0)
		}
		hi = newLimb(A.base, nil)
		lCarry = newLimb(RBX, nil)
	case 3:
		if !fixedmod {
			modulus = tape.newReprAtParam(size, "p", rLast.asRegister(), 0)
		}
		hi = newLimb(A.base, nil)
		lCarry = newLimb(RBX, nil)
	case 4:
		if !fixedmod {
			modulus = tape.newReprAtParam(size, "p", rLast.asRegister(), 0)
		}
		hi = newLimb(A.base, nil)
		lCarry = newLimb(RBX, nil)
	case 5:
		hi = newLimb(A.base, nil)
		lCarry = newLimb(RBX, nil)
		if !fixedmod {
			s := tape.stack.extend(1, true)
			w := W.updateIndex(-1).next(_NO_ITER)
			t := newLimb(w.s, nil)
			w.moveTo(s, _ASSIGN)
			modulus = tape.newReprAtParam(size, "p", t.asRegister(), 0)
		}
	case 6:
		W.next(_ITER).moveTo(A.base, _ASSIGN)
		W.next(_ITER).moveTo(RBX, _ASSIGN)
		//
		if !fixedmod {
			W.updateIndex(9)
		} else {
			W.updateIndex(10)
		}
		Stack.updateIndex(0)
		//
		w := W.next(_ITER)
		hi = newLimb(w.s, nil)
		w.moveTo(Stack.next(_ITER), _ASSIGN)
		w = W.next(_ITER)
		lCarry = newLimb(w.s, nil)
		w.moveTo(Stack.next(_ITER), _ASSIGN)
		if !fixedmod {
			s := tape.stack.extend(1, true)
			w := W.next(_ITER)
			t := newLimb(w.s, nil)
			w.moveTo(s, _ASSIGN)
			modulus = tape.newReprAtParam(size, "p", t.s.(Register), 0)
		}
	case 7:
		W.next(_ITER).moveTo(A.base, _ASSIGN)
		W.next(_ITER).moveTo(RBX, _ASSIGN)
		//
		w2 := W.next(_ITER)
		w3 := W.next(_ITER)
		//
		if !fixedmod {
			W.updateIndex(9)
		} else {
			W.updateIndex(10)
		}
		Stack.updateIndex(0)
		//
		w10 := W.next(_ITER)
		t := w10.clone()
		w10.moveTo(Stack.next(_ITER), _ASSIGN)
		w2.moveTo(t, _ASSIGN)
		//
		w11 := W.next(_ITER)
		t = w11.clone()
		w11.moveTo(Stack.next(_ITER), _ASSIGN)
		w3.moveTo(t, _ASSIGN)
		//
		w := W.next(_ITER)
		hi = newLimb(w.s, nil)
		w.moveTo(Stack.next(_ITER), _ASSIGN)
		//
		w = W.next(_ITER)
		lCarry = newLimb(w.s, nil)
		w.moveTo(Stack.next(_ITER), _ASSIGN)
		//
		if !fixedmod {
			s := tape.stack.extend(1, true)
			w := W.next(_ITER)
			t := newLimb(w.s, nil)
			w.moveTo(s, _ASSIGN)
			modulus = tape.newReprAtParam(size, "p", t.s.(Register), 0)
		}
	case 8:
		Stack.updateIndex(0)
		W.next(_ITER).moveTo(A.base, _ASSIGN)
		W.next(_ITER).moveTo(RBX, _ASSIGN)
		//
		w2 := W.next(_ITER)
		w3 := W.next(_ITER)
		w4 := W.next(_ITER)
		w5 := W.next(_ITER)
		//
		if !fixedmod {
			W.updateIndex(9)
		} else {
			W.updateIndex(10)
		}
		Stack.updateIndex(0)
		//
		w10 := W.next(_ITER)
		t := w10.clone()
		w10.moveTo(Stack.next(_ITER), _ASSIGN)
		w2.moveTo(t, _ASSIGN)
		//
		w11 := W.next(_ITER)
		t = w11.clone()
		w11.moveTo(Stack.next(_ITER), _ASSIGN)
		w3.moveTo(t, _ASSIGN)
		//
		w12 := W.next(_ITER)
		t = w12.clone()
		w12.moveTo(Stack.next(_ITER), _ASSIGN)
		w4.moveTo(t, _ASSIGN)
		//
		w13 := W.next(_ITER)
		t = w13.clone()
		w13.moveTo(Stack.next(_ITER), _ASSIGN)
		w5.moveTo(t, _ASSIGN)
		//
		w := W.next(_ITER)
		hi = newLimb(w.s, nil)
		w.moveTo(Stack.next(_ITER), _ASSIGN)
		//
		w = W.next(_ITER)
		lCarry = newLimb(w.s, nil)
		w.moveTo(Stack.next(_ITER), _ASSIGN)
		//
		if !fixedmod {
			s := tape.stack.extend(1, true)
			w := W.next(_ITER)
			t := newLimb(w.s, nil)
			w.moveTo(s, _ASSIGN)
			modulus = tape.newReprAtParam(size, "p", t.s.(Register), 0)
		}
	default:
	}
	W.updateIndex(0)
	mont48Adx(tape, W, inp, modulus, hi, lCarry, fixedmod)
	tape.ret()
	RET()
	comment("end")
}
