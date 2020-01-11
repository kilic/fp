package x86

import (
	"testing"

	. "github.com/mmcloughlin/avo/build"
)

func TestGen(t *testing.T) {
	genMontMul48Adx(6, true, true)
	Generate()
}
