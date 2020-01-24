package x86

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/reg"
)

var mlo = RAX
var mhi = RDX

var supportedBitSizes = []int{
	64,
	128,
	192,
	256,
	320,
	384,
	448,
	512,
}

type bitFlags []int

func (flag *bitFlags) String() string {
	return "bit size flag"
}

func (flag *bitFlags) Set(value string) error {
	i, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}
	*flag = append(*flag, i)
	return nil
}

func GenX86All(output string) error {
	// a hack for avo output
	file := filepath.Join(output, "x86_arithmetic.s")
	if err := flag.Set("out", file); err != nil {
		return err
	}
	if err := os.MkdirAll(output, os.ModePerm); err != nil {
		return err
	}
	fixedmod, single, archTag := false, false, true
	for _, bitSize := range supportedBitSizes {
		limbSize := bitSize / 64
		generateCopy(limbSize, single)
		generateEq(limbSize, single)
		generateCmp(limbSize, single)
		generateAdd(limbSize, fixedmod, single)
		generateAddNoCar(limbSize, single)
		generateDouble(limbSize, fixedmod, single)
		generateSub(limbSize, fixedmod, single)
		generateSubNoCar(limbSize, single)
		generateNeg(limbSize, fixedmod, single)
		generateMul2(limbSize, single)
		generateDiv2(limbSize, single)
		if limbSize != 1 {
			genMontMulAdx(limbSize, fixedmod, single)
			genMontMulNoAdx(limbSize, fixedmod, single, archTag)
		}
	}
	Generate()
	appendSingleLimbMultiplicationCode(file)
	appendIsEvenCode(file)
	pretty(file)
	return nil
}

func GenX86(output string, bitSize int, arch string, fixedmod bool, single bool) error {
	// a hack for avo output
	file := filepath.Join(output, "arithmetic.s")
	if err := flag.Set("out", file); err != nil {
		return err
	}
	// Package("github.com/kilic/fp/" + output)
	limbSize := bitSize / 64
	if bitSize%64 != 0 {
		return fmt.Errorf(fmt.Sprintf("bad bit size, %d\n", bitSize))
	}
	if limbSize < 2 || limbSize > 8 {
		return fmt.Errorf("limb size %d not implemented\n", limbSize)
	}
	generateCopy(limbSize, single)
	generateEq(limbSize, single)
	generateCmp(limbSize, single)
	generateAdd(limbSize, fixedmod, single)
	generateAddNoCar(limbSize, single)
	generateDouble(limbSize, fixedmod, single)
	generateSub(limbSize, fixedmod, single)
	generateSubNoCar(limbSize, single)
	generateNeg(limbSize, fixedmod, single)
	switch arch {
	case "ADX":
		genMontMulAdx(limbSize, fixedmod, single)
	default:
		genMontMulNoAdx(limbSize, fixedmod, single, false)
	}
	Generate()
	pretty(file)
	return nil
}

func comment(str string) {
	Commentf("| \n\n/* %s \t\t\t\t*/\n", str)
}

func appendIsEvenCode(filename string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.WriteString(isEvenCode); err != nil {
		panic(err)
	}
}

func appendSingleLimbMultiplicationCode(filename string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.WriteString(singleLimbMultiplicationCode); err != nil {
		panic(err)
	}
	if _, err = f.WriteString(singleLimbMultiplicationNonAdxBmi2Code); err != nil {
		panic(err)
	}
}

func pretty(filename string) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln(err)
	}
	lines := strings.Split(string(input), "\n")
	for i, _ := range lines {
		lines[i] = strings.Replace(lines[i], "0x0000000000000000", "0x00", -1)
		lines[i] = strings.Replace(lines[i], "0x0000000000", "0x00", -1)
		lines[i] = strings.Replace(lines[i], "0x00000000", "0x00", -1)
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filename, []byte(output), 0600)
	if err != nil {
		log.Fatalln(err)
	}
}
