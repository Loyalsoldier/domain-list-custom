package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Instance is the main instance for converting domain lists
type Instance struct {
	Config    *Config
	Container Container
}

// NewInstance creates a new Instance
func NewInstance() (*Instance, error) {
	return &Instance{
		Container: NewSimpleContainer(),
	}, nil
}

// InitConfig initializes the instance with a config file
func (i *Instance) InitConfig(configFile string) error {
	var configBytes []byte
	var err error

	configFile = strings.TrimSpace(configFile)
	if configFile == "" {
		return fmt.Errorf("config file is required")
	}

	// Check if it's a URL or local file
	if strings.HasPrefix(strings.ToLower(configFile), "http://") ||
		strings.HasPrefix(strings.ToLower(configFile), "https://") {
		// Download from URL
		resp, err := http.Get(configFile)
		if err != nil {
			return fmt.Errorf("failed to download config: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to download config: status code %d", resp.StatusCode)
		}

		configBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}
	} else {
		// Read from local file
		configBytes, err = os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Parse config
	var config Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	i.Config = &config
	return nil
}

// Run runs the conversion process
func (i *Instance) Run() error {
	if i.Config == nil {
		return fmt.Errorf("config is not initialized")
	}

	// Process input
	log.Println("Processing input...")
	for idx, inputConfig := range i.Config.Input {
		log.Printf("  [%d/%d] type: %s, action: %s", idx+1, len(i.Config.Input), inputConfig.Type, inputConfig.Action)
		
		converter, err := inputConfig.GetInputConverter()
		if err != nil {
			return fmt.Errorf("failed to get input converter: %w", err)
		}

		newContainer, err := converter.Input(i.Container)
		if err != nil {
			return fmt.Errorf("failed to process input [type: %s, action: %s]: %w", 
				inputConfig.Type, inputConfig.Action, err)
		}

		if newContainer != nil {
			i.Container = newContainer
		}
	}

	log.Printf("Processed %d entries\n", i.Container.Len())

	// Process output
	log.Println("Processing output...")
	for idx, outputConfig := range i.Config.Output {
		log.Printf("  [%d/%d] type: %s, action: %s", idx+1, len(i.Config.Output), outputConfig.Type, outputConfig.Action)
		
		converter, err := outputConfig.GetOutputConverter()
		if err != nil {
			return fmt.Errorf("failed to get output converter: %w", err)
		}

		if err := converter.Output(i.Container); err != nil {
			return fmt.Errorf("failed to process output [type: %s, action: %s]: %w", 
				outputConfig.Type, outputConfig.Action, err)
		}
	}

	log.Println("Done!")
	return nil
}
