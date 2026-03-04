package classifier

import (
	"strings"

	"github.com/M-Gregoire/claude-hook-guard/pkg/internal"
)

// ActionType represents the type of action a tool performs
type ActionType string

const (
	// ActionRead represents read-only operations
	ActionRead ActionType = "read"
	// ActionWrite represents write operations
	ActionWrite ActionType = "write"
)

// ToolFamily represents a logical grouping of related tools
type ToolFamily string

// ToolInfo contains classification information for a tool
type ToolInfo struct {
	ActionType ActionType
	Family     ToolFamily
}

// Classifier holds loaded tool classifications
type Classifier struct {
	toolMap map[string]ToolInfo
}

// NewClassifier creates a new classifier with families loaded from a directory
func NewClassifier(familiesDir string) (*Classifier, error) {
	_, toolMap, err := LoadFamilies(familiesDir)
	if err != nil {
		return nil, err
	}

	return &Classifier{toolMap: toolMap}, nil
}

// Classify returns the action type and tool family for a given tool
func (c *Classifier) Classify(toolName string) (ActionType, ToolFamily, bool) {
	info, ok := c.toolMap[toolName]
	if !ok {
		// Unknown tool - return false
		return "", "", false
	}
	return info.ActionType, info.Family, true
}

// ClassifyBashCommand attempts to classify a bash command
func (c *Classifier) ClassifyBashCommand(command string) (ActionType, ToolFamily) {
	// Extract the base command (first word)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ActionRead, ""
	}

	baseCmd := parts[0]
	internal.Debug("classifying bash command", "base_cmd", baseCmd, "full_command", command)

	// Check if we have classification for this command
	if info, ok := c.toolMap[baseCmd]; ok {
		internal.Debug("found classification", "cmd", baseCmd, "family", info.Family, "action", info.ActionType)
		// For git, refine action type based on subcommand
		if baseCmd == "git" && len(parts) > 1 {
			return classifyGitCommand(parts[1]), info.Family
		}
		return info.ActionType, info.Family
	}

	internal.Debug("no classification found", "cmd", baseCmd, "toolmap_size", len(c.toolMap))

	// Check for redirection operators (indicates write)
	if containsAny(command, []string{">", ">>", "tee "}) {
		return ActionWrite, ""
	}

	// Unknown command - default to read for safety
	return ActionRead, ""
}

// classifyGitCommand determines if a git subcommand is read or write
func classifyGitCommand(subcommand string) ActionType {
	readCommands := map[string]bool{
		"status": true, "log": true, "diff": true, "show": true,
		"branch": true, "remote": true, "fetch": true, "ls-files": true,
		"ls-remote": true, "rev-parse": true, "describe": true,
	}

	if readCommands[subcommand] {
		return ActionRead
	}

	// push, commit, add, merge, rebase, etc. are write
	return ActionWrite
}

// containsAny checks if a string contains any of the substrings
func containsAny(str string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(str, substr) {
			return true
		}
	}
	return false
}

// Global classifier instance (initialized on demand)
var globalClassifier *Classifier

// InitGlobalClassifier initializes the global classifier
func InitGlobalClassifier(familiesDir string) error {
	var err error
	globalClassifier, err = NewClassifier(familiesDir)
	return err
}

// Classify uses the global classifier (for backward compatibility)
func Classify(toolName string) (ActionType, ToolFamily, bool) {
	if globalClassifier == nil {
		// No families loaded - return unknown
		return "", "", false
	}
	return globalClassifier.Classify(toolName)
}

// ClassifyBashCommand uses the global classifier (for backward compatibility)
func ClassifyBashCommand(command string) (ActionType, ToolFamily) {
	if globalClassifier == nil {
		// No families loaded - analyze command directly
		if containsAny(command, []string{">", ">>", "tee "}) {
			return ActionWrite, ""
		}
		return ActionRead, ""
	}
	return globalClassifier.ClassifyBashCommand(command)
}
