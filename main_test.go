package main

import (
	"fmt"
	"testing"
)

func TestFindStruct(t *testing.T) {
	ss := findPkg("demo/pkg2", "People")

	for _, s := range ss {
		fmt.Printf("%s:\n", s.Name)
		for _, v := range s.Fields {
			fmt.Printf("    |%s|%s|%s|\n", v.Name, v.Type, v.Description)
		}
	}
}
