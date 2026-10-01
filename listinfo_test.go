package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
)

func TestParseRule(t *testing.T) {
	tests := []struct {
		line  string
		kind  router.Domain_Type
		value string
		attrs int
	}{
		{"Example.COM", router.Domain_RootDomain, "example.com", 0},
		{"full:Example.COM", router.Domain_Full, "example.com", 0},
		{"keyword:Example", router.Domain_Plain, "example", 0},
		{`regexp:^(?:Example|Other)\.com$`, router.Domain_Regex, `^(?:Example|Other)\.com$`, 0},
		{"domain:Example.COM\t@CN  @ads", router.Domain_RootDomain, "example.com", 2},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			rule, err := NewListInfo().parseRule(tt.line)
			if err != nil {
				t.Fatal(err)
			}
			if rule.Type != tt.kind || rule.Value != tt.value || len(rule.Attribute) != tt.attrs {
				t.Fatalf("unexpected rule: %v", rule)
			}
			if tt.attrs > 0 && rule.Attribute[0].GetKey() != "cn" {
				t.Fatalf("attribute was not normalized: %v", rule.Attribute)
			}
		})
	}
	for _, line := range []string{"", "domain:", "full:", "regexp:", "unknown:example.com", "example.com @", "example.com cn", "include:", "include: @cn", "include:child @", "include:child @cn @"} {
		t.Run("invalid/"+line, func(t *testing.T) {
			if _, err := NewListInfo().parseRule(line); err == nil {
				t.Fatalf("expected error for %q", line)
			}
		})
	}
	for _, attr := range []string{"", "@", "cn"} {
		if _, err := NewListInfo().parseAttribute(attr); err == nil {
			t.Errorf("expected error for attribute %q", attr)
		}
	}
}

func TestProcessListErrorLocation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "example")
	if err := os.WriteFile(path, []byte("# comment\n\nfull:\n"), 0600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	err = NewListInfo().ProcessList(file)
	if err == nil || !strings.Contains(err.Error(), path+":3:") {
		t.Fatalf("expected filename and line number, got %v", err)
	}
}

func testList(t *testing.T, name fileName, lines ...string) *ListInfo {
	t.Helper()
	list := NewListInfo()
	list.Name = name
	for _, line := range lines {
		rule, err := list.parseRule(line)
		if err != nil {
			t.Fatal(err)
		}
		if rule != nil {
			list.classifyRule(rule)
		}
	}
	return list
}

func TestFlattenInclusions(t *testing.T) {
	for _, includes := range [][]string{
		{"include:child @cn @ads", "include:child @cn"},
		{"include:child @cn", "include:child", "include:child"},
	} {
		t.Run(strings.Join(includes, ","), func(t *testing.T) {
			parent := testList(t, "PARENT", includes...)
			child := testList(t, "CHILD", "example.com @cn @ads", "other.com @cn2")
			lists := ListInfoMap{"PARENT": parent, "CHILD": child}
			if err := lists.FlattenAndGenUniqueDomainList(); err != nil {
				t.Fatal(err)
			}
			want := 1
			if len(includes) == 3 {
				want = 2
			}
			if len(parent.AttributeRuleUniqueList) != want {
				t.Fatalf("included %d rules, want %d", len(parent.AttributeRuleUniqueList), want)
			}
		})
	}
}

func TestToGFWList(t *testing.T) {
	list := testList(t, "EXAMPLE", "full:exact.example", "example.com", "keyword:word", `regexp:^other\.example$`)
	lists := ListInfoMap{"EXAMPLE": list}
	if err := lists.FlattenAndGenUniqueDomainList(); err != nil {
		t.Fatal(err)
	}
	list.ToGeoSite(nil)
	got := string(list.ToGFWList())
	for _, want := range []string{
		"[AutoProxy 0.2.9]\n",
		" CST\n",
		"|http://exact.example^\n",
		"|https://exact.example^\n",
		"||example.com^\n",
		"word\n",
		"/^other\\.example$/\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q from GFWList", want)
		}
	}
}
