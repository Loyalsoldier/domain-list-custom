package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

type fileName string

type attribute string

// GetDataDir returns the path to the "data" directory used to generate lists.
// Usage order:
// 1. The datapath that user set when running the program
// 2. The default path "./data" (data directory in the current working directory) if exists
// 3. The path to the data directory of project `v2fly/domain-list-community` in GOPATH mode
func GetDataDir() string {
	if *dataPath != "" { // Use dataPath option if set by user
		fmt.Printf("Use domain list files in '%s' directory.\n", *dataPath)
		return *dataPath
	}

	defaultDataDir := filepath.Join("./", "data")
	if _, err := os.Stat(defaultDataDir); !os.IsNotExist(err) { // Use "./data" directory if exists
		fmt.Printf("Use domain list files in '%s' directory.\n", defaultDataDir)
		return defaultDataDir
	}

	return filepath.Join(GetGOPATH(), "src", "github.com", "v2fly", "domain-list-community", "data")
}

// envFile returns the name of the Go environment configuration file.
// Copy from https://github.com/golang/go/blob/c4f2a9788a7be04daf931ac54382fbe2cb754938/src/cmd/go/internal/cfg/cfg.go#L150-L166
func envFile() (string, error) {
	if file := os.Getenv("GOENV"); file != "" {
		if file == "off" {
			return "", fmt.Errorf("GOENV=off")
		}
		return file, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", fmt.Errorf("missing user-config dir")
	}
	return filepath.Join(dir, "go", "env"), nil
}

// GetRuntimeEnv returns the value of runtime environment variable,
// that is set by running following command: `go env -w key=value`.
func GetRuntimeEnv(key string) (string, error) {
	file, err := envFile()
	if err != nil {
		return "", err
	}
	if file == "" {
		return "", fmt.Errorf("missing runtime env file")
	}
	var data []byte
	var runtimeEnv string
	data, readErr := os.ReadFile(file)
	if readErr != nil {
		return "", readErr
	}
	envStrings := strings.Split(string(data), "\n")
	for _, envItem := range envStrings {
		envItem = strings.TrimSuffix(envItem, "\r")
		envKeyValue := strings.Split(envItem, "=")
		if strings.EqualFold(strings.TrimSpace(envKeyValue[0]), key) {
			runtimeEnv = strings.TrimSpace(envKeyValue[1])
		}
	}
	return runtimeEnv, nil
}

// GetGOPATH returns GOPATH environment variable as a string. It will NOT be empty.
func GetGOPATH() string {
	// The one set by user explicitly by `export GOPATH=/path` or `env GOPATH=/path command`
	GOPATH := os.Getenv("GOPATH")
	if GOPATH == "" {
		var err error
		// The one set by user by running `go env -w GOPATH=/path`
		GOPATH, err = GetRuntimeEnv("GOPATH")
		if err != nil {
			// The default one that Golang uses
			return build.Default.GOPATH
		}
		if GOPATH == "" {
			return build.Default.GOPATH
		}
		return GOPATH
	}
	return GOPATH
}

// isEmpty checks if the rule that has been trimmed out spaces is empty
func isEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// removeComment removes comments in the rule
func removeComment(line string) string {
	idx := strings.Index(line, "#")
	if idx == -1 {
		return line
	}
	return strings.TrimSpace(line[:idx])
}
