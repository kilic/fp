package main

var fieldElementTestTemplates = []string{

	testEncDec,
	testAddition,
	testSubtraction,
	testDoubling,
	testMontgomerry,
	testExponentiation,
	testInversion}

const (
	testMain = `
var n int

func TestMain(m *testing.M) {
iter := flag.Int("iter", 1000, "# of iters")
flag.Parse()
n = *iter
m.Run()}
`

	testEncDec = `
t.Run("Encoding & Decoding", func(t *testing.T) {
var field *{{ $FIELD }}
for i := 0; i < n; i++ {
for true {
p, err := rand.Prime(rand.Reader, {{ $N_BIT }})
if err != nil {
t.Fatal(err) }
field = New{{ $FIELD }}(p.Bytes())
if field != nil {
break } } }
t.Run("1", func(t *testing.T) {
bytes := []byte{
0, }
if !new({{ $FE }}).Unmarshal(bytes).Equals(&{{ $FE }}{0}) {
t.Errorf("bad encoding\n") } })
t.Run("2", func(t *testing.T) {
bytes := []byte{
254, 253, }
if new({{ $FE }}).Unmarshal(bytes).Equals(&{{ $FE }}{0xfe, 0xfd}) {
t.Errorf("bad encoding\n") } })
t.Run("3", func(t *testing.T) {
var a, b {{ $FE }}
er1 := field.RandElement(&a, rand.Reader)
if er1 != nil {
t.Fatal(er1) }
bytes := make([]byte, {{ $N_LIMB }} *8)
a.Marshal(bytes[:])
b.Unmarshal(bytes[:])
if !a.Equals(&b) {
t.Errorf("bad encoding or decoding\n") } })
t.Run("4", func(t *testing.T) {
var a {{ $FE }}
er1 := field.RandElement(&a, rand.Reader)
if er1 != nil {
t.Fatal(er1) }
b, er1 := new({{ $FE }}).SetString(a.String())
if er1 != nil {
t.Fatal(er1) }
if !a.Equals(b) {
t.Errorf("bad encoding or decoding\n") } })
t.Run("5", func(t *testing.T) {
var a {{ $FE }}
er1 := field.RandElement(&a, rand.Reader)
if er1 != nil {
t.Fatal(er1) }
b := new({{ $FE }}).SetBig(a.Big())
if er1 != nil {
t.Fatal(er1) }
if !a.Equals(b) {
t.Errorf("bad encoding or decoding\n") } }) })
`

	testAddition = `
t.Run("Addition", func(t *testing.T) {
var a, b, c, u, v {{ $FE }}
zero := new({{ $FE }}).SetUint(0)
for i := 0; i < n; i++ {
var field *{{ $FIELD }}
for true {
p, err := rand.Prime(rand.Reader, {{ $N_BIT }})
if err != nil {
t.Fatal(err)}
field = New{{ $FIELD }}(p.Bytes())
if field != nil { 
break }}
er1 := field.RandElement(&a, rand.Reader)
er2 := field.RandElement(&b, rand.Reader)
er3 := field.RandElement(&c, rand.Reader)
if er1 != nil || er2 != nil || er3 != nil {
t.Fatal(er1, er2, er3)}
field.Add(&u, &a, &b)
field.Add(&u, &u, &c)
field.Add(&v, &b, &c)
field.Add(&v, &v, &a)
if !u.Equals(&v) {
t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n",a, b, c, u, v)}
field.Add(&u, &a, &b)
field.Add(&v, &b, &a)
if !u.Equals(&v) {
t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv:%s\n", a, b, u, v)}
field.Add(&u, &a, zero)
if !u.Equals(&a) {
t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n",a, u)}
field.Neg(&u, &a)
field.Add(&u, &u, &a)
if !u.Equals(zero) {
t.Fatalf("Bad Negation\na:%s", a.String())} }})
`

	testSubtraction = `
t.Run("Subtraction", func(t *testing.T) {
var a, b, c, u, v {{ $FE }}
zero := new({{ $FE }}).SetUint(0)
for i := 0; i < n; i++ {
var field *{{ $FIELD }}
for true {
p, err := rand.Prime(rand.Reader, {{ $N_BIT }})
if err != nil {
t.Fatal(err)}
field = New{{ $FIELD }}(p.Bytes())
if field != nil { 
break }}
er1 := field.RandElement(&a, rand.Reader)
er2 := field.RandElement(&b, rand.Reader)
er3 := field.RandElement(&c, rand.Reader)
if er1 != nil || er2 != nil || er3 != nil {
t.Fatal(er1, er2, er3)}
field.Sub(&u, &a, &c)
field.Sub(&u, &u, &b)
field.Sub(&v, &a, &b)
field.Sub(&v, &v, &c)
if !u.Equals(&v) {
t.Fatalf("Additive associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv:%s\n", a, b, c, u, v)}
field.Sub(&u, &a, zero)
if !u.Equals(&a) {
t.Fatalf("Additive identity does not hold\na: %s\nu: %s\n", a, u)}
field.Sub(&u, &a, &b)
field.Sub(&v, &b, &a)
field.Add(&u, &u, &v)
if !u.Equals(zero) {
t.Fatalf("Additive commutativity does not hold\na: %s\nb: %s\nu: %s\nv: %s", a, b, u, v)}
field.Sub(&u, &a, &b)
field.Sub(&v, &b, &a)
field.Neg(&v, &v)
if !u.Equals(&u) {
t.Fatalf("Bad Negation\na:%s", a.String())} }})
`

	testDoubling = `
t.Run("Doubling", func(t *testing.T) {
var a, u, v {{ $FE }}
for i := 0; i < n; i++ {
var field *{{ $FIELD }}
for true {
p, err := rand.Prime(rand.Reader, {{ $N_BIT }})
if err != nil {
t.Fatal(err)}
field = New{{ $FIELD }}(p.Bytes())
if field != nil {
break}}
err := field.RandElement(&a, rand.Reader)
if err != nil {
t.Fatal(err)}
field.Double(&u, &a)
field.Add(&v, &a, &a)
if !u.Equals(&v) {
t.Fatalf("Bad doubling\na: %s\nu: %s\nv: %s\n", a, u, v)} }})
`

	testMontgomerry = `
t.Run("Montgomerry", func(t *testing.T) {
var a, b, c, u, v, w {{ $FE }}
zero := new({{ $FE }}).SetUint(0)
one := new({{ $FE }}).SetUint(1)
for i := 0; i < n; i++ {
var field *{{ $FIELD }}
for true {
p, err := rand.Prime(rand.Reader, {{ $N_BIT }})
if err != nil {
t.Fatal(err)}
field = New{{ $FIELD }}(p.Bytes())
if field != nil {
break }}
er1 := field.RandElement(&a, rand.Reader)
er2 := field.RandElement(&b, rand.Reader)
er3 := field.RandElement(&c, rand.Reader)
if er1 != nil || er2 != nil || er3 != nil {
t.Fatal(er1, er2, er3) }
field.Mont(&u, zero)
if !u.Equals(zero) {
t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P) }
field.Demont(&u, zero)
if !u.Equals(zero) {
t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P) }
field.Mont(&u, one)
if !u.Equals(field.r1) {
t.Fatalf("Bad Montgomerry encoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P) }
field.Demont(&u, field.r1)
if !u.Equals(one) {
t.Fatalf("Bad Montgomerry decoding\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P) }
field.Mul(&u, &a, zero)
if !u.Equals(zero) {
t.Fatalf("Bad zero element\na: %s\nu: %s\np: %s\n", a, u, field.P) }
field.Mul(&u, &a, one)
field.Mul(&u, &u, field.r2)
if !u.Equals(&a) {
t.Fatalf("Multiplication identity does not hold, expected to equal itself\nu: %s\np: %s\n", u, field.P) }
field.Mul(&u, field.r2, one)
if !u.Equals(field.r1) {
t.Fatalf("Multiplication identity does not hold, expected to equal r1\nu: %s\np: %s\n", u, field.P) }
field.Mul(&u, &a, &b)
field.Mul(&u, &u, &c)
field.Mul(&v, &b, &c)
field.Mul(&v, &v, &a)
if !u.Equals(&v) {
t.Fatalf("Multiplicative associativity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.P) }
field.Add(&u, &a, &b)
field.Mul(&u, &c, &u)
field.Mul(&w, &a, &c)
field.Mul(&v, &b, &c)
field.Add(&v, &v, &w)
if !u.Equals(&v) {
t.Fatalf("Distributivity does not hold\na: %s\nb: %s\nc: %s\nu: %s\nv: %s\np: %s\n", a, b, c, u, v, field.P) } }})
`

	testExponentiation = `
t.Run("Exponentiation", func(t *testing.T) {
var a, u, v {{ $FE }}
bytes := make([]byte, {{ $N_LIMB }} *8)
for i := 0; i < n; i++ {
var field *{{ $FIELD }}
for true {
p, err := rand.Prime(rand.Reader, {{ $N_BIT }})
if err != nil {
t.Fatal(err) }
field = New{{ $FIELD }}(p.Bytes())
if field != nil {
break }}
er1 := field.RandElement(&a, rand.Reader)
if er1 != nil {
t.Fatal(er1)}
field.Exp(&u, &a, big.NewInt(0))
if !u.Equals(field.r1) {
t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)}
field.Exp(&u, &a, big.NewInt(1))
if !u.Equals(&a) {
t.Fatalf("Bad exponentiation, expected to equal a\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)}
field.Mul(&u, &a, &a)
field.Mul(&u, &u, &u)
field.Mul(&u, &u, &u)
field.Exp(&v, &a, big.NewInt(8))
if !u.Equals(&v) {
t.Fatalf("Bad exponentiation\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P)}
p := new(big.Int).SetBytes(field.P.Marshal(bytes))
field.Exp(&u, &a, p)
if !u.Equals(&a) {
t.Fatalf("Bad exponentiation, expected to equal itself\nu: %s\na: %s\np: %s\n", u, a, field.P)}
field.Exp(&u, &a, p.Sub(p, big.NewInt(1)))
if !u.Equals(field.r1) {
t.Fatalf("Bad exponentiation, expected to equal r1\nu: %s\na: %s\nr1: %s\np: %s\n", u, a, field.r1, field.P) }
}})
`

	testInversion = `
t.Run("Inversion", func(t *testing.T) {
var u, a, v {{ $FE }}
one := new({{ $FE }}).SetUint(1)
bytes := make([]byte, {{ $N_LIMB }} *8)
for i := 0; i < n; i++ {
var field *{{ $FIELD }}
for true {
p, err := rand.Prime(rand.Reader, {{ $N_BIT }})
if err != nil {
t.Fatal(err) }
field = New{{ $FIELD }}(p.Bytes())
if field != nil {
break }}
er1 := field.RandElement(&a, rand.Reader)
if er1 != nil {
t.Fatal(er1) }
field.InvMontUp(&u, &a)
field.Mul(&u, &u, &a)
if !u.Equals(field.r1) {
t.Fatalf("Bad inversion, expected to equal r1\nu: %s\nr1: %s\np: %s\n", u, field.r1, field.P) }
field.Mont(&u, &a)
field.InvMontDown(&v, &u)
field.Mul(&v, &v, &u)
if !v.Equals(one) {
t.Fatalf("Bad inversion, expected to equal 1\nu: %s\nv: %s\na: %s\np: %s\n", u, v, a, field.P) }
p := new(big.Int).SetBytes(field.P.Marshal(bytes))
field.Exp(&u, &a, p.Sub(p, big.NewInt(2)))
field.InvMontUp(&v, &a)
if !v.Equals(&u) {
t.Fatalf("Bad inversion") }
field.InvEEA(&u, &a)
field.Mul(&u, &u, &a)
field.Mul(&u, &u, field.r2)
if !u.Equals(one) {
t.Fatalf("Bad inversion") }} })
`
	benches = `
func Benchmark{{ $FIELD }}(t *testing.B) {
var a, b, c {{ $FE }}
var field *{{ $FIELD }}
for true {
p, err := rand.Prime(rand.Reader, {{ $N_BIT }})
if err != nil {
t.Fatal(err) }
field = New{{ $FIELD }}(p.Bytes())
if field != nil {
break }}
er1 := field.RandElement(&a, rand.Reader)
er2 := field.RandElement(&b, rand.Reader)
if er1 != nil || er2 != nil {
t.Fatal(er1, er2)}
t.Run("Addition", func(t *testing.B) {
t.ResetTimer()
for i := 0; i < t.N; i++ {
field.Add(&c, &a, &b) }})
t.Run("Subtraction", func(t *testing.B) {
t.ResetTimer()
for i := 0; i < t.N; i++ {
field.Sub(&c, &a, &b) }})
t.Run("Doubling", func(t *testing.B) {
t.ResetTimer()
for i := 0; i < t.N; i++ {
field.Double(&c, &a) }})
t.Run("Multiplication", func(t *testing.B) {
t.ResetTimer()
for i := 0; i < t.N; i++ {
field.Mul(&c, &a, &b) }})
// t.Run("Squaring", func(t *testing.B) {
// t.ResetTimer()
// for i := 0; i < t.N; i++ {
// field.Square(&c, &a) }})
t.Run("Inversion", func(t *testing.B) {
t.ResetTimer()
for i := 0; i < t.N; i++ {
field.InvMontUp(&c, &a) }})
t.Run("Exponentiation", func(t *testing.B) {
bytes := make([]byte, {{ $N_LIMB }} *8)
e := new(big.Int).SetBytes(field.P.Marshal(bytes))
t.ResetTimer()
for i := 0; i < t.N; i++ {
field.Exp(&c, &a, e) }})}
`
)
