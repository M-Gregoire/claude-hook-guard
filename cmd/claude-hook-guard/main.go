package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/M-Gregoire/claude-hook-guard/pkg/classifier"
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
	cfg, err := loadConfigWithLogging(configPath, verbose)
	if err != nil {
		return err
	}

	decisionLogger, err := setupLogger(cfg, verbose)
	if err != nil {
		return err
	}
	defer func() {
		if err := decisionLogger.Close(); err != nil {
			log.Printf("Warning: failed to close logger: %v", err)
		}
	}()

	input, err := parseHookInput(verbose)
	if err != nil {
		return err
	}

	decision, reason, matchedRule, err := evaluateInput(cfg, &input, verbose)
	if err != nil {
		return err
	}

	if err := decisionLogger.Log(&input, decision, reason, matchedRule); err != nil {
		log.Printf("Warning: failed to log decision: %v", err)
	}

	return outputDecision(decision, reason, verbose)
}

func loadConfigWithLogging(configPath string, verbose bool) (*config.Config, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	if verbose {
		log.Printf("Loaded %d rules from %s", len(cfg.Rules), configPath)
	}

	// Initialize global classifier with families
	if err := classifier.InitGlobalClassifier(cfg.FamiliesDir); err != nil {
		return nil, fmt.Errorf("failed to load tool families: %w", err)
	}
	if verbose {
		familiesDir := cfg.FamiliesDir
		if familiesDir == "" {
			familiesDir = "~/.config/claude-hook-guard/families"
		}
		log.Printf("Loaded tool families from %s", familiesDir)
	}

	return cfg, nil
}

func setupLogger(cfg *config.Config, verbose bool) (*logger.Logger, error) {
	logFile := cfg.Logging.File
	if logFile == "" {
		logFile = os.ExpandEnv("$HOME/.claude/claude-hook-guard.log")
	}

	decisionLogger, err := logger.New(cfg.Logging.Enabled, logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	if verbose && cfg.Logging.Enabled {
		log.Printf("Logging enabled to: %s", logFile)
	}
	return decisionLogger, nil
}

func parseHookInput(verbose bool) (hook.Input, error) {
	inputData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return hook.Input{}, fmt.Errorf("failed to read stdin: %w", err)
	}

	var input hook.Input
	if err := json.Unmarshal(inputData, &input); err != nil {
		return hook.Input{}, fmt.Errorf("failed to parse hook input: %w", err)
	}

	if verbose {
		log.Printf("Processing tool: %s", input.ToolName)
	}
	return input, nil
}

func evaluateInput(cfg *config.Config, input *hook.Input, verbose bool) (hook.Decision, string, string, error) {
	m := matcher.New(cfg)
	decision, reason, matchedRule, err := m.Evaluate(input)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to evaluate rules: %w", err)
	}

	if verbose {
		log.Printf("Decision: %s, Reason: %s, Matched: %s", decision, reason, matchedRule)
	}
	return decision, reason, matchedRule, nil
}

func outputDecision(decision hook.Decision, reason string, verbose bool) error {
	// If decision is to ignore (no rules matched), exit without output
	if decision == hook.DecisionIgnore {
		if verbose {
			log.Printf("No rules matched, passing through to Claude Code")
		}
		return nil
	}

	output := hook.Output{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       decision,
			PermissionDecisionReason: reason,
		},
	}

	outputJSON, err := output.OutputJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	fmt.Println(string(outputJSON))
	return nil
}
