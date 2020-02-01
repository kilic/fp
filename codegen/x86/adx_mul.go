package x86

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/reg"
)

var RSize int

func mulADX(size int) {
	funcName := "mul"
	// TEXT(funcName, NOSPLIT, fmt.Sprintf("func(c *[%d]uint64, a, b, p *[%d]uint64, inp uint64)", 2*2, 2))
	TEXT(funcName, NOSPLIT, fmt.Sprintf("func(c *[%d]uint64, a, b *[%d]uint64)", size*2, size))
	commentHeader("inputs")
	tape := newTape(_NO_SWAP, RAX, RBX, RDX)
	A := tape.newReprAtParam(size, "a", tape.di(), 0)
	B := tape.newReprAtParam(size, "b", tape.si(), 0)
	R := tape.newReprAllocGPRs(RSize)
	R.debug("R")
	if size > RSize*2 {
		panic("only two-partial multiplication is allowed")
	}
	var W *repr
	if size > RSize {
		Wr := partialMulADX(tape, A, B, R).commentState("Wr").debug("Wr")
		tape.moveToStack(Wr).commentState("Wr @ stack").debug("W should be in stack")
		// tmp solution
		A := tape.newReprAtParam(size, "a", tape.di(), 0)
		Wl := partialMulADX(tape, A, B, R).commentState("Wl").debug("Wl")
		Wr.commentState("Wr")
		Wl.setSwap(tape.ax())
		W = combinePartialResults(tape, Wr, Wl).commentState("W combined").debug("W combined")
	} else {
		W = partialMulADX(tape, A, B, R).commentState("W").debug("W")
	}
	_ = W
	// C := tape.newReprAtParam(size*2, "c", mlo, 0)
	// W.setSwap(mhi)
	// for i := 0; i < W.size; i++ {
	// 	W.next().moveTo(C.next(), _NO_ASSIGN)
	// }
	tape.ret()
}

func partialMulADX(tape *tape, A, B, R *repr) *repr {
	stack := tape.stack

	if A.size != B.size {
		panic("operands should be in same size")
	}
	if A.size == 0 {
		panic("bad size, A")
	}
	if B.size == 0 {
		panic("bad size, B")
	}
	size := B.size
	// We process second operand only at size of R
	// i = B.index
	// W = A * B[i, min(i + R.size, B.size)]
	span := size - B.i
	if span > R.size {
		span = R.size
	}
	resultWindow := size + span
	resultOffset := B.i
	W := tape.newReprNoAlloc(size * 2)
	W.updateIndex(resultOffset)
	// We will be using stack to save the result when all gp registers are busy.
	// Keep in mind that if usingStack is true registers will be rotated processing thru A.
	var usingStack bool = R.size+1 < resultWindow
	ax := tape.ax()
	bx := tape.bx()
	ai := tape.dx()
	ax.xorself() // clear flags
	for i := 0; i < A.size; i++ {
		firstI, lastI := i == 0, i == A.size-1
		needStack := R.size < (A.size - 1 + span - i)
		//////////////////
		commentA(i, ai)
		//////////////////
		A.next().moveTo(ai, _NO_ASSIGN)
		if !firstI {
			if !lastI {
				// clear flags
				// span size == R.size is an edge case
				// if span size and registers size equal
				// we should clear with an idle register
				// otherwise clear with next register which is also
				// required for clearing register itself for
				// following use at the end of this iter.
				if span == R.size {
					ax.xorself()
				} else {
					R.next().xorself()
				}
			} else { // use base of operand B at the last iter
				A.base.xorself()
			}
		}
		// align B
		B.updateIndex(resultOffset)
		for j := 0; j < span; j++ {
			firstJ, lastJ := j == 0, j == span-1
			//////////////////
			commentAiBj(i, B.i)
			//////////////////
			if usingStack && !(firstI && firstJ) {
				R.updateIndex(i + j - 1)
			} else {
				R.updateIndex(i + j)
			}
			if firstI {
				if firstJ {
					if usingStack {
						B.next().mulx(ax, R.next())
						s := stack.next()
						ax.moveTo(s, _NO_ASSIGN)
						W.next().set(s.clone())
					} else {
						Ra := R.next()
						B.next().mulx(Ra, R.next())
						W.next().set(Ra)
					}
				} else {
					Ra, Rb := R.next(), R.next()
					B.next().mulx(ax, Rb)
					Ra.adcxq(ax)
					Rb.addCarryIf(lastJ)
				}
			} else {
				var Rb *limb
				Ra := R.next()
				if lastI && lastJ {
					Rb = A.base.clone()
				} else {
					Rb = R.next()
				}
				B.next().mulx(ax, bx)
				Ra.adoxq(ax)
				if lastJ {
					if lastI {
						Rb.adoxq(bx).addCarry()
					} else {
						Rb.adoxq(Rb).adcxq(bx)
					}
				} else {
					Rb.adcxq(bx)
				}
				// stack
				if firstJ {
					if usingStack && needStack {
						s := stack.next()
						Ra.moveTo(s, _NO_ASSIGN)
						W.next().set(s.clone())
						Ra.clearIf(span == R.size) // an edge case detailed above
					} else if !lastI {
						W.next().set(Ra)
					}
				}
			}
		}
	}
	commentHeader("\t\t\t")
	// Get back to the beginning of last iteration
	R.rotate(-span)
	for W.i != resultOffset+resultWindow-1 { // Go until last one
		W.next().set(R.next())
	}
	// Highest limb of the result was kept in carry register
	wLast := W.next()
	if W.i != (resultOffset+resultWindow)%(size*2) {
		panic("result should go end of the span")
	}
	wLast.set(A.base.clone())
	return W
}
