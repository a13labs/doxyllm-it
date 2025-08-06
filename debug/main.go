package main

import (
"fmt"
"os"
"path/filepath"
"gopkg.in/yaml.v2"
)

type GroupConfig struct {
	Name         string   `yaml:"name"`
	Title        string   `yaml:"title"`
	Description  string   `yaml:"description"`
	Files        []string `yaml:"files"`
	GenerateDefgroup bool `yaml:"generateDefgroup"`
}

type DoxyllmConfig struct {
	Global string            `yaml:"global,omitempty"`
	Files  map[string]string `yaml:"files,omitempty"`
	Ignore []string          `yaml:"ignore,omitempty"`
	Groups map[string]GroupConfig `yaml:"groups,omitempty"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: debug <config-file> <target-file>")
		os.Exit(1)
	}
	
	configFile := os.Args[1]
	targetFile := os.Args[2]
	
	// Read configuration
	content, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}
	
	var config DoxyllmConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		fmt.Printf("Error parsing YAML: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("=== Configuration Debug ===\n")
	fmt.Printf("Config file: %s\n", configFile)
	fmt.Printf("Target file: %s\n", targetFile)
	fmt.Printf("Groups found: %d\n", len(config.Groups))
	
	for groupName, group := range config.Groups {
		fmt.Printf("\nGroup: %s\n", groupName)
		fmt.Printf("  Name: %s\n", group.Name)
		fmt.Printf("  Title: %s\n", group.Title)
		fmt.Printf("  GenerateDefgroup: %t\n", group.GenerateDefgroup)
		fmt.Printf("  Files: %v\n", group.Files)
		
		// Test file matching
		configDir := filepath.Dir(configFile)
		fmt.Printf("  Config dir: %s\n", configDir)
		
		relPath, err := filepath.Rel(configDir, targetFile)
		if err != nil {
			relPath = filepath.Base(targetFile)
		}
		fmt.Printf("  Relative path: %s\n", relPath)
		fmt.Printf("  Base name: %s\n", filepath.Base(targetFile))
		
		for _, groupFile := range group.Files {
			matched, err := filepath.Match(groupFile, relPath)
			fmt.Printf("  Match '%s' with '%s': %t (err: %v)\n", groupFile, relPath, matched, err)
			
			if groupFile == relPath || groupFile == filepath.Base(targetFile) {
				fmt.Printf("  Exact match with '%s': true\n", groupFile)
			}
		}
	}
}
