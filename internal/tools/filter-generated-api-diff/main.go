// Command filter-generated-api-diff keeps apidiff findings for generated
// model types that are reachable through aliases in the public root package.
package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const generatedImportSuffix = "/internal/generated/intercom"

func main() {
	if len(os.Args) != 2 {
		_, _ = fmt.Fprintln(os.Stderr, "usage: filter-generated-api-diff <module-root>")
		os.Exit(2)
	}

	reachable, err := reachableGeneratedTypes(os.Args[1])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "find reachable generated types: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if affectedType, ok := affectedGeneratedType(line); ok {
			if _, reachable := reachable[affectedType]; reachable {
				fmt.Println(line)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "read apidiff report: %v\n", err)
		os.Exit(1)
	}
}

func reachableGeneratedTypes(moduleRoot string) (map[string]struct{}, error) {
	targets, err := generatedPublicTargets(moduleRoot)
	if err != nil {
		return nil, err
	}

	typeExpressions, err := generatedTypeExpressions(filepath.Join(moduleRoot, "internal", "generated", "intercom"))
	if err != nil {
		return nil, err
	}

	reachable := make(map[string]struct{}, len(targets))
	queue := make([]string, 0, len(targets))
	for target := range targets {
		if _, exists := typeExpressions[target]; exists {
			queue = append(queue, target)
		}
	}

	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if _, seen := reachable[name]; seen {
			continue
		}
		reachable[name] = struct{}{}

		dependencies := make(map[string]struct{})
		for _, expression := range typeExpressions[name] {
			collectTypeDependencies(expression, dependencies)
		}
		for dependency := range dependencies {
			if _, exists := typeExpressions[dependency]; exists {
				queue = append(queue, dependency)
			}
		}
	}

	return reachable, nil
}

type rootExpression struct {
	expression       ast.Expr
	generatedAliases map[string]struct{}
}

type rootReferences struct {
	generated map[string]struct{}
	local     map[string]struct{}
}

func generatedPublicTargets(moduleRoot string) (map[string]struct{}, error) {
	rootTypes := make(map[string][]rootExpression)
	publicTypes := make(map[string]struct{})
	var publicExpressions []rootExpression

	entries, err := os.ReadDir(moduleRoot)
	if err != nil {
		return nil, err
	}

	fileset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fileset, filepath.Join(moduleRoot, name), nil, 0)
		if err != nil {
			return nil, err
		}

		generatedAliases := make(map[string]struct{})
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil || !strings.HasSuffix(importPath, generatedImportSuffix) {
				continue
			}
			alias := filepath.Base(importPath)
			if spec.Name != nil {
				alias = spec.Name.Name
			}
			generatedAliases[alias] = struct{}{}
		}

		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.GenDecl:
				switch declaration.Tok {
				case token.TYPE:
					for _, rawSpec := range declaration.Specs {
						typeSpec, ok := rawSpec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						rootTypes[typeSpec.Name.Name] = append(rootTypes[typeSpec.Name.Name], rootExpression{
							expression:       typeSpec.Type,
							generatedAliases: generatedAliases,
						})
						if ast.IsExported(typeSpec.Name.Name) {
							publicTypes[typeSpec.Name.Name] = struct{}{}
						}
					}
				case token.CONST, token.VAR:
					for _, rawSpec := range declaration.Specs {
						valueSpec, ok := rawSpec.(*ast.ValueSpec)
						if !ok || valueSpec.Type == nil || !valueSpecIsExported(valueSpec) {
							continue
						}
						publicExpressions = append(publicExpressions, rootExpression{
							expression:       valueSpec.Type,
							generatedAliases: generatedAliases,
						})
					}
				}
			case *ast.FuncDecl:
				if !ast.IsExported(declaration.Name.Name) {
					continue
				}
				expression := rootExpression{
					expression:       declaration.Type,
					generatedAliases: generatedAliases,
				}
				if receiver, ok := receiverTypeName(declaration.Recv); ok {
					rootTypes[receiver] = append(rootTypes[receiver], expression)
				} else {
					publicExpressions = append(publicExpressions, expression)
				}
			}
		}
	}

	targets := make(map[string]struct{})
	queue := make([]string, 0, len(publicTypes))
	for name := range publicTypes {
		queue = append(queue, name)
	}
	for _, expression := range publicExpressions {
		references := collectRootReferences(expression)
		for generated := range references.generated {
			targets[generated] = struct{}{}
		}
		for local := range references.local {
			if _, exists := rootTypes[local]; exists {
				queue = append(queue, local)
			}
		}
	}

	visited := make(map[string]struct{})
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if _, seen := visited[name]; seen {
			continue
		}
		visited[name] = struct{}{}

		for _, expression := range rootTypes[name] {
			references := collectRootReferences(expression)
			for generated := range references.generated {
				targets[generated] = struct{}{}
			}
			for local := range references.local {
				if _, exists := rootTypes[local]; exists {
					queue = append(queue, local)
				}
			}
		}
	}

	return targets, nil
}

