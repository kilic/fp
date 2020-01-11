#!/bin/bash -e
N_FUZZ=50
GEN_DIR='./generated'

field_sizes=(256 320 384 448 512)

# ADX backend
ARCH='ADX'
for BIT_SIZE in "${field_sizes[@]}"
do
  # option B, fixed modulus
  go run . -output $GEN_DIR -opt B -bit $BIT_SIZE -arch $ARCH
  goreturns -w -p $GEN_DIR
  go test ./generated -v -fuzz $N_FUZZ
  # option C, non fixed modulus
  go run . -output $GEN_DIR -opt C -bit $BIT_SIZE -arch $ARCH
  goreturns -w -p $GEN_DIR
  go test ./generated -v -fuzz $N_FUZZ
done

# non ADX backend
for BIT_SIZE in "${field_sizes[@]}"
do
  # option B, fixed modulus
  go run . -output $GEN_DIR -opt B -bit $BIT_SIZE
  goreturns -w -p $GEN_DIR
  go test ./generated -v -fuzz $N_FUZZ
  # option C, non fixed modulus
  go run . -output $GEN_DIR -opt C -bit $BIT_SIZE
  goreturns -w -p $GEN_DIR
  go test ./generated -v -fuzz $N_FUZZ
done
