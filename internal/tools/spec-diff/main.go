package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var httpMethods = map[string]bool{
	"get":     true,
	"put":     true,
	"post":    true,
	"delete":  true,
	"options": true,
	"head":    true,
	"patch":   true,
	"trace":   true,
}

type specSummary struct {
	Operations map[string]operationSummary
	Schemas    map[string]string
}

type operationSummary struct {
	Method      string
	Path        string
	OperationID string
	Summary     string
	Description string
	Responses   map[string]bool
}

func main() {
	var outputPath string
	flag.StringVar(&outputPath, "out", "", "optional path to write markdown summary")
	flag.Parse()

	if flag.NArg() != 2 {
		_, _ = fmt.Fprintf(os.Stderr, "usage: spec-diff [-out summary.md] <old-spec.yaml> <new-spec.yaml>\n")
		os.Exit(2)
	}

	summary, err := diffSpecs(flag.Arg(0), flag.Arg(1))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "spec-diff: %v\n", err)
		os.Exit(1)
	}

	if outputPath == "" {
		fmt.Print(summary)
		return
	}
	if err := os.WriteFile(outputPath, []byte(summary), 0o644); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "spec-diff: write summary: %v\n", err)
		os.Exit(1)
	}
}

func diffSpecs(oldPath, newPath string) (string, error) {
	oldSpec, err := readSpec(oldPath)
	if err != nil {
		return "", err
	}
	newSpec, err := readSpec(newPath)
	if err != nil {
		return "", err
	}

	var out strings.Builder
	out.WriteString("## OpenAPI diff summary\n\n")

	removedOperations := removedKeys(oldSpec.Operations, newSpec.Operations)
	removedSchemas := removedKeys(oldSpec.Schemas, newSpec.Schemas)
	removedResponses := responseChanges(oldSpec.Operations, newSpec.Operations, false)
	writeSection(&out, "Breaking candidates", append(append(formatList("Removed operation", removedOperations), formatList("Removed schema", removedSchemas)...), removedResponses...))

	addedOperations := addedKeys(oldSpec.Operations, newSpec.Operations)
	addedSchemas := addedKeys(oldSpec.Schemas, newSpec.Schemas)
	addedResponses := responseChanges(oldSpec.Operations, newSpec.Operations, true)
	writeSection(&out, "Additive changes", append(append(formatList("Added operation", addedOperations), formatList("Added schema", addedSchemas)...), addedResponses...))

	writeSection(&out, "Documentation-only candidates", documentationChanges(oldSpec.Operations, newSpec.Operations))
	writeSection(&out, "Other schema changes", changedSchemas(oldSpec.Schemas, newSpec.Schemas))

	out.WriteString("Generated code freshness is still enforced separately by `make generate-check`.\n")
	return out.String(), nil
}

func readSpec(path string) (specSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return specSummary{}, fmt.Errorf("read %s: %w", path, err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return specSummary{}, fmt.Errorf("parse %s: %w", path, err)
	}

	return specSummary{
		Operations: operationsFromSpec(&root),
		Schemas:    schemasFromSpec(&root),
	}, nil
}

func operationsFromSpec(spec *yaml.Node) map[string]operationSummary {
	paths := lookup(spec, "paths")
	operations := map[string]operationSummary{}
	if paths == nil || paths.Kind != yaml.MappingNode {
		return operations
	}

	for i := 0; i < len(paths.Content)-1; i += 2 {
		path := paths.Content[i].Value
		pathItem := paths.Content[i+1]
		if pathItem.Kind != yaml.MappingNode {
			continue
		}
		for j := 0; j < len(pathItem.Content)-1; j += 2 {
			method := pathItem.Content[j].Value
			if !httpMethods[method] {
				continue
			}
			operation := pathItem.Content[j+1]
			id := scalarValue(lookup(operation, "operationId"))
			key := method + " " + path
			if id != "" {
				key = id
			}
			operations[key] = operationSummary{
				Method:      strings.ToUpper(method),
				Path:        path,
				OperationID: id,
				Summary:     scalarValue(lookup(operation, "summary")),
				Description: scalarValue(lookup(operation, "description")),
				Responses:   responseCodes(operation),
			}
		}
	}
	return operations
}

