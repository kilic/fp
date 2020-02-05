`fp` generates prime fields, field elements and x86 optimized, high speed field operations.

## Generating Field

Example commands to generate fields can be found [here](codegen/example.sh). 

There are four generation modes.

### A. Predefined Modulus

Given modulus input you get a field implementation with its precomputed constants.

```sh
# maybe you would like to generate a field for a BLS12-381 elliptic curve implementation
MODULUS=0x1a0111ea397fe69a4b1ba7b6434bacd764774b84f38512bf6730d2a0f6b0f6241eabfffeb153ffffb9feffffffffaaab
go run . -output $GEN_DIR -bit 384 -opt A -modulus $MODULUS
```

### B. Random Field

Option B helps to generate a random field with random prime modulus at desired bit length.

### C. Arbitrary modulus

With this option you get a field implementation where you feed the modulus while construction of a field in runtime.

### D. Generic

In generic case, field elements are `unsafe pointers`. This helps us to decide size of field element and its arithmetic functions in runtime. It also helps us to represent field element with single type independent from their size. [Generic field implementation](generic/field.go) is already generated.

```go
type fieldElement = unsafe.Pointer

// Given limb size returns field element filled with zero 
func newFieldElement(limbSize int) (fieldElement, error) {
  switch limbSize {
  case 1:
    return unsafe.Pointer(&[1]uint64{}), nil
  case 2:
   return unsafe.Pointer(&[2]uint64{}), nil
 case 3:
   return unsafe.Pointer(&[3]uint64{}), nil
  ...
  case 12:
    return unsafe.Pointer(&[12]uint64{}), nil
  ...
  ...
}
```

Simply provide a modulus in bytes to initialize a field.

```go
pStr := "0x73eda753299d7d483339d80809a1d80553bda402fffe5bfeffffffff00000001"
pBytes, _ := hex.DecodeString(pStr[2:])
field := newField(pBytes)
```

## Benchmark

Benchmarked on 2,7 GHz i5 machine

Table below shows multiplication delays corresponding field sizes in bits.

```
128:  10.5 ns/op
192:  15.8 ns/op
256:  23.9 ns/op
320:  33.3 ns/op
384:  43.4 ns/op
448:  59.2 ns/op
512:  67.2 ns/op
576:  81.9 ns/op
640:  103 ns/op
704:  130 ns/op
768:  153 ns/op
832:  177 ns/op
896:  202 ns/op
960:  228 ns/op
1024: 256 ns/op
```