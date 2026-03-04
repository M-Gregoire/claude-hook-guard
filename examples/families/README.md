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

**mcpAtlassianRead.yaml** - Read-only Atlassian MCP operations (21 tools)
- Jira: Get issues, search, view projects
- Confluence: Get pages, search, view comments
- Use this family to auto-approve safe read operations on Jira/Confluence

**mcpAtlassianWrite.yaml** - Write/modify Atlassian MCP operations (10 tools)
- Jira: Create/edit issues, add comments, transition tickets
- Confluence: Create/update pages, add comments
- Use this family to require approval or block write operations

### MCP Datadog Families

**mcpDatadogRead.yaml** - Read-only Datadog MCP operations (18 tools)
- Analyze: Logs
- Get: Incidents, metrics, notebooks, traces
- Search: Dashboards, events, hosts, logs, metrics, monitors, notebooks, RUM events, services, spans
- Use this family to auto-approve safe read operations on Datadog observability platform

**mcpDatadogWrite.yaml** - Write/modify Datadog MCP operations (2 tools)
- Create/edit notebooks
- Use this family to require approval for modifications

## Usage

### Copy Example Families

To use example families, copy them to your families directory:

```bash
# Atlassian (Jira/Confluence)
cp examples/families/mcpAtlassian*.yaml ~/.config/claude-hook-guard/families/

# Datadog (observability platform)
cp examples/families/mcpDatadog*.yaml ~/.config/claude-hook-guard/families/
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

  # Auto-approve Datadog read operations
  - name: allow-datadog-reads
    priority: 100
    match:
      tool_family:
        equals: "mcpDatadogRead"
    action: allow
    reason: Safe read operation on Datadog observability platform

  # Require approval for Datadog writes
  - name: ask-datadog-writes
    priority: 100
    match:
      tool_family:
        equals: "mcpDatadogWrite"
    action: ask
    reason: Datadog modification requires approval
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
