package v2ray

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Loyalsoldier/domain-list-custom/lib"
	router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

const (
	TypeGeositeOut = "v2rayGeoSite"
	DescGeositeOut = "Convert domain lists to V2Ray geosite format"
)

func init() {
	lib.RegisterOutputConfigCreator(TypeGeositeOut, func(action lib.Action, data json.RawMessage) (lib.OutputConverter, error) {
		return newGeositeOut(action, data)
	})
	lib.RegisterOutputConverter(TypeGeositeOut, &GeositeOut{
		Description: DescGeositeOut,
	})
}

type GeositeOut struct {
	Type            string
	Action          lib.Action
	Description     string
	OutputDir       string
	OutputName      string
	Want            []string
	Exclude         []string
	ExcludeAttrs    map[string]map[string]bool
	GFWListOutput   string
}

func newGeositeOut(action lib.Action, data json.RawMessage) (lib.OutputConverter, error) {
	var tmp struct {
		OutputDir     string   `json:"outputDir"`
		OutputName    string   `json:"outputName"`
		Want          []string `json:"wantedList"`
		Exclude       []string `json:"excludedList"`
		ExcludeAttrs  string   `json:"excludeAttrs"`
		GFWListOutput string   `json:"gfwlistOutput"`
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &tmp); err != nil {
			return nil, err
		}
	}

	if tmp.OutputDir == "" {
		tmp.OutputDir = "./output"
	}

	if tmp.OutputName == "" {
		tmp.OutputName = "geosite.dat"
	}

	// Process exclude attributes
	excludeAttrsMap := make(map[string]map[string]bool)
	if tmp.ExcludeAttrs != "" {
		exFilenameAttrSlice := strings.Split(tmp.ExcludeAttrs, ",")
		for _, exFilenameAttr := range exFilenameAttrSlice {
			exFilenameAttr = strings.TrimSpace(exFilenameAttr)
			exFilenameAttrMap := strings.Split(exFilenameAttr, "@")
			filename := strings.ToUpper(strings.TrimSpace(exFilenameAttrMap[0]))
			excludeAttrsMap[filename] = make(map[string]bool)
			for _, attr := range exFilenameAttrMap[1:] {
				attr = strings.TrimSpace(attr)
				if len(attr) > 0 {
					excludeAttrsMap[filename][attr] = true
				}
			}
		}
	}

	return &GeositeOut{
		Type:          TypeGeositeOut,
		Action:        action,
		Description:   DescGeositeOut,
		OutputDir:     tmp.OutputDir,
		OutputName:    tmp.OutputName,
		Want:          tmp.Want,
		Exclude:       tmp.Exclude,
		ExcludeAttrs:  excludeAttrsMap,
		GFWListOutput: tmp.GFWListOutput,
	}, nil
}

func (g *GeositeOut) GetType() string {
	return g.Type
}

func (g *GeositeOut) GetAction() lib.Action {
	return g.Action
}

func (g *GeositeOut) GetDescription() string {
	return g.Description
}

func (g *GeositeOut) Output(container lib.Container) error {
	// Create output directory
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate geosite list
	geositeList := g.toGeoSiteList(container)
	if geositeList == nil {
		return fmt.Errorf("failed to generate geosite list")
	}

	// Marshal to protobuf
	protoBytes, err := proto.Marshal(geositeList)
	if err != nil {
		return fmt.Errorf("failed to marshal geosite list: %w", err)
	}

	// Write dat file
	filepath := filepath.Join(g.OutputDir, g.OutputName)
	if err := os.WriteFile(filepath, protoBytes, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filepath, err)
	}

	log.Printf("✅ Generated %s\n", g.OutputName)

	// Generate GFWList if specified
	if g.GFWListOutput != "" {
		if err := g.generateGFWList(container); err != nil {
			return fmt.Errorf("failed to generate GFWList: %w", err)
		}
	}

	return nil
}

func (g *GeositeOut) toGeoSiteList(container lib.Container) *router.GeoSiteList {
	geositeList := new(router.GeoSiteList)

	for _, name := range g.filterAndSortList(container) {
		entry, found := container.GetEntry(name)
		if !found {
			log.Printf("❌ entry %s not found\n", name)
			continue
		}

		geosite := g.toGeoSite(entry)
		if geosite != nil {
			geositeList.Entry = append(geositeList.Entry, geosite)
		}
	}

	return geositeList
}

