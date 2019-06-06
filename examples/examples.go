// Package examples defines a bunch of interfaces on which make.go.mock is run.
//
// Check out examples.go to see examples of go:generate directives, and the code
// they generate.
//
// See also the documentation for package github.com/tcard/make.go.mock/examples/generated.
// That's the code generated for the interfaces in this package.
//
// Check also the provided example in this package's documentation.
package examples

import (
	"os"
)

//go:generate make.go.mock -v -type MyInterface -dst mock_MyInterface_test.go
//go:generate make.go.mock -v -type MyInterface -dst in_test_pkg_test.go -dstpkg examples_test
//go:generate make.go.mock -v -type MyInterface -as DifferentName
//go:generate make.go.mock -v -type MyInterface -as MyInterfaceInCustomFile -dst custom_file_name_test.go
//go:generate make.go.mock -v -type MyInterface -dst generated/generated.go
//go:generate make.go.mock -v -type MyInterface -dst generated -dstpkg generated_test

type MyInterface interface {
	Embedded
	Boring()
	ReturnSomethingAtLeast() int
	ShouldBeFun(int, map[string]map[MyStruct]bool, ...chan<- <-chan struct{}) (int, error)
	StdSomething(f *os.File, ints ...int) (named bool)
}

type Embedded interface {
	EmbeddedMethod()
}

type MyStruct struct {
	SomeField int
	File      *os.File
}

//go:generate make.go.mock -v -type MyFunc

type MyFunc func(a, b, c int, x bool, multi ...MyStruct) (ok bool, err error)
