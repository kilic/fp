#!/bin/bash -e
N_FUZZ=100


limb_sizes=(2 3 4 5 6 7 8 9 10 11 12 13 14 15 16)

for LIMB_SIZE in "${limb_sizes[@]}"
do
  go run ./test/ -limb $LIMB_SIZE -noadx -fixed
  go test ./debug/ -run Hard -fuzz $N_FUZZ -v
done

for LIMB_SIZE in "${limb_sizes[@]}"
do
  go run ./test/ -limb $LIMB_SIZE -noadx
  go test ./debug/ -run Hard -fuzz $N_FUZZ -v
done


