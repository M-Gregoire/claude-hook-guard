# Example Tool Families

This directory contains example tool family definitions that you can copy to your personal families directory.

## Default Families

These families are automatically created in `~/.config/claude-hook-guard/families/` on first run:
- `search.yaml` - Search tools (Grep, Glob, grep, rg, find)
- `edit.yaml` - Edit operations (Edit, sed, awk, vim)
- `file.yaml` - File operations (Read, Write, cat, touch, cp, mv)
- `git.yaml` - Git operations (git, gh, git-lfs)
- `gotools.yaml` - Go development tools (gofmt, gocyclo, golangci-lint, staticcheck)

## Example Families (Not Created by Default)

### MCP Atlassian Families

**mcpAtlassianRead.yaml** - Read-only Atlassian MCP operations
- Jira: Get issues, search, view projects
- Confluence: Get pages, search, view comments
- Use this family to auto-approve safe read operations on Jira/Confluence

**mcpAtlassianWrite.yaml** - Write/modify Atlassian MCP operations
- Jira: Create/edit issues, add comments, transition tickets
- Confluence: Create/update pages, add comments
- Use this family to require approval or block write operations

## Usage

### Copy Example Families

To use an example family, copy it to your families directory:

```bash
cp examples/families/mcpAtlassianRead.yaml ~/.config/claude-hook-guard/families/
cp examples/families/mcpAtlassianWrite.yaml ~/.config/claude-hook-guard/families/
```

### Configure Rules

Then add rules to your config to use these families:

```yaml
rules:
  # Auto-approve Atlassian read operations
  - name: allow-atlassian-reads
    priority: 100
    match:
      tool_family:
        equals: "mcpAtlassianRead"
    action: allow
    reason: Safe read operation on Atlassian services

  # Require approval for Atlassian writes
  - name: ask-atlassian-writes
    priority: 100
    match:
      tool_family:
        equals: "mcpAtlassianWrite"
    action: ask
    reason: Atlassian modification requires approval
```

## Creating Custom Families

You can create your own tool families:

```yaml
name: my-custom-family
description: Description of what this family does
claude_tools:
  - ToolName1
  - ToolName2
shell_commands:
  - command1
  - command2
```

Save as `~/.config/claude-hook-guard/families/my-custom-family.yaml` and restart the hook.
