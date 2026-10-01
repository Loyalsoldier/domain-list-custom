package main

import "testing"

func TestDomainTrieInsert(t *testing.T) {
	trie := NewDomainTrie()
	for _, tt := range []struct {
		domain string
		want   bool
	}{
		{"child.example.com", true},
		{"example.com", true},
		{"example.com", false},
		{"other.example.com", false},
		{"child.example.com", false},
		{"other.com", true},
		{"com", true},
		{"another.com", false},
		{"example.net", true},
	} {
		got, err := trie.Insert(tt.domain)
		if err != nil || got != tt.want {
			t.Errorf("Insert(%q) = %v, %v; want %v", tt.domain, got, err, tt.want)
		}
	}
	if _, err := trie.Insert(""); err == nil {
		t.Fatal("expected error for empty domain")
	}
}
