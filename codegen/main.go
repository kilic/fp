package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kilic/fp/codegen/gocode"
	"github.com/kilic/fp/codegen/x86"
)

func main() {

	// _ = out
	options := `Options:
A : Generate a fixed modulus field (no modulus input for backend)
B : Generate a fixed modulus random field (no modulus input for backend)
C : Generate a arbitrary modulus field for a fixed bit size
D : Generate all implemented backends only
`

	var output string
	var bitSize int
	var modulus string
	var opt string
	var arch string

	flag.StringVar(&output, "output", "tmp", "output directory")
	flag.IntVar(&bitSize, "bit", 0, "bit size of the field")
	flag.StringVar(&modulus, "modulus", "", "bit size of the field")
	flag.StringVar(&opt, "opt", "", options)
	flag.StringVar(&arch, "arch", "", "")
	flag.Parse()

	output = filepath.Clean(output)
	s, err := os.Stat(output)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(output, os.ModePerm); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		if !s.IsDir() {
			panic("output should be directory")
		}
	}

	var fixedmod bool
	switch opt {
	case "A":
		err := gocode.GenField(output, bitSize, modulus, opt)
		if err != nil {
			panic(err)
		}
		fixedmod := true
		single := true
		err = x86.GenX86(output, bitSize, arch, fixedmod, single)
		if err != nil {
			panic(err)
		}
	case "B":
		err := gocode.GenField(output, bitSize, modulus, opt)
		if err != nil {
			panic(err)
		}
		fixedmod := true
		single := true
		err = x86.GenX86(output, bitSize, arch, fixedmod, single)
		if err != nil {
			panic(err)
		}
	case "C":
		err := gocode.GenField(output, bitSize, modulus, opt)
		if err != nil {
			panic(err)
		}
		fixedmod = false
		single := true
		err = x86.GenX86(output, bitSize, arch, fixedmod, single)
		if err != nil {
			panic(err)
		}
	case "D":
		var supportedLimbSizes = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
		gocode.GenDeclerationsForMultiple(output, supportedLimbSizes)
		err := x86.GenX86All(output)
		if err != nil {
			panic(err)
		}
	default:
		panic(fmt.Sprintf("no such option %s\n" + opt))
	}
}
