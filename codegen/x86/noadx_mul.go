package x86

func partialMulNoADX(tape *tape, A, B, R *repr, ai, carry *limb) *repr {
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
	if usingStack { // First register will be assigned, no need to clear.
		R.clearFrom(1)
	} else { // First two register will be assigned, no need to clear.
		R.clearFrom(2)
	}

	// Switch A and B if span size = 1
	if span == 1 {
		commentB(B.size-1, ai)
		B.updateIndex(resultOffset)
		B.next().moveTo(ai, _NO_ASSIGN)
	}
	carry.clearIf(span == 2 || span == 1)
	for i := 0; i < A.size; i++ {
		lastI := i == A.size-1
		//////////////////
		if span > 1 {
			commentA(i, ai)
		}
		//////////////////
		if span != 1 {
			A.next().moveTo(ai, _NO_ASSIGN)
		}
		carry.clearIf(i != 0 && span > 2)
		// Bring back B
		B.updateIndex(resultOffset)
		for j := 0; j < span; j++ {
			//////////////////
			commentAiBj(i, B.i)
			//////////////////
			var Ra *limb
			if i+j == 0 {
				// if using stack and i,j = 0,0
				// low limb of result goes to stack
				// otherwise save it to the register
				if usingStack {
					Ra = R.next()
					// ai * bj
					// mov low -> stack (W0)
					// mov hig -> Ra
					s := stack.next()
					if span > 1 {
						B.next().mul(ai, s, Ra, _MUL_MOVE)
					} else {
						A.next().mul(ai, s, Ra, _MUL_MOVE)
					}
					W.next().set(s)
				} else {
					Ra = R.next()
					Rb := R.next()
					// ai * bj
					// move low -> Ra
					// move hig -> Rb
					if span > 1 {
						B.next().mul(ai, Ra, Rb, _MUL_MOVE)
					} else {
						A.next().mul(ai, Ra, Rb, _MUL_MOVE)
					}
					W.next().set(Ra)
				}
			} else {
				if usingStack {
					R.updateIndex(i + j - 1)
				} else {
					R.updateIndex(i + j)
				}
				// if it is very last multiplication
				// high limb of result goes to register which was before used for carry
				if i == A.size-1 && j == span-1 {
					Ra = R.next()
					if span > 1 {
						B.next().mul(ai, Ra, carry, _MUL_ADD)
					} else {
						A.next().mul(ai, Ra, carry, _MUL_ADD)
					}
				} else {
					Ra = R.next()
					Rb := R.next()
					if span > 1 {
						B.next().mul(ai, Ra, Rb, _MUL_ADD)
					} else {
						A.next().mul(ai, Ra, Rb, _MUL_ADD)
					}
				}
			}
			// No carry operations involved in first round of multiplication
			if i != 0 {
				if j == 0 {
					if span != 1 { // fix: should be obvious
						R.next().addCarryIf(span != 2 || !lastI) // this is third register (Rc)
						carry.addCarryIf(span > 2 || lastI)
					}
					// Ra should be moved to stack if no gp register space left
					// Ra will be used in rotation for higher limbs of result
					var needStack bool
					needStack = R.size < (A.size - 1 + span - i)
					if i < A.size-1 { // if it is last round do keep it in register
						// W_(i+j)
						if usingStack && needStack {
							s := stack.next()
							Ra.moveTo(s, _NO_ASSIGN).clear()
							W.next().set(s)
						} else {
							// if not using stack there is no rotation
							W.next().set(Ra)
						}
					}
				} else if j != 0 && j != span-2 && j != span-1 {
					R.next().add(carry, _CARRY)
					carry.clear().addCarry()
				} else if j == span-2 {
					if i == A.size-1 { // if last i
						carry.addCarry()
					} else {
						R.next().add(carry, _CARRY) // this is third register (Rc)
					}
				}
			}
		}
	}
	commentHeader("\t\t\t")
	// Get back to the beginning of last iteration
	R.rotate(-span)
	for W.i != resultOffset+resultWindow-1 { // go until last one
		W.next().set(R.next())
	}
	// Highest limb of the result was kept in carry register
	wLast := W.next()
	if W.i != (resultOffset+resultWindow)%(size*2) {
		panic("result should go end of the span")
	}
	wLast.set(carry)
	return W
}
