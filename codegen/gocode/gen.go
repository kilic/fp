package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"text/template"
)

func main() {
	out := flag.String("out", "", "")
	from := flag.Int("from", 4, "")
	to := flag.Int("to", 16, "")
	genFields := flag.Bool("field", false, "")
	genDeclarations := flag.Bool("decl", false, "")
	globalModulus := flag.Bool("globmod", false, "")
	flag.Parse()
	if *genFields {
		GenerateFieldElements(*out+"/"+"field_elements.go", *from, *to)
		GenerateFields(*out+"/"+"fields.go", *from, *to, *globalModulus)
		GenerateFieldElementTests(*out+"/"+"field_test.go", *from, *to)
	}
	if *genDeclarations {
		GenerateDeclerations(*out+"/"+"arithmetic_decl.go", *from, *to, *globalModulus)
		GenerateTypes(*out+"/"+"types.go", *from, *to)
	}
}

type feSize struct {
	Bit          int
	Limb         int
	Bytes        int
	Iter         int
	FieldElement string
	Field        string
	GlobMod      bool
}

func pkg(name string) string {
	return fmt.Sprintf("package %s\n", name)
}

func imports(str string, imports []string) string {
	if len(imports) > 0 {
		str += fmt.Sprintf("%s\n", "import (")
		for _, imprt := range imports {
			str += fmt.Sprintf("\"%s\"\n", imprt)
		}
		str += fmt.Sprintf("%s\n", ")")
	}
	return str
}

func generate(declerations string, templates []string, funcs template.FuncMap, data interface{}) (string, error) {
	codeStr := ""
	acc := declerations + "\n"
	for _, t := range templates {
		acc += t + "\n"
	}
	template, err := template.New("").Funcs(funcs).Parse(acc)
	if err != nil {
		return "", err
	}
	buffer := new(bytes.Buffer)
	err = template.Execute(buffer, data)
	if err != nil {
		return "", err
	}
	codeStr += fmt.Sprintf("\n%s", buffer.String())
	return codeStr, nil
}

func GenerateFieldElements(out string, from, to int) {
	codeStr := pkg("fp")
	codeStr = imports(codeStr, []string{"math/big", "math/bits", "io", "fmt", "encoding/hex"})
	for i := from; i <= to; i++ {
		data := feSize{
			Limb:         i,
			Bit:          64 * i,
			FieldElement: fmt.Sprintf("Fe%d", 64*i),
			Field:        fmt.Sprintf("Field%d", 64*i),
			Bytes:        i * 8,
		}
		declerations := "" +
			"{{ $N_LIMB := .Limb }}" +
			"{{ $N_BIT := .Bit }}" +
			"{{ $FE := .FieldElement }}" +
			"{{ $FIELD := .Field }}" +
			"{{ $N_BYTES := .Bytes }}"
		if generated, err := generate(declerations, fieldElementTemplates, utilFuncs, data); err != nil {
			panic(err)
		} else {
			codeStr += "\n" + generated
		}
	}
	if err := ioutil.WriteFile(out, []byte(codeStr), 0600); err != nil {
		panic(err)
	}
}

func GenerateFields(out string, from, to int, globalModulus bool) {
	codeStr := pkg("fp")
	codeStr = imports(codeStr, []string{"fmt", "math/big", "io", "crypto/rand"})
	for i := from; i <= to; i++ {
		data := feSize{
			Limb:         i,
			Bit:          64 * i,
			FieldElement: fmt.Sprintf("Fe%d", 64*i),
			Field:        fmt.Sprintf("Field%d", 64*i),
			Bytes:        i * 8,
			GlobMod:      globalModulus,
		}
		declerations := "" +
			"{{ $N_LIMB := .Limb }}" +
			"{{ $N_BIT := .Bit }}" +
			"{{ $FE := .FieldElement }}" +
			"{{ $FIELD := .Field }}" +
			"{{ $N_BYTES := .Bytes }}" +
			"{{ $GlobMod := .GlobMod }}"
		if generated, err := generate(declerations, fieldTemplates, utilFuncs, data); err != nil {
			panic(err)
		} else {
			codeStr += "\n" + generated
		}
	}
	if err := ioutil.WriteFile(out, []byte(codeStr), 0600); err != nil {
		panic(err)
	}
}

