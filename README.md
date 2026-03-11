# claude-hook-guard

[![CI](https://github.com/M-Gregoire/claude-hook-guard/workflows/CI/badge.svg)](https://github.com/M-Gregoire/claude-hook-guard/actions)
[![Release](https://img.shields.io/github/v/release/M-Gregoire/claude-hook-guard)](https://github.com/M-Gregoire/claude-hook-guard/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/M-Gregoire/claude-hook-guard)](https://goreportcard.com/report/github.com/M-Gregoire/claude-hook-guard)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/M-Gregoire/claude-hook-guard/blob/master/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/M-Gregoire/claude-hook-guard)](https://github.com/M-Gregoire/claude-hook-guard/blob/master/go.mod)

A flexible, rule-based permission system for Claude Code hooks. Make intelligent decisions about allowing or denying tool operations based on tool names, parameters, working directory, and more.

## ⚠️ Security Model

**claude-hook-guard is a workflow guardrail, not a security boundary.**

This tool is designed to:
- ✅ Prevent **accidental** unwanted operations
- ✅ Automate approval of **trusted** workflows
- ✅ Provide **audit logs** of Claude's actions
- ✅ Enforce **organizational policies** for routine tasks

This tool is **NOT** designed to:
- ❌ Prevent **deliberate bypass attempts**
- ❌ Provide **security isolation** or sandboxing
- ❌ Protect against **adversarial behavior**

### Why It's Easy to Bypass

The hook operates at the tool invocation level, not the system call level. Example bypass:

```bash
# Even if reading /etc/sensitive is blocked:
# 1. Write a program that reads the file
cat > bypass.go << 'EOF'
package main
import ("fmt"; "os")
func main() {
    data, _ := os.ReadFile("/etc/sensitive")
    fmt.Println(string(data))
}
EOF

# 2. Compile it (allowed - just running go build)
go build -o bypass

# 3. Execute it (allowed - just running ./bypass)
./bypass  # Reads /etc/sensitive, bypassing your rule!
```

The hook cannot see what compiled binaries or scripts do internally.

### Path Matching is Best-Effort

**Determining which path a command operates on is fundamentally difficult and unreliable.**

The hook extracts paths in this order:
1. Explicit parameters (`file_path`, `path`) from tools like Read, Write, Edit
2. Current working directory (CWD) as fallback for Bash commands

For shell commands, the hook **does not** attempt to parse command arguments to extract paths because:

- **Syntax ambiguity**: Is `-p /path` a flag with argument, or a boolean flag followed by a positional argument?
- **Semantic ambiguity**: Which path matters? `rsync /src /dst` has both source and destination. `docker run -v /host:/container` has two paths.
- **Multiple paths**: `cat file1 file2 file3` operates on multiple files
- **No explicit path**: `npm install` operates in CWD but has no path argument

**Recommendation:** Path-based rules work best for:
- Tools with explicit path parameters (Read, Write, Edit, Grep, Glob)
- CWD-based matching for shell commands (e.g., "allow all operations when CWD is ~/dev/")
- Blocking entire command patterns regardless of path (e.g., "deny all `rm -rf`")

**Do not rely on path matching as a security control.** It's a convenience feature for workflow automation, not a security boundary.

### When to Use This Tool

**Good use cases:**
- Prevent accidental `rm -rf` in the wrong directory
- Auto-approve safe read operations in your development directories
- Require confirmation before git pushes to main/master
- Auto-approve document creation skills
- Log all MCP operations for compliance

**Bad use cases:**
- Preventing a malicious actor from accessing sensitive files
- Enforcing security policies against untrusted code
- Sandboxing Claude from system resources

### For Real Security Isolation

If you need actual security boundaries, use OS-level sandboxing:
- **Containers:** Docker, Podman
- **VMs:** Virtual machines with network isolation
- **OS Sandboxing:** macOS sandbox profiles, Linux seccomp/AppArmor
- **Dedicated environments:** Separate development environments for sensitive work

## Features

- **Semantic matching**: Match by `action_type` (read/write) and `tool_family` (search/edit/file/git)
- **MCP support**: Control permissions for MCP (Model Context Protocol) server operations
- **Rule-based matching**: Define complex permission rules using YAML configuration
- **Advanced string matching**: Supports regex, prefix/suffix, contains, and exact matching
- **Priority system**: Control rule evaluation order with priorities
- **Path matching**: Match on file paths across different tools
- **Three decision types**: `allow`, `deny`, or `ask` (prompt user)
- **Decision logging**: JSON-formatted logs of all permission decisions
- **Command substitution expansion**: Automatically evaluate `$()` sub-commands against rules

## Installation

### Using go install (recommended)

```bash
go install github.com/M-Gregoire/claude-hook-guard/cmd/claude-hook-guard@latest
```

### From source

```bash
git clone https://github.com/M-Gregoire/claude-hook-guard.git
cd claude-hook-guard
go build -o claude-hook-guard ./cmd/claude-hook-guard
```

### From release binaries

Download pre-built binaries from the [releases page](https://github.com/M-Gregoire/claude-hook-guard/releases).

## Configuration

Create a config file at `~/.config/claude-hook-guard/config.yaml`:

```yaml
logging:
  enabled: true
  file: $HOME/.claude/claude-hook-guard.log

rules:
  # Allow all operations in ~/org/projects
  - name: allow-org-projects
    priority: 150
    match:
      action_type:
        one_of: ["read", "write"]
      path:
        regex: "^(/Users/.+/org/projects/|~/org/projects/)"
    action: allow
    reason: Documentation directory

  # Allow read operations in ~/src
  - name: allow-read-src
    priority: 100
    match:
      action_type:
        equals: "read"
      path:
        regex: "^(/Users/.+/src/|~/src/)"
    action: allow
    reason: Safe read operation in source directory

  # Deny destructive operations
  - name: deny-destructive
    priority: 200
    match:
      action_type:
        equals: "write"
      parameters:
        command:
          regex: "rm\\s+(-[^\\s]*r[^\\s]*f|--recursive.*--force)"
    action: deny
    reason: Destructive command blocked for safety
```

See `examples/config.yaml` for a complete example.

## Usage with Claude Code

Update your Claude Code hooks configuration in `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": ".*",
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/claude-hook-guard",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
```

## Rule Configuration

### Semantic Matching

Match tools by their semantic meaning rather than specific names:

**Action Types:**
- `read`: Read-only operations (Read, Grep, Glob, git status, cat file, ls, etc.)
- `write`: Write operations (Write, Edit, git commit, sed, rm, touch, cat > file, etc.)

Note: Commands with output redirection (`>`, `>>`) are classified as write operations.

**Tool Families:**
- `search`: Search tools (Grep, Glob, grep, rg, ag, find)
- `file`: File operations (Read, Write)
- `edit`: Edit operations (Edit, sed, awk, vim)
- `git`: Git commands (via Bash)
- `shell`: Shell commands (cat, touch, ls, rm, mv, cp, etc.)

**Example: Allow all writes in ~/org/projects**
```yaml
- name: allow-org-writes
  match:
    action_type:
      equals: "write"
    path:
      prefix: "/Users/user/org/projects/"
  action: allow
```

**Example: Allow search tools anywhere**
```yaml
- name: allow-search
  match:
    tool_family:
      equals: "search"
  action: allow
```

### Skill Matching

Control permissions for Claude Code skills (e.g., from https://github.com/anthropics/skills):

Skills are invoked with `tool_name: "Skill"` and the skill name in `parameters.skill`.

**Use fully qualified names** (`plugin:skill`) to avoid ambiguity when multiple plugins have skills with the same name.

**Example: Auto-approve document creation skills**
```yaml
- name: allow-document-skills
  match:
    tool_name:
      equals: "Skill"
    parameters:
      skill:
        one_of: ["document-skills:pdf", "document-skills:docx", "document-skills:pptx"]
  action: allow
  reason: Safe document creation skill from official Anthropic plugin
```

**Example: Match all skills from a trusted plugin**
```yaml
- name: allow-all-document-skills
  match:
    tool_name:
      equals: "Skill"
    parameters:
      skill:
        prefix: "document-skills:"
  action: allow
  reason: All skills from trusted document-skills plugin
```

**Example: Require approval for builder skills**
```yaml
- name: ask-builder-skills
  match:
    tool_name:
      equals: "Skill"
    parameters:
      skill:
        one_of: ["mcp-builder", "skill-creator", "web-artifacts-builder"]
  action: ask
  reason: Builder skill requires approval
```

### MCP (Model Context Protocol) Matching

Control permissions for MCP server operations by matching on server name and tool name:

**MCP Server**: The MCP server providing the tool (e.g., `atlassian`, `github`)
**MCP Tool**: The specific operation on that server (e.g., `searchJiraIssuesUsingJql`, `getJiraIssue`)

**Example: Allow all read operations from atlassian MCP**
```yaml
- name: allow-atlassian-reads
  match:
    mcp_server:
      equals: "atlassian"
    mcp_tool:
      regex: "^(get|search|list|fetch|read).*"
  action: allow
  reason: Safe read operation on Atlassian services
```

**Example: Require approval for Atlassian writes**
```yaml
- name: ask-atlassian-writes
  match:
    mcp_server:
      equals: "atlassian"
    mcp_tool:
      regex: "^(create|update|delete|edit|add).*"
  action: ask
  reason: Modification to Atlassian resources requires approval
```

**Example: Blanket approve all operations from a trusted MCP server**
```yaml
- name: allow-trusted-mcp
  match:
    mcp_server:
      equals: "trusted-server"
  action: allow
  reason: Trusted MCP server
```

### String Matching

String matchers support multiple matching strategies:

- `equals`: Exact string match
- `regex`: Regular expression match
- `one_of`: Match any string in the list
- `contains`: String contains substring
- `prefix`: String starts with prefix
- `suffix`: String ends with suffix

### Rule Structure

```yaml
- name: rule-name
  description: Optional description
  priority: 100  # Higher priority evaluated first
  match:
    action_type:       # Optional: "read" or "write"
      equals: "read"
    tool_family:       # Optional: "search", "edit", "file", "git", "shell"
      equals: "search"
    tool_name:         # Optional: specific tool name
      equals: "Bash"
    mcp_server:        # Optional: MCP server name (e.g., "atlassian")
      equals: "atlassian"
    mcp_tool:          # Optional: MCP tool name (e.g., "searchJiraIssuesUsingJql")
      regex: "^search.*"
    path:              # Optional: matches file_path, path parameter, or command
      regex: "^/Users/.+/src/"
    cwd:               # Optional: current working directory
      prefix: "/Users/user/src/"
    parameters:        # Optional: specific parameter matching
      command:
        regex: "git.*"
  action: allow        # or deny, ask
  reason: Optional reason shown to user
```

**Note:** You can use `action_type`/`tool_family` for semantic matching OR `tool_name` for specific tools. For MCP tools, use `mcp_server` and `mcp_tool` matchers. For skills, match on `tool_name: "Skill"` and use `parameters.skill` to match the skill name. Semantic matching is recommended for maintainability.

## Command Substitution Expansion

When a Bash command contains `$()` substitutions, Claude Code shows an approval prompt by default (e.g., `AUTH="$(ddtool auth token ...)" http GET url`). With `expand_command_substitutions: true`, claude-hook-guard evaluates each sub-command independently against your rules. If the main command **and** all sub-commands each match an `allow` rule, the whole command is auto-approved — no prompt needed.

Enable in your config:

```yaml
expand_command_substitutions: true
```

Then add rules for each command you trust:

```yaml
rules:
  # Allow ddtool for auth token generation
  - name: allow-ddtool
    match:
      parameters:
        command: { regex: "^ddtool " }
    action: allow

  # Allow httpie for API testing
  - name: allow-http
    match:
      parameters:
        command: { regex: "^http " }
    action: allow
```

With these rules, `AUTH="$(ddtool auth token foo --http-header)" http GET https://example.com "$AUTH"` is auto-approved because both `ddtool` and `http` are individually allowed.

**Combining logic:**

| Main command | Sub-commands | Result |
|---|---|---|
| `allow` | all `allow` | **allow** |
| `allow` | any `ask` or no rule | **ask** |
| `allow` | any `deny` | **deny** |
| `deny` | any | **deny** |
| `ask` / no rule | any | **ask** |

Nested substitutions (e.g., `$(cmd $(inner))`) are handled recursively. Default: `false`.

## Logging

Enable decision logging to track all permission decisions:

```yaml
logging:
  enabled: true
  file: $HOME/.claude/claude-hook-guard.log
```

The log file contains JSON entries with:
- Timestamp
- Tool name and input parameters
- Working directory
- Decision (allow/deny/ask)
- Reason for the decision
- Matched rule name

View recent decisions:
```bash
tail -f ~/.claude/claude-hook-guard.log | jq
```

Analyze decisions:
```bash
# Count decisions by type
jq -r '.decision' ~/.claude/claude-hook-guard.log | sort | uniq -c

# Show all denied operations
jq 'select(.decision == "deny")' ~/.claude/claude-hook-guard.log

# Show operations for a specific tool
jq 'select(.tool_name == "Bash")' ~/.claude/claude-hook-guard.log
```

## Testing

Test your hook manually:

```bash
echo '{"tool_name":"Grep","tool_input":{"pattern":"test","path":"/Users/user/src/project"}}' | \
  ./claude-hook-guard -verbose
```

## License

MIT
