package x86

var singleLimbMultiplicationCode = `
// func mul1(c *[1]uint64, a *[1]uint64, b *[1]uint64, p *[1]uint64, inp uint64)
TEXT ·mul1(SB), NOSPLIT, $0-40

/* inputs 								*/

	MOVQ a+8(FP), DI
	MOVQ b+16(FP), SI

/* multiplication 				*/

	MOVQ (SI), DX
	MULXQ (DI), R8, R9

/* montgommery reduction	*/

	MOVQ p+24(FP), R15
  MOVQ  R8, DX
	MULXQ inp+32(FP), DX, DI

  MULXQ (R15), AX, DI
  ADDQ AX, R8
	ADCQ DI, R9
	ADCQ $0x00, R8

/* modular reduction 			*/

	MOVQ R9, AX
	SUBQ (R15), AX
	SBBQ $0x00, R8

/* out 										*/

	MOVQ    c+0(FP), DI
	CMOVQCC AX, R9
	MOVQ    R9, (DI)
	RET

/* end 				*/
`

var singleLimbMultiplicationNonAdxBmi2Code = `
// func mul_no_adx_bmi2_1(c *[1]uint64, a *[1]uint64, b *[1]uint64, p *[1]uint64, inp uint64)
TEXT ·mul_no_adx_bmi2_1(SB), NOSPLIT, $0-40

/* inputs 										*/

	MOVQ a+8(FP), DI
	MOVQ b+16(FP), SI

	// | 

/* multiplication 						*/

	MOVQ (SI), CX
	MOVQ (DI), AX
	MULQ CX
	MOVQ AX, R8
	MOVQ DX, R9

/* montgommery reduction 			*/

	MOVQ p+24(FP), R15

	MOVQ R8, AX
	MULQ inp+32(FP)
	MOVQ AX, CX

	MOVQ (R15), AX
	MULQ CX
	ADDQ AX, R8
	ADCQ DX, R9
	ADCQ $0x00, R8

/* modular reduction 				*/

	MOVQ R9, AX
	SUBQ (R15), AX
	SBBQ $0x00, R8

/* out 											*/

	MOVQ    c+0(FP), DI
	CMOVQCC AX, R9
	MOVQ    R9, (DI)
	RET

/* end 											*/
`
