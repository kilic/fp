package x86

import (
	"errors"
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

var debugOn = false

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
	base  *limb
	swap  GPPhysical
}

func newReprEmpty(size int) *repr {
	number := make([]*limb, size)
	for i := 0; i < size; i++ {
		number[i] = newLimbEmpty()
	}
	return &repr{number, 0, size, nil, nil}
}

func (r *repr) ops() []Op {
	ops := make([]Op, r.size)
	for i, l := range r.limbs {
		ops[i] = l.s
	}
	return ops
}

func (r *repr) debug(desc string) *repr {
	if debugOn {
		fmt.Printf("--------------\n")
		fmt.Printf("Repr\n%s\n", desc)
		fmt.Printf("Size: %d\n", r.size)
		p := r.i
		for i := 0; i < len(r.limbs); i++ {
			fmt.Printf("[%d]: ", i)
			limb := r.limbs[i]
			if limb == nil || limb.s == nil {
				fmt.Printf("NN")
			} else {
				fmt.Printf("%s", limb.String())
			}
			if i == p {
				fmt.Printf("\t*")
			}
			fmt.Printf("\n")
		}
		fmt.Printf("--------------\n")
	}
	return r
}

func (r *repr) debug2(desc string, j int) *repr {
	if debugOn {
		fmt.Printf("--------------\n")
		fmt.Printf("Repr\n%s\n", desc)
		fmt.Printf("Size: %d\n", r.size)
		p := r.i
		var k int
		countK := false
		for i := 0; i < len(r.limbs); i++ {
			fmt.Printf("[%d]: ", i)
			limb := r.limbs[i]
			if limb == nil || limb.s == nil {
				fmt.Printf("NN")
			} else {
				fmt.Printf("%s", limb.String())
			}
			if i == p {
				fmt.Printf("\t*")
				countK = true
			}
			if k == j {
				fmt.Printf("\t&")
			}
			if countK {
				k++
			}
			fmt.Printf("\n")
		}
		fmt.Printf("--------------\n")
	}
	return r
}

func (r *repr) commentCurrent(name string) {
	Commentf("| %s%d @ %s", name, r.i, r.get().String())
}

func (r *repr) commentPrevious(name string) {
	Commentf("| %s%d @ %s", name, (r.i-1+r.size)%r.size, r.at(r.i-1).String())
}

func (r *repr) commentNext(name string) {
	Commentf("| %s%d @ %s", name, (r.i+1)%r.size, r.at(r.i+1).String())
}

// load will cause changing of source index
func (r *repr) load(src *repr) *repr {
	for i := 0; i < r.size; i++ {
		r.next().load(src.next(), nil)
	}
	return r
}

func (r *repr) set(l *limb) {
	r.limbs[r.i] = l
}

func (r *repr) setSwap(swap *limb) *repr {
	for i := 0; i < r.size; i++ {
		r.limbs[i].setSwap(swap)
	}
	return r
}

func (r *repr) updateIndex(i int) *limb {
	r.i = (i + r.size) % r.size
	return r.get()
}

// slice will not pass source index
func (r *repr) slice(from, to int) *repr {
	size := to - from
	if size < 0 || from < 0 || to < 0 {
		panic(errors.New(""))
	}
	dst := newReprEmpty(size)
	r.updateIndex(from)
	for i := 0; i < size; i++ {
		dst.limbs[i] = r.next()
	}
	return dst
}

func (r *repr) reverse() *repr {
	limbs := []*limb{}
	for i := r.size - 1; i >= 0; i-- {
		limbs = append(limbs, r.limbs[i])
	}
	r.i = (-r.i + r.size) % r.size
	r.limbs = limbs
	return r
}

func (r *repr) copy(from, to int) *repr {
	size := to - from
	if size < 0 || from < 0 || to < 0 {
		panic(errors.New(""))
	}
	dst := newReprEmpty(size)
	r.updateIndex(from)
	for i := 0; i < size; i++ {
		dst.next().set(r.next().clone())
	}
	return dst
}

func (r *repr) clone() *repr {
	dst := newReprEmpty(r.size)
	for i := 0; i < r.size; i++ {
		dst.next().set(r.next().clone())
	}
	return dst
}

func (r *repr) registerSlice() *repr {
	r2 := newReprEmpty(0)
	for i := 0; i < r.size; i++ {
		l := r.at(i)
		if l.atReg() {
			r2.extend(l)
		}
	}
	return r2
}

func (r *repr) stackSlice() *repr {
	r2 := newReprEmpty(0)
	for i := 0; i < r.size; i++ {
		l := r.at(i)
		if l.atMem() {
			r2.extend(l)
		}
	}
	return r2
}

func (r *repr) next() *limb {
	i := r.i
	r.i = (r.i + 1) % r.size
	return r.limbs[i]
}

func (r *repr) adjustIndex() *repr {
	r.updateIndex(0)
	for i := 0; i < r.size; i++ {
		if !r.limbs[i].isEmpty() {
			r.updateIndex(i)
			return r
		}
	}
	return r
}