func GenerateDeclerations(out string, from, to int, globalModulus bool) {
	codeStr := pkg("fp")
	// https://github.com/mmcloughlinto/avo/issues/60
	// function declaration in avo with TEXT function
	// does not support external types.
	// So we have generate stubs in advance.
	for i := from; i <= to; i++ {
		if globalModulus {
			codeStr += fmt.Sprintf("func add%d(c, a, b *Fe%d)\n\n", i, i*64)
			codeStr += fmt.Sprintf("func addn%d(a, b *Fe%d) uint64\n\n", i, i*64)
			codeStr += fmt.Sprintf("func sub%d(c, a, b *Fe%d)\n\n", i, i*64)
			codeStr += fmt.Sprintf("func subn%d(a, b *Fe%d) uint64\n\n", i, i*64)
			codeStr += fmt.Sprintf("func neg%d(c, a *Fe%d)\n\n", i, i*64)
			codeStr += fmt.Sprintf("func double%d(c, a *Fe%d)\n\n", i, i*64)
			codeStr += fmt.Sprintf("func mul%d(c *[%d]uint64, a, b *Fe%d)\n\n", i, i*2, i*64)
			codeStr += fmt.Sprintf("func square%d(c *[%d]uint64, a *Fe%d)\n\n", i, i*2, i*64)
			codeStr += fmt.Sprintf("func mont%d(c *Fe%d, w *[%d]uint64)\n\n", i, i*64, i*2)
			codeStr += fmt.Sprintf("func montmul%d(c, a, b *Fe%d)\n\n", i, i*64)
			codeStr += fmt.Sprintf("func montsquare%d(c, a *Fe%d)\n\n", i, i*64)
		} else {
			codeStr += fmt.Sprintf("func add%d(c, a, b, p *Fe%d)\n\n", i, i*64)
			codeStr += fmt.Sprintf("func addn%d(a, b *Fe%d) uint64\n\n", i, i*64)
			codeStr += fmt.Sprintf("func sub%d(c, a, b, p *Fe%d)\n\n", i, i*64)
			codeStr += fmt.Sprintf("func subn%d(a, b *Fe%d) uint64\n\n", i, i*64)
			codeStr += fmt.Sprintf("func neg%d(c, a, p *Fe%d)\n\n", i, i*64)
			codeStr += fmt.Sprintf("func double%d(c, a, p *Fe%d)\n\n", i, i*64)
			codeStr += fmt.Sprintf("func mul%d(c *[%d]uint64, a, b *Fe%d)\n\n", i, i*2, i*64)
			codeStr += fmt.Sprintf("func square%d(c *[%d]uint64, a, p *Fe%d)\n\n", i, i*2, i*64)
			codeStr += fmt.Sprintf("func mont%d(c *Fe%d, w *[%d]uint64, p *Fe%d,inp uint64)\n\n", i, i*64, i*2, i*64)
			codeStr += fmt.Sprintf("func montmul%d(c, a, b, p *Fe%d, inp uint64)\n\n", i, i*64)
			codeStr += fmt.Sprintf("func montsquare%d(c, a, p *Fe%d, inp uint64)\n\n", i, i*64)
		}
	}
	if err := ioutil.WriteFile(out, []byte(codeStr), 0600); err != nil {
		panic(err)
	}
}

func GenerateTypes(out string, from, to int) {
	codeStr := pkg("fp")
	for i := from; i <= to; i++ {
		codeStr += fmt.Sprintf("type Fe%d [%d]uint64\n", i*64, i)
	}
	if err := ioutil.WriteFile(out, []byte(codeStr), 0600); err != nil {
		panic(err)
	}
}

var utilFuncs = map[string]interface{}{
	"iterUp":   iterUp,
	"iterDown": iterDown,
	"decr":     decr,
	"mul":      mul,
}

func iterUp(from int, n int) []int {
	it := make([]int, n-from)
	for i := 0; i < len(it); i++ {
		it[i] = i + from
	}
	return it
}

func iterDown(n int) []int {
	it := make([]int, n)
	for i := 0; i < n; i++ {
		it[i] = n - 1 - i
	}
	return it
}

func decr(n int) int {
	return n - 1
}

func mul(n, m int) int {
	return n * m
}
