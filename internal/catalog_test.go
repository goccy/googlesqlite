package internal

import (
	"testing"

	googlesql "github.com/goccy/go-googlesql"
	"github.com/google/go-cmp/cmp"
)

// TestSubCatalogCarriesNoBuiltinFunctions pins that a sub-catalog is
// created empty, while the root keeps the builtin set. resetCatalog
// rebuilds every sub-catalog on each DROP, so a sub-catalog built with
// the builtins cost one full builtin catalog per project and dataset per
// drop, and grew the wasm heap past 2 GiB under sustained DDL.
func TestSubCatalogCarriesNoBuiltinFunctions(t *testing.T) {
	root := newSimpleCatalog("root")
	if root == nil {
		t.Fatal("newSimpleCatalog returned nil")
	}
	rootNames, err := root.FunctionNames()
	if err != nil {
		t.Fatalf("root FunctionNames: %v", err)
	}
	if len(rootNames) == 0 {
		t.Fatal("the root catalog holds no builtin functions; nothing below would be asserted")
	}

	c := &Catalog{subCatalogs: map[*googlesql.SimpleCatalog]map[string]*googlesql.SimpleCatalog{}}
	for _, path := range [][]string{{"project"}, {"project", "dataset"}} {
		cat := root
		for _, name := range path {
			cat = c.getOrCreateSubCatalog(cat, name)
			if cat == nil {
				t.Fatalf("getOrCreateSubCatalog(%v) returned nil", path)
			}
		}
		names, err := cat.FunctionNames()
		if err != nil {
			t.Fatalf("FunctionNames of %v: %v", path, err)
		}
		if len(names) != 0 {
			t.Errorf("sub-catalog %v holds %d functions, want 0 (the root holds %d)", path, len(names), len(rootNames))
		}
	}
}

func TestNormalizeTablePath(t *testing.T) {
	c := &Catalog{}

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "already normalized path",
			input:    []string{"project", "dataset", "table"},
			expected: []string{"project", "dataset", "table"},
		},
		{
			name:     "single element with dots",
			input:    []string{"project.dataset.table"},
			expected: []string{"project", "dataset", "table"},
		},
		{
			name:     "mixed normalized and dot-separated",
			input:    []string{"project", "dataset.table"},
			expected: []string{"project", "dataset", "table"},
		},
		{
			name:     "two element path",
			input:    []string{"dataset", "table"},
			expected: []string{"dataset", "table"},
		},
		{
			name:     "single element no dots",
			input:    []string{"table"},
			expected: []string{"table"},
		},
		{
			name:     "empty path",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "project with hyphen",
			input:    []string{"my-project.dataset.table"},
			expected: []string{"my-project", "dataset", "table"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.normalizeTablePath(tt.input)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("normalizeTablePath(%v) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}
