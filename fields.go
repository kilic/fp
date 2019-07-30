package fp

import (
	"crypto/rand"
	"io"
	"math/big"
)

var inp4 uint64
var modulus4 Fe256

type Field256 struct {
	// p2  = p-2
	// r1  = r modp
	// r2  = r^2 modp
	pBig *big.Int
	r1   *Fe256
	r2   *Fe256
	P    *Fe256
}

func NewField256(p []byte) *Field256 {
	if len(p) > 256 {
		return nil
	}
	modulus4 = *new(Fe256).Unmarshal(p)
	pBig := new(big.Int).SetBytes(p)
	inpT := new(big.Int).ModInverse(new(big.Int).Neg(pBig), new(big.Int).SetBit(new(big.Int), 64, 1))
	if inpT == nil {
		return nil
	}
	inp4 = inpT.Uint64()
	r1Big := new(big.Int).SetBit(new(big.Int), 256, 1)
	r1 := new(Fe256).SetBig(new(big.Int).Mod(r1Big, pBig))
	r2 := new(Fe256).SetBig(new(big.Int).Exp(r1Big, new(big.Int).SetUint64(2), pBig))
	return &Field256{
		pBig: pBig,
		r1:   r1,
		r2:   r2,
		P:    &modulus4}
}

func (f *Field256) NewElementFromBytes(in []byte) *Fe256 {
	fe := new(Fe256).Unmarshal(in)
	f.Mul(fe, fe, f.r2)
	return fe
}

func (f *Field256) NewElementFromUint(in uint64) *Fe256 {
	fe := &Fe256{in}
	if in == 0 {
		return fe
	}
	montmul4(fe, fe, f.r2)
	return fe
}

func (f *Field256) NewElementFromBig(in *big.Int) *Fe256 {
	fe := new(Fe256).SetBig(in)
	montmul4(fe, fe, f.r2)
	return fe
}

func (f *Field256) NewElementFromString(in string) (*Fe256, error) {
	fe, err := new(Fe256).SetString(in)
	if err != nil {
		return nil, err
	}
	montmul4(fe, fe, f.r2)
	return fe, nil
}

func (f *Field256) Zero() *Fe256 {
	return new(Fe256).SetUint(0)
}

func (f *Field256) One() *Fe256 {
	return new(Fe256).Set(f.r1)
}

func (f *Field256) Copy(dst *Fe256, src *Fe256) *Fe256 {
	return dst.Set(src)
}

func (f *Field256) RandElement(fe *Fe256, r io.Reader) error {
	bi, err := rand.Int(r, f.pBig)
	if err != nil {
		return err
	}
	fe.SetBig(bi)
	return nil
}

func (f *Field256) Mont(c, a *Fe256) {
	montmul4(c, a, f.r2)
}

func (f *Field256) Demont(c, a *Fe256) {
	mont4(c, &[8]uint64{
		a[0], a[1], a[2], a[3]})
}

func (f *Field256) Add(c, a, b *Fe256) {
	add4(c, a, b)
}

func (f *Field256) Double(c, a *Fe256) {
	double4(c, a)
}

func (f *Field256) Sub(c, a, b *Fe256) {
	sub4(c, a, b)
}

func (f *Field256) Neg(c, a *Fe256) {
	neg4(c, a)
}

func (f *Field256) Square(c, a *Fe256) {
	montmul4(c, a, a)
}

func (f *Field256) Mul(c, a, b *Fe256) {
	montmul4(c, a, b)
}

func (f *Field256) Exp(c, a *Fe256, e *big.Int) {
	z := new(Fe256).Set(f.r1)
	for i := e.BitLen(); i >= 0; i-- {
		montmul4(z, z, z)
		if e.Bit(i) == 1 {
			montmul4(z, z, a)
		}
	}
	c.Set(z)
}

func (f *Field256) InvMontUp(inv, fe *Fe256) {
	u := new(Fe256).Set(&modulus4)
	v := new(Fe256).Set(fe)
	s := &Fe256{1, 0, 0, 0}
	r := &Fe256{0, 0, 0, 0}
	var k int
	var z uint64
	var found = false
	// Phase 1
	for i := 0; i < 256*2; i++ {
		if v.IsZero() {
			found = true
			break
		}
		if u.IsEven() {
			u.div2(0)
			s.mul2()
		} else if v.IsEven() {
			v.div2(0)
			z += r.mul2()
		} else if u.Cmp(v) == 1 {
			subn4(u, v)
			u.div2(0)
			addn4(r, s)
			s.mul2()
		} else {
			subn4(v, u)
			v.div2(0)
			addn4(s, r)
			z += r.mul2()
		}
		k += 1
	}
	if found && k > 256 {
		if r.Cmp(&modulus4) != -1 || z > 0 {
			subn4(r, &modulus4)
		}
		u.Set(&modulus4)
		subn4(u, r)
		// Phase 2
		for i := k; i < 256*2; i++ {
			double4(u, u)
		}
		inv.Set(u)
	} else {
		inv.Set(&Fe256{0, 0, 0, 0})
	}
}

