package getters

import (
	"go/ast"

	"github.com/lukmanhafidz/testgen/helper"
	"github.com/lukmanhafidz/testgen/parser"
)

func GetFuncType(filepath string) []*ast.FuncDecl {
	return parser.ParseFuncAST(filepath)
}

func GetStructType(filepath string) (*ast.StructType, string) {
	var structType *ast.StructType
	var structName string

	typeSpec := parser.ParseTypeAST(structType, filepath)

	structName = typeSpec.Name.Name
	// Check if the type is an Interface
	if _, ok := typeSpec.Type.(*ast.StructType); ok {
		structType = typeSpec.Type.(*ast.StructType)
		// fmt.Printf("- %s\n", methodsName)
	}
	return structType, structName
}

func GetInterfaceType(filepath string) (*ast.InterfaceType, string) {
	var interfaceType *ast.InterfaceType

	typeSpec := parser.ParseTypeAST(interfaceType, filepath)

	interfaceName := typeSpec.Name.Name
	// Check if the type is an Interface
	if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
		interfaceType = typeSpec.Type.(*ast.InterfaceType)
		// fmt.Printf("- %s\n", methodsName)
	}

	return interfaceType, interfaceName
}

func GetModelsFromUsecase(fn *ast.FuncDecl) map[string]string {
	result := map[string]string{}

	// params
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			helper.ReadExprType(field.Type, result)
		}
	}

	// return values
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			helper.ReadExprType(field.Type, result)
		}
	}

	// body
	if fn.Body != nil {
		ast.Inspect(fn.Body, func(n ast.Node) bool {

			switch x := n.(type) {

			case *ast.CompositeLit:
				helper.ReadExprType(x.Type, result)

			case *ast.CallExpr:
				if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "new" {
					if len(x.Args) > 0 {
						helper.ReadExprType(x.Args[0], result)
					}
				}
			}

			return true
		})
	}

	return result
}
