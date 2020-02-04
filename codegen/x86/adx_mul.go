package x86

func partialMulADX(tape *tape, A, B, R *repr) *repr {
	stack := tape.stack
	size := B.size
	// We process second operand only at size of R
	// i = B.index
	// W = A * B[i, min(i + R.size, B.size)]
	span := size - B.i
	useBaseOfInputForLastLimb := true
	if span > R.size {
		span = R.size
		useBaseOfInputForLastLimb = false
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
	var needStack bool
	for i := 0; i < A.size; i++ {
		firstI, lastI := i == 0, i == A.size-1
		if useBaseOfInputForLastLimb {
			needStack = R.size < (A.size - 1 + span - i)
		} else {
			needStack = R.size < (A.size + span - i)
		}
		//////////////////
		commentA(i, ai)
		//////////////////
		A.next().moveTo(ai, _NO_ASSIGN)
		if !firstI {
			if !lastI || !useBaseOfInputForLastLimb {
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
				if lastI && lastJ && useBaseOfInputForLastLimb {
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
	R.rotate(-span)
	for W.i != resultOffset+resultWindow-1 { // Go until last one
		W.next().set(R.next())
	}
	// Highest limb of the result was kept in carry register
	wLast := W.next()
	if W.i != (resultOffset+resultWindow)%(size*2) {
		panic("result should go end of the span")
	}
	if useBaseOfInputForLastLimb {
		wLast.set(A.base.clone())
	} else {
		wLast.set(R.next())

	}
	return W
}
