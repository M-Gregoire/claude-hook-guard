package hook

import "testing"

func TestParseMCPTool(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		wantServer string
		wantTool   string
		wantIsMCP  bool
	}{
		{
			name:       "valid MCP tool",
			toolName:   "mcp__atlassian__searchJiraIssuesUsingJql",
			wantServer: "atlassian",
			wantTool:   "searchJiraIssuesUsingJql",
			wantIsMCP:  true,
		},
		{
			name:       "valid MCP tool with underscores in tool name",
			toolName:   "mcp__atlassian__get_jira_issue",
			wantServer: "atlassian",
			wantTool:   "get_jira_issue",
			wantIsMCP:  true,
		},
		{
			name:       "non-MCP tool",
			toolName:   "Bash",
			wantServer: "",
			wantTool:   "",
			wantIsMCP:  false,
		},
		{
			name:       "non-MCP tool with underscores",
			toolName:   "some_tool_name",
			wantServer: "",
			wantTool:   "",
			wantIsMCP:  false,
		},
		{
			name:       "malformed MCP tool - missing parts",
			toolName:   "mcp__atlassian",
			wantServer: "",
			wantTool:   "",
			wantIsMCP:  false,
		},
		{
			name:       "malformed MCP tool - only one part",
			toolName:   "mcp__",
			wantServer: "",
			wantTool:   "",
			wantIsMCP:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotServer, gotTool, gotIsMCP := ParseMCPTool(tt.toolName)
			if gotServer != tt.wantServer {
				t.Errorf("ParseMCPTool() gotServer = %v, want %v", gotServer, tt.wantServer)
			}
			if gotTool != tt.wantTool {
				t.Errorf("ParseMCPTool() gotTool = %v, want %v", gotTool, tt.wantTool)
			}
			if gotIsMCP != tt.wantIsMCP {
				t.Errorf("ParseMCPTool() gotIsMCP = %v, want %v", gotIsMCP, tt.wantIsMCP)
			}
		})
	}
}