func (f *Field256) InvMontDown(inv, fe *Fe256) {
	u := new(Fe256).Set(&modulus4)
	v := new(Fe256).Set(fe)
	s := &Fe256{1, 0, 0, 0}
	r := &Fe256{0, 0, 0, 0}
	var k int
	var z uint64
	var found = false
	// Phase 1
	for i := 0; i < 256*2; i++ {
		if v.IsZero() {
			found = true
			break
		}
		if u.IsEven() {
			u.div2(0)
			s.mul2()
		} else if v.IsEven() {
			v.div2(0)
			z += r.mul2()
		} else if u.Cmp(v) == 1 {
			subn4(u, v)
			u.div2(0)
			addn4(r, s)
			s.mul2()
		} else {
			subn4(v, u)
			v.div2(0)
			addn4(s, r)
			z += r.mul2()
		}
		k += 1
	}
	if found && k > 256 {
		if r.Cmp(&modulus4) != -1 || z > 0 {
			subn4(r, &modulus4)
		}
		u.Set(&modulus4)
		subn4(u, r)
		// Phase 2
		var e uint64
		for i := 0; i < k-256; i++ {
			if u.IsEven() {
				u.div2(0)
			} else {
				e = addn4(u, &modulus4)
				u.div2(e)
			}
		}
		inv.Set(u)
	} else {
		inv.Set(&Fe256{0, 0, 0, 0})
	}
}

func (f *Field256) InvEEA(inv, fe *Fe256) {
	u := new(Fe256).Set(fe)
	v := new(Fe256).Set(&modulus4)
	p := new(Fe256).Set(&modulus4)
	x1 := &Fe256{1}
	x2 := &Fe256{0}
	var e uint64
	for !u.IsOne() && !v.IsOne() {
		for u.IsEven() {
			u.div2(0)
			if x1.IsEven() {
				x1.div2(0)
			} else {
				e = addn4(x1, p)
				x1.div2(e)
			}
		}
		for v.IsEven() {
			v.div2(0)
			if x2.IsEven() {
				x2.div2(0)
			} else {
				e = addn4(x2, p)
				x2.div2(e)
			}
		}
		if u.Cmp(v) == -1 {
			subn4(v, u)
			sub4(x2, x2, x1)
		} else {
			subn4(u, v)
			sub4(x1, x1, x2)
		}
	}
	if u.IsOne() {
		inv.Set(x1)
		return
	}
	inv.Set(x2)
}

var inp5 uint64
var modulus5 Fe320

type Field320 struct {
	// p2  = p-2
	// r1  = r modp
	// r2  = r^2 modp
	pBig *big.Int
	r1   *Fe320
	r2   *Fe320
	P    *Fe320
}

func NewField320(p []byte) *Field320 {
	if len(p) > 320 {
		return nil
	}
	modulus5 = *new(Fe320).Unmarshal(p)
	pBig := new(big.Int).SetBytes(p)
	inpT := new(big.Int).ModInverse(new(big.Int).Neg(pBig), new(big.Int).SetBit(new(big.Int), 64, 1))
	if inpT == nil {
		return nil
	}
	inp5 = inpT.Uint64()
	r1Big := new(big.Int).SetBit(new(big.Int), 320, 1)
	r1 := new(Fe320).SetBig(new(big.Int).Mod(r1Big, pBig))
	r2 := new(Fe320).SetBig(new(big.Int).Exp(r1Big, new(big.Int).SetUint64(2), pBig))
	return &Field320{
		pBig: pBig,
		r1:   r1,
		r2:   r2,
		P:    &modulus5}
}

func (f *Field320) NewElementFromBytes(in []byte) *Fe320 {
	fe := new(Fe320).Unmarshal(in)
	f.Mul(fe, fe, f.r2)
	return fe
}

func (f *Field320) NewElementFromUint(in uint64) *Fe320 {
	fe := &Fe320{in}
	if in == 0 {
		return fe
	}
	montmul5(fe, fe, f.r2)
	return fe
}

func (f *Field320) NewElementFromBig(in *big.Int) *Fe320 {
	fe := new(Fe320).SetBig(in)
	montmul5(fe, fe, f.r2)
	return fe
}

