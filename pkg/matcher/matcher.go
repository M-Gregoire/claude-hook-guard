package matcher

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/M-Gregoire/claude-hook-guard/pkg/classifier"
	"github.com/M-Gregoire/claude-hook-guard/pkg/config"
	"github.com/M-Gregoire/claude-hook-guard/pkg/hook"
)

// Matcher evaluates rules against hook inputs
type Matcher struct {
	config *config.Config
}

// New creates a new matcher with the given configuration
func New(cfg *config.Config) *Matcher {
	return &Matcher{config: cfg}
}

// Evaluate evaluates all rules against the input and returns the decision, reason, and matched rule name
func (m *Matcher) Evaluate(input *hook.Input) (hook.Decision, string, string, error) {
	// Sort rules by priority (higher first)
	rules := make([]config.Rule, len(m.config.Rules))
	copy(rules, m.config.Rules)
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	// Evaluate each rule in priority order
	for _, rule := range rules {
		matches, err := m.matchRule(rule, input)
		if err != nil {
			return hook.DecisionAsk, "", "", fmt.Errorf("error evaluating rule %s: %w", rule.Name, err)
		}

		if matches {
			decision := hook.Decision(rule.Action)
			reason := rule.Reason
			if reason == "" {
				reason = fmt.Sprintf("Matched rule: %s", rule.Name)
			}
			return decision, reason, rule.Name, nil
		}
	}

	// No rules matched - pass through to Claude Code
	return hook.DecisionIgnore, "No matching rules", "", nil
}

// matchRule checks if a rule matches the input
func (m *Matcher) matchRule(rule config.Rule, input *hook.Input) (bool, error) {
	return m.matchesToolName(rule, input) &&
		m.matchesActionTypeAndFamily(rule, input) &&
		m.matchesMCP(rule, input) &&
		m.matchesPath(rule, input) &&
		m.matchesCWD(rule, input) &&
		m.matchesParameters(rule, input), nil
}

// matchesToolName checks if the tool name matches
func (m *Matcher) matchesToolName(rule config.Rule, input *hook.Input) bool {
	if rule.Match.ToolName == nil {
		return true
	}
	return matchString(rule.Match.ToolName, input.ToolName)
}

// matchesMCP checks if MCP server and tool match
func (m *Matcher) matchesMCP(rule config.Rule, input *hook.Input) bool {
	if rule.Match.MCPServer == nil && rule.Match.MCPTool == nil {
		return true
	}

	// Parse the tool name to extract MCP server and tool
	server, tool, isMCP := hook.ParseMCPTool(input.ToolName)
	if !isMCP {
		// Tool is not an MCP tool, only match if neither MCP field is specified
		return rule.Match.MCPServer == nil && rule.Match.MCPTool == nil
	}

	// Check MCP server if specified
	if rule.Match.MCPServer != nil {
		if !matchString(rule.Match.MCPServer, server) {
			return false
		}
	}

	// Check MCP tool if specified
	if rule.Match.MCPTool != nil {
		if !matchString(rule.Match.MCPTool, tool) {
			return false
		}
	}

	return true
}

// matchesActionTypeAndFamily checks if action type and tool family match
func (m *Matcher) matchesActionTypeAndFamily(rule config.Rule, input *hook.Input) bool {
	if rule.Match.ActionType == nil && rule.Match.ToolFamily == nil {
		return true
	}

	actionType, toolFamily := m.classifyTool(input)

	if rule.Match.ActionType != nil {
		if !matchString(rule.Match.ActionType, string(actionType)) {
			return false
		}
		// If action_type matched, also check tool_family if specified
		if rule.Match.ToolFamily != nil {
			return matchString(rule.Match.ToolFamily, string(toolFamily))
		}
		return true
	}

	// Check tool family even if action_type not specified
	return matchString(rule.Match.ToolFamily, string(toolFamily))
}