func (r *repr) rotate(s int) *limb {
	i := r.i
	r.i = (r.i + s) % r.size
	r.i = (r.i + r.size) % r.size
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
	j := (i + r.size) % r.size
	return r.limbs[j]
}

func (r *repr) atMem() bool {
	return r.get().atMem()
}

func (r *repr) atReg() bool {
	return r.get().atReg()
}

func (r *repr) clearFrom(ofs int) *repr {
	for i := 0; i < r.size; i++ {
		if i >= ofs {
			r.at(i).clear()
		}
	}
	r.updateIndex(0)
	return r
}

func (r *repr) commentState(desc string) *repr {
	breakAt := (r.size / 2) - 1
	var state string = "\t// "
	for i := 0; i < r.size; i++ {
		ri := r.at(i)
		var s string
		if ri == nil || ri.s == nil {
			s = fmt.Sprintf("| %-3d %-16s", i, "-")
		} else {
			s = fmt.Sprintf("| %-3d %-16v", i, r.at(i).String())
		}
		if len(s) > 16 {
			s = s[:16]
		}
		state += s
		if i == breakAt {
			state += "\n\t// "
		}
	}
	Commentf("| \n\t// | %s\n%s\n\n", desc, state)
	return r
}

func (r *repr) mul(iter bool, op, lo, hi *limb, addOrMove bool) {
	if iter {
		r.next().mul(op, lo, hi, addOrMove)
		return
	}
	r.get().mul(op, lo, hi, addOrMove)
}

func (r *repr) extend(l *limb) *repr {
	r.size = r.size + 1
	r.limbs = append(r.limbs, l)
	return r
}

type limb struct {
	s    Op
	swap Op
	tag  string
}

func newLimbEmpty() *limb {
	return &limb{}
}

func newLimb(op Op) *limb {
	return &limb{
		s: op,
	}
}

func (l *limb) setSwap(swap *limb) *limb {
	l.swap = swap.s
	return l
}

func (l *limb) String() string {
	if l.atMem() || l.atReg() {
		return l.s.Asm()
	}
	return "NN"
}

func (l *limb) set(op *limb) *limb {
	l.s = op.s
	return l
}

func (l *limb) atStack() bool { return IsMem(l.s) }

func (l *limb) atMem() bool { return IsMem(l.s) }

func (l *limb) atReg() bool { return IsRegister(l.s) }

func (l *limb) clone() *limb {
	return newLimb(l.s)
}

func (l *limb) comment(tag string, index int) *limb {
	Commentf("| %s%d @ %s", tag, index, l.String())
	return l
}

func (l *limb) assertAtReg(desc string) *limb {
	if !l.atReg() {
		panic(desc)
	}
	return l
}

func (l *limb) assertAtMem(desc string) *limb {
	if !l.atMem() {
		panic(desc)
	}
	return l
}

func (l *limb) delete() *limb {
	s := l.clone()
	l.s = nil
	l.swap = nil
	return s
}

func (l *limb) op() Op {
	return l.s
}

func (l *limb) load(src, dst *limb) {
	if dst != nil {
		if dst.atMem() && src.atMem() {
			MOVQ(src.s, l.swap)
			MOVQ(l.swap, dst.s)
		} else {
			MOVQ(src.s, dst.s)
		}
		l.set(dst)
		return
	}
	if src.atMem() && l.atMem() {
		if src.String() != l.String() {
			MOVQ(src.s, l.swap)
			MOVQ(l.swap, l.s)
		}
		return
	}
	MOVQ(src.s, l.s)
}

func (l *limb) moveTo(dst *limb, assign bool) *limb {
	if assign {
		l.load(l, dst)
		return l
	}
	if l.atMem() && dst.atMem() {
		MOVQ(l.s, l.swap)
		MOVQ(l.swap, dst.s)
		return l
	}
	MOVQ(l.s, dst.s)
	return l
}

func (l *limb) move(dst *limb) *limb {
	if l.atMem() && dst.atMem() {
		MOVQ(l.s, l.swap)
		MOVQ(l.swap, dst.s)
		return l
	}
	MOVQ(l.s, dst.s)
	return l
}

func (l *limb) moveIf(c bool, dst *limb) *limb {
	if c {
		return l.move(dst)
	}
	return l
}

func (l *limb) moveAssign(dst *limb) *limb {
	l.load(l, dst)
	return l
}

func (l *limb) moveAssignIf(c bool, dst *limb) *limb {
	if c {
		return l.moveAssign(dst)
	}
	return l
}

func (l *limb) moveIfNotCF(dst *limb) *limb {
	if dst.atMem() {
		MOVQ(dst.s, l.swap)
		CMOVQCC(l.s, l.swap)
		MOVQ(l.swap, dst.s)
		return l
	}
	CMOVQCC(l.s, dst.s)
	return l
}

// if CF == 0:
// 	mov(LMB, dst)
// else:
//	mov(AUX, dst)
func (l *limb) moveIfNotCFAux(aux *limb, dst *limb) {
	// fmt.Println("cc", l, aux, dst)
	// return

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
	// MOVQ(aux.s, aux.swap)
	// CMOVQCC(l.s, aux.swap)
	// MOVQ(aux.swap, dst.s)
	MOVQ(aux.s, l.swap)
	CMOVQCC(l.s, l.swap)
	MOVQ(l.swap, dst.s)
}

