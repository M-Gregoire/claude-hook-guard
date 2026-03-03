# claude-hook-guard

A flexible, rule-based permission system for Claude Code hooks. Make intelligent decisions about allowing or denying tool operations based on tool names, parameters, working directory, and more.

## Features

- **Rule-based matching**: Define complex permission rules using YAML configuration
- **Advanced string matching**: Supports regex, prefix/suffix, contains, and exact matching
- **Priority system**: Control rule evaluation order with priorities
- **Parameter inspection**: Match on tool parameters and command options
- **Three decision types**: `allow`, `deny`, or `ask` (prompt user)

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
  # Allow read operations in ~/src
  - name: allow-read-src
    priority: 100
    match:
      tool_name:
        one_of: ["Read", "Grep", "Glob"]
      parameters:
        path:
          regex: "^(/Users/.+/src/|~/src/)"
    action: allow
    reason: Safe read operation in source directory

  # Deny destructive operations
  - name: deny-destructive
    priority: 200
    match:
      tool_name:
        equals: "Bash"
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
    tool_name:
      equals: "Bash"
    cwd:
      prefix: "/Users/user/src/"
    parameters:
      command:
        regex: "git.*"
  action: allow  # or deny, ask
  reason: Optional reason shown to user
```

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
