package main

import (
	"fmt"
	"go/ast"
	"log"
	"os"
	"strings"

	"github.com/lukmanhafidz/testgen/getters"
	"github.com/lukmanhafidz/testgen/helper"
	"github.com/lukmanhafidz/testgen/parser"

	"github.com/dave/jennifer/jen"
)

func main() {
	var structFields, modelsPtr, interfaceFuncs, resultsFunc, usecaseParams []jen.Code
	var structName, mainPkg string
	funcReturn := jen.Dict{}

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	currentDirParts := strings.Split(currentDir, "/")
	mainPkg = currentDirParts[len(currentDirParts)-1]
	usecasePkg := parser.ParsePackageAST()
	fileName := getters.GetPathFile()

	f := jen.NewFile("testgen")
	f.Add()

	mocksPath := mainPkg + "/mocks"

	structType, structTypeName := getters.GetStructType()
	structName = structTypeName
	if strings.Contains(structName, "Usecase") {
		structName = strings.ReplaceAll(structName, "Usecase", "TestRepo")
	}

	for _, structVar := range structType.Fields.List {
		dataType := ""
		if structDataType, ok := structVar.Type.(*ast.StarExpr); ok { //Pointer Struct
			selectorStruct := structDataType.X.(*ast.SelectorExpr)
			dataType = "*" + selectorStruct.X.(*ast.Ident).Name + "." + selectorStruct.Sel.Name

		} else if structDataType, ok := structVar.Type.(*ast.SelectorExpr); ok {
			selectorStruct := structDataType.X.(*ast.Ident)
			dataType = selectorStruct.Name + "." + structDataType.Sel.Name
		}

		dataTypeStr := strings.Split(dataType, ".")
		if string(dataTypeStr[1][0]) == "I" && !strings.Contains(dataTypeStr[1], "Usecase") {
			structFields = append(structFields, jen.
				Id(structVar.Names[0].Name).
				Op("*").
				Qual(mocksPath, dataTypeStr[1]))

			modelsPtr = append(modelsPtr, jen.
				Id(structVar.Names[0].Name).
				Op(":=").
				Id("new").
				Params(jen.Qual(mocksPath, dataTypeStr[1])).
				Line())

			funcReturn[jen.Id(structVar.Names[0].Name)] = jen.Id(structVar.Names[0].Name)

			usecaseParams = append(usecaseParams, jen.
				Id(structTypeName+"Test").
				Dot(structVar.Names[0].Name))

		} else {
			usecaseParams = append(usecaseParams, jen.Id(structVar.Names[0].Name))
		}
	}

	f.Type().
		Id(structName).
		Struct(structFields...)

	funcDecl := new(ast.FuncDecl)
	funcTypes := getters.GetFuncType()

	usecaseModelMap := map[string]string{}
	importAliasMap := map[string]bool{}

	for _, funcs := range funcTypes {
		if strings.Contains(funcs.Name.Name, "New") {
			funcDecl = funcs
			continue
		}

		// must be method
		if funcs.Recv == nil {
			continue
		}

		usecaseModels := getters.GetModelsFromUsecase(funcs)
		for modelPkg, modelType := range usecaseModels {
			if _, ok := usecaseModelMap[modelPkg]; !ok {
				usecaseModelMap[modelPkg] = modelType
			}
		}
	}

	for modelPkg := range usecaseModelMap {
		pkgName := modelPkg
		if strings.Contains(modelPkg, ".") {
			pkgName = strings.Split(modelPkg, ".")[0]
		}

		usecaseImports := parser.ParseImportAST()
		for importPath, importAlias := range usecaseImports {
			importPath = strings.Trim(importPath, `\"`)

			if importAlias != "" {
				importAliasMap[importAlias] = true
			}

			if importAlias == pkgName {
				f.ImportAlias(importPath, importAlias)
				usecaseModelMap[modelPkg] = importPath

			} else if strings.Contains(importPath, pkgName) && !importAliasMap[importAlias] {
				f.ImportAlias(importPath, pkgName)
				usecaseModelMap[modelPkg] = importPath
				break
			}
		}
	}

	// log.Println(usecaseModelMap) //log for debug
	f.Var().Defs(helper.GenerateMockVars(usecaseModelMap)...)

	f.Func().
		Id("initNewData").
		Params().
		Block(helper.GenerateInit(usecaseModelMap)...).
		Line()

	constructName := funcDecl.Name.Name + "TestRepo"
	if strings.Contains(funcDecl.Name.Name, "Usecase") {
		constructName = strings.ReplaceAll(funcDecl.Name.Name, "Usecase", "TestRepo")
	}

	f.Func().
		Id(constructName).
		Params().
		Op("*").
		Id(structName).
		Block(
			jen.Add(modelsPtr...),
			jen.Return(jen.
				Op("&").
				Id(structName).
				Values(funcReturn)),
		).
		Line()

	interfaceType, interfaceName := getters.GetInterfaceType()
	for _, methodList := range interfaceType.Methods.List {
		methodName := "Test" + methodList.Names[0].Name
		interfaceFuncs = append(interfaceFuncs, jen.
			Line().
			Func().
			Params(jen.
				Id("r").
				Op("*").
				Id(structName)).
			Id(methodName).
			Params(jen.Id("t").Op("*").Qual("testing", "T"),
				jen.Id("usecase").Qual(mainPkg+"/"+usecasePkg, interfaceName)).
			Bool().
			Block(
				jen.Id("result").Op(":=").True(),
				jen.Return(jen.Id("result"))))

		resultName := "test" + methodList.Names[0].Name + "Result"
		testFunc := jen.Statement{jen.
			Id(resultName).
			Op(":=").
			Id(structTypeName+"Test").
			Dot(methodName).
			Params(jen.Id("t"), jen.Id(structTypeName)).
			Line().
			Qual("github.com/stretchr/testify/assert", "Equal").
			Params(
				jen.Id("t"),
				jen.True(),
				jen.Id(resultName)).
			Line().
			Id("initNewData").
			Params()}

		resultsFunc = append(resultsFunc, testFunc.Line().Line())
	}

	usecaseName := strings.ReplaceAll(funcDecl.Name.Name, "New", "Test")
	f.Func().
		Id(usecaseName).
		Params(jen.
			Id("t").
			Op("*").
			Qual("testing", "T")).
		Block(jen.Id(structTypeName+"Test").Op(":=").Id(constructName).Params(),
			jen.Id(structTypeName).Op(":=").Qual(mainPkg+"/"+usecasePkg, funcDecl.Name.Name).Params(usecaseParams...).Line(),
			jen.Add(resultsFunc...))

	f.Add(interfaceFuncs...)

	fileName = strings.Trim(fileName, ".go")
	err = f.Save("tests/" + fileName + "_test.go")
	if err != nil {
		fmt.Printf("error: %v", err)
	}
}

/*TODO:
- implement usecase
- scan usecase without naming
- faker data
*/
