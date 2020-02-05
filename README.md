
# Usage

## Code Genetation

Examples can be found here. There are four generation modes.

## A. Predefined Modulus

Given modulus input you get a field implementation with its precomputed constants.

```sh
# maybe you would like to generate a field for a BLS12-381 elliptic curve implementation
MODULUS=0x1a0111ea397fe69a4b1ba7b6434bacd764774b84f38512bf6730d2a0f6b0f6241eabfffeb153ffffb9feffffffffaaab
go run . -output $GEN_DIR -bit 384 -opt A -modulus $MODULUS
```

## B. Random Field

For testing or playing purposes you can get a field implementation with random prime modulus and its precomputed constant values at desired bit lenght.

## C. Arbitrary modulus

With this option you get a field implementation where you feed the modulus while construction of a field in runtime.

## D. Generic

In generic case, field elements are `unsafe pointers`. This helps us to decide size of field element and its arithmetic functions in runtime. It also help us to represent field element with single type independent from their size.

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
