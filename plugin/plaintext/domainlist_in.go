package plaintext

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Loyalsoldier/domain-list-custom/lib"
	router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
)

const (
	TypeDomainListIn = "domainlist"
	DescDomainListIn = "Convert domain list to other formats"
)

func init() {
	lib.RegisterInputConfigCreator(TypeDomainListIn, func(action lib.Action, data json.RawMessage) (lib.InputConverter, error) {
		return newDomainListIn(action, data)
	})
	lib.RegisterInputConverter(TypeDomainListIn, &DomainListIn{
		Description: DescDomainListIn,
	})
}

type DomainListIn struct {
	Type        string
	Action      lib.Action
	Description string
	DataDir     string
	Want        map[string]bool
}

type fileInfo struct {
	Name                  string
	HasInclusion          bool
	InclusionAttributeMap map[string][]string
	Domains               []*router.Domain
}

func newDomainListIn(action lib.Action, data json.RawMessage) (lib.InputConverter, error) {
	var tmp struct {
		DataDir string   `json:"dataDir"`
		Want    []string `json:"wantedList"`
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &tmp); err != nil {
			return nil, err
		}
	}

	if tmp.DataDir == "" {
		return nil, fmt.Errorf("dataDir is required")
	}

	// Filter wanted list
	wantList := make(map[string]bool)
	for _, want := range tmp.Want {
		if want = strings.ToUpper(strings.TrimSpace(want)); want != "" {
			wantList[want] = true
		}
	}

	return &DomainListIn{
		Type:        TypeDomainListIn,
		Action:      action,
		Description: DescDomainListIn,
		DataDir:     tmp.DataDir,
		Want:        wantList,
	}, nil
}

func (d *DomainListIn) GetType() string {
	return d.Type
}

func (d *DomainListIn) GetAction() lib.Action {
	return d.Action
}

func (d *DomainListIn) GetDescription() string {
	return d.Description
}

