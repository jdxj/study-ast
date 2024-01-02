package main

import (
	"fmt"
	"testing"
)

func TestFindStruct(t *testing.T) {
	ss := findStruct("demo/pkg3", "Object")

	for _, s := range ss {
		fmt.Printf("%s:\n", s.Name)
		for _, v := range s.Fields {
			fmt.Printf("    |%s|%s|%s|\n", v.Name, v.Type, v.Description)
		}
	}
}
