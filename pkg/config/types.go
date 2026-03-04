package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the hook guard configuration
type Config struct {
	FamiliesDir string  `yaml:"families_dir,omitempty"`
	Logging     Logging `yaml:"logging,omitempty"`
	Rules       []Rule  `yaml:"rules"`
}

// Logging configures decision logging
type Logging struct {
	Enabled bool   `yaml:"enabled"`
	File    string `yaml:"file,omitempty"`
}

// Rule represents a single permission rule
type Rule struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Match       Match  `yaml:"match"`
	Action      string `yaml:"action"` // allow, deny, ask
	Reason      string `yaml:"reason,omitempty"`
	Priority    int    `yaml:"priority,omitempty"` // Higher priority rules evaluated first
}

// Match defines the conditions for a rule to match
type Match struct {
	ToolName   *StringMatcher         `yaml:"tool_name,omitempty"`
	ActionType *StringMatcher         `yaml:"action_type,omitempty"` // read, write
	ToolFamily *StringMatcher         `yaml:"tool_family,omitempty"` // search, edit, file, git, etc.
	MCPServer  *StringMatcher         `yaml:"mcp_server,omitempty"`  // MCP server name (e.g., "atlassian")
	MCPTool    *StringMatcher         `yaml:"mcp_tool,omitempty"`    // MCP tool name (e.g., "searchJiraIssuesUsingJql")
	CWD        *StringMatcher         `yaml:"cwd,omitempty"`
	Path       *StringMatcher         `yaml:"path,omitempty"` // Matches file_path, path, or command path
	Parameters map[string]interface{} `yaml:"parameters,omitempty"`
}

// StringMatcher allows flexible string matching
type StringMatcher struct {
	Equals   string   `yaml:"equals,omitempty"`
	Regex    string   `yaml:"regex,omitempty"`
	OneOf    []string `yaml:"one_of,omitempty"`
	Contains string   `yaml:"contains,omitempty"`
	Prefix   string   `yaml:"prefix,omitempty"`
	Suffix   string   `yaml:"suffix,omitempty"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}
