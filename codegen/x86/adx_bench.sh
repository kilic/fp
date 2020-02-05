#!/bin/bash -e

limb_sizes=(2 3 4 5 6 7 8 9 10 11 12 13 14 15 16)

for LIMB_SIZE in "${limb_sizes[@]}"
do
  echo $LIMB_SIZE
  go run ./test/ -limb $LIMB_SIZE
  go test ./debug/ -run ^$ -bench=. -v
done