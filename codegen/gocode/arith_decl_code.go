package gocode

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
