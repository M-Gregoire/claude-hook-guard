package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/M-Gregoire/claude-hook-guard/pkg/hook"
)

// Logger handles decision logging
type Logger struct {
	enabled bool
	file    *os.File
}

// New creates a new logger
func New(enabled bool, logPath string) (*Logger, error) {
	if !enabled {
		return &Logger{enabled: false}, nil
	}

	// Expand environment variables in path
	logPath = os.ExpandEnv(logPath)

	// Open log file in append mode
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		enabled: true,
		file:    file,
	}, nil
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
	CWD       string                 `json:"cwd,omitempty"`
	Decision  hook.Decision          `json:"decision"`
	Reason    string                 `json:"reason"`
	MatchedBy string                 `json:"matched_by,omitempty"`
}

// Log logs a decision
func (l *Logger) Log(input *hook.Input, decision hook.Decision, reason string, matchedRule string) error {
	if !l.enabled || l.file == nil {
		return nil
	}

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		ToolName:  input.ToolName,
		ToolInput: input.ToolInput,
		CWD:       input.CWD,
		Decision:  decision,
		Reason:    reason,
		MatchedBy: matchedRule,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	if _, err := l.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	return nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