func (f *Field320) NewElementFromString(in string) (*Fe320, error) {
	fe, err := new(Fe320).SetString(in)
	if err != nil {
		return nil, err
	}
	montmul5(fe, fe, f.r2)
	return fe, nil
}

func (f *Field320) Zero() *Fe320 {
	return new(Fe320).SetUint(0)
}

func (f *Field320) One() *Fe320 {
	return new(Fe320).Set(f.r1)
}

func (f *Field320) Copy(dst *Fe320, src *Fe320) *Fe320 {
	return dst.Set(src)
}

func (f *Field320) RandElement(fe *Fe320, r io.Reader) error {
	bi, err := rand.Int(r, f.pBig)
	if err != nil {
		return err
	}
	fe.SetBig(bi)
	return nil
}

func (f *Field320) Mont(c, a *Fe320) {
	montmul5(c, a, f.r2)
}

func (f *Field320) Demont(c, a *Fe320) {
	mont5(c, &[10]uint64{
		a[0], a[1], a[2], a[3], a[4]})
}

func (f *Field320) Add(c, a, b *Fe320) {
	add5(c, a, b)
}

func (f *Field320) Double(c, a *Fe320) {
	double5(c, a)
}

func (f *Field320) Sub(c, a, b *Fe320) {
	sub5(c, a, b)
}

func (f *Field320) Neg(c, a *Fe320) {
	neg5(c, a)
}

func (f *Field320) Square(c, a *Fe320) {
	montmul5(c, a, a)
}

func (f *Field320) Mul(c, a, b *Fe320) {
	montmul5(c, a, b)
}

func (f *Field320) Exp(c, a *Fe320, e *big.Int) {
	z := new(Fe320).Set(f.r1)
	for i := e.BitLen(); i >= 0; i-- {
		montmul5(z, z, z)
		if e.Bit(i) == 1 {
			montmul5(z, z, a)
		}
	}
	c.Set(z)
}

func (f *Field320) InvMontUp(inv, fe *Fe320) {
	u := new(Fe320).Set(&modulus5)
	v := new(Fe320).Set(fe)
	s := &Fe320{1, 0, 0, 0}
	r := &Fe320{0, 0, 0, 0}
	var k int
	var z uint64
	var found = false
	// Phase 1
	for i := 0; i < 320*2; i++ {
		if v.IsZero() {
			found = true
			break
		}
		if u.IsEven() {
			u.div2(0)
			s.mul2()
		} else if v.IsEven() {
			v.div2(0)
			z += r.mul2()
		} else if u.Cmp(v) == 1 {
			subn5(u, v)
			u.div2(0)
			addn5(r, s)
			s.mul2()
		} else {
			subn5(v, u)
			v.div2(0)
			addn5(s, r)
			z += r.mul2()
		}
		k += 1
	}
	if found && k > 320 {
		if r.Cmp(&modulus5) != -1 || z > 0 {
			subn5(r, &modulus5)
		}
		u.Set(&modulus5)
		subn5(u, r)
		// Phase 2
		for i := k; i < 320*2; i++ {
			double5(u, u)
		}
		inv.Set(u)
	} else {
		inv.Set(&Fe320{0, 0, 0, 0})
	}
}

func (f *Field320) InvMontDown(inv, fe *Fe320) {
	u := new(Fe320).Set(&modulus5)
	v := new(Fe320).Set(fe)
	s := &Fe320{1, 0, 0, 0}
	r := &Fe320{0, 0, 0, 0}
	var k int
	var z uint64
	var found = false
	// Phase 1
	for i := 0; i < 320*2; i++ {
		if v.IsZero() {
			found = true
			break
		}
		if u.IsEven() {
			u.div2(0)
			s.mul2()
		} else if v.IsEven() {
			v.div2(0)
			z += r.mul2()
		} else if u.Cmp(v) == 1 {
			subn5(u, v)
			u.div2(0)
			addn5(r, s)
			s.mul2()
		} else {
			subn5(v, u)
			v.div2(0)
			addn5(s, r)
			z += r.mul2()
		}
		k += 1
	}
	if found && k > 320 {
		if r.Cmp(&modulus5) != -1 || z > 0 {
			subn5(r, &modulus5)
		}
		u.Set(&modulus5)
		subn5(u, r)
		// Phase 2
		var e uint64
		for i := 0; i < k-320; i++ {
			if u.IsEven() {
				u.div2(0)
			} else {
				e = addn5(u, &modulus5)
				u.div2(e)
			}
		}
		inv.Set(u)
	} else {
		inv.Set(&Fe320{0, 0, 0, 0})
	}
}

