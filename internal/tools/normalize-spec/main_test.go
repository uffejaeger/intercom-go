package main

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPatchConversationSourceAuthorCompatibility(t *testing.T) {
	var spec yaml.Node
	if err := yaml.Unmarshal([]byte(`
components:
  schemas:
    conversation_source:
      properties:
        author:
          $ref: '#/components/schemas/conversation_source_author'
`), &spec); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}

	patchConversationSourceAuthorCompatibility(&spec)

	authorRef := lookup(
		&spec,
		"components",
		"schemas",
		"conversation_source",
		"properties",
		"author",
		"$ref",
	)
	if got, want := scalarValue(authorRef), "#/components/schemas/conversation_part_author"; got != want {
		t.Fatalf("author ref = %q, want %q", got, want)
	}
}

func TestPatchConversationSourceAuthorCompatibilityLeavesUnexpectedRef(t *testing.T) {
	var spec yaml.Node
	if err := yaml.Unmarshal([]byte(`
components:
  schemas:
    conversation_source:
      properties:
        author:
          $ref: '#/components/schemas/custom_author'
`), &spec); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}

	patchConversationSourceAuthorCompatibility(&spec)

	authorRef := lookup(
		&spec,
		"components",
		"schemas",
		"conversation_source",
		"properties",
		"author",
		"$ref",
	)
	if got, want := scalarValue(authorRef), "#/components/schemas/custom_author"; got != want {
		t.Fatalf("author ref = %q, want %q", got, want)
	}
}
