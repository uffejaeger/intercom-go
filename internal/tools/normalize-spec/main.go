package main

import (
	"fmt"
	"os"
	"strings"

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