func (f *Field320) InvEEA(inv, fe *Fe320) {
	u := new(Fe320).Set(fe)
	v := new(Fe320).Set(&modulus5)
	p := new(Fe320).Set(&modulus5)
	x1 := &Fe320{1}
	x2 := &Fe320{0}
	var e uint64
	for !u.IsOne() && !v.IsOne() {
		for u.IsEven() {
			u.div2(0)
			if x1.IsEven() {
				x1.div2(0)
			} else {
				e = addn5(x1, p)
				x1.div2(e)
			}
		}
		for v.IsEven() {
			v.div2(0)
			if x2.IsEven() {
				x2.div2(0)
			} else {
				e = addn5(x2, p)
				x2.div2(e)
			}
		}
		if u.Cmp(v) == -1 {
			subn5(v, u)
			sub5(x2, x2, x1)
		} else {
			subn5(u, v)
			sub5(x1, x1, x2)
		}
	}
	if u.IsOne() {
		inv.Set(x1)
		return
	}
	inv.Set(x2)
}

var inp6 uint64
var modulus6 Fe384

type Field384 struct {
	// p2  = p-2
	// r1  = r modp
	// r2  = r^2 modp
	pBig *big.Int
	r1   *Fe384
	r2   *Fe384
	P    *Fe384
}

func NewField384(p []byte) *Field384 {
	if len(p) > 384 {
		return nil
	}
	modulus6 = *new(Fe384).Unmarshal(p)
	pBig := new(big.Int).SetBytes(p)
	inpT := new(big.Int).ModInverse(new(big.Int).Neg(pBig), new(big.Int).SetBit(new(big.Int), 64, 1))
	if inpT == nil {
		return nil
	}
	inp6 = inpT.Uint64()
	r1Big := new(big.Int).SetBit(new(big.Int), 384, 1)
	r1 := new(Fe384).SetBig(new(big.Int).Mod(r1Big, pBig))
	r2 := new(Fe384).SetBig(new(big.Int).Exp(r1Big, new(big.Int).SetUint64(2), pBig))
	return &Field384{
		pBig: pBig,
		r1:   r1,
		r2:   r2,
		P:    &modulus6}
}

func (f *Field384) NewElementFromBytes(in []byte) *Fe384 {
	fe := new(Fe384).Unmarshal(in)
	f.Mul(fe, fe, f.r2)
	return fe
}

func (f *Field384) NewElementFromUint(in uint64) *Fe384 {
	fe := &Fe384{in}
	if in == 0 {
		return fe
	}
	montmul6(fe, fe, f.r2)
	return fe
}

func (f *Field384) NewElementFromBig(in *big.Int) *Fe384 {
	fe := new(Fe384).SetBig(in)
	montmul6(fe, fe, f.r2)
	return fe
}

func (f *Field384) NewElementFromString(in string) (*Fe384, error) {
	fe, err := new(Fe384).SetString(in)
	if err != nil {
		return nil, err
	}
	montmul6(fe, fe, f.r2)
	return fe, nil
}

func (f *Field384) Zero() *Fe384 {
	return new(Fe384).SetUint(0)
}

func (f *Field384) One() *Fe384 {
	return new(Fe384).Set(f.r1)
}

func (f *Field384) Copy(dst *Fe384, src *Fe384) *Fe384 {
	return dst.Set(src)
}

func (f *Field384) RandElement(fe *Fe384, r io.Reader) error {
	bi, err := rand.Int(r, f.pBig)
	if err != nil {
		return err
	}
	fe.SetBig(bi)
	return nil
}

func (f *Field384) Mont(c, a *Fe384) {
	montmul6(c, a, f.r2)
}

func (f *Field384) Demont(c, a *Fe384) {
	mont6(c, &[12]uint64{
		a[0], a[1], a[2], a[3], a[4], a[5]})
}

func (f *Field384) Add(c, a, b *Fe384) {
	add6(c, a, b)
}

func (f *Field384) Double(c, a *Fe384) {
	double6(c, a)
}

func (f *Field384) Sub(c, a, b *Fe384) {
	sub6(c, a, b)
}

func (f *Field384) Neg(c, a *Fe384) {
	neg6(c, a)
}

func (f *Field384) Square(c, a *Fe384) {
	montmul6(c, a, a)
}

func (f *Field384) Mul(c, a, b *Fe384) {
	montmul6(c, a, b)
}

func (f *Field384) Exp(c, a *Fe384, e *big.Int) {
	z := new(Fe384).Set(f.r1)
	for i := e.BitLen(); i >= 0; i-- {
		montmul6(z, z, z)
		if e.Bit(i) == 1 {
			montmul6(z, z, a)
		}
	}
	c.Set(z)
}