func (d *DomainListIn) Input(container lib.Container) (lib.Container, error) {
	// Read all files from data directory
	fileInfoMap := make(map[string]*fileInfo)
	
	err := filepath.Walk(d.DataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		filename := strings.ToUpper(filepath.Base(path))
		
		// Filter by wanted list if specified
		if len(d.Want) > 0 && !d.Want[filename] {
			return nil
		}

		fileData, err := d.processFile(path, filename)
		if err != nil {
			return fmt.Errorf("failed to process file %s: %w", path, err)
		}

		fileInfoMap[filename] = fileData
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Process inclusions
	if err := d.processInclusions(fileInfoMap); err != nil {
		return nil, err
	}

	// Add entries to container
	for filename, fileData := range fileInfoMap {
		entry := lib.NewEntry(filename)
		entry.AddDomains(fileData.Domains)
		
		if err := container.Add(entry); err != nil {
			return nil, err
		}
	}

	return container, nil
}

func (d *DomainListIn) processFile(path string, filename string) (*fileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info := &fileInfo{
		Name:                  filename,
		InclusionAttributeMap: make(map[string][]string),
		Domains:               make([]*router.Domain, 0),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		
		if lib.IsEmpty(line) {
			continue
		}
		
		line = lib.RemoveComment(line)
		if lib.IsEmpty(line) {
			continue
		}

		// Parse rule
		domain, isInclusion, err := d.parseRule(line, info)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rule '%s': %w", line, err)
		}

		if !isInclusion && domain != nil {
			info.Domains = append(info.Domains, domain)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return info, nil
}

func (d *DomainListIn) parseRule(line string, info *fileInfo) (*router.Domain, bool, error) {
	line = strings.TrimSpace(line)

	if line == "" {
		return nil, false, errors.New("empty line")
	}

	// Parse include rule
	if strings.HasPrefix(line, "include:") {
		d.parseInclusion(line, info)
		return nil, true, nil
	}

	parts := strings.Split(line, " ")
	ruleWithType := strings.TrimSpace(parts[0])
	if ruleWithType == "" {
		return nil, false, errors.New("empty rule")
	}

	var domain router.Domain
	if err := d.parseTypeRule(ruleWithType, &domain); err != nil {
		return nil, false, err
	}

	// Parse attributes
	for _, attrString := range parts[1:] {
		if attrString = strings.TrimSpace(attrString); attrString != "" {
			attr, err := d.parseAttribute(attrString)
			if err != nil {
				return nil, false, err
			}
			domain.Attribute = append(domain.Attribute, attr)
		}
	}

	return &domain, false, nil
}

func (d *DomainListIn) parseInclusion(inclusion string, info *fileInfo) {
	inclusionVal := strings.TrimPrefix(strings.TrimSpace(inclusion), "include:")
	info.HasInclusion = true
	
	inclusionValSlice := strings.Split(inclusionVal, "@")
	filename := strings.ToUpper(strings.TrimSpace(inclusionValSlice[0]))
	
	switch len(inclusionValSlice) {
	case 1:
		// Inclusion without attribute
		info.InclusionAttributeMap[filename] = append(info.InclusionAttributeMap[filename], "@")
	default:
		// Inclusion with attributes
		for _, attr := range inclusionValSlice[1:] {
			attr = strings.ToLower(strings.TrimSpace(attr))
			if attr != "" {
				info.InclusionAttributeMap[filename] = append(info.InclusionAttributeMap[filename], "@"+attr)
			}
		}
	}
}

func (d *DomainListIn) parseTypeRule(domain string, rule *router.Domain) error {
	kv := strings.Split(domain, ":")
	switch len(kv) {
	case 1:
		// Line without type prefix
		rule.Type = router.Domain_RootDomain
		rule.Value = strings.ToLower(strings.TrimSpace(kv[0]))
	case 2:
		// Line with type prefix
		ruleType := strings.TrimSpace(kv[0])
		ruleVal := strings.TrimSpace(kv[1])
		rule.Value = strings.ToLower(ruleVal)
		
		switch strings.ToLower(ruleType) {
		case "full":
			rule.Type = router.Domain_Full
		case "domain":
			rule.Type = router.Domain_RootDomain
		case "keyword":
			rule.Type = router.Domain_Plain
		case "regexp":
			rule.Type = router.Domain_Regex
			rule.Value = ruleVal // Keep original case for regex
		default:
			return fmt.Errorf("unknown domain type: %s", ruleType)
		}
	}
	return nil
}

func (d *DomainListIn) parseAttribute(attr string) (*router.Domain_Attribute, error) {
	if attr[0] != '@' {
		return nil, fmt.Errorf("invalid attribute: %s", attr)
	}
	attr = attr[1:] // Trim attribute prefix '@'

	var attribute router.Domain_Attribute
	attribute.Key = strings.ToLower(attr)
	attribute.TypedValue = &router.Domain_Attribute_BoolValue{BoolValue: true}
	
	return &attribute, nil
}

func (d *DomainListIn) processInclusions(fileInfoMap map[string]*fileInfo) error {
	// Build dependency levels
	processed := make(map[string]bool)
	
	for len(processed) < len(fileInfoMap) {
		changed := false
		
		for filename, info := range fileInfoMap {
			if processed[filename] {
				continue
			}

			// Check if all dependencies are processed
			canProcess := true
			if info.HasInclusion {
				for depName := range info.InclusionAttributeMap {
					if !processed[depName] {
						canProcess = false
						break
					}
				}
			}

			if canProcess || !info.HasInclusion {
				// Process inclusions
				if info.HasInclusion {
					for depName, attrs := range info.InclusionAttributeMap {
						depInfo := fileInfoMap[depName]
						if depInfo == nil {
							return fmt.Errorf("included file %s not found", depName)
						}

						for _, attrWanted := range attrs {
							if attrWanted == "@" {
								// Include all domains
								info.Domains = append(info.Domains, depInfo.Domains...)
							} else {
								// Include domains with specific attribute
								for _, domain := range depInfo.Domains {
									for _, attr := range domain.Attribute {
										if "@"+attr.GetKey() == attrWanted {
											info.Domains = append(info.Domains, domain)
											break
										}
									}
								}
							}
						}
					}
				}
				
				processed[filename] = true
				changed = true
			}
		}

		if !changed {
			// Circular dependency detected
			var unprocessed []string
			for filename := range fileInfoMap {
				if !processed[filename] {
					unprocessed = append(unprocessed, filename)
				}
			}
			return fmt.Errorf("circular dependency detected in files: %v", unprocessed)
		}
	}

	return nil
}
