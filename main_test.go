package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

func TestParseExcludedAttributes(t *testing.T) {
	got := parseExcludedAttributes(" CN@ADS ,cn@!CN, example@cn@ads")
	for name, attrs := range map[fileName][]attribute{"CN": {"ads", "!cn"}, "EXAMPLE": {"cn", "ads"}} {
		for _, attr := range attrs {
			if !got[name][attr] {
				t.Errorf("missing exclusion %s@%s", name, attr)
			}
		}
	}
	if len(parseExcludedAttributes("")) != 0 {
		t.Fatal("empty exclusions should produce an empty map")
	}
}

func TestRunExports(t *testing.T) {
	oldData, oldName, oldOutput := *dataPath, *datName, *outputPath
	oldExports, oldExcludes, oldGFW := *exportLists, *excludeAttrs, *toGFWList
	t.Cleanup(func() {
		*dataPath, *datName, *outputPath = oldData, oldName, oldOutput
		*exportLists, *excludeAttrs, *toGFWList = oldExports, oldExcludes, oldGFW
	})
	*dataPath, *datName, *outputPath = t.TempDir(), "geosite.dat", filepath.Join(t.TempDir(), "publish")
	*exportLists, *excludeAttrs, *toGFWList = " example ", "example@ADS,example@!CN", " example "
	source := filepath.Join(*dataPath, "example")
	if err := os.WriteFile(source, []byte("keep.example\ndrop.example @ads\nother.example @!cn\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(*outputPath, 0755); err != nil {
		t.Fatal(err)
	}
	gfwPath := filepath.Join(*outputPath, "gfwlist.txt")
	if err := os.WriteFile(gfwPath, []byte(strings.Repeat("stale bytes", 1024)), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(*outputPath, *datName))
	if err != nil {
		t.Fatal(err)
	}
	var sites router.GeoSiteList
	if err := proto.Unmarshal(data, &sites); err != nil {
		t.Fatal(err)
	}
	if len(sites.Entry) != 1 || len(sites.Entry[0].Domain) != 1 || sites.Entry[0].Domain[0].Value != "keep.example" {
		t.Fatalf("unexpected geosite: %v", &sites)
	}
	plaintext, err := os.ReadFile(filepath.Join(*outputPath, "example.txt"))
	if err != nil || string(plaintext) != "domain:keep.example\n" {
		t.Fatalf("unexpected plaintext: %q, %v", plaintext, err)
	}
	encoded, err := os.ReadFile(gfwPath)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		t.Fatalf("invalid base64 output: %v", err)
	}
	if !strings.HasSuffix(string(decoded), "||keep.example^\n") || strings.Contains(string(decoded), "drop.example") {
		t.Fatalf("unexpected GFWList: %q", decoded)
	}

	*toGFWList = ""
	*outputPath = filepath.Join(t.TempDir(), "publish")
	if err := run(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(*outputPath, "gfwlist.txt")); !os.IsNotExist(err) {
		t.Fatalf("disabled GFWList export created a file: %v", err)
	}
	*toGFWList = "missing"
	if err := run(); err == nil {
		t.Fatal("expected an error for a missing GFWList source")
	}
	*dataPath = filepath.Join(t.TempDir(), "missing")
	if err := run(); err == nil {
		t.Fatal("expected an error for a missing data directory")
	}
}