func (f *Field384) InvMontUp(inv, fe *Fe384) {
	u := new(Fe384).Set(&modulus6)
	v := new(Fe384).Set(fe)
	s := &Fe384{1, 0, 0, 0}
	r := &Fe384{0, 0, 0, 0}
	var k int
	var z uint64
	var found = false
	// Phase 1
	for i := 0; i < 384*2; i++ {
		if v.IsZero() {
			found = true
			break
		}
		if u.IsEven() {
			u.div2(0)
			s.mul2()
		} else if v.IsEven() {
			v.div2(0)
			z += r.mul2()
		} else if u.Cmp(v) == 1 {
			subn6(u, v)
			u.div2(0)
			addn6(r, s)
			s.mul2()
		} else {
			subn6(v, u)
			v.div2(0)
			addn6(s, r)
			z += r.mul2()
		}
		k += 1
	}
	if found && k > 384 {
		if r.Cmp(&modulus6) != -1 || z > 0 {
			subn6(r, &modulus6)
		}
		u.Set(&modulus6)
		subn6(u, r)
		// Phase 2
		for i := k; i < 384*2; i++ {
			double6(u, u)
		}
		inv.Set(u)
	} else {
		inv.Set(&Fe384{0, 0, 0, 0})
	}
}

func (f *Field384) InvMontDown(inv, fe *Fe384) {
	u := new(Fe384).Set(&modulus6)
	v := new(Fe384).Set(fe)
	s := &Fe384{1, 0, 0, 0}
	r := &Fe384{0, 0, 0, 0}
	var k int
	var z uint64
	var found = false
	// Phase 1
	for i := 0; i < 384*2; i++ {
		if v.IsZero() {
			found = true
			break
		}
		if u.IsEven() {
			u.div2(0)
			s.mul2()
		} else if v.IsEven() {
			v.div2(0)
			z += r.mul2()
		} else if u.Cmp(v) == 1 {
			subn6(u, v)
			u.div2(0)
			addn6(r, s)
			s.mul2()
		} else {
			subn6(v, u)
			v.div2(0)
			addn6(s, r)
			z += r.mul2()
		}
		k += 1
	}
	if found && k > 384 {
		if r.Cmp(&modulus6) != -1 || z > 0 {
			subn6(r, &modulus6)
		}
		u.Set(&modulus6)
		subn6(u, r)
		// Phase 2
		var e uint64
		for i := 0; i < k-384; i++ {
			if u.IsEven() {
				u.div2(0)
			} else {
				e = addn6(u, &modulus6)
				u.div2(e)
			}
		}
		inv.Set(u)
	} else {
		inv.Set(&Fe384{0, 0, 0, 0})
	}
}

func (f *Field384) InvEEA(inv, fe *Fe384) {
	u := new(Fe384).Set(fe)
	v := new(Fe384).Set(&modulus6)
	p := new(Fe384).Set(&modulus6)
	x1 := &Fe384{1}
	x2 := &Fe384{0}
	var e uint64
	for !u.IsOne() && !v.IsOne() {
		for u.IsEven() {
			u.div2(0)
			if x1.IsEven() {
				x1.div2(0)
			} else {
				e = addn6(x1, p)
				x1.div2(e)
			}
		}
		for v.IsEven() {
			v.div2(0)
			if x2.IsEven() {
				x2.div2(0)
			} else {
				e = addn6(x2, p)
				x2.div2(e)
			}
		}
		if u.Cmp(v) == -1 {
			subn6(v, u)
			sub6(x2, x2, x1)
		} else {
			subn6(u, v)
			sub6(x1, x1, x2)
		}
	}
	if u.IsOne() {
		inv.Set(x1)
		return
	}
	inv.Set(x2)
}

var inp7 uint64
var modulus7 Fe448

type Field448 struct {
	// p2  = p-2
	// r1  = r modp
	// r2  = r^2 modp
	pBig *big.Int
	r1   *Fe448
	r2   *Fe448
	P    *Fe448
}

func NewField448(p []byte) *Field448 {
	if len(p) > 448 {
		return nil
	}
	modulus7 = *new(Fe448).Unmarshal(p)
	pBig := new(big.Int).SetBytes(p)
	inpT := new(big.Int).ModInverse(new(big.Int).Neg(pBig), new(big.Int).SetBit(new(big.Int), 64, 1))
	if inpT == nil {
		return nil
	}
	inp7 = inpT.Uint64()
	r1Big := new(big.Int).SetBit(new(big.Int), 448, 1)
	r1 := new(Fe448).SetBig(new(big.Int).Mod(r1Big, pBig))
	r2 := new(Fe448).SetBig(new(big.Int).Exp(r1Big, new(big.Int).SetUint64(2), pBig))
	return &Field448{
		pBig: pBig,
		r1:   r1,
		r2:   r2,
		P:    &modulus7}
}

