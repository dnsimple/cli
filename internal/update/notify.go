package update

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dnsimple/cli/internal/output"
	"github.com/fatih/color"
)

// Opts configures the update check behavior.
type Opts struct {
	CurrentVersion string
	// IsTerminal reports whether both the standard output stream and the
	// standard error stream are terminals.
	IsTerminal bool
	Debug      bool
	Args       []string
}

// ShouldCheck evaluates whether the update check should run.
func ShouldCheck(opts Opts) bool {
	if opts.CurrentVersion == "dev" {
		return false
	}
	if os.Getenv("DNSIMPLE_NO_UPDATE_CHECK") != "" {
		return false
	}
	if !opts.IsTerminal {
		return false
	}
	if isCI() {
		return false
	}
	for _, arg := range opts.Args {
		switch arg {
		case "--version", "-v", "--help", "-h", "completion":
			return false
		}
		// Stop scanning at the first non-flag argument (the subcommand)
		// unless it's a flag we're interested in.
		if !strings.HasPrefix(arg, "-") && arg != "completion" {
			break
		}
	}
	return true
}

// isCI reports whether the CLI runs on a continuous integration service.
func isCI() bool {
	return os.Getenv("CI") != "" ||
		os.Getenv("BUILD_NUMBER") != "" ||
		os.Getenv("RUN_ID") != ""
}

// CheckAsync launches a background version check and returns a channel
// that will receive the result (or nil on error/no update).
func CheckAsync(ctx context.Context, currentVersion string, debug bool) <-chan *CheckResult {
	ch := make(chan *CheckResult, 1)
	logf := debugLogger(debug)

	go func() {
		defer func() {
			// Ensure we always send something so the receiver doesn't block.
			select {
			case ch <- nil:
			default:
			}
		}()

		state, err := LoadState()
		if err != nil {
			logf("failed to load state: %v", err)
			return
		}

		var latestVersion string

		if !state.IsStale(DefaultCheckInterval) {
			latestVersion = state.LatestVersion
			logf("using cached version %s (checked %s ago)", latestVersion, time.Since(state.CheckedAt).Truncate(time.Second))
		} else {
			logf("state is stale, fetching latest version from %s", DefaultReleaseURL)

			fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			latestVersion, err = FetchLatestVersion(fetchCtx, DefaultReleaseURL)
			if err != nil {
				logf("failed to fetch latest version: %v", err)
				return
			}
			logf("fetched latest version: %s", latestVersion)

			state.LatestVersion = latestVersion
			state.CheckedAt = time.Now()
			_ = state.Save() // Best-effort save.
		}

		if latestVersion == "" || !CompareVersions(currentVersion, latestVersion) {
			logf("no update available (current=%s, latest=%s)", currentVersion, latestVersion)
			return
		}

		exe, _ := os.Executable()
		if exe != "" {
			resolved, err := filepath.EvalSymlinks(exe)
			if err == nil {
				exe = resolved
			}
		}

		method := DetectInstallMethod(exe)
		logf("update available %s -> %s (install method: %s, executable: %s)", currentVersion, latestVersion, installMethodName(method), exe)

		ch <- &CheckResult{
			CurrentVersion:  currentVersion,
			LatestVersion:   latestVersion,
			UpdateAvailable: true,
			InstallMethod:   method,
		}
	}()

	return ch
}

func debugLogger(enabled bool) func(string, ...any) {
	if !enabled {
		return func(string, ...any) {}
	}
	return func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, "[update] "+format+"\n", args...)
	}
}

func installMethodName(m InstallMethod) string {
	switch m {
	case InstallMethodHomebrew:
		return "homebrew"
	case InstallMethodScript:
		return "script"
	case InstallMethodPowerShell:
		return "powershell"
	case InstallMethodGo:
		return "go"
	default:
		return "unknown"
	}
}

// PrintNotice writes the update notification to the given writer.
func PrintNotice(w io.Writer, result *CheckResult, useColor bool) {
	if result == nil || !result.UpdateAvailable {
		return
	}

	yellow := output.NewColor(useColor, color.FgYellow)
	cyan := output.NewColor(useColor, color.FgCyan)

	fmt.Fprintf(w, "\n\n%s %s → %s\n",
		yellow.Sprint("A new release of dnsimple is available:"),
		cyan.Sprint(result.CurrentVersion),
		cyan.Sprint(result.LatestVersion))

	if cmd := UpgradeCommand(result.InstallMethod); cmd != "" {
		fmt.Fprintf(w, "To upgrade, run: %s\n", cmd)
	}

	fmt.Fprintf(w, "%s\n\n", yellow.Sprint(ReleaseURL(result.LatestVersion)))
}
