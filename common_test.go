package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetRuntimeEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), "env")
	t.Setenv("GOENV", path)
	if err := os.WriteFile(path, []byte("GOPATH\nOTHER=value\nGOPATH=/path=with=equals\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := GetRuntimeEnv("GOPATH")
	if err != nil || got != "/path=with=equals" {
		t.Fatalf("GetRuntimeEnv() = %q, %v", got, err)
	}
	got, err = GetRuntimeEnv("MISSING")
	if err != nil || got != "" {
		t.Fatalf("GetRuntimeEnv(MISSING) = %q, %v", got, err)
	}
	t.Setenv("GOENV", "off")
	if _, err := GetRuntimeEnv("GOPATH"); err == nil {
		t.Fatal("expected an error with GOENV=off")
	}
}

func TestGetDataDir(t *testing.T) {
	t.Chdir(t.TempDir())
	old := *dataPath
	t.Cleanup(func() { *dataPath = old })
	*dataPath = "custom"
	if got := GetDataDir(); got != "custom" {
		t.Fatalf("explicit path = %q", got)
	}

	*dataPath = ""
	first, second := t.TempDir(), t.TempDir()
	t.Setenv("GOPATH", first+string(os.PathListSeparator)+second)
	want := filepath.Join(first, "src", "github.com", "v2fly", "domain-list-community", "data")
	if got := GetDataDir(); got != want {
		t.Fatalf("fallback path = %q, want %q", got, want)
	}
	if err := os.Mkdir("data", 0755); err != nil {
		t.Fatal(err)
	}
	if got := GetDataDir(); got != "data" {
		t.Fatalf("local path = %q", got)
	}
}