func (f *Field448) NewElementFromBytes(in []byte) *Fe448 {
	fe := new(Fe448).Unmarshal(in)
	f.Mul(fe, fe, f.r2)
	return fe
}

func (f *Field448) NewElementFromUint(in uint64) *Fe448 {
	fe := &Fe448{in}
	if in == 0 {
		return fe
	}
	montmul7(fe, fe, f.r2)
	return fe
}

func (f *Field448) NewElementFromBig(in *big.Int) *Fe448 {
	fe := new(Fe448).SetBig(in)
	montmul7(fe, fe, f.r2)
	return fe
}

func (f *Field448) NewElementFromString(in string) (*Fe448, error) {
	fe, err := new(Fe448).SetString(in)
	if err != nil {
		return nil, err
	}
	montmul7(fe, fe, f.r2)
	return fe, nil
}

func (f *Field448) Zero() *Fe448 {
	return new(Fe448).SetUint(0)
}

func (f *Field448) One() *Fe448 {
	return new(Fe448).Set(f.r1)
}

func (f *Field448) Copy(dst *Fe448, src *Fe448) *Fe448 {
	return dst.Set(src)
}

func (f *Field448) RandElement(fe *Fe448, r io.Reader) error {
	bi, err := rand.Int(r, f.pBig)
	if err != nil {
		return err
	}
	fe.SetBig(bi)
	return nil
}

func (f *Field448) Mont(c, a *Fe448) {
	montmul7(c, a, f.r2)
}

func (f *Field448) Demont(c, a *Fe448) {
	mont7(c, &[14]uint64{
		a[0], a[1], a[2], a[3], a[4], a[5], a[6]})
}

func (f *Field448) Add(c, a, b *Fe448) {
	add7(c, a, b)
}

func (f *Field448) Double(c, a *Fe448) {
	double7(c, a)
}

func (f *Field448) Sub(c, a, b *Fe448) {
	sub7(c, a, b)
}

func (f *Field448) Neg(c, a *Fe448) {
	neg7(c, a)
}

func (f *Field448) Square(c, a *Fe448) {
	montmul7(c, a, a)
}

func (f *Field448) Mul(c, a, b *Fe448) {
	montmul7(c, a, b)
}

func (f *Field448) Exp(c, a *Fe448, e *big.Int) {
	z := new(Fe448).Set(f.r1)
	for i := e.BitLen(); i >= 0; i-- {
		montmul7(z, z, z)
		if e.Bit(i) == 1 {
			montmul7(z, z, a)
		}
	}
	c.Set(z)
}

func (f *Field448) InvMontUp(inv, fe *Fe448) {
	u := new(Fe448).Set(&modulus7)
	v := new(Fe448).Set(fe)
	s := &Fe448{1, 0, 0, 0}
	r := &Fe448{0, 0, 0, 0}
	var k int
	var z uint64
	var found = false
	// Phase 1
	for i := 0; i < 448*2; i++ {
		if v.IsZero() {
			found = true
			break
		}
		if u.IsEven() {
			u.div2(0)
			s.mul2()
		} else if v.IsEven() {
			v.div2(0)
			z += r.mul2()
		} else if u.Cmp(v) == 1 {
			subn7(u, v)
			u.div2(0)
			addn7(r, s)
			s.mul2()
		} else {
			subn7(v, u)
			v.div2(0)
			addn7(s, r)
			z += r.mul2()
		}
		k += 1
	}
	if found && k > 448 {
		if r.Cmp(&modulus7) != -1 || z > 0 {
			subn7(r, &modulus7)
		}
		u.Set(&modulus7)
		subn7(u, r)
		// Phase 2
		for i := k; i < 448*2; i++ {
			double7(u, u)
		}
		inv.Set(u)
	} else {
		inv.Set(&Fe448{0, 0, 0, 0})
	}
}

