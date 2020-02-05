#!/bin/bash -e
N_FUZZ=5
GEN_DIR='./generated'

field_sizes=(\
128 192 256 320 384 448 512 \
576 640 704 768 832 896 960 \
1024
)

# ADX backend
ARCH='ADX'
for BIT_SIZE in "${field_sizes[@]}"
do
  # option B, fixed modulus
  echo 'B' $BIT_SIZE $ARCH
  go run . -output $GEN_DIR -opt B -bit $BIT_SIZE -arch $ARCH
  goreturns -w -p $GEN_DIR
  go test ./generated -fuzz $N_FUZZ
  # option C, non fixed modulus
  echo 'C' $BIT_SIZE $ARCH
  go run . -output $GEN_DIR -opt C -bit $BIT_SIZE -arch $ARCH
  goreturns -w -p $GEN_DIR
  go test ./generated -fuzz $N_FUZZ
done

# non ADX backend
for BIT_SIZE in "${field_sizes[@]}"
do
  # option B, fixed modulus
  echo 'B' $BIT_SIZE $ARCH fixed
  go run . -output $GEN_DIR -opt B -bit $BIT_SIZE
  goreturns -w -p $GEN_DIR
  go test ./generated -fuzz $N_FUZZ
  # option C, non fixed modulus
  echo 'C' $BIT_SIZE $ARCH fixed
  go run . -output $GEN_DIR -opt C -bit $BIT_SIZE
  goreturns -w -p $GEN_DIR
  go test ./generated -fuzz $N_FUZZ
done
