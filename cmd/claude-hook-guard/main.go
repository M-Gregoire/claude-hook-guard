package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/M-Gregoire/claude-hook-guard/pkg/config"
	"github.com/M-Gregoire/claude-hook-guard/pkg/hook"
	"github.com/M-Gregoire/claude-hook-guard/pkg/logger"
	"github.com/M-Gregoire/claude-hook-guard/pkg/matcher"
)

func main() {
	configPath := flag.String("config", os.ExpandEnv("$HOME/.config/claude-hook-guard/config.yaml"), "Path to configuration file")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	if err := run(*configPath, *verbose); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(configPath string, verbose bool) error {
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if verbose {
		log.Printf("Loaded %d rules from %s", len(cfg.Rules), configPath)
	}

	// Initialize logger
	logFile := cfg.Logging.File
	if logFile == "" {
		logFile = os.ExpandEnv("$HOME/.claude/claude-hook-guard.log")
	}

	decisionLogger, err := logger.New(cfg.Logging.Enabled, logFile)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer decisionLogger.Close()

	if verbose && cfg.Logging.Enabled {
		log.Printf("Logging enabled to: %s", logFile)
	}

	// Read hook input from stdin
	inputData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	var input hook.Input
	if err := json.Unmarshal(inputData, &input); err != nil {
		return fmt.Errorf("failed to parse hook input: %w", err)
	}

	if verbose {
		log.Printf("Processing tool: %s", input.ToolName)
	}

	// Create matcher and evaluate
	m := matcher.New(cfg)
	decision, reason, matchedRule, err := m.Evaluate(&input)
	if err != nil {
		return fmt.Errorf("failed to evaluate rules: %w", err)
	}

	if verbose {
		log.Printf("Decision: %s, Reason: %s, Matched: %s", decision, reason, matchedRule)
	}

	// Log the decision
	if err := decisionLogger.Log(&input, decision, reason, matchedRule); err != nil {
		log.Printf("Warning: failed to log decision: %v", err)
	}

	// If decision is to ignore (no rules matched), exit without output
	// This allows Claude Code to show normal permission prompts with "Approve for this session"
	if decision == hook.DecisionIgnore {
		if verbose {
			log.Printf("No rules matched, passing through to Claude Code")
		}
		return nil
	}

	// Build output
	output := hook.Output{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       decision,
			PermissionDecisionReason: reason,
		},
	}

	// Output JSON
	outputJSON, err := output.OutputJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	fmt.Println(string(outputJSON))
	return nil
}
