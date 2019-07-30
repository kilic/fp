// +build ignore

package main

var fieldTemplates = []string{
	fTmplConstants,
	fTmplFieldDef,
	fTmplNew,
	fTmplNewFeBytes,
	fTmplNewFeUint,
	fTmplNewFeBig,
	fTmplNewFeString,
	fTmplZero,
	fTmplOne,
	fTmplCopy,
	fTmplRand,
	fTmplMont,
	fTmplDemont,
	fTmplAdd,
	fTmplDouble,
	fTmplSub,
	fTmplNeg,
	fTmplSquare,
	fTmplMul,
	fTmplExp,
	fTmplInvMontUp,
	fTemplInvMontDown,
	fTmplInvEEA,
}

const (
	fTmplConstants = `
var inp{{ $N_LIMB }} uint64
var modulus{{ $N_LIMB }} Fe{{ $N_BIT }} 
`

	fTmplFieldDef = `
type Field{{ $N_BIT }} struct {
// p2  = p-2
// r1  = r modp
// r2  = r^2 modp
pBig *big.Int
r1  *Fe{{ $N_BIT }} 
r2  *Fe{{ $N_BIT }} 
P   *Fe{{ $N_BIT }} }
`

	fTmplNew = `
func NewField{{ $N_BIT }}(p []byte) *Field{{ $N_BIT }} {
if len(p) > {{ $N_BIT }} {
return nil }
modulus{{ $N_LIMB }} = *new(Fe{{ $N_BIT }}).Unmarshal(p)
pBig := new(big.Int).SetBytes(p)
inpT := new(big.Int).ModInverse(new(big.Int).Neg(pBig), new(big.Int).SetBit(new(big.Int), 64, 1))
if inpT == nil {
return nil }
inp{{ $N_LIMB }} = inpT.Uint64()
r1Big := new(big.Int).SetBit(new(big.Int), {{ $N_BIT }}, 1)
r1 := new(Fe{{ $N_BIT }}).SetBig(new(big.Int).Mod(r1Big, pBig))
r2 := new(Fe{{ $N_BIT }}).SetBig(new(big.Int).Exp(r1Big, new(big.Int).SetUint64(2), pBig))
return &Field{{ $N_BIT }}{
pBig: pBig,
r1:   r1,
r2:   r2,
P:    &modulus{{ $N_LIMB }}, }}
`

	fTmplNewFeBytes = `
func (f *Field{{ $N_BIT }}) NewElementFromBytes(in []byte) *Fe{{ $N_BIT }} {
fe := new(Fe{{ $N_BIT }}).Unmarshal(in)
f.Mul(fe, fe, f.r2)
return fe }
`

	fTmplNewFeUint = `
func (f *Field{{ $N_BIT }}) NewElementFromUint(in uint64) *Fe{{ $N_BIT }} {
fe := &Fe{{ $N_BIT }}{in}
if in == 0 {
return fe }
montmul{{ $N_LIMB }}(fe, fe, f.r2)
return fe }
`

	fTmplNewFeBig = `
func (f *Field{{ $N_BIT }}) NewElementFromBig(in *big.Int) *Fe{{ $N_BIT }} {
fe := new(Fe{{ $N_BIT }}).SetBig(in)
montmul{{ $N_LIMB }}(fe, fe, f.r2)
return fe }
`

	fTmplNewFeString = `
func (f *Field{{ $N_BIT }}) NewElementFromString(in string) (*Fe{{ $N_BIT }}, error) {
fe, err := new(Fe{{ $N_BIT }}).SetString(in)
if err != nil {
return nil, err }
montmul{{ $N_LIMB }}(fe, fe, f.r2)
return fe, nil }
`

	fTmplZero = `
func (f *Field{{ $N_BIT }}) Zero() *Fe{{ $N_BIT }} {
return new(Fe{{ $N_BIT }}).SetUint(0) }
`

	fTmplOne = `
func (f *Field{{ $N_BIT }}) One() *Fe{{ $N_BIT }} {
return new(Fe{{ $N_BIT }}).Set(f.r1) }
`

	fTmplCopy = `
func (f *Field{{ $N_BIT }}) Copy(dst *Fe{{ $N_BIT }}, src *Fe{{ $N_BIT }}) *Fe{{ $N_BIT }} {
return dst.Set(src) }
`

	fTmplRand = `
func (f *Field{{ $N_BIT }}) RandElement(fe *Fe{{ $N_BIT }}, r io.Reader) error {
bi, err := rand.Int(r, f.pBig)
if err != nil {
return err }
fe.SetBig(bi)
return nil }		
`
	fTmplMont = `
func (f *Field{{ $N_BIT }}) Mont(c, a *Fe{{ $N_BIT }}) {
montmul{{ $N_LIMB }}(c, a, f.r2) }
`

	fTmplDemont = `
func (f *Field{{ $N_BIT }}) Demont(c, a *Fe{{ $N_BIT }}) {
mont{{ $N_LIMB }}(c, &[ {{ mul $N_LIMB 2 }} ]uint64{
{{ range $x := iterUp 0 $N_LIMB }} a[ {{$x}} ], {{end}} }) }
`

	fTmplAdd = `
func (f *Field{{ $N_BIT }}) Add(c, a, b *Fe{{ $N_BIT }}) {
add{{ $N_LIMB }}(c, a, b) }
`

	fTmplDouble = `
func (f *Field{{ $N_BIT }}) Double(c, a *Fe{{ $N_BIT }}) {
double{{ $N_LIMB }}(c, a) }
`

	fTmplSub = `
func (f *Field{{ $N_BIT }}) Sub(c, a, b *Fe{{ $N_BIT }}) {
sub{{ $N_LIMB }}(c, a, b) }
`

	fTmplNeg = `
func (f *Field{{ $N_BIT }}) Neg(c, a *Fe{{ $N_BIT }}) {
neg{{ $N_LIMB }}(c, a) }
`

	fTmplSquare = `
func (f *Field{{ $N_BIT }}) Square(c, a *Fe{{ $N_BIT }}) {
montmul{{ $N_LIMB }}(c, a, a) }
`

	fTmplMul = `
func (f *Field{{ $N_BIT }}) Mul(c, a, b *Fe{{ $N_BIT }}) {
montmul{{ $N_LIMB }}(c, a, b) }
`

	fTmplExp = `
func (f *Field{{ $N_BIT }}) Exp(c, a*Fe{{ $N_BIT }}, e *big.Int) {
z := new(Fe{{ $N_BIT }}).Set(f.r1)
for i := e.BitLen(); i >= 0; i-- {
montmul{{ $N_LIMB }}(z, z, z)
if e.Bit(i) == 1 {
montmul{{ $N_LIMB }}(z, z, a) } }
c.Set(z) }
`

	fTmplInvEEA = `
func (f *Field{{ $N_BIT }}) InvEEA(inv, fe *Fe{{ $N_BIT }}) {
u := new(Fe{{ $N_BIT }}).Set(fe)
v := new(Fe{{ $N_BIT }}).Set(&modulus{{ $N_LIMB }})
p := new(Fe{{ $N_BIT }}).Set(&modulus{{ $N_LIMB }})
x1 := &Fe{{ $N_BIT }}{1}
x2 := &Fe{{ $N_BIT }}{0}
var e uint64
for !u.IsOne() && !v.IsOne() {
for u.IsEven() {
u.div2(0)
if x1.IsEven() {
x1.div2(0)
} else {
e = addn{{ $N_LIMB }}(x1, p)
x1.div2(e) }}
for v.IsEven() {
v.div2(0)
if x2.IsEven() {
x2.div2(0)
} else {
e = addn{{ $N_LIMB }}(x2, p)
x2.div2(e) }}
if u.Cmp(v) == -1 {
subn{{ $N_LIMB }}(v, u)
sub{{ $N_LIMB }}(x2, x2, x1)
} else {
subn{{ $N_LIMB }}(u, v)
sub{{ $N_LIMB }}(x1, x1, x2) }}
if u.IsOne() {
inv.Set(x1)
return }
inv.Set(x2)}
`

	fTmplInvMontUp = `
func (f *Field{{ $N_BIT }}) InvMontUp(inv, fe *Fe{{ $N_BIT }}) {
u := new(Fe{{ $N_BIT }}).Set(&modulus{{ $N_LIMB }})
v := new(Fe{{ $N_BIT }}).Set(fe)
s := &Fe{{ $N_BIT }}{1, 0, 0, 0}
r := &Fe{{ $N_BIT }}{0, 0, 0, 0}
var k int
var z uint64
var found = false
// Phase 1
for i := 0; i < {{ $N_BIT }} *2; i++ {
if v.IsZero() {
found = true
break }
if u.IsEven() {
u.div2(0)
s.mul2()
} else if v.IsEven() {
v.div2(0)
z += r.mul2()
} else if u.Cmp(v) == 1 {
subn{{ $N_LIMB }}(u, v)
u.div2(0)
addn{{ $N_LIMB }}(r, s)
s.mul2()
} else {
subn{{ $N_LIMB }}(v, u)
v.div2(0)
addn{{ $N_LIMB }}(s, r)
z += r.mul2() }
k += 1 }
if found && k > {{ $N_BIT }} {
if r.Cmp(&modulus{{ $N_LIMB }}) != -1 || z > 0 {
subn{{ $N_LIMB }}(r, &modulus{{ $N_LIMB }}) }
u.Set(&modulus{{ $N_LIMB }})
subn{{ $N_LIMB }}(u, r)
// Phase 2
for i := k; i < {{ $N_BIT }}*2; i++ {
double{{ $N_LIMB }}(u, u) }
inv.Set(u)
} else {
inv.Set(&Fe{{ $N_BIT }}{0, 0, 0, 0}) }}
`

	fTemplInvMontDown = `
func (f *Field{{ $N_BIT }}) InvMontDown(inv, fe *Fe{{ $N_BIT }}) {
u := new(Fe{{ $N_BIT }}).Set(&modulus{{ $N_LIMB }})
v := new(Fe{{ $N_BIT }}).Set(fe)
s := &Fe{{ $N_BIT }}{1, 0, 0, 0}
r := &Fe{{ $N_BIT }}{0, 0, 0, 0}
var k int
var z uint64
var found = false
// Phase 1
for i := 0; i < {{ $N_BIT }}*2; i++ {
if v.IsZero() {
found = true
break }
if u.IsEven() {
u.div2(0)
s.mul2()
} else if v.IsEven() {
v.div2(0)
z += r.mul2()
} else if u.Cmp(v) == 1 {
subn{{ $N_LIMB }}(u, v)
u.div2(0)
addn{{ $N_LIMB }}(r, s)
s.mul2()
} else {
subn{{ $N_LIMB }}(v, u)
v.div2(0)
addn{{ $N_LIMB }}(s, r)
z += r.mul2() }
k += 1 }
if found && k > {{ $N_BIT }} {
if r.Cmp(&modulus{{ $N_LIMB }}) != -1 || z > 0 {
subn{{ $N_LIMB }}(r, &modulus{{ $N_LIMB }}) }
u.Set(&modulus{{ $N_LIMB }})
subn{{ $N_LIMB }}(u, r)
// Phase 2
var e uint64
for i := 0; i < k-{{ $N_BIT }}; i++ {
if u.IsEven() {
u.div2(0)
} else {
e = addn{{ $N_LIMB }}(u, &modulus{{ $N_LIMB }})
u.div2(e) }}
inv.Set(u)
} else {
inv.Set(&Fe{{ $N_BIT }}{0, 0, 0, 0}) }}
`
)
