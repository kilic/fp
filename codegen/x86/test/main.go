package main

import (
	"flag"

	"github.com/kilic/fp/codegen/x86"
)

func main() {
	_limb := flag.Int("limb", 1, "# of iters")
	_fixed := flag.Bool("fixed", false, "# of iters")
	_noadx := flag.Bool("noadx", false, "# of iters")
	_logs := flag.Bool("logs", false, "# of iters")
	flag.Parse()
	var limbs = *_limb
	var fixed = *_fixed
	var noadx = *_noadx
	var logs = *_logs
	x86.GenDebugTest(limbs, fixed, noadx, logs)
}
