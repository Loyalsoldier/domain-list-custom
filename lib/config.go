package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	inputConfigCreatorCache  = make(map[string]inputConfigCreator)
	outputConfigCreatorCache = make(map[string]outputConfigCreator)
	inputConverterCache      = make(map[string]InputConverter)
	outputConverterCache     = make(map[string]OutputConverter)
)

type inputConfigCreator func(Action, json.RawMessage) (InputConverter, error)

type outputConfigCreator func(Action, json.RawMessage) (OutputConverter, error)

func RegisterInputConfigCreator(id string, fn inputConfigCreator) error {
	id = strings.ToLower(id)
	if _, found := inputConfigCreatorCache[id]; found {
		return errors.New("config creator has already been registered")
	}
	inputConfigCreatorCache[id] = fn
	return nil
}

func createInputConfig(id string, action Action, data json.RawMessage) (InputConverter, error) {
	id = strings.ToLower(id)
	fn, found := inputConfigCreatorCache[id]
	if !found {
		return nil, errors.New("unknown config type")
	}
	return fn(action, data)
}

func RegisterOutputConfigCreator(id string, fn outputConfigCreator) error {
	id = strings.ToLower(id)
	if _, found := outputConfigCreatorCache[id]; found {
		return errors.New("config creator has already been registered")
	}
	outputConfigCreatorCache[id] = fn
	return nil
}

func createOutputConfig(id string, action Action, data json.RawMessage) (OutputConverter, error) {
	id = strings.ToLower(id)
	fn, found := outputConfigCreatorCache[id]
	if !found {
		return nil, errors.New("unknown config type")
	}
	return fn(action, data)
}

func RegisterInputConverter(id string, converter InputConverter) error {
	id = strings.ToLower(id)
	if _, found := inputConverterCache[id]; found {
		return errors.New("converter has already been registered")
	}
	inputConverterCache[id] = converter
	return nil
}

func RegisterOutputConverter(id string, converter OutputConverter) error {
	id = strings.ToLower(id)
	if _, found := outputConverterCache[id]; found {
		return errors.New("converter has already been registered")
	}
	outputConverterCache[id] = converter
	return nil
}

// Config is the configuration for converting domain lists
type Config struct {
	Input  []ConfigItem `json:"input"`
	Output []ConfigItem `json:"output"`
}

// ConfigItem is a single input or output configuration
type ConfigItem struct {
	Type   string          `json:"type"`
	Action string          `json:"action"`
	Args   json.RawMessage `json:"args"`
}

// UnmarshalJSON unmarshals a ConfigItem from JSON
func (c *ConfigItem) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Type   string          `json:"type"`
		Action string          `json:"action"`
		Args   json.RawMessage `json:"args"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	tmp.Type = strings.TrimSpace(tmp.Type)
	tmp.Action = strings.TrimSpace(tmp.Action)

	if tmp.Type == "" {
		return fmt.Errorf("type is required")
	}
	if tmp.Action == "" {
		return fmt.Errorf("action is required")
	}

	action := Action(strings.ToLower(tmp.Action))
	if !ActionsRegistry[action] {
		return fmt.Errorf("unknown action: %s", tmp.Action)
	}

	c.Type = tmp.Type
	c.Action = tmp.Action
	c.Args = tmp.Args

	return nil
}

// GetInputConverter returns an InputConverter for the ConfigItem
func (c *ConfigItem) GetInputConverter() (InputConverter, error) {
	action := Action(strings.ToLower(c.Action))
	return createInputConfig(c.Type, action, c.Args)
}

// GetOutputConverter returns an OutputConverter for the ConfigItem
func (c *ConfigItem) GetOutputConverter() (OutputConverter, error) {
	action := Action(strings.ToLower(c.Action))
	return createOutputConfig(c.Type, action, c.Args)
}