func collectRootReferences(root rootExpression) rootReferences {
	references := rootReferences{
		generated: make(map[string]struct{}),
		local:     make(map[string]struct{}),
	}
	collectRootExpressionReferences(root.expression, root.generatedAliases, references)
	return references
}

func collectRootExpressionReferences(expression ast.Expr, generatedAliases map[string]struct{}, references rootReferences) {
	switch expression := expression.(type) {
	case *ast.Ident:
		references.local[expression.Name] = struct{}{}
	case *ast.SelectorExpr:
		packageName, ok := expression.X.(*ast.Ident)
		if ok {
			if _, generated := generatedAliases[packageName.Name]; generated {
				references.generated[expression.Sel.Name] = struct{}{}
			}
		}
	case *ast.ParenExpr:
		collectRootExpressionReferences(expression.X, generatedAliases, references)
	case *ast.StarExpr:
		collectRootExpressionReferences(expression.X, generatedAliases, references)
	case *ast.ArrayType:
		collectRootExpressionReferences(expression.Elt, generatedAliases, references)
	case *ast.MapType:
		collectRootExpressionReferences(expression.Key, generatedAliases, references)
		collectRootExpressionReferences(expression.Value, generatedAliases, references)
	case *ast.ChanType:
		collectRootExpressionReferences(expression.Value, generatedAliases, references)
	case *ast.Ellipsis:
		collectRootExpressionReferences(expression.Elt, generatedAliases, references)
	case *ast.IndexExpr:
		collectRootExpressionReferences(expression.X, generatedAliases, references)
		collectRootExpressionReferences(expression.Index, generatedAliases, references)
	case *ast.IndexListExpr:
		collectRootExpressionReferences(expression.X, generatedAliases, references)
		for _, index := range expression.Indices {
			collectRootExpressionReferences(index, generatedAliases, references)
		}
	case *ast.StructType:
		for _, field := range expression.Fields.List {
			if fieldIsExported(field) {
				collectRootExpressionReferences(field.Type, generatedAliases, references)
			}
		}
	case *ast.InterfaceType:
		for _, method := range expression.Methods.List {
			if fieldIsExported(method) {
				collectRootExpressionReferences(method.Type, generatedAliases, references)
			}
		}
	case *ast.FuncType:
		collectRootFieldListReferences(expression.TypeParams, generatedAliases, references)
		collectRootFieldListReferences(expression.Params, generatedAliases, references)
		collectRootFieldListReferences(expression.Results, generatedAliases, references)
	}
}

func collectRootFieldListReferences(fields *ast.FieldList, generatedAliases map[string]struct{}, references rootReferences) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		collectRootExpressionReferences(field.Type, generatedAliases, references)
	}
}

func fieldIsExported(field *ast.Field) bool {
	if len(field.Names) == 0 {
		return true
	}
	for _, name := range field.Names {
		if ast.IsExported(name.Name) {
			return true
		}
	}
	return false
}

func valueSpecIsExported(spec *ast.ValueSpec) bool {
	for _, name := range spec.Names {
		if ast.IsExported(name.Name) {
			return true
		}
	}
	return false
}

