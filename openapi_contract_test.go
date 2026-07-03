package intercom

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestGeneratedOperationsAreAccountedFor(t *testing.T) {
	generatedOperations := generatedResponseOperations(t)
	wrappedOperations := wrappedGeneratedOperations(t)
	exceptions := map[string]string{
		"LisDataEvents": "DataEvents.List uses Client.NewRequest and Client.Do for explicit identifier validation and query encoding.",
	}

	var missing []string
	for operation := range generatedOperations {
		if wrappedOperations[operation] || exceptions[operation] != "" {
			continue
		}
		missing = append(missing, operation)
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("generated operations missing public wrapper or explicit exception:\n%s", strings.Join(missing, "\n"))
	}

	var staleExceptions []string
	for operation := range exceptions {
		if !generatedOperations[operation] {
			staleExceptions = append(staleExceptions, operation)
		}
	}
	sort.Strings(staleExceptions)
	if len(staleExceptions) > 0 {
		t.Fatalf("contract exceptions no longer match generated operations:\n%s", strings.Join(staleExceptions, "\n"))
	}
}

func generatedResponseOperations(t *testing.T) map[string]bool {
	t.Helper()

	file := parseGoFile(t, filepath.Join("internal", "generated", "intercom", "client.gen.go"))
	operations := map[string]bool{}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "ClientWithResponsesInterface" {
				continue
			}
			iface, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				t.Fatalf("ClientWithResponsesInterface is %T, want *ast.InterfaceType", typeSpec.Type)
			}
			for _, field := range iface.Methods.List {
				for _, name := range field.Names {
					if operation, ok := canonicalGeneratedOperation(name.Name); ok {
						operations[operation] = true
					}
				}
			}
		}
	}

	if len(operations) == 0 {
		t.Fatal("no generated response operations found")
	}
	return operations
}

func wrappedGeneratedOperations(t *testing.T) map[string]bool {
	t.Helper()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob Go files: %v", err)
	}

	operations := map[string]bool{}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		file := parseGoFile(t, path)
		ast.Inspect(file, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if !isGeneratedClientSelector(selector.X) {
				return true
			}
			operation, ok := canonicalGeneratedOperation(selector.Sel.Name)
			if !ok {
				return true
			}
			operations[operation] = true
			return true
		})
	}

	if len(operations) == 0 {
		t.Fatal("no wrapped generated operations found")
	}
	return operations
}

func parseGoFile(t *testing.T, path string) *ast.File {
	t.Helper()

	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return file
}

func isGeneratedClientSelector(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "generated"
}

func canonicalGeneratedOperation(method string) (string, bool) {
	operation, ok := strings.CutSuffix(method, "WithResponse")
	if !ok {
		return "", false
	}
	operation, _ = strings.CutSuffix(operation, "WithBody")
	if operation == "" {
		return "", false
	}
	return operation, true
}
