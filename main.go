// Command make.go.mock generates type-safe mocks for Go interfaces and functions.
//
// See README.md for details.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tcard/make.go.mock/internal/makegomock"
)

func main() {
	typeName := flag.String("type", "", "name of the type to mock")
	as := flag.String("as", "", "base name for generated identifiers; leave blank for default")
	dst := flag.String("dst", "", "path of the generated file; pass a dir for default file name; leave black for same dir")
	dstPkgName := flag.String("dstpkg", "", "package name for the generated file; leave blank to infer")
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()

	if *typeName == "" {
		exit("expected non-empty -type")
	}

	dstFilePath, err := makegomock.GenerateToFileFromFile(
		*dst,
		*dstPkgName,
		*as,
		os.Getenv("GOFILE"),
		os.Getenv("GOPACKAGE"),
		*typeName,
	)
	nilOrExit(err, "%s")

	if *verbose {
		fmt.Fprintf(os.Stderr, "generated %s\n", dstFilePath)
	}
}

func nilOrExit(err error, s string, args ...interface{}) {
	if err != nil {
		if errs, ok := err.(makegomock.Errors); ok {
			for _, err := range errs.Errs {
				fmt.Fprintln(os.Stderr, err)
			}
			err = errs.Err
		}
		exit(s, append(args, err)...)
	}
}

func exit(s string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Errorf(s, args...))
	os.Exit(1)
}
