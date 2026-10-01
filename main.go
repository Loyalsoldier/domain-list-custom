package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"
)

var (
	dataPath     = flag.String("datapath", "", "Path to your custom 'data' directory (defaults to ./data, then domain-list-community in GOPATH)")
	datName      = flag.String("datname", "geosite.dat", "Name of the generated dat file")
	outputPath   = flag.String("outputpath", "./publish", "Output path to the generated files")
	exportLists  = flag.String("exportlists", "category-ads-all,tld-cn,cn,geolocation-cn,tld-!cn,geolocation-!cn,private,apple,icloud,google,steam", "Lists to be exported in plaintext format, separated by ',' comma")
	excludeAttrs = flag.String("excludeattrs", "cn@!cn@ads,geolocation-cn@!cn@ads,geolocation-!cn@cn@ads", "Exclude rules with certain attributes in certain lists, separated by ',' comma, support multiple attributes in one list. Example: geolocation-!cn@cn@ads,geolocation-cn@!cn")
	toGFWList    = flag.String("togfwlist", "geolocation-!cn", "List to be exported in GFWList format")
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed:", err)
		os.Exit(1)
	}
}

func run() error {
	dir := GetDataDir()
	listInfoMap := make(ListInfoMap)

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if err := listInfoMap.Marshal(path); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if err := listInfoMap.FlattenAndGenUniqueDomainList(); err != nil {
		return err
	}

	// Process and split *excludeAttrs
	excludeAttrsInFile := parseExcludedAttributes(*excludeAttrs)

	// Process and split *exportLists
	var exportListsSlice []string
	if *exportLists != "" {
		tempSlice := strings.Split(*exportLists, ",")
		for _, exportList := range tempSlice {
			exportList = strings.TrimSpace(exportList)
			if len(exportList) > 0 {
				exportListsSlice = append(exportListsSlice, exportList)
			}
		}
	}

	// Generate geosite.dat
	protoBytes, err := proto.Marshal(listInfoMap.ToProto(excludeAttrsInFile))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*outputPath, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*outputPath, *datName), protoBytes, 0644); err != nil {
		return err
	}
	fmt.Printf("%s has been generated successfully in '%s'.\n", *datName, *outputPath)

	// Generate plaintext list files
	filePlainTextBytesMap, err := listInfoMap.ToPlainText(exportListsSlice)
	if err != nil {
		return err
	}
	for filename, plaintextBytes := range filePlainTextBytesMap {
		filename += ".txt"
		if err := os.WriteFile(filepath.Join(*outputPath, filename), plaintextBytes, 0644); err != nil {
			return err
		}
		fmt.Printf("%s has been generated successfully in '%s'.\n", filename, *outputPath)
	}

	// Generate gfwlist.txt
	if listName := strings.TrimSpace(*toGFWList); listName != "" {
		gfwlistBytes, err := listInfoMap.ToGFWList(listName)
		if err != nil {
			return err
		}
		encoded := base64.StdEncoding.EncodeToString(gfwlistBytes)
		if err := os.WriteFile(filepath.Join(*outputPath, "gfwlist.txt"), []byte(encoded), 0644); err != nil {
			return err
		}
		fmt.Printf("gfwlist.txt has been generated successfully in '%s'.\n", *outputPath)
	}
	return nil
}

func parseExcludedAttributes(value string) map[fileName]map[attribute]bool {
	excludeAttrsInFile := make(map[fileName]map[attribute]bool)
	if value != "" {
		exFilenameAttrSlice := strings.Split(value, ",")
		for _, exFilenameAttr := range exFilenameAttrSlice {
			exFilenameAttr = strings.TrimSpace(exFilenameAttr)
			exFilenameAttrMap := strings.Split(exFilenameAttr, "@")
			filename := fileName(strings.ToUpper(strings.TrimSpace(exFilenameAttrMap[0])))
			if excludeAttrsInFile[filename] == nil {
				excludeAttrsInFile[filename] = make(map[attribute]bool)
			}
			for _, attr := range exFilenameAttrMap[1:] {
				attr = strings.ToLower(strings.TrimSpace(attr))
				if len(attr) > 0 {
					excludeAttrsInFile[filename][attribute(attr)] = true
				}
			}
		}
	}

	return excludeAttrsInFile
}
