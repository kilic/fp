package x86

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
)

func assert(c bool, desc string) {
	if !c {
		panic(desc)
	}
}

func comment(str string) {
	Commentf("| %s", str)
}

func commentHeader(str string) {
	s := fmt.Sprintf("%-40s", str)

	s = s[:40]
	Commentf("| \n\n/* %s*/\n", s)
}

func commentSeperator() {
	commentHeader("")
}

func commentU(i int) {
	comment(fmt.Sprintf("| u%[1]d = w%[1]d * inp", i))
}

func commentJ(j int) {
	comment(fmt.Sprintf("j%[1]d\n", j))
}

func commentI(i int) {
	commentHeader(fmt.Sprintf("i = %d", i))
}

func commentA(i int, ai *limb) {
	commentI(i)
	comment(fmt.Sprintf("a%d @ %s", i, ai.String()))
}

func commentB(i int, bi *limb) {
	commentI(i)
	comment(fmt.Sprintf("b%d @ %s", i, bi.String()))
}

func commentAiBj(i, j int) {
	comment(fmt.Sprintf("a%d * b%d ", i, j))
}
