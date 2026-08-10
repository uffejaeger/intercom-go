package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAffectedGeneratedType(t *testing.T) {
	tests := []struct {
		line string
		want string
		ok   bool
	}{
		{line: "- ConversationSourceSchema.Author: changed from A to B", want: "ConversationSourceSchema", ok: true},
		{line: "- (*ExampleSchema).Validate: removed", want: "ExampleSchema", ok: true},
		{line: "- ExampleSchema: removed", want: "ExampleSchema", ok: true},
		{line: "not an apidiff finding", ok: false},
	}

	for _, test := range tests {
		got, ok := affectedGeneratedType(test.line)
		if got != test.want || ok != test.ok {
			t.Errorf("affectedGeneratedType(%q) = (%q, %t), want (%q, %t)", test.line, got, ok, test.want, test.ok)
		}
	}
}

func TestReachableGeneratedTypes(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "aliases.go", `package example

import gen "example.test/sdk/internal/generated/intercom"

type Public = gen.Root
`)
	writeTestFile(t, root, "internal/generated/intercom/models.go", `package intercom

type Root struct {
	Nested *Nested
}

type Nested struct {
	Value string
}

type MethodResult struct {
	Nested *MethodNested
}

type MethodNested struct {
	Value string
}

type MethodParameter struct {
	Value string
}

type Unreachable struct {
	Ignored bool
}

func (Root) Result() MethodResult {
	return MethodResult{}
}

func (*Root) Accept(MethodParameter) {}
`)

	reachable, err := reachableGeneratedTypes(root)
	if err != nil {
		t.Fatalf("reachableGeneratedTypes: %v", err)
	}
	for _, name := range []string{"Root", "Nested", "MethodResult", "MethodNested", "MethodParameter"} {
		if _, ok := reachable[name]; !ok {
			t.Errorf("%s is not reachable", name)
		}
	}
	if _, ok := reachable["Unreachable"]; ok {
		t.Error("Unreachable unexpectedly reported as reachable")
	}
}

func writeTestFile(t *testing.T, root, name, contents string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
