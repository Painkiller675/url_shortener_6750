// Package exitmainchecker contains check for usage of exit function in main function
package testlint

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ErrExitMainCheckAnalyzer - an object of analysis.Analyzer type to create a custom analyzer for the multichecker
var ErrExitMainCheckAnalyzer = &analysis.Analyzer{
	Name: "exitMainCheck",
	Doc:  "check call os.Exit in func main() of package main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {

		// tests are generating build cache that has main package, ignoring such files
		if fullPath := pass.Fset.Position(file.Pos()).String(); strings.Contains(fullPath, "go-build") {
			continue
		}

		if pass.Pkg.Name() != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			mainDecl, isFuncDecl := node.(*ast.FuncDecl)
			if !isFuncDecl {
				return true
			}

			if mainDecl.Name.Name != "main" {
				return false
			}

			ast.Inspect(mainDecl, func(node ast.Node) bool {
				callExpr, isCallExpr := node.(*ast.CallExpr)
				if !isCallExpr {
					return true
				}

				s, isSelectorExpr := callExpr.Fun.(*ast.SelectorExpr)
				if !isSelectorExpr {
					return true
				}

				if s.Sel.Name == "Exit" {
					ident := s.X.(*ast.Ident)
					if ident.Name == "os" {
						pass.Reportf(s.Pos(), "exit call in main function")
					}
				}

				return false
			})

			return false
		})
	}

	return nil, nil //nolint:nilnil
}
