package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"reflect"
	"strings"
)

var (
	// go mod name
	modName = "github.com/jdxj/study-ast"
	pkgPath string
	name    string
	// 解析缓存
	pkgMap = make(map[string]*ast.Package)
)

func main() {
	flag.StringVar(&modName, "mod-name", "", "go mod name")
	flag.StringVar(&pkgPath, "pkg-path", "", "target struct pkg path without mod-name")
	flag.StringVar(&name, "name", "", "struct name")
	flag.Parse()

	if modName == "" || pkgPath == "" || name == "" {
		return
	}

	ss := findStruct(pkgPath, name)
	for _, s := range ss {
		fmt.Printf("%s:\n", s.Name)
		for _, v := range s.Fields {
			fmt.Printf("    |%s|%s|%s|\n", v.Name, v.Type, v.Description)
		}
	}
}

// Struct 描述了其所含字段的描述
type Struct struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name        string
	Type        string
	Description string
}

// findStruct 在指定的 pkgPath 中寻找 structName,
// 返回值中第0个为该 structName, 剩余的是其所依赖的 struct.
func findStruct(pkgPath, structName string) []Struct {
	_, ok := pkgMap[pkgPath]
	if !ok {
		tfs := token.NewFileSet()
		pkgs, err := parser.ParseDir(tfs, pkgPath, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}

		// pkgPath 为某个包的路径, 所以 pkgs 中只有一个值
		for _, astPkg := range pkgs {
			pkgMap[pkgPath] = astPkg
		}
	}
	return findStructInPkg(pkgPath, pkgMap[pkgPath], structName)
}

func findStructInPkg(curPkgPath string, astPkg *ast.Package, structName string) []Struct {
	var ss []Struct
	for _, file := range astPkg.Files {
		ss = append(ss, findStructInFile(file, curPkgPath, structName)...)
	}
	return ss
}

func findStructInFile(file *ast.File, curPkgPath, structName string) []Struct {
	var (
		ss        []Struct
		importMap = make(map[string]string)
	)
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		switch genDecl.Tok {
		case token.IMPORT:
			// 记录当前文件的 import
			for _, spec := range genDecl.Specs {
				importSpec := spec.(*ast.ImportSpec)
				pkgPath := strings.Trim(importSpec.Path.Value, `"`)
				pkgPath = strings.TrimPrefix(pkgPath, modName+"/")

				pkg := path.Base(pkgPath)
				if importSpec.Name != nil {
					pkg = importSpec.Name.Name
				}
				importMap[pkg] = pkgPath
			}

		case token.TYPE:
			// 寻找 struct 定义
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if typeSpec.Name.Name != structName {
					continue
				}
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				// 创建 Struct
				fields, ssIn := getFields(importMap, structType, structName, curPkgPath)
				s := Struct{
					Name:   structName,
					Fields: fields,
				}

				ss = append(ss, s)
				ss = append(ss, ssIn...)
				// 递归查找
				ss = append(ss, getDeepStruct(importMap, structType, curPkgPath)...)
			}
		}

	}
	return ss
}

// getFields 获取当前 struct 的 field
func getFields(importMap map[string]string, parentStructType *ast.StructType, parentStructName, curPkgPath string) ([]Field, []Struct) {
	var (
		fields = make([]Field, 0, len(parentStructType.Fields.List))
		ssOut  []Struct
	)
	for _, field := range parentStructType.Fields.List {
		pkgPath, structName, typ, structType := unwrapFieldType(importMap, curPkgPath, field)
		// 匿名字段
		if field.Names == nil {
			ssIn := findStruct(pkgPath, structName)
			// 第一个是 structName 本身, 将其 fields 赋给当前 struct
			// len(ssIn) != 0, 否则是不能编译的 go 代码
			fields = append(fields, ssIn[0].Fields...)
			// 剩下的是 structName 所依赖的
			ssOut = append(ssOut, ssIn[1:]...)
			continue
		}

		// 匿名 struct
		if structType != nil {
			structName = fmt.Sprintf(".%s.%s", parentStructName, structName)
			fields, ssIn := getFields(importMap, structType, structName, curPkgPath)
			s := Struct{
				Name:   structName,
				Fields: fields,
			}
			ssOut = append(ssOut, s)
			ssOut = append(ssOut, ssIn...)
		}

		// 普通字段
		fields = append(fields, Field{
			Name: getTag(field),
			//Type:        getType(field),
			Type:        typ,
			Description: getDescription(field),
		})
	}
	return fields, ssOut
}

func getDeepStruct(importMap map[string]string, structType *ast.StructType, curPkgPath string) []Struct {
	var ss []Struct
	for _, field := range structType.Fields.List {
		// 跳过匿名
		if field.Names == nil {
			continue
		}

		pkgPath, structName, _, st := unwrapFieldType(importMap, curPkgPath, field)
		if st != nil {
			continue
		}
		ss = append(ss, findStruct(pkgPath, structName)...)
	}
	return ss
}

func getTag(field *ast.Field) string {
	if field.Tag == nil {
		return ""
	}
	tag := strings.Trim(field.Tag.Value, "`")
	return reflect.StructTag(tag).Get("json")
}

func unwrapFieldType(importMap map[string]string, curPkgPath string, field *ast.Field) (
	pkgPath, structName, literal string, structType *ast.StructType) {

	pkgPath = curPkgPath
	if len(field.Names) != 0 {
		// 匿名 struct 使用字段名
		structName = field.Names[0].Name
	}
	var (
		expr   = field.Type
		prefix string
	)
	for {
		switch exprType := expr.(type) {
		case *ast.ArrayType:
			expr = exprType.Elt
			prefix += "[]"
		case *ast.StarExpr:
			expr = exprType.X
		case *ast.SelectorExpr:
			pkgPath = importMap[exprType.X.(*ast.Ident).Name]
			expr = exprType.Sel
		case *ast.StructType:
			literal = prefix + "struct"
			structType = exprType
			return
		case *ast.InterfaceType:
			literal = prefix + "Object"
			structName = "interface"
			return
		case *ast.Ident:
			structName = exprType.Name
			literal = prefix + exprType.Name
			return
		}
	}
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
