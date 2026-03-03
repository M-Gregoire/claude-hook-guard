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

	// No rules matched - default to ask
	return hook.DecisionAsk, "No matching rules", "", nil
}

// matchRule checks if a rule matches the input
func (m *Matcher) matchRule(rule config.Rule, input *hook.Input) (bool, error) {
	// Check tool name
	if rule.Match.ToolName != nil {
		if !matchString(rule.Match.ToolName, input.ToolName) {
			return false, nil
		}
	}

	// Check action type
	if rule.Match.ActionType != nil {
		actionType, toolFamily := m.classifyTool(input)
		if !matchString(rule.Match.ActionType, string(actionType)) {
			return false, nil
		}
		// If action_type matched, also check tool_family if specified
		if rule.Match.ToolFamily != nil {
			if !matchString(rule.Match.ToolFamily, string(toolFamily)) {
				return false, nil
			}
		}
	} else if rule.Match.ToolFamily != nil {
		// Check tool family even if action_type not specified
		_, toolFamily := m.classifyTool(input)
		if !matchString(rule.Match.ToolFamily, string(toolFamily)) {
			return false, nil
		}
	}

	// Check path (looks in file_path, path parameters, or CWD)
	if rule.Match.Path != nil {
		pathToCheck := m.extractPath(input)
		if pathToCheck == "" || !matchString(rule.Match.Path, pathToCheck) {
			return false, nil
		}
	}

	// Check CWD
	if rule.Match.CWD != nil {
		if !matchString(rule.Match.CWD, input.CWD) {
			return false, nil
		}
	}

	// Check parameters
	if rule.Match.Parameters != nil {
		if !matchParameters(rule.Match.Parameters, input.ToolInput) {
			return false, nil
		}
	}

	return true, nil
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
	// For Bash commands, extract path from command if possible
	if input.ToolName == "Bash" {
		if cmd, ok := input.ToolInput["command"].(string); ok {
			// Try to extract paths from command (simple heuristic)
			// This is a basic implementation - could be enhanced
			return cmd
		}
	}
	// Fall back to CWD
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

		// Simple equality check for now
		// TODO: Add more sophisticated matching for nested objects
		if fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", expectedValue) {
			return false
		}
	}

	return true
}
