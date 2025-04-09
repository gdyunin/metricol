package exitcheck

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var ExitInMainAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for call os.Exit in main",
	Run:  run,
}

func run(p *analysis.Pass) (any, error) {
	if p.Pkg.Name() != "main" {
		return nil, nil //nolint
	}

	for _, file := range p.Files {
		filename := p.Fset.File(file.Pos()).Name()
		if strings.Contains(filename, ".cache") {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Name.Name != "main" {
				return true
			}

			ast.Inspect(fd, funcContainExitInspector(p))

			return false
		})
	}
	return nil, nil //nolint
}

func funcContainExitInspector(p *analysis.Pass) func(ast.Node) bool {
	return func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		obj := p.TypesInfo.Uses[sel.Sel]
		if obj == nil {
			return true
		}

		if obj.Pkg() == nil || obj.Pkg().Path() != "os" {
			return true
		}

		if obj.Name() != "Exit" {
			return true
		}

		// - Вы указали уровень английского intermediate. Составьте предложение со словом "allow"
		// - Allow, kto eto zvonit?
		p.Reportf(ce.Lparen, "don't allow call os.Exit in main function")

		return true
	}
}
