package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) != 3 {
		_, _ = fmt.Fprintf(os.Stderr, "usage: normalize-spec <input> <output>\n")
		os.Exit(2)
	}

	input, err := os.ReadFile(os.Args[1])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		os.Exit(1)
	}

	var spec yaml.Node
	if err := yaml.Unmarshal(input, &spec); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "parse spec: %v\n", err)
		os.Exit(1)
	}

	patchPathParameters(&spec)
	patchComponentGoNames(&spec)
	patchPropertyGoNames(&spec)

	output, err := yaml.Marshal(&spec)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "marshal spec: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(os.Args[2], output, 0o644); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}

func patchPropertyGoNames(spec *yaml.Node) {
	schemas := lookup(spec, "components", "schemas")
	if schemas == nil || schemas.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(schemas.Content)-1; i += 2 {
		patchPropertyGoNamesInSchema(schemas.Content[i+1])
	}
}

func patchPropertyGoNamesInSchema(schema *yaml.Node) {
	properties := lookup(schema, "properties")
	if properties != nil && properties.Kind == yaml.MappingNode {
		for i := 0; i < len(properties.Content)-1; i += 2 {
			propertyName := properties.Content[i].Value
			property := properties.Content[i+1]
			if property.Kind == yaml.MappingNode && lookup(property, "x-go-name") == nil {
				property.Content = append(property.Content,
					&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "x-go-name"},
					&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: pascal(propertyName)},
				)
			}
			wrapRefSiblingProperty(property)
			patchPropertyGoNamesInSchema(property)
		}
	}

	for _, key := range []string{"items", "additionalProperties"} {
		if child := lookup(schema, key); child != nil {
			patchPropertyGoNamesInSchema(child)
		}
	}

	for _, key := range []string{"allOf", "anyOf", "oneOf"} {
		children := lookup(schema, key)
		if children == nil || children.Kind != yaml.SequenceNode {
			continue
		}
		for _, child := range children.Content {
			patchPropertyGoNamesInSchema(child)
		}
	}
}

func wrapRefSiblingProperty(property *yaml.Node) {
	ref := lookup(property, "$ref")
	if property == nil || property.Kind != yaml.MappingNode || ref == nil || len(property.Content) <= 2 {
		return
	}

	content := make([]*yaml.Node, 0, len(property.Content)+2)
	for i := 0; i < len(property.Content)-1; i += 2 {
		if property.Content[i].Value == "$ref" {
			continue
		}
		content = append(content, property.Content[i], property.Content[i+1])
	}
	content = append(content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "allOf"},
		&yaml.Node{
			Kind: yaml.SequenceNode,
			Tag:  "!!seq",
			Content: []*yaml.Node{
				{
					Kind: yaml.MappingNode,
					Tag:  "!!map",
					Content: []*yaml.Node{
						{Kind: yaml.ScalarNode, Tag: "!!str", Value: "$ref"},
						{Kind: yaml.ScalarNode, Tag: "!!str", Value: ref.Value},
					},
				},
			},
		},
	)
	property.Content = content
}

func patchComponentGoNames(spec *yaml.Node) {
	schemas := lookup(spec, "components", "schemas")
	if schemas == nil || schemas.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(schemas.Content)-1; i += 2 {
		schemaName := schemas.Content[i].Value
		schema := schemas.Content[i+1]
		if schema.Kind != yaml.MappingNode || lookup(schema, "x-go-name") != nil {
			continue
		}
		if schemaType := lookup(schema, "type"); scalarValue(schemaType) != "object" {
			continue
		}
		schema.Content = append(schema.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "x-go-name"},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: pascal(schemaName) + "Schema"},
		)
	}
}

func pascal(input string) string {
	var builder strings.Builder
	upperNext := true
	for _, r := range input {
		if r == '_' || r == '-' || r == ' ' || r == '.' {
			upperNext = true
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			upperNext = true
			continue
		}
		if upperNext {
			builder.WriteRune(unicode.ToUpper(r))
			upperNext = false
			continue
		}
		builder.WriteRune(r)
	}

	result := builder.String()
	if result == "" {
		return "Anonymous"
	}
	if first := rune(result[0]); unicode.IsDigit(first) {
		return "Schema" + result
	}
	return result
}

func patchPathParameters(spec *yaml.Node) {
	paths := lookup(spec, "paths")
	if paths == nil || paths.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(paths.Content)-1; i += 2 {
		path := paths.Content[i].Value
		pathItem := paths.Content[i+1]
		for _, parameterName := range pathParameterNames(path) {
			for _, method := range []string{"get", "put", "post", "delete", "patch"} {
				parameters := lookup(pathItem, method, "parameters")
				if parameters == nil || parameters.Kind != yaml.SequenceNode {
					continue
				}
				for _, parameter := range parameters.Content {
					name := lookup(parameter, "name")
					in := lookup(parameter, "in")
					if scalarValue(name) == parameterName && scalarValue(in) != "path" {
						in.Value = "path"
					}
				}
			}
		}
	}
}

func pathParameterNames(path string) []string {
	var names []string
	for {
		start := strings.Index(path, "{")
		if start == -1 {
			return names
		}
		end := strings.Index(path[start:], "}")
		if end == -1 {
			return names
		}
		names = append(names, path[start+1:start+end])
		path = path[start+end+1:]
	}
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
