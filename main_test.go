package main

import (
	"testing"
)

func TestFindStruct(t *testing.T) {
	pkgMap, err := findPkg("./demo/pkg2")
	if err != nil {
		t.Fatalf("%+v\n", err)
	}

	typeSpecs := findStruct(pkgMap, "People")
	printField(typeSpecs)

}