func (f *Field448) InvMontDown(inv, fe *Fe448) {
	u := new(Fe448).Set(&modulus7)
	v := new(Fe448).Set(fe)
	s := &Fe448{1, 0, 0, 0}
	r := &Fe448{0, 0, 0, 0}
	var k int
	var z uint64
	var found = false
	// Phase 1
	for i := 0; i < 448*2; i++ {
		if v.IsZero() {
			found = true
			break
		}
		if u.IsEven() {
			u.div2(0)
			s.mul2()
		} else if v.IsEven() {
			v.div2(0)
			z += r.mul2()
		} else if u.Cmp(v) == 1 {
			subn7(u, v)
			u.div2(0)
			addn7(r, s)
			s.mul2()
		} else {
			subn7(v, u)
			v.div2(0)
			addn7(s, r)
			z += r.mul2()
		}
		k += 1
	}
	if found && k > 448 {
		if r.Cmp(&modulus7) != -1 || z > 0 {
			subn7(r, &modulus7)
		}
		u.Set(&modulus7)
		subn7(u, r)
		// Phase 2
		var e uint64
		for i := 0; i < k-448; i++ {
			if u.IsEven() {
				u.div2(0)
			} else {
				e = addn7(u, &modulus7)
				u.div2(e)
			}
		}
		inv.Set(u)
	} else {
		inv.Set(&Fe448{0, 0, 0, 0})
	}
}

func (f *Field448) InvEEA(inv, fe *Fe448) {
	u := new(Fe448).Set(fe)
	v := new(Fe448).Set(&modulus7)
	p := new(Fe448).Set(&modulus7)
	x1 := &Fe448{1}
	x2 := &Fe448{0}
	var e uint64
	for !u.IsOne() && !v.IsOne() {
		for u.IsEven() {
			u.div2(0)
			if x1.IsEven() {
				x1.div2(0)
			} else {
				e = addn7(x1, p)
				x1.div2(e)
			}
		}
		for v.IsEven() {
			v.div2(0)
			if x2.IsEven() {
				x2.div2(0)
			} else {
				e = addn7(x2, p)
				x2.div2(e)
			}
		}
		if u.Cmp(v) == -1 {
			subn7(v, u)
			sub7(x2, x2, x1)
		} else {
			subn7(u, v)
			sub7(x1, x1, x2)
		}
	}
	if u.IsOne() {
		inv.Set(x1)
		return
	}
	inv.Set(x2)
}

var inp8 uint64
var modulus8 Fe512

type Field512 struct {
	// p2  = p-2
	// r1  = r modp
	// r2  = r^2 modp
	pBig *big.Int
	r1   *Fe512
	r2   *Fe512
	P    *Fe512
}

func NewField512(p []byte) *Field512 {
	if len(p) > 512 {
		return nil
	}
	modulus8 = *new(Fe512).Unmarshal(p)
	pBig := new(big.Int).SetBytes(p)
	inpT := new(big.Int).ModInverse(new(big.Int).Neg(pBig), new(big.Int).SetBit(new(big.Int), 64, 1))
	if inpT == nil {
		return nil
	}
	inp8 = inpT.Uint64()
	r1Big := new(big.Int).SetBit(new(big.Int), 512, 1)
	r1 := new(Fe512).SetBig(new(big.Int).Mod(r1Big, pBig))
	r2 := new(Fe512).SetBig(new(big.Int).Exp(r1Big, new(big.Int).SetUint64(2), pBig))
	return &Field512{
		pBig: pBig,
		r1:   r1,
		r2:   r2,
		P:    &modulus8}
}

func (f *Field512) NewElementFromBytes(in []byte) *Fe512 {
	fe := new(Fe512).Unmarshal(in)
	f.Mul(fe, fe, f.r2)
	return fe
}

func (f *Field512) NewElementFromUint(in uint64) *Fe512 {
	fe := &Fe512{in}
	if in == 0 {
		return fe
	}
	montmul8(fe, fe, f.r2)
	return fe
}

func (f *Field512) NewElementFromBig(in *big.Int) *Fe512 {
	fe := new(Fe512).SetBig(in)
	montmul8(fe, fe, f.r2)
	return fe
}

func (f *Field512) NewElementFromString(in string) (*Fe512, error) {
	fe, err := new(Fe512).SetString(in)
	if err != nil {
		return nil, err
	}
	montmul8(fe, fe, f.r2)
	return fe, nil
}

func (f *Field512) Zero() *Fe512 {
	return new(Fe512).SetUint(0)
}

func (f *Field512) One() *Fe512 {
	return new(Fe512).Set(f.r1)
}

func (f *Field512) Copy(dst *Fe512, src *Fe512) *Fe512 {
	return dst.Set(src)
}

func (f *Field512) RandElement(fe *Fe512, r io.Reader) error {
	bi, err := rand.Int(r, f.pBig)
	if err != nil {
		return err
	}
	fe.SetBig(bi)
	return nil
}

func (f *Field512) Mont(c, a *Fe512) {
	montmul8(c, a, f.r2)
}

func (f *Field512) Demont(c, a *Fe512) {
	mont8(c, &[16]uint64{
		a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7]})
}

