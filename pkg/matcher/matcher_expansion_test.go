package matcher_test

import (
	"testing"

	"github.com/M-Gregoire/claude-hook-guard/pkg/config"
	"github.com/M-Gregoire/claude-hook-guard/pkg/hook"
	"github.com/M-Gregoire/claude-hook-guard/pkg/matcher"
)

func makeExpansionMatcher(rules []config.Rule) *matcher.Matcher {
	return matcher.New(&config.Config{
		ExpandCommandSubstitutions: true,
		Rules:                      rules,
	})
}

func expansionInput(command string) *hook.Input {
	return &hook.Input{
		ToolName: "Bash",
		ToolInput: map[string]interface{}{
			"command": command,
		},
		CWD: "/tmp",
	}
}

// allowRulePrefix matches commands that start with the given prefix (more precise than contains)
func allowRulePrefix(name, cmdPrefix string) config.Rule {
	return config.Rule{
		Name:   name,
		Action: "allow",
		Match: config.Match{
			ToolName: &config.StringMatcher{Equals: "Bash"},
			Parameters: map[string]interface{}{
				"command": map[string]interface{}{"regex": "^" + cmdPrefix},
			},
		},
	}
}

func allowRule(name, cmdContains string) config.Rule {
	return config.Rule{
		Name:   name,
		Action: "allow",
		Match: config.Match{
			ToolName: &config.StringMatcher{Equals: "Bash"},
			Parameters: map[string]interface{}{
				"command": map[string]interface{}{"contains": cmdContains},
			},
		},
	}
}

func denyRule(name, cmdContains string) config.Rule {
	return config.Rule{
		Name:   name,
		Action: "deny",
		Match: config.Match{
			ToolName: &config.StringMatcher{Equals: "Bash"},
			Parameters: map[string]interface{}{
				"command": map[string]interface{}{"contains": cmdContains},
			},
		},
	}
}

func TestEvaluateWithExpansion_AllAllow(t *testing.T) {
	m := makeExpansionMatcher([]config.Rule{
		allowRule("allow-ddtool", "ddtool"),
		allowRule("allow-http", "http"),
	})

	decision, _, _, err := m.Evaluate(expansionInput(
		`AUTH="$(ddtool auth token foo --http-header)" http GET https://example.com "$AUTH"`,
	))
	if err != nil {
		t.Fatal(err)
	}
	if decision != hook.DecisionAllow {
		t.Errorf("got %v, want allow", decision)
	}
}

func TestEvaluateWithExpansion_SubCommandNotMatched(t *testing.T) {
	// Only http is allowed (by prefix), ddtool has no rule → sub-command is ignored→ask
	m := makeExpansionMatcher([]config.Rule{
		allowRulePrefix("allow-http", "http"),
	})

	decision, _, _, err := m.Evaluate(expansionInput(
		`AUTH="$(ddtool auth token foo --http-header)" http GET https://example.com "$AUTH"`,
	))
	if err != nil {
		t.Fatal(err)
	}
	if decision != hook.DecisionAsk {
		t.Errorf("got %v, want ask", decision)
	}
}

func TestEvaluateWithExpansion_MainCommandDenied(t *testing.T) {
	m := makeExpansionMatcher([]config.Rule{
		allowRule("allow-ddtool", "ddtool"),
		denyRule("deny-http", "http"),
	})

	decision, _, _, err := m.Evaluate(expansionInput(
		`AUTH="$(ddtool auth token foo --http-header)" http GET https://example.com "$AUTH"`,
	))
	if err != nil {
		t.Fatal(err)
	}
	if decision != hook.DecisionDeny {
		t.Errorf("got %v, want deny", decision)
	}
}

func TestEvaluateWithExpansion_SubCommandDenied(t *testing.T) {
	m := makeExpansionMatcher([]config.Rule{
		denyRule("deny-ddtool", "ddtool"),
		allowRule("allow-http", "http"),
	})

	decision, _, _, err := m.Evaluate(expansionInput(
		`AUTH="$(ddtool auth token foo --http-header)" http GET https://example.com "$AUTH"`,
	))
	if err != nil {
		t.Fatal(err)
	}
	if decision != hook.DecisionDeny {
		t.Errorf("got %v, want deny", decision)
	}
}

func TestEvaluateWithExpansion_DisabledByDefault(t *testing.T) {
	// With expand_command_substitutions: false, $() commands are NOT expanded
	m := matcher.New(&config.Config{
		ExpandCommandSubstitutions: false,
		Rules: []config.Rule{
			allowRule("allow-http", "http"),
		},
	})

	// Without expansion, the whole raw command is matched — "http" is in it, so it matches allow
	// But the key point is: no sub-command evaluation occurs
	decision, _, _, err := m.Evaluate(expansionInput(
		`AUTH="$(ddtool auth token foo --http-header)" http GET https://example.com "$AUTH"`,
	))
	if err != nil {
		t.Fatal(err)
	}
	// The raw command contains "http" so the allow rule matches directly
	if decision != hook.DecisionAllow {
		t.Errorf("got %v, want allow (rule matched raw command)", decision)
	}
}

func TestEvaluateWithExpansion_NoSubstitution(t *testing.T) {
	m := makeExpansionMatcher([]config.Rule{
		allowRule("allow-http", "http"),
	})

	// No $() — should use normal evaluateRules path
	decision, _, _, err := m.Evaluate(expansionInput("http GET https://example.com"))
	if err != nil {
		t.Fatal(err)
	}
	if decision != hook.DecisionAllow {
		t.Errorf("got %v, want allow", decision)
	}
}

func TestEvaluateWithExpansion_NestedSubstitution(t *testing.T) {
	// Nested: $(cat $(find . -name foo))
	// Outer sub: "cat $(find . -name foo)" — contains "cat" and nested "find . -name foo"
	m := makeExpansionMatcher([]config.Rule{
		allowRule("allow-echo", "echo"),
		allowRule("allow-cat", "cat"),
		allowRule("allow-find", "find"),
	})

	decision, _, _, err := m.Evaluate(expansionInput("echo $(cat $(find . -name foo))"))
	if err != nil {
		t.Fatal(err)
	}
	if decision != hook.DecisionAllow {
		t.Errorf("got %v, want allow", decision)
	}
}
