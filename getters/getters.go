package getters

import (
	"flag"
	"go/ast"
	"log"

	"github.com/lukmanhafidz/testgen/constants"
	"github.com/lukmanhafidz/testgen/helper"
	"github.com/lukmanhafidz/testgen/parser"
)

func GetFuncType() []*ast.FuncDecl {
	return parser.ParseFuncAST()
}

func GetStructType() (*ast.StructType, string) {
	var structType *ast.StructType
	var structName string

	typeSpec := parser.ParseTypeAST(structType)

	structName = typeSpec.Name.Name
	// Check if the type is an Interface
	if _, ok := typeSpec.Type.(*ast.StructType); ok {
		structType = typeSpec.Type.(*ast.StructType)
		// fmt.Printf("- %s\n", methodsName)
	}
	return structType, structName
}

func GetInterfaceType() (*ast.InterfaceType, string) {
	var interfaceType *ast.InterfaceType

	typeSpec := parser.ParseTypeAST(interfaceType)

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

func GetPathFile() string {
	var fileName, filePath string
	flag.StringVar(&fileName, "filename", "", "name of a usecase file")
	flag.StringVar(&filePath, "filepath", "", "path of a usecase file")
	flag.Parse()

	if fileName == "" {
		log.Fatal("flag filename not found")
	}
	if filePath == "" {
		log.Fatal("flag filepath not found")
	}

	constants.FileName = fileName
	constants.FilePath = filePath

	return fileName
}