func (f *Field512) Add(c, a, b *Fe512) {
	add8(c, a, b)
}

func (f *Field512) Double(c, a *Fe512) {
	double8(c, a)
}

func (f *Field512) Sub(c, a, b *Fe512) {
	sub8(c, a, b)
}

func (f *Field512) Neg(c, a *Fe512) {
	neg8(c, a)
}

func (f *Field512) Square(c, a *Fe512) {
	montmul8(c, a, a)
}

func (f *Field512) Mul(c, a, b *Fe512) {
	montmul8(c, a, b)
}

func (f *Field512) Exp(c, a *Fe512, e *big.Int) {
	z := new(Fe512).Set(f.r1)
	for i := e.BitLen(); i >= 0; i-- {
		montmul8(z, z, z)
		if e.Bit(i) == 1 {
			montmul8(z, z, a)
		}
	}
	c.Set(z)
}

func (f *Field512) InvMontUp(inv, fe *Fe512) {
	u := new(Fe512).Set(&modulus8)
	v := new(Fe512).Set(fe)
	s := &Fe512{1, 0, 0, 0}
	r := &Fe512{0, 0, 0, 0}
	var k int
	var z uint64
	var found = false
	// Phase 1
	for i := 0; i < 512*2; i++ {
		if v.IsZero() {
			found = true
			break
		}
		if u.IsEven() {
			u.div2(0)
			s.mul2()
		} else if v.IsEven() {
			v.div2(0)
			z += r.mul2()
		} else if u.Cmp(v) == 1 {
			subn8(u, v)
			u.div2(0)
			addn8(r, s)
			s.mul2()
		} else {
			subn8(v, u)
			v.div2(0)
			addn8(s, r)
			z += r.mul2()
		}
		k += 1
	}
	if found && k > 512 {
		if r.Cmp(&modulus8) != -1 || z > 0 {
			subn8(r, &modulus8)
		}
		u.Set(&modulus8)
		subn8(u, r)
		// Phase 2
		for i := k; i < 512*2; i++ {
			double8(u, u)
		}
		inv.Set(u)
	} else {
		inv.Set(&Fe512{0, 0, 0, 0})
	}
}

func (f *Field512) InvMontDown(inv, fe *Fe512) {
	u := new(Fe512).Set(&modulus8)
	v := new(Fe512).Set(fe)
	s := &Fe512{1, 0, 0, 0}
	r := &Fe512{0, 0, 0, 0}
	var k int
	var z uint64
	var found = false
	// Phase 1
	for i := 0; i < 512*2; i++ {
		if v.IsZero() {
			found = true
			break
		}
		if u.IsEven() {
			u.div2(0)
			s.mul2()
		} else if v.IsEven() {
			v.div2(0)
			z += r.mul2()
		} else if u.Cmp(v) == 1 {
			subn8(u, v)
			u.div2(0)
			addn8(r, s)
			s.mul2()
		} else {
			subn8(v, u)
			v.div2(0)
			addn8(s, r)
			z += r.mul2()
		}
		k += 1
	}
	if found && k > 512 {
		if r.Cmp(&modulus8) != -1 || z > 0 {
			subn8(r, &modulus8)
		}
		u.Set(&modulus8)
		subn8(u, r)
		// Phase 2
		var e uint64
		for i := 0; i < k-512; i++ {
			if u.IsEven() {
				u.div2(0)
			} else {
				e = addn8(u, &modulus8)
				u.div2(e)
			}
		}
		inv.Set(u)
	} else {
		inv.Set(&Fe512{0, 0, 0, 0})
	}
}

func (f *Field512) InvEEA(inv, fe *Fe512) {
	u := new(Fe512).Set(fe)
	v := new(Fe512).Set(&modulus8)
	p := new(Fe512).Set(&modulus8)
	x1 := &Fe512{1}
	x2 := &Fe512{0}
	var e uint64
	for !u.IsOne() && !v.IsOne() {
		for u.IsEven() {
			u.div2(0)
			if x1.IsEven() {
				x1.div2(0)
			} else {
				e = addn8(x1, p)
				x1.div2(e)
			}
		}
		for v.IsEven() {
			v.div2(0)
			if x2.IsEven() {
				x2.div2(0)
			} else {
				e = addn8(x2, p)
				x2.div2(e)
			}
		}
		if u.Cmp(v) == -1 {
			subn8(v, u)
			sub8(x2, x2, x1)
		} else {
			subn8(u, v)
			sub8(x1, x1, x2)
		}
	}
	if u.IsOne() {
		inv.Set(x1)
		return
	}
	inv.Set(x2)
}
