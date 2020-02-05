#!/bin/bash -e
N_FUZZ=1000
GEN_DIR='./generated'
ARCH='ADX'

### I would like to generate,

###     Option A
######################################
### 384 bit field with given modulus
#
# MODULUS=0x1a0111ea397fe69a4b1ba7b6434bacd764774b84f38512bf6730d2a0f6b0f6241eabfffeb153ffffb9feffffffffaaab
# go run . -output $GEN_DIR -bit 384 -opt A -modulus $MODULUS -arch $ARCH
#


##     Option B
######################################
## 384 bit field with a random modulus

# go run . -output $GEN_DIR -opt B \
# -bit 384 \
# -arch $ARCH


###     Option C
#######################################
### 384 bit field with not-predefined modulus
# go run . -output ./generated -opt C \
# -bit 384 \
# -arch $ARCH

###     Option D
#######################################
### x86 backends for all supported bit sizes and architectures (adx or w/o adx)
go run . -output $GEN_DIR -opt D 

#######################################
# format the code
goreturns -w -p $GEN_DIR
# run the test
go test ./generated -v -fuzz $N_FUZZ
