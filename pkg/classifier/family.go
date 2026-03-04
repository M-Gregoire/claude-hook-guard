package classifier

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/M-Gregoire/claude-hook-guard/pkg/internal"
	"gopkg.in/yaml.v3"
)

//go:embed families/search.yaml
var defaultSearchFamily string

//go:embed families/edit.yaml
var defaultEditFamily string

//go:embed families/file.yaml
var defaultFileFamily string

//go:embed families/git.yaml
var defaultGitFamily string

//go:embed families/gotools.yaml
var defaultGoToolsFamily string

// FamilyDefinition represents a tool family loaded from YAML
type FamilyDefinition struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	ClaudeTools  []string `yaml:"claude_tools"`
	ShellCommands []string `yaml:"shell_commands"`
}

// LoadFamilies loads all family definitions from a directory
func LoadFamilies(familiesDir string) (map[string]ToolFamily, map[string]ToolInfo, error) {
	if familiesDir == "" {
		home, _ := os.UserHomeDir()
		familiesDir = filepath.Join(home, ".config", "claude-hook-guard", "families")
	}

	// Create default families if directory doesn't exist
	if _, err := os.Stat(familiesDir); os.IsNotExist(err) {
		if err := createDefaultFamilies(familiesDir); err != nil {
			return nil, nil, err
		}
	}

	// Map of tool name -> ToolInfo
	toolMap := make(map[string]ToolInfo)
	// Map of family name -> ToolFamily (for validation)
	familyNames := make(map[string]ToolFamily)

	// Read all .yaml files in the families directory
	entries, err := os.ReadDir(familiesDir)
	if err != nil {
		return nil, nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		familyPath := filepath.Join(familiesDir, entry.Name())
		family, err := loadFamilyFile(familyPath)
		if err != nil {
			internal.Error("failed to load family file", "file", entry.Name(), "error", err)
			continue // Skip invalid files
		}

		familyType := ToolFamily(family.Name)
		familyNames[family.Name] = familyType

		// Add Claude Code tools
		for _, tool := range family.ClaudeTools {
			actionType := inferActionType(tool)
			toolMap[tool] = ToolInfo{
				ActionType: actionType,
				Family:     familyType,
			}
		}

		// Add shell commands (will be matched when running via Bash)
		for _, cmd := range family.ShellCommands {
			actionType := inferActionType(cmd)
			toolMap[cmd] = ToolInfo{
				ActionType: actionType,
				Family:     familyType,
			}
		}
	}

	internal.Debug("loaded tool families", "tools", len(toolMap), "families", len(familyNames))
	return familyNames, toolMap, nil
}

func loadFamilyFile(path string) (*FamilyDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var family FamilyDefinition
	if err := yaml.Unmarshal(data, &family); err != nil {
		return nil, err
	}

	return &family, nil
}

// createDefaultFamilies creates default family files in the specified directory
func createDefaultFamilies(familiesDir string) error {
	// Create directory
	if err := os.MkdirAll(familiesDir, 0755); err != nil {
		return err
	}

	// Write default family files
	defaultFamilies := map[string]string{
		"search.yaml":  defaultSearchFamily,
		"edit.yaml":    defaultEditFamily,
		"file.yaml":    defaultFileFamily,
		"git.yaml":     defaultGitFamily,
		"gotools.yaml": defaultGoToolsFamily,
	}

	for filename, content := range defaultFamilies {
		path := filepath.Join(familiesDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

// inferActionType infers the action type based on tool name
func inferActionType(toolName string) ActionType {
	// Read-only tools
	readTools := map[string]bool{
		"Read": true, "Grep": true, "Glob": true,
		"cat": true, "grep": true, "rg": true, "ag": true,
		"find": true, "fd": true, "ls": true,
		"gocyclo": true, "golangci-lint": true, "staticcheck": true,
	}

	if readTools[toolName] {
		return ActionRead
	}

	// Everything else defaults to write
	return ActionWrite
}
