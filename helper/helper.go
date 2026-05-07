package helper

import (
	"flag"
	"go/ast"
	"log"
	"strings"

	"github.com/dave/jennifer/jen"
)

var (
	FileName = ""
	Package  = ""
)

func ReadExprType(expr ast.Expr, result map[string]string) {

	switch t := expr.(type) {

	case *ast.SelectorExpr:
		// model.User
		if pkg, ok := t.X.(*ast.Ident); ok {
			result[pkg.Name+"."+t.Sel.Name] = "struct"
		}

	case *ast.StarExpr:
		// *model.User
		ReadExprType(t.X, result)

	case *ast.ArrayType:
		// []model.User
		ReadExprType(t.Elt, result)

	case *ast.MapType:
		ReadExprType(t.Key, result)
		ReadExprType(t.Value, result)

	case *ast.Ellipsis:
		ReadExprType(t.Elt, result)
	}
}

func GenerateMockVars(models map[string]string) []jen.Code {
	var codes []jen.Code

	for full, importPath := range models {

		part := strings.Split(full, ".")
		pkg := part[0]
		name := part[1]

		varName := "mock" + name

		if importPath != "" {
			codes = append(codes,
				jen.Id(varName).Qual(importPath, name),
			)
		} else {
			codes = append(codes,
				jen.Id(varName).Id(pkg).Dot(name),
			)
		}
	}

	return codes
}

func GenerateInit(models map[string]string) []jen.Code {
	var body []jen.Code

	for full := range models {

		part := strings.Split(full, ".")
		// pkg := part[0]
		name := part[1]

		varName := "mock" + name

		body = append(body,
			jen.Qual("github.com/go-faker/faker/v4", "FakeData").Params(jen.Op("&").Id(varName)),
		)
	}

	return body
}

func GetFilePath() string {
	flag.StringVar(&FileName, "filename", "", "name of a usecase file")
	flag.StringVar(&Package, "package", "", "package of a usecase file")
	flag.Parse()

	if FileName == "" {
		log.Fatal("flag filename not found")
	}
	if Package == "" {
		log.Fatal("flag package not found")
	}

	return Package + FileName
}
