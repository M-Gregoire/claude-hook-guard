package classifier

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

const (
	// FamilySearch represents search tools (grep, rg, ag, find, Grep, Glob)
	FamilySearch ToolFamily = "search"
	// FamilyEdit represents edit operations (Edit, sed, awk)
	FamilyEdit ToolFamily = "edit"
	// FamilyFile represents file operations (Read, Write, cat, touch)
	FamilyFile ToolFamily = "file"
	// FamilyGit represents git commands via Bash
	FamilyGit ToolFamily = "git"
	// FamilyShell represents other bash commands
	FamilyShell ToolFamily = "shell"
)

// ToolInfo contains classification information for a tool
type ToolInfo struct {
	ActionType ActionType
	Family     ToolFamily
}

// toolClassification maps tool names to their classification
var toolClassification = map[string]ToolInfo{
	// Claude Code tools - File operations
	"Read":  {ActionRead, FamilyFile},
	"Write": {ActionWrite, FamilyFile},
	"Edit":  {ActionWrite, FamilyEdit},

	// Claude Code tools - Search operations
	"Grep": {ActionRead, FamilySearch},
	"Glob": {ActionRead, FamilySearch},

	// Bash is special - classified based on command content
	"Bash": {ActionRead, FamilyShell}, // Default, can be overridden
}

// Classify returns the action type and tool family for a given tool
func Classify(toolName string) (ActionType, ToolFamily, bool) {
	info, ok := toolClassification[toolName]
	if !ok {
		// Unknown tool - default to read/shell for safety
		return ActionRead, FamilyShell, false
	}
	return info.ActionType, info.Family, true
}

// ClassifyBashCommand attempts to classify a bash command
func ClassifyBashCommand(command string) (ActionType, ToolFamily) {
	// Check for git commands
	if containsAny(command, []string{"git "}) {
		// Git read operations
		if containsAny(command, []string{"git status", "git log", "git diff", "git show", "git branch", "git remote", "git fetch"}) {
			return ActionRead, FamilyGit
		}
		// Git write operations
		if containsAny(command, []string{"git push", "git commit", "git add", "git merge", "git rebase"}) {
			return ActionWrite, FamilyGit
		}
		return ActionRead, FamilyGit // Default git to read
	}

	// Check for search commands
	if containsAny(command, []string{"grep ", "rg ", "ag ", "find ", "locate "}) {
		return ActionRead, FamilySearch
	}

	// Check for edit commands
	if containsAny(command, []string{"sed ", "awk ", "vim ", "nano ", "emacs "}) {
		return ActionWrite, FamilyEdit
	}

	// Check for write operations
	// Redirection operators indicate write
	if containsAny(command, []string{">", ">>", "tee "}) {
		return ActionWrite, FamilyShell
	}
	// File manipulation commands
	if containsAny(command, []string{"rm ", "mv ", "cp ", "touch ", "mkdir ", "chmod ", "chown "}) {
		return ActionWrite, FamilyShell
	}

	// Default to read for safety
	return ActionRead, FamilyShell
}

// containsAny checks if the string contains any of the substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

// contains is a simple substring check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
