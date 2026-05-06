package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"

	"github.com/lukmanhafidz/testgen/constants"
)

func ParseFileAST() *ast.File {
	fset := token.NewFileSet()
	filePath := constants.FilePath + constants.FileName

	node, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	return node
}

func ParseTypeAST(typeAST ast.Expr) *ast.TypeSpec {
	var result *ast.TypeSpec

	node := ParseFileAST()
	// Walk the AST to find type declarations
	ast.Inspect(node, func(n ast.Node) bool {
		// Look for Type Specifications (e.g., type Name interface)
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok || result != nil {
			return true
		}

		switch typeAST.(type) {
		case *ast.InterfaceType:
			if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
				result = typeSpec
				fmt.Printf("- %s\n", typeSpec.Name.Name)
			}

		case *ast.StructType:
			if _, ok := typeSpec.Type.(*ast.StructType); ok {
				result = typeSpec
				fmt.Printf("- %s\n", typeSpec.Name.Name)
			}

		default:
			fmt.Println("Unknown type")
			return false
		}

		return true
	})

	return result
}

func ParseFuncAST() []*ast.FuncDecl {
	var result []*ast.FuncDecl

	node := ParseFileAST()
	// Walk the AST to find func declarations
	ast.Inspect(node, func(n ast.Node) bool {
		// Look for Type Specifications (e.g., type Name interface)
		funcSpec, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		fmt.Printf("- %s\n", funcSpec.Name)

		result = append(result, funcSpec)
		return true
	})

	return result
}

func ParseImportAST() map[string]string {
	result := map[string]string{}

	node := ParseFileAST()
	ast.Inspect(node, func(n ast.Node) bool {
		// Check if the node is a General Declaration (import, const, var, type)
		if decl, ok := n.(*ast.GenDecl); ok && decl.Tok == token.IMPORT {
			for _, spec := range decl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					result[importSpec.Path.Value] = ""
					if importSpec.Name != nil {
						result[importSpec.Path.Value] = importSpec.Name.Name
					}
				}
				// fmt.Printf("Import Path: %s\n", importSpec.Path.Value)
			}
		}
		return true
	})

	return result
}

func ParsePackageAST() string {
	result := ""

	node := ParseFileAST()
	ast.Inspect(node, func(n ast.Node) bool {
		if pkg, ok := n.(*ast.File); ok {
			result = pkg.Name.Name
		}

		return true
	})

	return result
}
