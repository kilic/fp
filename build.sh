#!/bin/bash -e
if [[ -z "$GEN_FROM" ]]; then
  GEN_FROM=4
fi
if [[ -z "$GEN_TO" ]]; then
  GEN_TO=$GEN_FROM
fi
if [[ -z "$GLOBAL_MODULUS" ]]; then
  GLOBAL_MODULUS=true
fi


BASE=$PWD

generated_files=(
  arithmetic_decl.go
  arithmetic.s
  field_elements.go
  field_test.go
  fields.go
  types.go
)

function clean() {
  rm -rf ./codegen/generated/*
  for file in "${generated_files[@]}"; do
    rm ./$file 2>/dev/null || true
  done
}

function move_to_base() {
  mkdir -p ./backup
  for file in "${generated_files[@]}"; do
    cp ./$file ./backup 2>/dev/null || true
    cp ./codegen/generated/$file . 2>/dev/null || true
  done
}

function generate_declerations() {
  mkdir -p ./codegen/generated
  rm ./codegen/generated/arithmetic_decl.go 2>/dev/null || true
  go run ./codegen/gocode/* -decl -out codegen/generated -from $GEN_FROM -to $GEN_TO -globmod=$GLOBAL_MODULUS
  goreturns -w ./codegen/generated/types.go

  echo -e "*** types are generated\n\t./codegen/generated/types.go\n"
  echo -e "*** declarations are generated\n\t./codegen/generated/arithmetic_decl.go\n"
}

function generate_gocode() {
  mkdir -p ./codegen/generated
  go run ./codegen/gocode/* -field -out ./codegen/generated -from $GEN_FROM -to $GEN_TO -globmod=$GLOBAL_MODULUS

  echo -e "*** field impls are generated\n\t\
./codegen/generated/field_elements.go\n\t\
./codegen/generated/fields.go\n\t\
./codegen/generated/field_test.go\n"

  goreturns -w -p ./codegen/generated
}

function generate_x86() {
  go run ./codegen/x86/* -out ./codegen/generated/arithmetic.s -from $GEN_FROM -to $GEN_TO -globmod=$GLOBAL_MODULUS
  echo -e "*** arithmetics are generated\n\t./codegen/generated/arithmetic.s\n"
  go vet ./codegen/generated
}

function generate_all() {
  rm -rf ./codegen/generated/*
  generate_declerations
  generate_x86
  generate_gocode
}

for i in "$@"; do
  case $i in
  '--gen')
    generate_all
    ;;
  '--test')
    go test ./codegen/generated -run '' -iter 10
    ;;
  '--bench')
    go test -benchmem -run=^$ -bench ./codegen/generated
    ;;
  '--mv')
    move_to_base
    ;;
  esac
done
