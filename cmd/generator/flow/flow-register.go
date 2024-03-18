/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"
)

const genTemplate = `// Code generated by go generate; DO NOT EDIT.
package {{.PackageName}}

func init() {
{{- range .Flows }}
	flowMap[f{{.Name}}] = {{.GeneratorName}}
{{- end }}
}
`

type Flow struct {
	Name          string
	GeneratorName string
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <source_file>", os.Args[0])
	}
	sourceFile := os.Args[1]

	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, sourceFile, nil, 0)
	if err != nil {
		log.Fatalf("Failed to parse source file: %v", err)
	}

	flows := []Flow{}

	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if len(fn.Type.Params.List) == 1 && len(fn.Type.Results.List) == 1 {
			if strings.HasSuffix(exprToString(fn.Type.Results.List[0].Type), "TaskFlow") {
				targetName := strings.ReplaceAll(fn.Name.Name, "Flow", "")
				flows = append(flows, Flow{
					Name:          targetName,
					GeneratorName: fn.Name.Name,
				})
			}
		}

		return true
	})

	tmpl, err := template.New("registration").Parse(genTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	outputFile := sourceFile[:len(sourceFile)-3] + "_gen.go"

	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	err = tmpl.Execute(f, struct {
		PackageName string
		Flows       []Flow
	}{
		PackageName: node.Name.Name,
		Flows:       flows,
	})
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	default:
		return fmt.Sprintf("unknown(%T)", e)
	}
}
