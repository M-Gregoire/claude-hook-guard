package expander_test

import (
	"testing"

	"github.com/M-Gregoire/claude-hook-guard/pkg/expander"
)

func TestExtractSubCommands(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		wantMain    string
		wantSubCmds []string
	}{
		{
			name:        "no substitution",
			command:     "http GET https://example.com",
			wantMain:    "http GET https://example.com",
			wantSubCmds: nil,
		},
		{
			name:        "simple substitution",
			command:     "echo $(date)",
			wantMain:    "echo",
			wantSubCmds: []string{"date"},
		},
		{
			name:        "variable assignment with substitution",
			command:     `AUTH="$(ddtool auth token foo --http-header)" http GET https://example.com "$AUTH"`,
			wantMain:    `AUTH="" http GET https://example.com "$AUTH"`,
			wantSubCmds: []string{"ddtool auth token foo --http-header"},
		},
		{
			name:        "multiple substitutions",
			command:     "curl $(get-url) --header $(get-auth)",
			wantMain:    "curl  --header",
			wantSubCmds: []string{"get-url", "get-auth"},
		},
		{
			name:        "nested substitution",
			command:     "echo $(cat $(find . -name foo))",
			wantMain:    "echo",
			wantSubCmds: []string{"cat $(find . -name foo)"},
		},
		{
			name:        "in double quotes",
			command:     `echo "$(date)"`,
			wantMain:    `echo ""`,
			wantSubCmds: []string{"date"},
		},
		{
			name:        "in single quotes - no expansion",
			command:     "echo '$(date)'",
			wantMain:    "echo '$(date)'",
			wantSubCmds: nil,
		},
		{
			name:        "empty substitution skipped",
			command:     "cmd $()",
			wantMain:    "cmd",
			wantSubCmds: nil,
		},
		{
			name:        "substitution with flags",
			command:     `AUTH="$(ddtool auth token rapid-seceng-cloud-security --datacenter us3.staging.dog --http-header)" http --ignore-stdin GET "https://scm.us3.staging.dog/health" "$AUTH"`,
			wantMain:    `AUTH="" http --ignore-stdin GET "https://scm.us3.staging.dog/health" "$AUTH"`,
			wantSubCmds: []string{"ddtool auth token rapid-seceng-cloud-security --datacenter us3.staging.dog --http-header"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMain, gotSubs := expander.ExtractSubCommands(tt.command)
			if gotMain != tt.wantMain {
				t.Errorf("mainCmd = %q, want %q", gotMain, tt.wantMain)
			}
			if len(gotSubs) != len(tt.wantSubCmds) {
				t.Errorf("subCmds = %v, want %v", gotSubs, tt.wantSubCmds)
				return
			}
			for i, sub := range gotSubs {
				if sub != tt.wantSubCmds[i] {
					t.Errorf("subCmds[%d] = %q, want %q", i, sub, tt.wantSubCmds[i])
				}
			}
		})
	}
}

func TestStripLeadingAssignments(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    string
	}{
		{
			name:    "no assignment",
			command: "http GET url",
			want:    "http GET url",
		},
		{
			name:    "single assignment",
			command: `AUTH="" http GET url`,
			want:    "http GET url",
		},
		{
			name:    "multiple assignments",
			command: `FOO=bar BAZ=qux http GET url`,
			want:    "http GET url",
		},
		{
			name:    "assignment only",
			command: `FOO=bar `,
			want:    "",
		},
		{
			name:    "empty string",
			command: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expander.StripLeadingAssignments(tt.command)
			if got != tt.want {
				t.Errorf("StripLeadingAssignments(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}
