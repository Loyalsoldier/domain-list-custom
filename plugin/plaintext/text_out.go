package plaintext

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Loyalsoldier/domain-list-custom/lib"
)

const (
	TypeTextOut = "text"
	DescTextOut = "Convert domain lists to plaintext format"
)

func init() {
	lib.RegisterOutputConfigCreator(TypeTextOut, func(action lib.Action, data json.RawMessage) (lib.OutputConverter, error) {
		return newTextOut(action, data)
	})
	lib.RegisterOutputConverter(TypeTextOut, &TextOut{
		Description: DescTextOut,
	})
}

type TextOut struct {
	Type        string
	Action      lib.Action
	Description string
	OutputDir   string
	OutputExt   string
	Want        []string
	Exclude     []string
}

func newTextOut(action lib.Action, data json.RawMessage) (lib.OutputConverter, error) {
	var tmp struct {
		OutputDir string   `json:"outputDir"`
		Want      []string `json:"wantedList"`
		Exclude   []string `json:"excludedList"`
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &tmp); err != nil {
			return nil, err
		}
	}

	if tmp.OutputDir == "" {
		tmp.OutputDir = "./output"
	}

	return &TextOut{
		Type:        TypeTextOut,
		Action:      action,
		Description: DescTextOut,
		OutputDir:   tmp.OutputDir,
		OutputExt:   ".txt",
		Want:        tmp.Want,
		Exclude:     tmp.Exclude,
	}, nil
}

func (t *TextOut) GetType() string {
	return t.Type
}

func (t *TextOut) GetAction() lib.Action {
	return t.Action
}

func (t *TextOut) GetDescription() string {
	return t.Description
}

func (t *TextOut) Output(container lib.Container) error {
	// Create output directory
	if err := os.MkdirAll(t.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, name := range t.filterAndSortList(container) {
		entry, found := container.GetEntry(name)
		if !found {
			log.Printf("❌ entry %s not found\n", name)
			continue
		}

		data, err := entry.MarshalText()
		if err != nil {
			return fmt.Errorf("failed to marshal entry %s: %w", name, err)
		}

		filename := strings.ToLower(entry.GetName()) + t.OutputExt
		filepath := filepath.Join(t.OutputDir, filename)
		
		if err := os.WriteFile(filepath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filepath, err)
		}

		log.Printf("✅ Generated %s\n", filename)
	}

	return nil
}

func (t *TextOut) filterAndSortList(container lib.Container) []string {
	excludeMap := make(map[string]bool)
	for _, exclude := range t.Exclude {
		if exclude = strings.ToUpper(strings.TrimSpace(exclude)); exclude != "" {
			excludeMap[exclude] = true
		}
	}

	wantList := make([]string, 0, len(t.Want))
	for _, want := range t.Want {
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
