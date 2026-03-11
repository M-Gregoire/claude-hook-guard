package expander

import (
	"regexp"
	"strings"
)

// leadingAssignmentRe matches a leading shell variable assignment token like VAR=value
var leadingAssignmentRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=\S*\s+`)

// ExtractSubCommands parses a bash command string and returns:
//   - mainCmd: the original command with all top-level $(...) spans replaced by ""
//   - subCmds: the inner contents of each top-level $(...) expression
//
// It uses a state machine to correctly handle nesting, single-quoted strings
// (no expansion), and double-quoted strings (expansion still occurs in bash).
func ExtractSubCommands(command string) (mainCmd string, subCmds []string) {
	if !strings.Contains(command, "$(") {
		return command, nil
	}

	type state int
	const (
		stateNormal state = iota
		stateInDollarParen
		stateInSingleQuote
		stateInDoubleQuote
	)

	var (
		result       strings.Builder
		current      strings.Builder
		stateStack   []state
		currentState = stateNormal
		depth        int
	)

	push := func(s state) {
		stateStack = append(stateStack, currentState)
		currentState = s
	}
	pop := func() {
		if len(stateStack) > 0 {
			currentState = stateStack[len(stateStack)-1]
			stateStack = stateStack[:len(stateStack)-1]
		}
	}

	i := 0
	for i < len(command) {
		ch := command[i]

		switch currentState {
		case stateNormal:
			if ch == '\'' {
				result.WriteByte(ch)
				push(stateInSingleQuote)
				i++
			} else if ch == '"' {
				result.WriteByte(ch)
				push(stateInDoubleQuote)
				i++
			} else if ch == '$' && i+1 < len(command) && command[i+1] == '(' {
				// Start of $(...) substitution — skip "$(" in main output
				push(stateInDollarParen)
				depth = 1
				current.Reset()
				i += 2
			} else {
				result.WriteByte(ch)
				i++
			}

		case stateInSingleQuote:
			result.WriteByte(ch)
			if ch == '\'' {
				pop()
			}
			i++

		case stateInDoubleQuote:
			if ch == '\\' && i+1 < len(command) {
				result.WriteByte(ch)
				result.WriteByte(command[i+1])
				i += 2
			} else if ch == '"' {
				result.WriteByte(ch)
				pop()
				i++
			} else if ch == '$' && i+1 < len(command) && command[i+1] == '(' {
				// Nested $() inside double quotes — still executed by bash
				push(stateInDollarParen)
				depth = 1
				current.Reset()
				i += 2
			} else {
				result.WriteByte(ch)
				i++
			}

		case stateInDollarParen:
			if ch == '(' {
				depth++
				current.WriteByte(ch)
				i++
			} else if ch == ')' {
				depth--
				if depth == 0 {
					// End of this substitution
					sub := strings.TrimSpace(current.String())
					if sub != "" {
						subCmds = append(subCmds, sub)
					}
					pop()
				} else {
					current.WriteByte(ch)
				}
				i++
			} else {
				current.WriteByte(ch)
				i++
			}
		}
	}

	mainCmd = strings.TrimSpace(result.String())
	return mainCmd, subCmds
}

// StripLeadingAssignments removes leading shell variable assignment tokens
// (e.g., VAR=value) from the start of a command string, returning the
// actual command to be executed.
//
// Example: `AUTH="" http GET url` → `http GET url`
func StripLeadingAssignments(command string) string {
	for {
		loc := leadingAssignmentRe.FindStringIndex(command)
		if loc == nil {
			break
		}
		command = command[loc[1]:]
	}
	return strings.TrimSpace(command)
}
