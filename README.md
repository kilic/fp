generated go code is better to be considered as boilerplate
you could tweak the templates and generate for your needs
we called 
while non static

Given prime number bit size, `fp` generates prime fields, field elements and x86 optimized, high speed field operations. Assembly code that performs field operations are generated with [avo](https://github.com/mmcloughlin/avo) library.

## Install, Build & Test

Fields from 256 to 512 bit length is available under generated tag. Install by using go modules:

```
go get github.com/kilic/fp@generated
```

If *go modules* is not preferred, fields should be generated locally. To install:

```
$ go get github.com/kilic/fp
```

Generate field implementations from, say, 4 to 8 into temporary directory *./codegen/generated*:

```
$ cd $GOPATH/github.com/kilic/fp
$ GEN_FROM=4 GEN_TO=8 ./build.sh --gen
```

Run tests and move to base directory:

```
$ ./build.sh --test
$ ./build.sh --mv
```

## Usage

Here is an example usage with 256 bit prime field:

```go

package main

import 	"github.com/kilic/fp"

// ...

  pStr := "0x73eda753299d7d483339d80809a1d80553bda402fffe5bfeffffffff00000001"
  pBytes, _ := hex.DecodeString(pStr[2:])
  field := fp.NewField256(pBytes)
  a := field.NewElementFromUint(2)
  b := field.NewElementFromUint(3)
  c := new(fp.Fe256)
  field.Mul(c, a, b)
  field.Demont(c, c)
  if !c.Equals(new(fp.Fe256).SetUint(6)) {
    log.Fatal(c)
  }
  
// ...

```


## Benchmarks

Benchmarked on 2,7 GHz i5.

```
256/Addition          5.79 ns/op
256/Multiplication    34.1 ns/op
320/Addition          6.92 ns/op
320/Multiplication    57.5 ns/op
384/Addition          8.95 ns/op
384/Multiplication    92.3 ns/op
448/Addition          10.8 ns/op
448/Multiplication    134 ns/op
512/Addition          12.5 ns/op
512/Multiplication    189 ns/op
```