package x86

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func isMem(op *limb) bool {
	return IsM64(op.s)
}

func isGp(op *limb) bool {
	return IsR64(op.s)
}

type tape struct {
	gpSet     *gpSet
	stack     *stack
	swap      GPPhysical
	limbTable map[string]*limb
	reprTable map[string]*repr
}

var (
	_NO_SWAP GPPhysical = nil
)

func newTape(reserve ...Op) *tape {
	gpSet := newGpSet(RCX, RAX, RDX, RBX, RDI, RSI, R8, R9, R10, R11, R12, R13, R14, R15)
	for i := 0; i < len(reserve); i++ {
		gpSet.alloc(newLimb(reserve[i]))
	}
	stack := newStack()
	return &tape{gpSet, stack, nil, make(map[string]*limb), make(map[string]*repr)}
}

func (t tape) newReprNoAlloc(size int) *repr {
	return newReprEmpty(size)
}

func (t tape) newReprAlloc(size int) *repr {
	r := newReprEmpty(size)
	for i := 0; i < size; i++ {
		r.limbs[i].set(t.next())
	}
	return r
}

func (t tape) newLimb() *limb {
	return t.next().clone()
}

func (t *tape) newReprAllocRemainingGPRs() *repr {
	r := newReprEmpty(t.gpSet.sizeFree())
	var i = 0
	for t.gpSet.sizeFree() != 0 {
		r.limbs[i].set(t.next())
		i++
	}
	return r
}

func (t *tape) newReprAllocGPRs(upto int) *repr {
	size := t.sizeFreeGp()
	if upto < size {
		size = upto
	}
	r := newReprEmpty(size)
	for i := 0; i < size; i++ {
		R := t.next()
		if !isGp(R) {
			panic("bad allocation processing")
		}
		r.next().set(R)
	}
	return r
}

func (t *tape) allocStack(size int) *repr {
	r := newReprEmpty(size)
	for i := 0; i < size; i++ {
		r.limbs[i].set(t.stack.next())
	}
	return r
}

// func (t *tape) newReprAtParam(size int, param string, dst Register, offset int) *repr {
func (t *tape) newReprAtParam(size int, param string, dst *limb, offset int) *repr {
	t.allocGp(dst)
	r, ok := dst.asPhysical()
	if !ok {
		panic("bad register for input")
	}
	return t.newReprAtMemory(
		size,
		Mem{
			Base: Load(Param(param), r),
		},
		offset,
	)
}

func (t *tape) ax() *limb {
	return newLimb(RAX)
}

func (t *tape) bx() *limb {
	return newLimb(RBX)
}

func (t *tape) cx() *limb {
	return newLimb(RCX)
}

func (t *tape) dx() *limb {
	return newLimb(RDX)
}

func (t *tape) si() *limb {
	return newLimb(RSI)
}

func (t *tape) di() *limb {
	return newLimb(RDI)
}

func (t *tape) newReprAtMemory(size int, base Mem, offset int) *repr {
	number := make([]*limb, size)
	for i := offset; i < size+offset; i++ {
		number[i-offset] = newLimb(base.Offset(int(i * 8)))
	}
	return &repr{number, 0, size, newLimb(base.Base), t.swap}
}

func (t *tape) newLimbAtMemory(size int, base Mem) *repr {
	number := make([]*limb, size)
	for i := 0; i < size; i++ {
		number[i] = newLimb(base.Offset(int(i * 8)))
	}
	return &repr{number, 0, size, newLimb(base.Base), t.swap}
}

func (t *tape) next() *limb {
	if op := t.gpSet.next(); op != nil {
		return op
	}
	return t.stack.next()
}

func (t *tape) alloc(ops ...*limb) *tape {
	for i := 0; i < len(ops); i++ {
		op := ops[i]
		if isGp(op) {
			t.gpSet.alloc(op)
		} else if isMem(op) {
			t.stack.free(op)
		}
	}
	return t
}

func (t *tape) free(ops ...*limb) *tape {
	for i := 0; i < len(ops); i++ {
		op := ops[i]
		if isGp(op) {
			t.gpSet.free(op)
		} else if isMem(op) {
			t.stack.free(op)
		}
	}
	return t
}

func (t *tape) freeIf(c bool, ops ...*limb) *tape {
	if c {
		return t.free(ops...)
	}
	return t
}

func (t *tape) allocGp(gps ...*limb) {
	t.gpSet.alloc(gps...)
}

func (t *tape) sizeFreeGp() int {
	return t.gpSet.sizeFree()
}

func (t *tape) moveToStack(r *repr) *repr {
	stack := t.stack
	for i := 0; i < r.size; i++ {
		ri := r.at(i)
		if ri != nil && ri.atReg() {
			ri.moveAssign(stack.next())
		}
	}
	return r
}

func (t *tape) ret() {
	t.stack.allocLocal()
	RET()
	commentHeader("end")
}

func (t *tape) moveAssign(r *limb) {
	a := t.next()
	t.free(r.clone())
	r.moveAssign(a)
}

func (t *tape) setLimbForKey(s string, limb *limb) *limb {
	t.limbTable[s] = limb
	return limb
}

