# claude-hook-guard

A flexible, rule-based permission system for Claude Code hooks. Make intelligent decisions about allowing or denying tool operations based on tool names, parameters, working directory, and more.

## Features

- **Semantic matching**: Match by `action_type` (read/write) and `tool_family` (search/edit/file/git)
- **Rule-based matching**: Define complex permission rules using YAML configuration
- **Advanced string matching**: Supports regex, prefix/suffix, contains, and exact matching
- **Priority system**: Control rule evaluation order with priorities
- **Path matching**: Match on file paths across different tools
- **Three decision types**: `allow`, `deny`, or `ask` (prompt user)
- **Decision logging**: JSON-formatted logs of all permission decisions

## Installation

```bash
cd ~/src/claude-hook-guard
go mod tidy
go build -o claude-hook-guard ./cmd/claude-hook-guard
```

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

**Note:** You can use `action_type`/`tool_family` for semantic matching OR `tool_name` for specific tools. Semantic matching is recommended for maintainability.

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
