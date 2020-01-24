package x86

var isEvenCode = `
TEXT ·is_even(SB), NOSPLIT, $0-9
	MOVQ a+0(FP), DI
	MOVB $0x00, ret+8(FP)
	MOVQ 0(DI), AX
	TESTQ $1, AX 
	JNZ ret
	MOVB $0x01, ret+8(FP)
ret:
	RET
`
