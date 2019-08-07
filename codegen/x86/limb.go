package main

import (
	"errors"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

const (
	_MUL_ADD   = true
	_MUL_MOVE  = false
	_ASSING    = true
	_NO_ASSIGN = false
	_NO_ITER   = false
	_ITER      = true
	_CARRY     = true
	_NO_CARRY  = false
	_ALLOC     = true
	_NO_ALLOC  = false
)

type repr struct {
	limbs []limb
	i     int
	size  int
	base  Register // set base if loaded from memory
}

func newReprEmpty(size int, swapReg Register) *repr {
	number := make([]limb, size)
	for i := 0; i < size; i++ {
		number[i] = newLimbEmpty(swapReg)
	}
	return &repr{number, 0, size, nil}
}

// load will cause changing of source index
func (r *repr) load(src *repr) *repr {
	for i := 0; i < r.size; i++ {
		r.next(_ITER).load(*src.next(_ITER), nil)
	}
	return r
}

func (r *repr) set(i int, op Op) {
	if isLimb(op) {
		r.limbs[i] = op.(limb)
		return
	}
	r.limbs[i].s = op
}

func (r *repr) updateIndex(i int) {
	r.i = i
}

// slice will cause changing of source index
func (r *repr) slice(from, to int) *repr {
	size := to - from
	if size < 0 || from < 0 || to < 0 {
		panic(errors.New(""))
	}
	dst := newReprEmpty(size, RBX)
	r.updateIndex(from)
	for i := 0; i < size; i++ {
		dst.next(_ITER).set(*r.next(_ITER))
	}
	return dst
}

func (r *repr) next(iter bool) *limb {
	i := r.i
	if iter {
		r.i = (r.i + 1) % r.size
	}
	return &r.limbs[i]
}

func (r *repr) mul(iter bool, op Op, lo Op, hi Op, addOrMove bool) {
	r.next(iter).mul(op, lo, hi, addOrMove)
}

func (r *repr) ops() []Op {
	ops := make([]Op, r.size)
	for i, l := range r.limbs {
		ops[i] = l
	}
	return ops
}

type limb struct {
	s       Op
	swapReg Register
}

func isLimb(op Op) bool { _, ok := op.(limb); return ok }

func newLimbEmpty(swapReg Register) limb {
	return limb{
		swapReg: swapReg,
	}
}

func newLimb(op Op, swapReg Register) limb {
	return limb{
		s:       op,
		swapReg: swapReg,
	}
}

func (l limb) Asm() string {
	return l.s.Asm()
}

func (l *limb) String() string {
	if l.atMem() || l.atReg() {
		return l.s.Asm()
	}
	return "NN"
}

func (l *limb) set(op Op) {
	if isLimb(op) {
		op = op.(limb).s
	}
	l.s = op
}

func (l *limb) atMem() bool { return IsMem(l.s) }

func (l *limb) atReg() bool { return IsRegister(l.s) }

func (l *limb) load(src Op, dst Op) {
	if isLimb(src) {
		src = src.(limb).s
	}
	if dst != nil {
		if isLimb(dst) {
			dst = dst.(limb).s
		}
		l.set(dst)
		if IsMem(dst) && IsMem(src) {
			MOVQ(src, RBX)
			MOVQ(RBX, dst)
			return
		}
		MOVQ(src, dst)
		return
	}
	if isMem(src) && l.atMem() {
		if src.Asm() != l.s.Asm() {
			MOVQ(src, RBX)
			MOVQ(RBX, l.s)
		}
		return
	}
	MOVQ(src, l.s)
}

func (l *limb) moveTo(dst Op, assing bool) {
	if isLimb(dst) {
		dst = dst.(limb).s
	}
	if assing {
		l.load(l.s, dst)
		return
	}
	if l.atMem() && IsMem(dst) {
		MOVQ(l.s, RBX)
		MOVQ(RBX, dst)
		return
	}
	MOVQ(l.s, dst)
}

func (l *limb) moveIfNotCF(dst limb) {

	if dst.atMem() {
		MOVQ(dst.s, l.swapReg)
		CMOVQCC(l.s, l.swapReg)
		MOVQ(l.swapReg, dst.s)
		return
	}
	CMOVQCC(l.s, dst.s)
}

// if CF == 0:
// 	mov(LMB, dst)
// else:
//	mov(AUX, dst)
func (l *limb) moveIfNotCFAux(aux limb, dst limb) {

	// Limb:	R	R M	M
	// Aux: 	R R R	R
	// Dest:  R	M R	M
	if aux.atReg() {
		CMOVQCC(l.s, aux.s)
		MOVQ(aux.s, dst.s)
		return
	}
	// Limb:	R M
	// Aux: 	M M
	// Dest:  R R
	if dst.atReg() {
		MOVQ(aux.s, dst.s)
		CMOVQCC(l.s, dst.s)
		return
	}
	// Limb:	R
	// Aux: 	M
	// Dest:  M
	if l.atReg() {
		CMOVQCS(aux.s, l.s)
		MOVQ(l.s, dst.s)
		return
	}
	// Limb:	M
	// Aux: 	M
	// Dest:  M
	MOVQ(aux.s, aux.swapReg)
	CMOVQCC(l.s, aux.swapReg)
	MOVQ(aux.swapReg, dst.s)
}

// loads subtractin result of (left + rigth)
// left operand is overwritten if stored at register
func (l *limb) loadAdd(left limb, rigth limb, brw bool) {
	Add := ADDQ
	if brw {
		Add = ADCQ
	}
	// Left :	R R M M
	// Right:	R M R M
	// Sum  : R R R R
	if l.atReg() {
		MOVQ(left.s, l.s)
		Add(rigth.s, l.s)
		return
	}
	// Left :	R R R
	// Right:	R M M
	// Sum  : M M M
	if left.atReg() {
		Add(rigth.s, left.s)
		MOVQ(left.s, l.s)
		return
	}
	// Left :	M
	// Right:	M
	// Sum  : M
	MOVQ(left.s, l.swapReg)
	Add(rigth.s, l.swapReg)
	MOVQ(l.swapReg, l.s)
}

// loads subtractin result of (left + rigth)
// left operand is overwritten if stored at register
func (l *limb) loadAddSafe(left limb, rigth limb, brw bool) {
	Add := ADDQ
	if brw {
		Add = ADCQ
	}
	// Left :	R R M M
	// Right:	R M R M
	// Sum  : R R R R
	if l.atReg() {
		MOVQ(left.s, l.s)
		Add(rigth.s, l.s)
		return
	}
	// Left :	M R R R
	// Right:	M R M M
	// Sum  : M M M M
	MOVQ(left.s, l.swapReg)
	Add(rigth.s, l.swapReg)
	MOVQ(l.swapReg, l.s)
}

// loads subtractin result of (left - rigth)
// left operand is overwritten if stored at register
func (l *limb) loadSub(left limb, rigth limb, brw bool) {
	Sub := SUBQ
	if brw {
		Sub = SBBQ
	}
	// Left :	R R M M
	// Right:	R M R M
	// Sub  : R R R R
	if l.atReg() {
		MOVQ(left.s, l.s)
		Sub(rigth.s, l.s)
		return
	}
	// Left :	R R R
	// Right:	R M M
	// Sub  : M M M
	if left.atReg() {
		Sub(rigth.s, left.s)
		MOVQ(left.s, l.s)
		return
	}
	// Left :	M
	// Right:	M
	// Sub  : M
	MOVQ(left.s, l.swapReg)
	Sub(rigth.s, l.swapReg)
	MOVQ(l.swapReg, l.s)
}

// loads subtractin result of (left - rigth)
func (l *limb) loadSubSafe(left limb, rigth limb, brw bool) {
	Sub := SUBQ
	if brw {
		Sub = SBBQ
	}
	// Left :	R R M M
	// Right:	R M R M
	// Sub  : R R R R
	if l.atReg() {
		MOVQ(left.s, l.s)
		Sub(rigth.s, l.s)
		return
	}
	// Left :	M R R R
	// Right:	M R M M
	// Sub  : M M M M
	MOVQ(left.s, l.swapReg)
	Sub(rigth.s, l.swapReg)
	MOVQ(l.swapReg, l.s)
}

func (l *limb) loadDouble(l2 limb, brw bool) {
	Add := ADDQ
	if brw {
		Add = ADCQ
	}
	if l.atReg() {
		MOVQ(l2.s, l.s)
		Add(l.s, l.s)
		return
	}
	MOVQ(l2.s, l.swapReg)
	Add(l.swapReg, l.swapReg)
	MOVQ(l.swapReg, l.s)
}

func (l *limb) mul(op Op, a0 Op, a1 Op, addOrMove bool) {
	MOVQ(l.s, RAX)
	MULQ(op)
	if addOrMove == _MUL_ADD {
		if a0 != nil {
			ADDQ(RAX, a0)
			if a1 != nil {
				ADCQ(RDX, a1)
			}
		}
		return
	}
	if a0 != nil {
		MOVQ(RAX, a0)
		if a1 != nil {
			MOVQ(RDX, a1)
		}
	}
}

func (l *limb) add(op Op, car bool) {
	operation := ADDQ
	if car {
		operation = ADCQ
	}
	operation(op, l.s)
}

func (l *limb) addCarry() {
	ADCQ(Imm(0), l.s)
}
