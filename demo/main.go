package main

import (
	"fmt"

	"github.com/jdxj/study-ast/demo/pkg1"
	"github.com/jdxj/study-ast/demo/pkg2"
)

func main() {
	p := &pkg2.People{
		Animal: &pkg1.Animal{Name: "jdxj"},
		Age:    8,
	}
	fmt.Printf("%+v\n", p)
}