// matchesPath checks if the path matches
func (m *Matcher) matchesPath(rule config.Rule, input *hook.Input) bool {
	if rule.Match.Path == nil {
		return true
	}
	pathToCheck := m.extractPath(input)
	return pathToCheck != "" && matchString(rule.Match.Path, pathToCheck)
}

// matchesCWD checks if the current working directory matches
func (m *Matcher) matchesCWD(rule config.Rule, input *hook.Input) bool {
	if rule.Match.CWD == nil {
		return true
	}
	return matchString(rule.Match.CWD, input.CWD)
}

// matchesParameters checks if the parameters match
func (m *Matcher) matchesParameters(rule config.Rule, input *hook.Input) bool {
	if rule.Match.Parameters == nil {
		return true
	}
	return matchParameters(rule.Match.Parameters, input.ToolInput)
}

// classifyTool determines the action type and tool family for an input
func (m *Matcher) classifyTool(input *hook.Input) (classifier.ActionType, classifier.ToolFamily) {
	// Special handling for Bash - classify based on command
	if input.ToolName == "Bash" {
		if cmd, ok := input.ToolInput["command"].(string); ok {
			return classifier.ClassifyBashCommand(cmd)
		}
	}

	actionType, toolFamily, _ := classifier.Classify(input.ToolName)
	return actionType, toolFamily
}

// extractPath extracts a file path from the input for matching
func (m *Matcher) extractPath(input *hook.Input) string {
	// Check common path parameters
	if path, ok := input.ToolInput["file_path"].(string); ok {
		return path
	}
	if path, ok := input.ToolInput["path"].(string); ok {
		return path
	}
	// Fall back to CWD for all other cases (including Bash commands)
	return input.CWD
}

// matchString checks if a string matches the given matcher
func matchString(matcher *config.StringMatcher, value string) bool {
	if matcher.Equals != "" {
		return value == matcher.Equals
	}

	if matcher.Regex != "" {
		re, err := regexp.Compile(matcher.Regex)
		if err != nil {
			return false
		}
		return re.MatchString(value)
	}

	if len(matcher.OneOf) > 0 {
		for _, option := range matcher.OneOf {
			if value == option {
				return true
			}
		}
		return false
	}

	if matcher.Contains != "" {
		return strings.Contains(value, matcher.Contains)
	}

	if matcher.Prefix != "" {
		return strings.HasPrefix(value, matcher.Prefix)
	}

	if matcher.Suffix != "" {
		return strings.HasSuffix(value, matcher.Suffix)
	}

	return true
}

// matchParameters checks if parameters match the given criteria
func matchParameters(criteria map[string]interface{}, params map[string]interface{}) bool {
	for key, expectedValue := range criteria {
		actualValue, exists := params[key]
		if !exists {
			return false
		}

		// Check if expectedValue is a map (StringMatcher from YAML)
		if criteriaMap, ok := expectedValue.(map[string]interface{}); ok {
			// Convert to StringMatcher and use matchString
			matcher := &config.StringMatcher{}
			if equals, ok := criteriaMap["equals"].(string); ok {
				matcher.Equals = equals
			}
			if regex, ok := criteriaMap["regex"].(string); ok {
				matcher.Regex = regex
			}
			if contains, ok := criteriaMap["contains"].(string); ok {
				matcher.Contains = contains
			}
			if prefix, ok := criteriaMap["prefix"].(string); ok {
				matcher.Prefix = prefix
			}
			if suffix, ok := criteriaMap["suffix"].(string); ok {
				matcher.Suffix = suffix
			}
			if oneOf, ok := criteriaMap["one_of"].([]interface{}); ok {
				matcher.OneOf = make([]string, len(oneOf))
				for i, v := range oneOf {
					if str, ok := v.(string); ok {
						matcher.OneOf[i] = str
					}
				}
			}

			// Match against the actual value
			actualStr := fmt.Sprintf("%v", actualValue)
			if !matchString(matcher, actualStr) {
				return false
			}
		} else {
			// Simple equality check for non-StringMatcher values
			if fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", expectedValue) {
				return false
			}
		}
	}

	return true
}
