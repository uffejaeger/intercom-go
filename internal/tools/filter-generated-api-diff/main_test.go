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

type PublicStruct struct {
	Direct gen.Direct
	ViaLocal localHelper
	hidden gen.HiddenField
}

type localHelper struct {
	Exposed gen.ViaLocal
	hidden gen.HiddenLocalField
}

type Service struct{}

func (*Service) PublicMethod(gen.PublicMethodParameter) gen.PublicMethodResult {
	return gen.PublicMethodResult{}
}

func (*Service) privateMethod(gen.PrivateMethodParameter) {}

func PublicFunction(gen.PublicFunctionParameter) gen.PublicFunctionResult {
	return gen.PublicFunctionResult{}
}

func privateFunction(gen.PrivateFunctionParameter) {}

type privateStruct struct {
	Exposed gen.PrivateStructField
}

type inferredLocalResult struct{}

func (inferredLocalResult) PublicMethod() gen.InferredMethodResult {
	return gen.InferredMethodResult{}
}

func PublicLocalResult() inferredLocalResult {
	return inferredLocalResult{}
}

var PublicVariable gen.PublicVariable
const PublicConstant gen.PublicConstant = "public"
var InferredPublicVariable = gen.InferredPublicVariable{}
const InferredPublicConstant = gen.InferredPublicConstant("public")
const InferredPublicConstantFromValue = gen.InferredPublicConstantValue
var InferredPublicConstructed = gen.NewInferredPublicConstructed()
var privateInferredVariable = gen.PrivateInferredVariable{}
const privateInferredConstant = gen.PrivateInferredConstant("private")
const privateInferredConstantFromValue = gen.PrivateInferredConstantValue
var privateInferredConstructed = gen.NewPrivateInferredConstructed()
var InferredPublicMixed, privateInferredMixed = gen.InferredPublicMixed{}, gen.PrivateInferredMixed{}
var privateVariable gen.PrivateVariable
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

type Direct struct{}
type ViaLocal struct{}
type HiddenField struct{}
type HiddenLocalField struct{}
type PublicMethodParameter struct{}
type PublicMethodResult struct{}
type PrivateMethodParameter struct{}
type PublicFunctionParameter struct{}
type PublicFunctionResult struct{}
type PrivateFunctionParameter struct{}
type PrivateStructField struct{}
type InferredMethodResult struct{}
type PublicVariable struct{}
type PublicConstant string
type InferredPublicVariable struct{}
type InferredPublicConstant string
const InferredPublicConstantValue InferredPublicConstant = "public"
type InferredPublicConstructed struct{}
type PrivateInferredVariable struct{}
type PrivateInferredConstant string
const PrivateInferredConstantValue PrivateInferredConstant = "private"
type PrivateInferredConstructed struct{}
type InferredPublicMixed struct{}
type PrivateInferredMixed struct{}
type PrivateVariable struct{}

type Unreachable struct {
	Ignored bool
}

func (Root) Result() MethodResult {
	return MethodResult{}
}

func (*Root) Accept(MethodParameter) {}

func NewInferredPublicConstructed() InferredPublicConstructed {
	return InferredPublicConstructed{}
}

func NewPrivateInferredConstructed() PrivateInferredConstructed {
	return PrivateInferredConstructed{}
}
`)

	reachable, err := reachableGeneratedTypes(root)
	if err != nil {
		t.Fatalf("reachableGeneratedTypes: %v", err)
	}
	for _, name := range []string{
		"Root",
		"Nested",
		"MethodResult",
		"MethodNested",
		"MethodParameter",
		"Direct",
		"ViaLocal",
		"PublicMethodParameter",
		"PublicMethodResult",
		"PublicFunctionParameter",
		"PublicFunctionResult",
		"InferredMethodResult",
		"PublicVariable",
		"PublicConstant",
		"InferredPublicVariable",
		"InferredPublicConstant",
		"InferredPublicConstructed",
		"InferredPublicMixed",
	} {
		if _, ok := reachable[name]; !ok {
			t.Errorf("%s is not reachable", name)
		}
	}
	for _, name := range []string{
		"HiddenField",
		"HiddenLocalField",
		"PrivateMethodParameter",
		"PrivateFunctionParameter",
		"PrivateStructField",
		"PrivateInferredVariable",
		"PrivateInferredConstant",
		"PrivateInferredConstructed",
		"PrivateInferredMixed",
		"PrivateVariable",
		"Unreachable",
	} {
		if _, ok := reachable[name]; ok {
			t.Errorf("%s unexpectedly reported as reachable", name)
		}
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