func (t *tape) lookupLimb(s string) *limb {
	return t.limbTable[s]
}

func (t *tape) setReprForKey(s string, r *repr) *repr {
	t.reprTable[s] = r
	return r
}

func (t *tape) lookupRepr(s string) *repr {
	return t.reprTable[s]
}

func (t *tape) debug() {
	fmt.Printf("--------------\n\n")
	fmt.Printf("Register Debug\n")
	t.gpSet.debug()
	t.stack.debug()
	if t.swap != nil {
		fmt.Printf("Swap: %s\n", t.swap.Asm())
	} else {
		fmt.Printf("No Swap\n")
	}
	fmt.Printf("--------------\n")
}

type gpSet struct {
	allocated map[GPPhysical]bool
	regs      map[int]GPPhysical
	size      int
}

func newGpSet(regs ...Op) *gpSet {
	allocated := make(map[GPPhysical]bool)
	regs_ := make(map[int]GPPhysical)
	for i, reg := range regs {
		if reg, ok := reg.(GPPhysical); ok {
			allocated[reg] = false
			regs_[i] = reg
		} else {
			panic("bad operand for general purpose set")
		}
	}
	return &gpSet{allocated: allocated, regs: regs_, size: len(regs)}
}

func (set *gpSet) alloc(regs ...*limb) {
	for _, reg := range regs {
		if reg, ok := reg.asPhysical(); ok {
			set.allocated[reg] = true
		}
	}
}

func (set *gpSet) free(regs ...*limb) {
	for _, reg := range regs {
		if reg, ok := reg.asPhysical(); ok {
			set.allocated[reg] = false
		}
	}
}

func (set *gpSet) freeAll() {
	for r := range set.allocated {
		set.allocated[r] = false
	}
}

func (set *gpSet) next() *limb {
	for i := 0; i < set.size; i++ {
		r := set.regs[i]
		if !set.allocated[r] {
			set.allocated[r] = true
			return newLimb(r)
		}
	}
	return nil
}

func (set *gpSet) sizeFree() int {
	c := 0
	for i := 0; i < set.size; i++ {
		r := set.regs[i]
		if !set.allocated[r] {
			c++
		}
	}
	return c
}

func (set *gpSet) sizeAllocated() int {
	c := 0
	for i := 0; i < set.size; i++ {
		r := set.regs[i]
		if set.allocated[r] {
			c++
		}
	}
	return c
}

func (set *gpSet) slice() []GPPhysical {
	regs := make([]GPPhysical, set.size)
	for i, r := range set.regs {
		regs[i] = r
	}
	return regs
}

func (set *gpSet) debug() {
	fmt.Printf("GP: %d/%d\n", set.sizeAllocated(), set.size)
	for i := 0; i < set.size; i++ {
		reg := set.regs[i]
		fmt.Printf("%s\t", reg.Asm())
		if set.allocated[reg] {
			fmt.Printf("ALLOC\n")
		} else {
			fmt.Printf("FREE\n")
		}
	}
	fmt.Printf("\n")
}

// stack manager with 8 byte slots
type stack struct {
	head      Mem
	allocated map[int]bool
	size      int
}

func newStack() *stack {
	allocated := make(map[int]bool)
	return &stack{
		head:      NewStackAddr(0),
		allocated: allocated,
		size:      0,
	}
}

func (s *stack) allocLocal() {
	AllocLocal(s.size * 8)
}

func (s *stack) allocLocalFineTuned(finetune int) {
	AllocLocal((s.size + finetune) * 8)
}

func (s *stack) extend(size int) Mem {
	offset := s.size * 8
	for i := s.size; i < s.size+size; i++ {
		s.allocated[i] = true
	}
	s.size += size
	return s.head.Offset(offset)
}

func (s *stack) next() *limb {
	// look up for free stack slot
	for i := 0; i < s.size; i++ {
		if !s.allocated[i] {
			s.allocated[i] = true
			m := s.head.Offset(8 * i)
			return newLimb(m)
		}
	}
	// else extend by one
	m := s.extend(1)
	return newLimb(m)
}

func (s *stack) free(mems ...*limb) {
	for _, l := range mems {
		if mem, ok := l.asMem(); ok {
			s.allocated[mem.Disp/8] = false
		}
	}
}

func (s stack) sizeFree() int {
	c := 0
	for i := 0; i < s.size; i++ {
		if !s.allocated[i] {
			c++
		}
	}
	return c
}

func (s *stack) sizeAllocated() int {
	c := 0
	for i := 0; i < s.size; i++ {
		if s.allocated[i] {
			c++
		}
	}
	return c
}

func (s *stack) freeAll() {
	for r := range s.allocated {
		s.allocated[r] = false
	}
}

func (s *stack) debug() {
	fmt.Printf("Stack: %d/%d\n", s.sizeAllocated(), s.size)
	for i := 0; i < s.size; i++ {
		fmt.Printf("%d\t", i*8)
		if s.allocated[i] {
			fmt.Printf("ALLOC\n")
		} else {
			fmt.Printf("FREE\n")
		}
	}
	fmt.Printf("\n")
}