func schemasFromSpec(spec *yaml.Node) map[string]string {
	schemas := lookup(spec, "components", "schemas")
	result := map[string]string{}
	if schemas == nil || schemas.Kind != yaml.MappingNode {
		return result
	}

	for i := 0; i < len(schemas.Content)-1; i += 2 {
		var buffer bytes.Buffer
		encoder := yaml.NewEncoder(&buffer)
		encoder.SetIndent(2)
		_ = encoder.Encode(schemas.Content[i+1])
		_ = encoder.Close()
		result[schemas.Content[i].Value] = buffer.String()
	}
	return result
}

func responseCodes(operation *yaml.Node) map[string]bool {
	responses := lookup(operation, "responses")
	result := map[string]bool{}
	if responses == nil || responses.Kind != yaml.MappingNode {
		return result
	}
	for i := 0; i < len(responses.Content)-1; i += 2 {
		result[responses.Content[i].Value] = true
	}
	return result
}

func addedKeys[T any](oldItems, newItems map[string]T) []string {
	var added []string
	for key := range newItems {
		if _, ok := oldItems[key]; !ok {
			added = append(added, key)
		}
	}
	sort.Strings(added)
	return added
}

func removedKeys[T any](oldItems, newItems map[string]T) []string {
	var removed []string
	for key := range oldItems {
		if _, ok := newItems[key]; !ok {
			removed = append(removed, key)
		}
	}
	sort.Strings(removed)
	return removed
}

func responseChanges(oldOperations, newOperations map[string]operationSummary, added bool) []string {
	var changes []string
	for key, oldOperation := range oldOperations {
		newOperation, ok := newOperations[key]
		if !ok {
			continue
		}
		var codes []string
		if added {
			codes = addedKeys(oldOperation.Responses, newOperation.Responses)
		} else {
			codes = removedKeys(oldOperation.Responses, newOperation.Responses)
		}
		for _, code := range codes {
			if added {
				changes = append(changes, fmt.Sprintf("Added response `%s` on `%s`", code, operationLabel(newOperation)))
			} else {
				changes = append(changes, fmt.Sprintf("Removed response `%s` from `%s`", code, operationLabel(oldOperation)))
			}
		}
	}
	sort.Strings(changes)
	return changes
}

func documentationChanges(oldOperations, newOperations map[string]operationSummary) []string {
	var changes []string
	for key, oldOperation := range oldOperations {
		newOperation, ok := newOperations[key]
		if !ok {
			continue
		}
		if oldOperation.Summary != newOperation.Summary || oldOperation.Description != newOperation.Description {
			changes = append(changes, "Changed operation docs for `"+operationLabel(newOperation)+"`")
		}
	}
	sort.Strings(changes)
	return changes
}

func changedSchemas(oldSchemas, newSchemas map[string]string) []string {
	var changes []string
	for key, oldSchema := range oldSchemas {
		newSchema, ok := newSchemas[key]
		if !ok {
			continue
		}
		if oldSchema != newSchema {
			changes = append(changes, "Changed schema `"+key+"`")
		}
	}
	sort.Strings(changes)
	return changes
}

func formatList(prefix string, values []string) []string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, prefix+" `"+value+"`")
	}
	return items
}

func writeSection(out *strings.Builder, title string, items []string) {
	out.WriteString("### " + title + "\n\n")
	if len(items) == 0 {
		out.WriteString("- None detected.\n\n")
		return
	}
	for _, item := range items {
		out.WriteString("- " + item + "\n")
	}
	out.WriteString("\n")
}

func operationLabel(operation operationSummary) string {
	if operation.OperationID != "" {
		return operation.OperationID
	}
	return operation.Method + " " + operation.Path
}

func lookup(node *yaml.Node, path ...string) *yaml.Node {
	current := node
	if current.Kind == yaml.DocumentNode && len(current.Content) == 1 {
		current = current.Content[0]
	}

	for _, key := range path {
		if current == nil || current.Kind != yaml.MappingNode {
			return nil
		}

		var next *yaml.Node
		for i := 0; i < len(current.Content)-1; i += 2 {
			if current.Content[i].Value == key {
				next = current.Content[i+1]
				break
			}
		}
		current = next
	}

	return current
}

func scalarValue(node *yaml.Node) string {
	if node == nil || node.Kind != yaml.ScalarNode {
		return ""
	}
	return node.Value
}
