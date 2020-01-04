package fp

import "unsafe"

func mul_two_4(a unsafe.Pointer)
func mul_two_5(a unsafe.Pointer)
func mul_two_6(a unsafe.Pointer)
func mul_two_7(a unsafe.Pointer)
func mul_two_8(a unsafe.Pointer)

func div_two_4(a unsafe.Pointer)
func div_two_5(a unsafe.Pointer)
func div_two_6(a unsafe.Pointer)
func div_two_7(a unsafe.Pointer)
func div_two_8(a unsafe.Pointer)

func is_even(a unsafe.Pointer) bool

//go:noescape
func mul4(c, a, b, p unsafe.Pointer, inp uint64)

//go:noescape
func mul5(c, a, b, p unsafe.Pointer, inp uint64)

//go:noescape
func mul6(c, a, b, p unsafe.Pointer, inp uint64)

//go:noescape
func mul7(c, a, b, p unsafe.Pointer, inp uint64)

//go:noescape
func mul8(c, a, b, p unsafe.Pointer, inp uint64)

//go:noescape
func add4(c, a, b, p unsafe.Pointer)

//go:noescape
func add5(c, a, b, p unsafe.Pointer)

//go:noescape
func add6(c, a, b, p unsafe.Pointer)

//go:noescape
func add7(c, a, b, p unsafe.Pointer)

//go:noescape
func add8(c, a, b, p unsafe.Pointer)

//go:noescape
func sub4(c, a, b, p unsafe.Pointer)

//go:noescape
func sub5(c, a, b, p unsafe.Pointer)

//go:noescape
func sub6(c, a, b, p unsafe.Pointer)

//go:noescape
func sub7(c, a, b, p unsafe.Pointer)

//go:noescape
func sub8(c, a, b, p unsafe.Pointer)

//go:noescape
func double4(c, a, p unsafe.Pointer)

//go:noescape
func double5(c, a, p unsafe.Pointer)

//go:noescape
func double6(c, a, p unsafe.Pointer)

//go:noescape
func double7(c, a, p unsafe.Pointer)

//go:noescape
func double8(c, a, p unsafe.Pointer)

//go:noescape
func _neg4(c, a, p unsafe.Pointer)

//go:noescape
func _neg5(c, a, p unsafe.Pointer)

//go:noescape
func _neg6(c, a, p unsafe.Pointer)

//go:noescape
func _neg7(c, a, p unsafe.Pointer)

//go:noescape
func _neg8(c, a, p unsafe.Pointer)

//go:noescape
func eq4(a, b unsafe.Pointer) bool

//go:noescape
func eq5(a, b unsafe.Pointer) bool

//go:noescape
func eq6(a, b unsafe.Pointer) bool

//go:noescape
func eq7(a, b unsafe.Pointer) bool

//go:noescape
func eq8(a, b unsafe.Pointer) bool

//go:noescape
func cpy4(dst, src unsafe.Pointer)

//go:noescape
func cpy5(dst, src unsafe.Pointer)

//go:noescape
func cpy6(dst, src unsafe.Pointer)

//go:noescape
func cpy7(dst, src unsafe.Pointer)

//go:noescape
func cpy8(dst, src unsafe.Pointer)

//go:noescape
func cmp4(a, b unsafe.Pointer) int8
func cmpx(a, b unsafe.Pointer) int8

//go:noescape
func cmp5(a, b unsafe.Pointer) int8

//go:noescape
func cmp6(a, b unsafe.Pointer) int8

//go:noescape
func cmp7(a, b unsafe.Pointer) int8

//go:noescape
func cmp8(a, b unsafe.Pointer) int8

//go:noescape
func addn4(a, b unsafe.Pointer) uint64

//go:noescape
func addn5(a, b unsafe.Pointer) uint64

//go:noescape
func addn6(a, b unsafe.Pointer) uint64

//go:noescape
func addn7(a, b unsafe.Pointer) uint64

//go:noescape
func addn8(a, b unsafe.Pointer) uint64

//go:noescape
func subn4(a, b unsafe.Pointer) uint64

//go:noescape
func subn5(a, b unsafe.Pointer) uint64

//go:noescape
func subn6(a, b unsafe.Pointer) uint64

//go:noescape
func subn7(a, b unsafe.Pointer) uint64

//go:noescape
func subn8(a, b unsafe.Pointer) uint64
