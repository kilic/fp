package gocode

import "fmt"

func arithmeticDeclerations(limbSize int, fixedModulus bool) string {
	code := ""
	if fixedModulus {
		code += "\n//go:noescape\nfunc add(c, a, b *fieldElement)\n" +
			"\n//go:noescape\nfunc addn(a, b *fieldElement) uint64\n" +
			"\n//go:noescape\nfunc sub(c, a, b *fieldElement)\n" +
			"\n//go:noescape\nfunc subn(a, b *fieldElement) uint64\n" +
			"\n//go:noescape\nfunc _neg(c, a *fieldElement)\n" +
			"\n//go:noescape\nfunc double(c, a *fieldElement)\n" +
			"\n//go:noescape\nfunc mul(c, a, b *fieldElement)\n"
	} else {
		code += "\n//go:noescape\nfunc add(c, a, b, p *fieldElement)\n" +
			"\n//go:noescape\nfunc addn(a, b *fieldElement) uint64\n" +
			"\n//go:noescape\nfunc sub(c, a, b, p *fieldElement)\n" +
			"\n//go:noescape\nfunc subn(a, b *fieldElement) uint64\n" +
			"\n//go:noescape\nfunc _neg(c, a, p *fieldElement)\n" +
			"\n//go:noescape\nfunc double(c, a, p *fieldElement)\n" +
			"\n//go:noescape\nfunc mul(c, a, b, p *fieldElement, inp uint64)\n"
	}
	return code
}

func arithmeticDeclerationsMultiple(limbSizes []int) string {
	var code string = `

//go:noescape
func is_even(a fieldElement) bool
`
	for i := 0; i < len(limbSizes); i++ {
		limbSize := limbSizes[i]

		code += fmt.Sprintf(`
//go:noescape
func eq%[1]d(a, b fieldElement) bool
	
//go:noescape
func mul_two_%[1]d(a fieldElement) uint64

//go:noescape
func div_two_%[1]d(a fieldElement)

//go:noescape
func cpy%[1]d(dst, src fieldElement)

//go:noescape
func cmp%[1]d(a, b fieldElement) int8

//go:noescape
func add%[1]d(c, a, b, p fieldElement)

//go:noescape
func addn%[1]d(a, b fieldElement) uint64

//go:noescape
func sub%[1]d(c, a, b, p fieldElement)

//go:noescape
func subn%[1]d(a, b fieldElement) uint64

//go:noescape
func _neg%[1]d(c, a, p fieldElement)

//go:noescape
func double%[1]d(c, a, p fieldElement)

//go:noescape
func mul%[1]d(c, a, b, p fieldElement, inp uint64)

//go:noescape
func mul_no_adx_bmi2_%[1]d(c, a, b, p fieldElement, inp uint64)
`, limbSize)

	}
	return code
}
