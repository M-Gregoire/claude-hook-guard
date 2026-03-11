# claude-hook-guard

A rule-based permission system for Claude Code hooks. Allows, denies, or requests approval for tool operations based on configurable rules.

**Config**: `~/.config/claude-hook-guard/config.yaml`
**Families dir**: `~/.config/claude-hook-guard/families/`
**Logs**: `~/.claude/claude-hook-guard.log`
**Binary**: `~/src/claude-hook-guard/claude-hook-guard` (or wherever installed)

> **Note**: This is a workflow guardrail, not a security boundary. It prevents accidental operations; deliberate bypass is always possible.

---

## Core Concepts

### Actions
- `allow` — Auto-approve without prompting
- `deny` — Block with a reason shown to the user
- `ask` — Force the normal Claude Code permission prompt

### Match Fields
All fields are optional and use [String Matchers](#string-matchers). All specified fields must match (AND logic).

| Field | Matches |
|-------|---------|
| `action_type` | `"read"` or `"write"` (semantic classification) |
| `tool_family` | Semantic family name (`"search"`, `"git"`, `"edit"`, `"file"`, ...) |
| `tool_name` | Claude tool name (`"Bash"`, `"Grep"`, `"Skill"`, ...) |
| `mcp_server` | MCP server name (`"atlassian"`, `"datadog"`) |
| `mcp_tool` | MCP tool name within the server |
| `cwd` | Current working directory |
| `path` | File path, command, or path parameter |
| `parameters.<key>` | Specific tool input parameter (e.g., `parameters.command`) |

### String Matchers
```yaml
field:
  equals: "exact-value"
  regex: "^pattern.*"
  one_of: ["val1", "val2"]
  contains: "substring"
  prefix: "/Users/"
  suffix: ".go"
```

### Built-in Families
- `search` — Grep, Glob, grep, rg, find, fd
- `edit` — Edit, sed, awk, vim, emacs
- `file` — Read, Write, cat, cp, mv, rm, mkdir
- `git` — git, gh, lazygit
- `gotools` — go, gopls, golangci-lint, etc.

---

## Adding a Permission Rule

Edit `~/.config/claude-hook-guard/config.yaml` and add to the `rules:` list:

```yaml
rules:
  - name: my-rule                    # Unique name (required)
    description: What this does      # Optional
    priority: 100                    # Higher = evaluated first (default: 0)
    match:
      action_type:
        equals: "read"
      path:
        prefix: "/Users/gregoire/src/"
    action: allow
    reason: Safe read in src
```

**Priority guidance**: Use 200+ for safety rules (deny), 100-150 for context-specific allows, 50-90 for broad defaults.

### Common Rule Patterns

**Allow reads in a directory:**
```yaml
- name: allow-read-src
  priority: 100
  match:
    action_type:
      equals: "read"
    path:
      regex: "^/Users/gregoire/(src|dd)/"
  action: allow
  reason: Safe read operation in source directory
```

**Deny destructive commands:**
```yaml
- name: deny-rm-rf
  priority: 200
  match:
    parameters:
      command:
        regex: "rm\\s+(-[^\\s]*r[^\\s]*f|--recursive.*--force)"
  action: deny
  reason: Destructive rm -rf blocked
```

**Block direct push to main:**
```yaml
- name: deny-git-push-main
  priority: 200
  match:
    tool_family:
      equals: "git"
    parameters:
      command:
        regex: "git push.*(origin )?(main|master)"
  action: deny
  reason: Direct push to main/master not allowed
```

**Allow all reads from an MCP server:**
```yaml
- name: allow-atlassian-reads
  priority: 100
  match:
    mcp_server:
      equals: "atlassian"
    mcp_tool:
      regex: "^(get|search|list|fetch|read).*"
  action: allow
  reason: Read-only Atlassian operations
```

**Require approval for MCP writes:**
```yaml
- name: ask-atlassian-writes
  priority: 100
  match:
    mcp_server:
      equals: "atlassian"
    mcp_tool:
      regex: "^(create|update|delete|edit|add).*"
  action: ask
  reason: Atlassian modification requires approval
```

**Allow specific skills:**
```yaml
- name: allow-document-skills
  priority: 100
  match:
    tool_name:
      equals: "Skill"
    parameters:
      skill:
        one_of: ["document-skills:pdf", "document-skills:docx"]
  action: allow
  reason: Trusted document skills
```

**Allow all skills from a plugin:**
```yaml
- name: allow-all-document-skills
  priority: 100
  match:
    tool_name:
      equals: "Skill"
    parameters:
      skill:
        prefix: "document-skills:"
  action: allow
  reason: All skills from trusted document-skills plugin
```

**Blanket trust a directory:**
```yaml
- name: allow-org-projects
  priority: 150
  match:
    action_type:
      one_of: ["read", "write"]
    path:
      regex: "^/Users/gregoire/org/projects/"
  action: allow
  reason: Documentation directory always allowed
```

---

## Command Substitution Expansion

When a Bash command contains `$()` substitutions (e.g., `AUTH="$(ddtool auth token ...)" http GET url`), Claude Code normally shows an approval prompt. With `expand_command_substitutions: true`, the guard evaluates each sub-command independently. If all parts match `allow` rules, the whole command is auto-approved.

**Enable in config:**
```yaml
expand_command_substitutions: true
```

**Add rules for trusted commands:**
```yaml
- name: allow-ddtool
  match:
    parameters:
      command: { regex: "^ddtool " }
  action: allow

- name: allow-http
  match:
    parameters:
      command: { regex: "^http " }
  action: allow
```

**Combining logic:** main + all subs must be `allow` → auto-approved. Any `deny` → denied. Any unmatched → prompts user. Nested `$()` handled recursively.

---

## Removing a Rule

Delete the corresponding rule block from `~/.config/claude-hook-guard/config.yaml`. The config is read on every hook invocation — no restart needed.

---

## Creating a Family

Families group related tools for semantic matching. Create a YAML file in `~/.config/claude-hook-guard/families/`:

```yaml
# ~/.config/claude-hook-guard/families/mytools.yaml
name: mytools
description: My custom tool group
claude_tools:
  - Bash
  - Write
shell_commands:
  - mytool
  - myother-cmd
```

Then reference it in rules:
```yaml
match:
  tool_family:
    equals: "mytools"
```

> MCP tools are best matched with `mcp_server`/`mcp_tool` matchers rather than families. Families work best for Claude tools and shell commands.

---

## Testing Rules

**Test a specific tool call manually:**
```bash
echo '{"tool_name":"Grep","tool_input":{"pattern":"test","path":"/Users/gregoire/src/project"},"cwd":"/Users/gregoire/src/project"}' | \
  ~/src/claude-hook-guard/claude-hook-guard -verbose
```

**Test a Bash command:**
```bash
echo '{"tool_name":"Bash","tool_input":{"command":"rm -rf /tmp/test"},"cwd":"/Users/gregoire/src"}' | \
  ~/src/claude-hook-guard/claude-hook-guard -verbose
```

**Test an MCP tool:**
```bash
echo '{"tool_name":"mcp__atlassian__createJiraIssue","tool_input":{},"cwd":"/Users/gregoire"}' | \
  ~/src/claude-hook-guard/claude-hook-guard -verbose
```

**Test a Skill invocation:**
```bash
echo '{"tool_name":"Skill","tool_input":{"skill":"document-skills:pdf"},"cwd":"/Users/gregoire"}' | \
  ~/src/claude-hook-guard/claude-hook-guard -verbose
```

Exit code 0 + no output = pass-through (no rule matched).
Exit code 0 + JSON output = rule matched, check `permissionDecision`.

---

## Viewing Logs

```bash
# Stream live decisions
tail -f ~/.claude/claude-hook-guard.log | jq

# Count by decision type
jq -r '.decision' ~/.claude/claude-hook-guard.log | sort | uniq -c

# Show denied operations
jq 'select(.decision == "deny")' ~/.claude/claude-hook-guard.log

# Show what triggered a specific rule
jq 'select(.matched_by == "deny-rm-rf")' ~/.claude/claude-hook-guard.log

# Show Bash commands that were allowed
jq 'select(.tool_name == "Bash" and .decision == "allow")' ~/.claude/claude-hook-guard.log

# Show ignored (pass-through) operations
jq 'select(.decision == "ignore")' ~/.claude/claude-hook-guard.log
```

---

## Building from Source

```bash
cd ~/src/claude-hook-guard
go build -o claude-hook-guard ./cmd/claude-hook-guard/
```

## Claude Code Hook Registration

In `~/.claude/settings.json`:
```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": ".*",
      "hooks": [{
        "type": "command",
        "command": "/Users/gregoire/src/claude-hook-guard/claude-hook-guard",
        "timeout": 5
      }]
    }]
  }
}
```