func generatedTypeExpressions(directory string) (map[string][]ast.Expr, error) {
	typeExpressions := make(map[string][]ast.Expr)
	fileset := token.NewFileSet()
	packages, err := parser.ParseDir(fileset, directory, func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}

	for _, pkg := range packages {
		for _, file := range pkg.Files {
			for _, declaration := range file.Decls {
				switch declaration := declaration.(type) {
				case *ast.GenDecl:
					if declaration.Tok != token.TYPE {
						continue
					}
					for _, rawSpec := range declaration.Specs {
						typeSpec, ok := rawSpec.(*ast.TypeSpec)
						if ok {
							typeExpressions[typeSpec.Name.Name] = append(typeExpressions[typeSpec.Name.Name], typeSpec.Type)
						}
					}
				case *ast.FuncDecl:
					receiver, ok := receiverTypeName(declaration.Recv)
					if ok {
						typeExpressions[receiver] = append(typeExpressions[receiver], declaration.Type)
					}
				}
			}
		}
	}

	return typeExpressions, nil
}

func receiverTypeName(receiver *ast.FieldList) (string, bool) {
	if receiver == nil || len(receiver.List) != 1 {
		return "", false
	}

	expression := receiver.List[0].Type
	for {
		switch typed := expression.(type) {
		case *ast.Ident:
			return typed.Name, true
		case *ast.ParenExpr:
			expression = typed.X
		case *ast.StarExpr:
			expression = typed.X
		case *ast.IndexExpr:
			expression = typed.X
		case *ast.IndexListExpr:
			expression = typed.X
		default:
			return "", false
		}
	}
}

func collectTypeDependencies(expression ast.Expr, dependencies map[string]struct{}) {
	switch expression := expression.(type) {
	case *ast.Ident:
		dependencies[expression.Name] = struct{}{}
	case *ast.ParenExpr:
		collectTypeDependencies(expression.X, dependencies)
	case *ast.StarExpr:
		collectTypeDependencies(expression.X, dependencies)
	case *ast.ArrayType:
		collectTypeDependencies(expression.Elt, dependencies)
	case *ast.MapType:
		collectTypeDependencies(expression.Key, dependencies)
		collectTypeDependencies(expression.Value, dependencies)
	case *ast.ChanType:
		collectTypeDependencies(expression.Value, dependencies)
	case *ast.Ellipsis:
		collectTypeDependencies(expression.Elt, dependencies)
	case *ast.IndexExpr:
		collectTypeDependencies(expression.X, dependencies)
		collectTypeDependencies(expression.Index, dependencies)
	case *ast.IndexListExpr:
		collectTypeDependencies(expression.X, dependencies)
		for _, index := range expression.Indices {
			collectTypeDependencies(index, dependencies)
		}
	case *ast.StructType:
		collectFieldListDependencies(expression.Fields, dependencies)
	case *ast.InterfaceType:
		collectFieldListDependencies(expression.Methods, dependencies)
	case *ast.FuncType:
		collectFieldListDependencies(expression.TypeParams, dependencies)
		collectFieldListDependencies(expression.Params, dependencies)
		collectFieldListDependencies(expression.Results, dependencies)
	}
}

func collectFieldListDependencies(fields *ast.FieldList, dependencies map[string]struct{}) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		collectTypeDependencies(field.Type, dependencies)
	}
}

func affectedGeneratedType(line string) (string, bool) {
	symbol, _, ok := strings.Cut(strings.TrimPrefix(line, "- "), ":")
	if !ok {
		return "", false
	}

	if strings.HasPrefix(symbol, "(*") {
		end := strings.Index(symbol, ").")
		if end == -1 {
			return "", false
		}
		return symbol[2:end], true
	}
	if strings.HasPrefix(symbol, "(") {
		end := strings.Index(symbol, ").")
		if end == -1 {
			return "", false
		}
		return symbol[1:end], true
	}
	if before, _, ok0 := strings.Cut(symbol, "."); ok0 {
		return before, true
	}
	return symbol, symbol != ""
}
