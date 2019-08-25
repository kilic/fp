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
	fTmplEqual,
	fTmplIsZero,
	fTmplToBytes,
	fTmplMont,
	fTmplDemont,
	fTmplAdd,
	fTmplDouble,
	fTmplSub,
	fTmplNeg,
	fTmplSquare,
	fTmplMul,
	fTmplExp,
	fTmplInverse,
	fTmplInvMontUp,
	fTemplInvMontDown,
	fTmplInvEEA,
}

const (
	fTmplConstants = `
{{- if $GlobMod }}
var inp{{ $N_LIMB }} uint64
var modulus{{ $N_LIMB }} {{ $FE }} 
{{- end }}
`

	fTmplFieldDef = `
type {{ $FIELD }} struct {
// r1  = r mod p
// r2  = r^2 mod p
// inp = -p^(-1) mod 2^64
pBig *big.Int
r1  *{{ $FE }} 
r2  *{{ $FE }} 
P   *{{ $FE }} 
inp uint64}
`

	fTmplNew = `
func New{{ $FIELD }}(p []byte) *{{ $FIELD }} {
if len(p) > {{ $N_BIT }} {
return nil }
pBig := new(big.Int).SetBytes(p)
inpT := new(big.Int).ModInverse(new(big.Int).Neg(pBig), new(big.Int).SetBit(new(big.Int), 64, 1))
if inpT == nil {
return nil}
inp := inpT.Uint64() 
r1, r2, modulus := &{{ $FE }}{}, &{{ $FE }}{}, &{{ $FE }}{}
modulus.FromBytes(p)
{{- if $GlobMod }} 
modulus{{ $N_LIMB }} = *modulus
inp{{ $N_LIMB }} = inp 
{{- end }}
r1Big := new(big.Int).SetBit(new(big.Int), {{ $N_BIT }}, 1)
r1.SetBig(new(big.Int).Mod(r1Big, pBig))
r2.SetBig(new(big.Int).Exp(r1Big, new(big.Int).SetUint64(2), pBig))
return &{{ $FIELD }}{
pBig: pBig,
r1:   r1,
r2:   r2,
P:    modulus,
inp:  inp}}
	`

	fTmplNewFeBytes = `
func (f *{{ $FIELD }}) NewElementFromBytes(in []byte) *{{ $FE }} {
fe := new({{ $FE }}).FromBytes(in)
f.Mul(fe, fe, f.r2)
return fe }
`

	fTmplNewFeUint = `
func (f *{{ $FIELD }}) NewElementFromUint(in uint64) *{{ $FE }} {
fe := &{{ $FE }}{in}
if in == 0 {
return fe }
f.Mul(fe, fe, f.r2)
return fe }
`

	fTmplNewFeBig = `
func (f *{{ $FIELD }}) NewElementFromBig(in *big.Int) *{{ $FE }} {
fe := new({{ $FE }}).SetBig(in)
f.Mul(fe, fe, f.r2)
return fe }
`

	fTmplNewFeString = `
func (f *{{ $FIELD }}) NewElementFromString(in string) (*{{ $FE }}, error) {
fe, err := new({{ $FE }}).SetString(in)
if err != nil {
return nil, err }
f.Mul(fe, fe, f.r2)
return fe, nil }
`

	fTmplZero = `
func (f *{{ $FIELD }}) Zero() *{{ $FE }} {
return new({{ $FE }}).SetUint(0) }
`

	fTmplOne = `
func (f *{{ $FIELD }}) One() *{{ $FE }} {
return new({{ $FE }}).Set(f.r1) }
`

	fTmplCopy = `
func (f *{{ $FIELD }}) Copy(dst *{{ $FE }}, src *{{ $FE }}) *{{ $FE }} {
return dst.Set(src) }
`

	fTmplRand = `
func (f *{{ $FIELD }}) RandElement(fe *{{ $FE }}, r io.Reader) (*{{ $FE }}, error) {
bi, err := rand.Int(r, f.pBig)
if err != nil {
return nil, err }
return fe.SetBig(bi), nil}		
`

	fTmplEqual = `
func (f *{{ $FIELD }}) Equal(a, b *{{ $FE }}) bool {
return a.Equals(b)
}`

	fTmplIsZero = `
func (f *{{ $FIELD }}) IsZero(a *{{ $FE }}) bool {
return a.IsZero()
}
`

	fTmplToBytes = `
func (f *{{ $FIELD }}) ToBytes(bytes []byte, fe *{{ $FE }}) ([]byte, error) {
if len(bytes) < {{ $N_BYTES }} {
return bytes, fmt.Errorf("output slice should be equal or larger than {{ $N_BYTES }} byte")}
fe2 := new({{ $FE }})
f.Demont(fe2, fe)
copy(bytes[:{{ $N_BYTES }}], fe2.Bytes())
return bytes, nil}
`

	fTmplMont = `
func (f *{{ $FIELD }}) Mont(c, a *{{ $FE }}) {
{{- if $GlobMod }} 
montmul{{ $N_LIMB }}(c, a, f.r2) 
{{- else }}
montmul{{ $N_LIMB }}(c, a, f.r2, f.P, f.inp) 
{{- end }}
}
`

	fTmplDemont = `
func (f *{{ $FIELD }}) Demont(c, a *{{ $FE }}) {
{{- if $GlobMod }} 
montmul{{ $N_LIMB }}(c, a, &{{ $FE }}{1}) 
{{- else }}
montmul{{ $N_LIMB }}(c, a, &{{ $FE }}{1}, f.P, f.inp) 
{{- end }}
}
`

	fTmplAdd = `
func (f *{{ $FIELD }}) Add(c, a, b *{{ $FE }}) {
{{- if $GlobMod }} 
add{{ $N_LIMB }}(c, a, b)
{{- else }}
add{{ $N_LIMB }}(c, a, b, f.P)
{{- end }}  }
`

	fTmplDouble = `
func (f *{{ $FIELD }}) Double(c, a *{{ $FE }}) {
{{- if $GlobMod }} 
double{{ $N_LIMB }}(c, a) 
{{- else }}
double{{ $N_LIMB }}(c, a, f.P) 
{{- end }} }
`

	fTmplSub = `
func (f *{{ $FIELD }}) Sub(c, a, b *{{ $FE }}) {
{{- if $GlobMod }} 
sub{{ $N_LIMB }}(c, a, b) 
{{- else }}
sub{{ $N_LIMB }}(c, a, b, f.P) 
{{- end }} }
`

	fTmplNeg = `
func (f *{{ $FIELD }}) Neg(c, a *{{ $FE }}) {
{{- if $GlobMod }}
neg{{ $N_LIMB }}(c, a) 
{{- else }}
neg{{ $N_LIMB }}(c, a, f.P) 
{{- end }}
}
`

	fTmplSquare = `
func (f *{{ $FIELD }}) Square(c, a *{{ $FE }}) {
{{- if $GlobMod }} 
montsquare{{ $N_LIMB }}(c, a) 
{{- else }}
montsquare{{ $N_LIMB }}(c, a, f.P, f.inp) 
{{- end }} 
}
`

	fTmplMul = `
func (f *{{ $FIELD }}) Mul(c, a, b *{{ $FE }}) {
{{- if $GlobMod }} 
montmul{{ $N_LIMB }}(c, a, b) 
{{- else }}
montmul{{ $N_LIMB }}(c, a, b, f.P, f.inp) 
{{- end }}
}
`

	fTmplExp = `
func (f *{{ $FIELD }}) Exp(c, a*{{ $FE }}, e *big.Int) {
z := new({{ $FE }}).Set(f.r1)
for i := e.BitLen(); i >= 0; i-- {
{{- if $GlobMod }} 
montmul{{ $N_LIMB }}(z, z, z)
{{- else }}
montmul{{ $N_LIMB }}(z, z, z, f.P, f.inp)
{{- end }}
if e.Bit(i) == 1 {
{{- if $GlobMod }}
montmul{{ $N_LIMB }}(z, z, a)
{{- else }}
montmul{{ $N_LIMB }}(z, z, a, f.P, f.inp)
{{- end }}
} }
c.Set(z) }
`

	fTmplInvEEA = `
func (f *{{ $FIELD }}) InvEEA(inv, fe *{{ $FE }}) {
u := new({{ $FE }}).Set(fe)
v := new({{ $FE }}).Set(f.P)
x1 := &{{ $FE }}{1}
x2 := &{{ $FE }}{0}
var e uint64
for !u.IsOne() && !v.IsOne() {
for u.IsEven() {
u.div2(0)
if x1.IsEven() {
x1.div2(0)
} else {
e = addn{{ $N_LIMB }}(x1, f.P)
x1.div2(e) }}
for v.IsEven() {
v.div2(0)
if x2.IsEven() {
x2.div2(0)
} else {
e = addn{{ $N_LIMB }}(x2, f.P)
x2.div2(e) }}
if u.Cmp(v) == -1 {
subn{{ $N_LIMB }}(v, u)
{{- if $GlobMod }} 
sub{{ $N_LIMB }}(x2, x2, x1)
{{- else }}
sub{{ $N_LIMB }}(x2, x2, x1, f.P) 
{{- end }}
} else {
subn{{ $N_LIMB }}(u, v) 
{{- if $GlobMod }}
sub{{ $N_LIMB }}(x1, x1, x2)  
{{- else }}
sub{{ $N_LIMB }}(x1, x1, x2, f.P) 
{{- end }} 
}}
if u.IsOne() {
inv.Set(x1)
return }
inv.Set(x2)}
`

	fTmplInverse = `
func (f *{{ $FIELD }}) Inverse(inv, fe *{{ $FE }}) {
f.InvMontDown(inv, fe)
}
`

	fTmplInvMontUp = `
func (f *{{ $FIELD }}) InvMontUp(inv, fe *{{ $FE }}) {
u := new({{ $FE }}).Set(f.P)
v := new({{ $FE }}).Set(fe)
s := &{{ $FE }}{1}
r := &{{ $FE }}{0}
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
if r.Cmp(f.P) != -1 || z > 0 {
subn{{ $N_LIMB }}(r, f.P) }
u.Set(f.P)
subn{{ $N_LIMB }}(u, r)
// Phase 2
for i := k; i < {{ $N_BIT }}*2; i++ {
{{- if $GlobMod }}
double{{ $N_LIMB }}(u, u) 
{{- else }}
double{{ $N_LIMB }}(u, u, f.P) 
{{- end }} }
inv.Set(u)
} else {
inv.Set(&{{ $FE }}{0}) }}
`

	fTemplInvMontDown = `
func (f *{{ $FIELD }}) InvMontDown(inv, fe *{{ $FE }}) {
u := new({{ $FE }}).Set(f.P)
v := new({{ $FE }}).Set(fe)
s := &{{ $FE }}{1}
r := &{{ $FE }}{0}
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
if r.Cmp(f.P) != -1 || z > 0 {
subn{{ $N_LIMB }}(r, f.P) }
u.Set(f.P)
subn{{ $N_LIMB }}(u, r)
// Phase 2
var e uint64
for i := 0; i < k-{{ $N_BIT }}; i++ {
if u.IsEven() {
u.div2(0)
} else {
e = addn{{ $N_LIMB }}(u, f.P)
u.div2(e) }}
inv.Set(u)
} else {
inv.Set(&{{ $FE }}{0}) }}
`
)
