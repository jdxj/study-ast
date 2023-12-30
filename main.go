package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
)

/*
 1. 解析 go 文件
 2. 查找指定 struct
 3. 读取字段 name type comment
 4. 如果某字段为struct
    a. 如果是本包
    b. 如果是其他包, 则解析对应包中的 go 文件
*/
func main() {
}

type Struct struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name        string
	Type        string
	Description string
}

func findPkg(path string) (map[string]*ast.Package, error) {
	var dirs []string
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(dirs) == 0 {
		return nil, nil
	}

	var (
		tfs    = token.NewFileSet()
		pkgMap = make(map[string]*ast.Package)
	)
	for _, p := range dirs {
		pkgs, err := parser.ParseDir(tfs, p, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		for pkg, astPkg := range pkgs {
			pkgPath := filepath.Join(p, pkg)
			pkgMap[pkgPath] = astPkg
		}
	}
	return pkgMap, nil
}

func findStruct(pkgMap map[string]*ast.Package, name ...string) []*ast.TypeSpec {
	if len(name) == 0 {
		return nil
	}
	nameMap := make(map[string]struct{})
	for _, v := range name {
		nameMap[v] = struct{}{}
	}

	var typeSpecs []*ast.TypeSpec
	for pkgPath, astPkg := range pkgMap {
		typeSpecs = append(typeSpecs, findStructInPkg(pkgPath, astPkg, nameMap)...)
	}
	return typeSpecs
}

func findStructInPkg(pkgPath string, astPkg *ast.Package, nameMap map[string]struct{}) []*ast.TypeSpec {
	var typeSpecs []*ast.TypeSpec
	for filePath, file := range astPkg.Files {
		typeSpecs = append(typeSpecs, findStructInFile(filePath, file, nameMap)...)
	}
	return typeSpecs
}

func findStructInFile(filePath string, file *ast.File, nameMap map[string]struct{}) []*ast.TypeSpec {
	var typeSpecs []*ast.TypeSpec
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			_, ok = nameMap[typeSpec.Name.Name]
			if !ok {
				continue
			}
			_, ok = typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			typeSpecs = append(typeSpecs, typeSpec)

		}
	}
	return typeSpecs
}

func printField(typeSpecs []*ast.TypeSpec) {
	for _, typeSpec := range typeSpecs {
		fmt.Printf("%s:\n", typeSpec.Name.Name)

		structType := typeSpec.Type.(*ast.StructType)
		for _, field := range structType.Fields.List {
			tag := getTag(field)
			typ := getType(field)
			desc := getDescription(field)
			fmt.Printf("    %s, %s, %s\n", tag, typ, desc)
		}
	}
}

func getTag(field *ast.Field) string {
	if field.Tag == nil {
		return ""
	}
	tag := strings.Trim(field.Tag.Value, "`")
	return reflect.StructTag(tag).Get("json")
}

func getType(field *ast.Field) string {
	var typ string
	switch expr := field.Type.(type) {
	case *ast.StarExpr:
		switch expr := expr.X.(type) {
		case *ast.SelectorExpr:
			typ = expr.Sel.Name
		}

	case *ast.Ident:
		typ = expr.Name
	}
	return typ
}

func getDescription(field *ast.Field) string {
	if field.Comment == nil {
		return ""
	}
	desc := field.Comment.List[0].Text
	desc = strings.Trim(desc, "//")
	desc = strings.Trim(desc, " ")
	return desc
}
