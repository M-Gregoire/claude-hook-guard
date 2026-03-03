package hook

import "encoding/json"

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
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
	DecisionAsk   Decision = "ask"
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