func (g *GeositeOut) toGeoSite(entry *lib.Entry) *router.GeoSite {
	geosite := new(router.GeoSite)
	geosite.CountryCode = strings.ToLower(entry.GetName())

	// Filter domains based on exclude attributes
	excludeAttrsMap := g.ExcludeAttrs[entry.GetName()]

	for _, domain := range entry.GetDomains() {
		// Check if domain should be excluded based on attributes
		if excludeAttrsMap != nil && len(domain.GetAttribute()) > 0 {
			shouldExclude := false
			for _, attr := range domain.GetAttribute() {
				if excludeAttrsMap[attr.GetKey()] {
					shouldExclude = true
					break
				}
			}
			if shouldExclude {
				continue
			}
		}

		geosite.Domain = append(geosite.Domain, domain)
	}

	return geosite
}

func (g *GeositeOut) generateGFWList(container lib.Container) error {
	// Find the entry for GFWList
	listName := strings.ToUpper(g.GFWListOutput)
	entry, found := container.GetEntry(listName)
	if !found {
		return fmt.Errorf("entry %s not found for GFWList generation", listName)
	}

	// Generate GFWList content
	loc, _ := time.LoadLocation("Asia/Shanghai")
	timeString := fmt.Sprintf("! Last Modified: %s\n", time.Now().In(loc).Format(time.RFC1123))

	gfwlistBytes := make([]byte, 0, 1024*512)
	gfwlistBytes = append(gfwlistBytes, []byte("[AutoProxy 0.2.9]\n")...)
	gfwlistBytes = append(gfwlistBytes, []byte(timeString)...)
	gfwlistBytes = append(gfwlistBytes, []byte("! Expires: 24h\n")...)
	gfwlistBytes = append(gfwlistBytes, []byte("! HomePage: https://github.com/Loyalsoldier/domain-list-custom\n")...)
	gfwlistBytes = append(gfwlistBytes, []byte("! GitHub URL: https://raw.githubusercontent.com/Loyalsoldier/domain-list-custom/release/gfwlist.txt\n")...)
	gfwlistBytes = append(gfwlistBytes, []byte("! jsdelivr URL: https://cdn.jsdelivr.net/gh/Loyalsoldier/domain-list-custom@release/gfwlist.txt\n")...)
	gfwlistBytes = append(gfwlistBytes, []byte("\n")...)

	for _, domain := range entry.GetDomains() {
		ruleVal := strings.TrimSpace(domain.GetValue())
		if len(ruleVal) == 0 {
			continue
		}

		switch domain.Type {
		case router.Domain_Full:
			gfwlistBytes = append(gfwlistBytes, []byte("|http://"+ruleVal+"\n")...)
			gfwlistBytes = append(gfwlistBytes, []byte("|https://"+ruleVal+"\n")...)
		case router.Domain_RootDomain:
			gfwlistBytes = append(gfwlistBytes, []byte("||"+ruleVal+"\n")...)
		case router.Domain_Plain:
			gfwlistBytes = append(gfwlistBytes, []byte(ruleVal+"\n")...)
		case router.Domain_Regex:
			gfwlistBytes = append(gfwlistBytes, []byte("/"+ruleVal+"/\n")...)
		}
	}

	// Encode to base64 and write to file
	filepath := filepath.Join(g.OutputDir, "gfwlist.txt")
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create gfwlist file: %w", err)
	}
	defer f.Close()

	encoder := base64.NewEncoder(base64.StdEncoding, f)
	defer encoder.Close()

	if _, err := encoder.Write(gfwlistBytes); err != nil {
		return fmt.Errorf("failed to write gfwlist: %w", err)
	}

	log.Printf("✅ Generated gfwlist.txt\n")
	return nil
}

func (g *GeositeOut) filterAndSortList(container lib.Container) []string {
	excludeMap := make(map[string]bool)
	for _, exclude := range g.Exclude {
		if exclude = strings.ToUpper(strings.TrimSpace(exclude)); exclude != "" {
			excludeMap[exclude] = true
		}
	}

	wantList := make([]string, 0, len(g.Want))
	for _, want := range g.Want {
		if want = strings.ToUpper(strings.TrimSpace(want)); want != "" && !excludeMap[want] {
			wantList = append(wantList, want)
		}
	}

	if len(wantList) > 0 {
		slices.Sort(wantList)
		return wantList
	}

	list := make([]string, 0, 300)
	for entry := range container.Loop() {
		name := entry.GetName()
		if excludeMap[name] {
			continue
		}
		list = append(list, name)
	}

	slices.Sort(list)
	return list
}