// loads subtractin result of (left + rigth)
// left operand is overwritten if stored at register
func (l *limb) loadAdd(left *limb, rigth *limb, brw bool) {
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
	MOVQ(left.s, l.swap)
	Add(rigth.s, l.swap)
	MOVQ(l.swap, l.s)
}

// loads subtractin result of (left + rigth)
// left operand is overwritten if stored at register
func (l *limb) loadAddSafe(left *limb, rigth *limb, brw bool) {
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
	MOVQ(left.s, l.swap)
	Add(rigth.s, l.swap)
	MOVQ(l.swap, l.s)
}

// loads subtractin result of (left - rigth)
// left operand is overwritten if stored at register
func (l *limb) loadSub(left *limb, rigth *limb, brw bool) {
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
	MOVQ(left.s, l.swap)
	Sub(rigth.s, l.swap)
	MOVQ(l.swap, l.s)
}

// loads subtractin result of (left - rigth)
func (l *limb) loadSubSafe(left *limb, rigth *limb, brw bool) {
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
	MOVQ(left.s, l.swap)
	Sub(rigth.s, l.swap)
	MOVQ(l.swap, l.s)
}

func (l *limb) loadDouble(l2 *limb, brw bool) {
	Add := ADDQ
	if brw {
		Add = ADCQ
	}
	if l.atReg() {
		MOVQ(l2.s, l.s)
		Add(l.s, l.s)
		return
	}
	MOVQ(l2.s, l.swap)
	Add(l.swap, l.swap)
	MOVQ(l.swap, l.s)
}

func (l *limb) mulx(lo, hi *limb) *limb {
	MULXQ(l.s, lo.s, hi.s)
	return l
}

func (l *limb) adcxq(a *limb) *limb {
	ADCXQ(a.s, l.s)
	return l
}

func (l *limb) adoxq(a *limb) *limb {
	ADOXQ(a.s, l.s)
	return l
}

func (l *limb) mul(op, a0, a1 *limb, addOrMove bool) {
	MOVQ(l.s, RAX)
	MULQ(op.s)
	if addOrMove == _MUL_ADD {
		if a0 != nil {
			ADDQ(RAX, a0.s)
		}
		if a1 != nil {
			ADCQ(RDX, a1.s)
		}
		return
	}
	if a0 != nil {
		MOVQ(RAX, a0.s)
	}
	if a1 != nil {
		MOVQ(RDX, a1.s)
	}
}

func (l *limb) add(op *limb, car bool) *limb {
	operation := ADDQ
	if car {
		operation = ADCQ
	}
	if op.atMem() && !l.atReg() {
		op.moveTo(newLimb(l.swap), _NO_ASSIGN)
		operation(l.swap, l.s)
	} else {
		operation(op.s, l.s)
	}
	return l
}

func (l *limb) adc(op *limb) *limb {
	operation := ADCQ
	if op.atMem() && !l.atReg() {
		op.moveTo(newLimb(l.swap), _NO_ASSIGN)
		operation(l.swap, l.s)
	} else {
		operation(op.s, l.s)
	}
	return l
}

func (l *limb) adcIf(c bool, op *limb) *limb {
	if c {
		return l.adc(op)
	}
	return l
}

func (l *limb) addNoCarry(op *limb) *limb {
	operation := ADDQ
	if op.atMem() && !l.atReg() {
		op.moveTo(newLimb(l.swap), _NO_ASSIGN)
		operation(l.swap, l.s)
	} else {
		operation(op.s, l.s)
	}
	return l
}

func (l *limb) addCarry() *limb {
	ADCQ(Imm(0), l.s)
	return l
}

func (l *limb) addCarryIf(c bool) *limb {
	if c {
		ADCQ(Imm(0), l.s)
	}
	return l
}

func (l *limb) clear() *limb {
	MOVQ(U64(0), l.s)
	return l
}

func (l *limb) isEmpty() bool {
	return l.s == nil
}

func (l *limb) asMem() (Mem, bool) {
	if m, ok := l.s.(Mem); ok {
		return m, ok
	}
	return Mem{}, false
}

func (l *limb) asRegister() (Register, bool) {
	if r, ok := l.s.(Register); ok {
		return r, ok
	}
	return nil, false
}

func (l *limb) asRegisterUnchecked() Register {
	return l.s.(Register)
}

func (l *limb) asPhysical() (GPPhysical, bool) {
	if r, ok := l.s.(GPPhysical); ok {
		return r, ok
	}
	return nil, false
}

func (l *limb) asPhysicalUnchecked() GPPhysical {
	return l.s.(GPPhysical)
}

func (l *limb) clearIf(c bool) *limb {
	if c {
		MOVQ(U64(0), l.s)
	}
	return l
}

func (l *limb) cmp(op *limb) {
	CMPQ(l.s, op.s)
}

func (l *limb) xorself() *limb {
	XORQ(l.s, l.s)
	return l
}
