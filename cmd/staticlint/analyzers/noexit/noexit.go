/*
Package noexit provides a static analyzer that prohibits direct calls to os.Exit
in the main function of package main.

Direct calls to os.Exit in main terminate the program immediately without running
deferred functions, preventing proper cleanup and graceful shutdown.

Instead of calling os.Exit directly, use logger.Fatal() or return an error from main.
*/
package noexit

import (
	"go/ast"
	"os"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is the noexit analyzer that checks for direct os.Exit calls in main.
var Analyzer = &analysis.Analyzer{
	Name:     "noexit",
	Doc:      "check for direct os.Exit calls in main function",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Only check package main
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Pre-filter to only function calls and function declarations
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.CallExpr)(nil),
	}

	inspect.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			// Pop from stack when leaving node
			return true
		}

		// We're entering node, it's already on the stack
		if _, ok := n.(*ast.CallExpr); ok {
			checkOsExitCall(pass, n, stack)
		}

		return true
	})

	return nil, nil
}

// checkOsExitCall checks if the given call expression is os.Exit and reports if in main.
func checkOsExitCall(pass *analysis.Pass, n ast.Node, stack []ast.Node) {
	callExpr, ok := n.(*ast.CallExpr)
	if !ok {
		return
	}

	// Check if it's a call to os.Exit
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	// Check for os.Exit
	ident, ok := selExpr.X.(*ast.Ident)
	if !ok || ident.Name != "os" || selExpr.Sel.Name != "Exit" {
		return
	}

	// Verify it's the os package, not a variable named 'os'
	obj := pass.TypesInfo.Uses[selExpr.Sel]
	if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != "os" {
		return
	}

	// Find the containing function in the stack
	var containingFn *ast.FuncDecl
	for _, node := range stack {
		if fn, ok := node.(*ast.FuncDecl); ok && fn != nil {
			containingFn = fn
			break
		}
	}

	if containingFn == nil {
		return
	}

	// Only report if it's the main function
	if isMainFunc(containingFn) {
		pass.Reportf(callExpr.Pos(), "direct os.Exit call in main function; use logger.Fatal or return error instead")
	}
}

// isMainFunc checks if the function declaration is the main function.
func isMainFunc(fn *ast.FuncDecl) bool {
	// Safety check
	if fn == nil || fn.Name == nil || fn.Type == nil {
		return false
	}

	// Check function name
	if fn.Name.Name != "main" {
		return false
	}

	// Check it's not a method (receiver is nil)
	if fn.Recv != nil {
		return false
	}

	// Check parameters (main has no parameters)
	if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 {
		return false
	}

	// Check return values (main has no return values)
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		return false
	}

	return true
}

// Ensure os package is imported
var _ = os.Exit
