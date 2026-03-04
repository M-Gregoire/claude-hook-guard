package hook

import (
	"encoding/json"
	"strings"
)

// Input represents the JSON input from Claude Code hooks
type Input struct {
	SessionID      string                 `json:"session_id"`
	TranscriptPath string                 `json:"transcript_path"`
	CWD            string                 `json:"cwd"`
	PermissionMode string                 `json:"permission_mode"`
	HookEventName  string                 `json:"hook_event_name"`
	ToolName       string                 `json:"tool_name"`
	ToolInput      map[string]interface{} `json:"tool_input"`
}

// Decision represents the permission decision
type Decision string

const (
	// DecisionAllow automatically approves the operation
	DecisionAllow Decision = "allow"
	// DecisionDeny blocks the operation
	DecisionDeny Decision = "deny"
	// DecisionAsk prompts the user for approval
	DecisionAsk Decision = "ask"
	// DecisionIgnore passes through to Claude Code (internal use)
	DecisionIgnore Decision = "ignore"
)

// Output represents the JSON output for Claude Code hooks
type Output struct {
	HookSpecificOutput HookSpecificOutput `json:"hookSpecificOutput"`
	SystemMessage      string             `json:"systemMessage,omitempty"`
}

// HookSpecificOutput contains the permission decision
type HookSpecificOutput struct {
	HookEventName            string   `json:"hookEventName,omitempty"`
	PermissionDecision       Decision `json:"permissionDecision"`
	PermissionDecisionReason string   `json:"permissionDecisionReason,omitempty"`
}

// OutputJSON returns the output as formatted JSON
func (o *Output) OutputJSON() ([]byte, error) {
	return json.MarshalIndent(o, "", "  ")
}

// ParseMCPTool parses an MCP tool name in the format "mcp__<server>__<tool>"
// and returns (server, tool, isMCP). If the tool name is not an MCP tool,
// isMCP will be false.
func ParseMCPTool(toolName string) (server string, tool string, isMCP bool) {
	if !strings.HasPrefix(toolName, "mcp__") {
		return "", "", false
	}

	parts := strings.SplitN(toolName, "__", 3)
	if len(parts) != 3 {
		return "", "", false
	}

	return parts[1], parts[2], true
}
