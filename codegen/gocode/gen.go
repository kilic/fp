package gocode

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
)

var supportedBitSizes = map[int]bool{
	128: true,
	192: true,
	256: true,
	320: true,
	384: true,
	448: true,
	512: true,
}

var supportedLimbSizes = []int{2, 3, 4, 5, 6, 7, 8}

func resolveBitSize(byteSize int) int {
	size := (byteSize / 8)
	if byteSize%8 != 0 {
		size += 1
	}
	return size * 64
}

func GenDeclerationsForMultiple(out string, limbSizes []int) {
	outDir := filepath.Clean(out)
	_limbSizes := limbSizes
	if _limbSizes == nil {
		_limbSizes = supportedLimbSizes
	}
	arithmeticDeclerationsCode := pkg("fp") + arithmeticDeclerationsMultiple(_limbSizes)
	writeToFile(arithmeticDeclerationsCode, filepath.Join(outDir, "arithmetic_decl.go"))
}

func GenField(out string, bitSize int, modulus string, opt string) error {

	var limbSize int
	var fixedModulus bool
	var modulusBig *big.Int
	outDir := filepath.Clean(out)
	switch opt {
	case "A":
		if modulus == "" {
			return fmt.Errorf("Modulus should be set for option A\n")
		}
		if len(modulus) < 2 || modulus[:2] != "0x" {
			return fmt.Errorf("Bad format for modulus\n")
		}
		bts, err := hex.DecodeString(modulus[2:])
		if err != nil {
			return err
		}
		bitSize := resolveBitSize(len(bts))
		if !supportedBitSizes[bitSize] {
			return fmt.Errorf("Bit size %d is not supported\n", bitSize)

		}
		modulusBig = new(big.Int).SetBytes(bts)
		limbSize = bitSize / 64
		fixedModulus = true
	case "B":
		if !supportedBitSizes[bitSize] {
			return fmt.Errorf("Bit size %d is not supported", bitSize)
		}
		limbSize = bitSize / 64
		var err error
		modulusBig, err = rand.Prime(rand.Reader, bitSize)
		if err != nil {
			panic(err)
		}
		fixedModulus = true
	case "C":
		if !supportedBitSizes[bitSize] {
			fmt.Printf("Bit size %d is not supported\n", bitSize)
			break
		}
		limbSize = bitSize / 64
		fixedModulus = false
	default:
		fmt.Printf("Do nothing. No such option %s\n\n", opt)
		flag.PrintDefaults()
	}

	arithmeticDeclerationsCode := pkg("fp") + arithmeticDeclerations(limbSize, fixedModulus)
	fieldElementImplCode := pkg("fp") + fieldElementImpl(limbSize)
	fieldImplCode := pkg("fp") + fieldImpl(limbSize, modulusBig)
	testCode := ""
	if fixedModulus {
		testCode = fieldTestFixedModulus
	} else {
		testCode = fieldTestNonFixedModulus
	}
	writeToFile(arithmeticDeclerationsCode, filepath.Join(outDir, "arithmetic_decl.go"))
	writeToFile(fieldElementImplCode, filepath.Join(outDir, "field_element.go"))
	writeToFile(fieldImplCode, filepath.Join(outDir, "field.go"))
	writeToFile(pkg("fp")+testCode, filepath.Join(outDir, "field_test.go"))
	return nil
}

func pkg(name string) string {
	return fmt.Sprintf("package %s\n", name)
}

func writeToFile(content string, out string) {
	if err := ioutil.WriteFile(out, []byte(content), 0600); err != nil {
		panic(err)
	}
}
