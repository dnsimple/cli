package update

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldCheck(t *testing.T) {
	base := Opts{
		CurrentVersion: "0.3.0",
		IsTerminal:     true,
		Args:           []string{"domains", "list"},
	}

	t.Run("normal invocation", func(t *testing.T) {
		t.Setenv("CI", "")
		t.Setenv("DNSIMPLE_NO_UPDATE_CHECK", "")
		assert.True(t, ShouldCheck(base))
	})

	t.Run("dev version", func(t *testing.T) {
		opts := base
		opts.CurrentVersion = "dev"
		assert.False(t, ShouldCheck(opts))
	})

	t.Run("env var suppresses", func(t *testing.T) {
		t.Setenv("DNSIMPLE_NO_UPDATE_CHECK", "1")
		assert.False(t, ShouldCheck(base))
	})

	t.Run("non-terminal", func(t *testing.T) {
		opts := base
		opts.IsTerminal = false
		assert.False(t, ShouldCheck(opts))
	})

	t.Run("CI env var", func(t *testing.T) {
		t.Setenv("CI", "true")
		assert.False(t, ShouldCheck(base))
	})

	t.Run("version flag", func(t *testing.T) {
		opts := base
		opts.Args = []string{"--version"}
		assert.False(t, ShouldCheck(opts))
	})

	t.Run("help flag", func(t *testing.T) {
		opts := base
		opts.Args = []string{"--help"}
		assert.False(t, ShouldCheck(opts))
	})

	t.Run("completion command", func(t *testing.T) {
		opts := base
		opts.Args = []string{"completion"}
		assert.False(t, ShouldCheck(opts))
	})
}

func TestPrintNotice(t *testing.T) {
	tests := []struct {
		name     string
		result   *CheckResult
		contains []string
	}{
		{
			"nil result",
			nil,
			nil,
		},
		{
			"no update",
			&CheckResult{UpdateAvailable: false},
			nil,
		},
		{
			"homebrew",
			&CheckResult{
				CurrentVersion:  "0.3.0",
				LatestVersion:   "0.4.0",
				UpdateAvailable: true,
				InstallMethod:   InstallMethodHomebrew,
			},
			[]string{"0.3.0", "0.4.0", "brew upgrade dnsimple"},
		},
		{
			"install script",
			&CheckResult{
				CurrentVersion:  "0.3.0",
				LatestVersion:   "0.4.0",
				UpdateAvailable: true,
				InstallMethod:   InstallMethodScript,
			},
			[]string{"0.3.0", "0.4.0", "curl -fsSL"},
		},
		{
			"go install",
			&CheckResult{
				CurrentVersion:  "0.3.0",
				LatestVersion:   "0.4.0",
				UpdateAvailable: true,
				InstallMethod:   InstallMethodGo,
			},
			[]string{"0.3.0", "0.4.0", "go install"},
		},
		{
			"unknown method",
			&CheckResult{
				CurrentVersion:  "0.3.0",
				LatestVersion:   "0.4.0",
				UpdateAvailable: true,
				InstallMethod:   InstallMethodUnknown,
			},
			[]string{"0.3.0", "0.4.0", "Visit", "releases/latest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintNotice(&buf, tt.result)

			if tt.contains == nil {
				assert.Empty(t, buf.String())
				return
			}

			output := buf.String()
			for _, s := range tt.contains {
				assert.Contains(t, output, s)
			}
		})
	}
}
