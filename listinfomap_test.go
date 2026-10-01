package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFlattenInvalidInclusions(t *testing.T) {
	tests := []struct {
		name  string
		lists map[fileName][]string
		err   string
	}{
		{"missing", map[fileName][]string{"A": {"include:missing"}}, "missing list MISSING"},
		{"self cycle", map[fileName][]string{"A": {"include:a"}}, "cyclic"},
		{"mutual cycle", map[fileName][]string{"A": {"include:b"}, "B": {"include:a"}}, "cyclic"},
		{"cycle with leaf", map[fileName][]string{"A": {"include:b"}, "B": {"include:a"}, "C": {"example.com"}}, "cyclic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lists := make(ListInfoMap)
			for name, lines := range tt.lists {
				lists[name] = testList(t, name, lines...)
			}
			if err := lists.FlattenAndGenUniqueDomainList(); err == nil || !strings.Contains(err.Error(), tt.err) {
				t.Fatalf("expected %q, got %v", tt.err, err)
			}
		})
	}
}

func TestFlattenNestedInclusions(t *testing.T) {
	lists := ListInfoMap{
		"A": testList(t, "A", "include:b", "sub.example.com"),
		"B": testList(t, "B", "include:c"),
		"C": testList(t, "C", "example.com"),
	}
	if err := lists.FlattenAndGenUniqueDomainList(); err != nil {
		t.Fatal(err)
	}
	rules := lists["A"].DomainTypeUniqueList
	if len(rules) != 1 || rules[0].Value != "example.com" {
		t.Fatalf("unexpected flattened rules: %v", rules)
	}
}

func TestMarshalDuplicateList(t *testing.T) {
	lists := make(ListInfoMap)
	for i := 0; i < 2; i++ {
		path := filepath.Join(t.TempDir(), "same")
		if err := os.WriteFile(path, []byte("example.com\n"), 0600); err != nil {
			t.Fatal(err)
		}
		err := lists.Marshal(path)
		if i == 0 && err != nil {
			t.Fatal(err)
		}
		if i == 1 && (err == nil || !strings.Contains(err.Error(), "duplicate list name")) {
			t.Fatalf("expected duplicate name error, got %v", err)
		}
	}
}
