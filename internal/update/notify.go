package update

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Opts configures the update check behavior.
type Opts struct {
	CurrentVersion string
	IsTerminal     bool
	Quiet          bool
	Args           []string
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
	if opts.Quiet {
		return false
	}
	if os.Getenv("CI") != "" {
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

// CheckAsync launches a background version check and returns a channel
// that will receive the result (or nil on error/no update).
func CheckAsync(ctx context.Context, currentVersion string) <-chan *CheckResult {
	ch := make(chan *CheckResult, 1)

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
			return
		}

		var latestVersion string

		if !state.IsStale(DefaultCheckInterval) {
			// Use cached version.
			latestVersion = state.LatestVersion
		} else {
			fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			latestVersion, err = FetchLatestVersion(fetchCtx, DefaultReleaseURL)
			if err != nil {
				return
			}

			state.LatestVersion = latestVersion
			state.CheckedAt = time.Now()
			_ = state.Save() // Best-effort save.
		}

		if latestVersion == "" || !CompareVersions(currentVersion, latestVersion) {
			return
		}

		exe, _ := os.Executable()
		if exe != "" {
			resolved, err := filepath.EvalSymlinks(exe)
			if err == nil {
				exe = resolved
			}
		}

		ch <- &CheckResult{
			CurrentVersion:  currentVersion,
			LatestVersion:   latestVersion,
			UpdateAvailable: true,
			InstallMethod:   DetectInstallMethod(exe),
		}
	}()

	return ch
}

// PrintNotice writes the update notification to the given writer.
func PrintNotice(w io.Writer, result *CheckResult) {
	if result == nil || !result.UpdateAvailable {
		return
	}

	fmt.Fprintf(w, "\nA new version of the DNSimple CLI is available: %s → %s\n",
		result.CurrentVersion, result.LatestVersion)

	if cmd := UpgradeCommand(result.InstallMethod); cmd != "" {
		fmt.Fprintf(w, "To upgrade, run: %s\n", cmd)
	} else {
		fmt.Fprintf(w, "Visit %s/latest to upgrade.\n", DefaultReleaseURL)
	}
}
