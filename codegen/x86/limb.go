package x86

import (
	"errors"
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

const (
	_MUL_ADD   = true
	_MUL_MOVE  = false
	_ASSIGN    = true
	_NO_ASSIGN = false
	_NO_ITER   = false
	_ITER      = true
	_CARRY     = true
	_NO_CARRY  = false
	_ALLOC     = true
	_NO_ALLOC  = false
)

type repr struct {
	limbs []*limb
	i     int
	size  int
	base  Register // set base if loaded from memory
	swap  GPPhysical
}

func newReprEmpty(size int, swap GPPhysical) *repr {
	number := make([]*limb, size)
	for i := 0; i < size; i++ {
		number[i] = newLimbEmpty(swap)
	}
	return &repr{number, 0, size, nil, swap}
}

// func (r *repr) ops() []Op {
// 	ops := []Op{}
// 	for i := 0; i < r.size; i++ {
// 		limb := r.limbs[i]
// 		if limb != nil {
// 			ops = append(ops, r.limbs[i])
// 		}
// 	}
// 	return ops
// }

func (r *repr) ops() []Op {
	ops := make([]Op, r.size)
	for i, l := range r.limbs {
		ops[i] = l.s
	}
	return ops
}

func (r *repr) debug() {
	fmt.Printf("--------------\n\n")
	fmt.Printf("Repr Debug\n")
	fmt.Printf("Size: %d\n", r.size)
	for i := 0; i < len(r.limbs); i++ {
		fmt.Printf("[%d]: ", i)
		limb := r.limbs[i]
		if limb == nil {
			fmt.Printf("notset\n")
		} else {
			fmt.Printf("%s\n", limb.Asm())
		}
	}
	fmt.Printf("--------------\n")
}

func (t *repr) commentCurrent(name string) {
	Commentf("| %s%d @ %s", name, t.i, t.next(_NO_ITER).Asm())
}

// load will cause changing of source index
func (r *repr) load(src *repr) *repr {
	for i := 0; i < r.size; i++ {
		r.next(_ITER).load(*src.next(_ITER), nil)
	}
	return r
}

// func (r *repr) set(i int, op Op) {
// 	if isLimb(op) {
// 		r.limbs[i] = op.(*limb)
// 		return
// 	}
// 	r.limbs[i].s = op
// }

func (r *repr) set(op Op) {
	if isLimb(op) {
		r.limbs[r.i] = op.(*limb)
		return
	}
	r.limbs[r.i].s = op
}

func (r *repr) setSwap(reg Register) {
	for i := 0; i < r.size; i++ {
		r.limbs[i].swapReg = reg
	}
}

func (r *repr) updateIndex(i int) *repr {
	r.i = (i + r.size) % r.size
	return r
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
		dst.next(_ITER).set(r.next(_ITER))
	}
	return dst
}

func (r *repr) next(iter bool) *limb {
	i := r.i
	if iter {
		r.i = (r.i + 1) % r.size
	}
	return r.limbs[i]
}

func (r *repr) previous() *limb {
	i := r.i
	r.i = ((r.i - 1) + r.size) % r.size
	return r.limbs[i]
}

func (r *repr) get() *limb {
	return r.limbs[r.i]
}

func (r *repr) at(i int) *limb {
	return r.limbs[i]
}

func (r *repr) last() *limb {
	return r.limbs[r.size-1]
}

func (r *repr) mul(iter bool, op Op, lo Op, hi Op, addOrMove bool) {
	r.next(iter).mul(op, lo, hi, addOrMove)
}

type limb struct {
	s       Op
	swapReg Register
}

func isLimb(op Op) bool { _, ok := op.(*limb); return ok }

func newLimbEmpty(swapReg Register) *limb {
	return &limb{
		swapReg: swapReg,
	}
}

func newLimb(op Op, swapReg Register) *limb {
	return &limb{
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

func (l *limb) set(op Op) *limb {
	if isLimb(op) {
		l.s = op.(*limb).s
		return l
	}
	l.s = op
	return l
}

func (l *limb) atMem() bool { return IsMem(l.s) }

func (l *limb) atReg() bool { return IsRegister(l.s) }

func (l *limb) clone() *limb {
	return newLimb(l.s, nil)
}

func (l *limb) asRegister() Register {
	return l.s.(Register)
}

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

// func (l *limb) moveTo(dst Op, assign bool) *limb {
// 	if isLimb(dst) {
// 		dst = dst.(*limb).s
// 	}
// 	if assign {
// 		l.load(l.s, dst)
// 		return l
// 	}
// 	if l.atMem() && IsMem(dst) {
// 		MOVQ(l.s, RBX)
// 		MOVQ(RBX, dst)
// 		return l
// 	}
// 	MOVQ(l.s, dst)
// 	return l
// }

func (l *limb) moveTo(dst Op, assign bool) *limb {
	if isLimb(dst) {
		dst = dst.(*limb).s
	}
	if assign {
		l.load(l.s, dst)
		return l
	}
	if l.atMem() && IsMem(dst) {
		MOVQ(l.s, l.swapReg)
		MOVQ(l.swapReg, dst)
		return l
	}
	MOVQ(l.s, dst)
	return l
}

func (l *limb) moveIfNotCF(dst limb) *limb {
	if dst.atMem() {
		MOVQ(dst.s, l.swapReg)
		CMOVQCC(l.s, l.swapReg)
		MOVQ(l.swapReg, dst.s)
		return l
	}
	CMOVQCC(l.s, dst.s)
	return l
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
	_a0, _a1, _op := a0, a1, op
	if isLimb(a0) {
		_a0 = a0.(*limb).s
	}
	if isLimb(a1) {
		_a1 = a1.(*limb).s
	}
	if isLimb(op) {
		_op = op.(*limb).s
	}
	MOVQ(l.s, RAX)
	MULQ(_op)
	if addOrMove == _MUL_ADD {
		if _a0 != nil {
			ADDQ(RAX, _a0)
		}
		if _a1 != nil {
			ADCQ(RDX, _a1)
		}
		return
	}
	if _a0 != nil {
		MOVQ(RAX, _a0)
	}
	if _a1 != nil {
		MOVQ(RDX, _a1)
	}
}

func (l *limb) add(op Op, car bool) {
	var _op Op = op
	if isLimb(op) {
		_op = op.(*limb).s
	}
	operation := ADDQ
	if car {
		operation = ADCQ
	}
	operation(_op, l.s)
}

func (l *limb) addTo(op Op, car bool) {
	var _op Op = op
	if isLimb(op) {
		_op = op.(*limb).s
	}
	operation := ADDQ
	if car {
		operation = ADCQ
	}
	operation(l.s, _op)
}

func (l *limb) addCarry() *limb {
	ADCQ(Imm(0), l.s)
	return l
}

func (l *limb) clear() *limb {
	MOVQ(U64(0), l.s)
	return l
}

func (l *limb) cmp(op Op) {
	CMPQ(l.s, op)
}

func (l *limb) xorself() *limb {
	XORQ(l.s, l.s)
	return l
}
