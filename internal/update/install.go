package update

import (
	"os"
	"path/filepath"
	"strings"
)

// InstallMethod represents how the CLI was installed.
type InstallMethod int

const (
	InstallMethodUnknown InstallMethod = iota
	InstallMethodHomebrew
	InstallMethodScript
	InstallMethodPowerShell
	InstallMethodGo
	InstallMethodWinget
)

// DetectInstallMethod determines the installation method from the executable path.
// The path should be the resolved (symlink-followed) path to the binary.
func DetectInstallMethod(executablePath string) InstallMethod {
	p := filepath.ToSlash(executablePath)

	if strings.Contains(p, "/Cellar/") || strings.Contains(p, "/homebrew/") {
		return InstallMethodHomebrew
	}

	if strings.Contains(p, "/.dnsimple/bin/") && strings.HasSuffix(strings.ToLower(p), ".exe") {
		return InstallMethodPowerShell
	}

	if strings.Contains(p, "/.dnsimple/bin/") {
		return InstallMethodScript
	}

	if strings.Contains(p, "/Microsoft/WinGet/Links/") ||
		strings.Contains(p, "/WinGet/Packages/") ||
		strings.Contains(p, "/WindowsApps/") {
		return InstallMethodWinget
	}

	// Check for Go bin directories.
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		if strings.HasPrefix(p, filepath.ToSlash(gopath)+"/bin/") {
			return InstallMethodGo
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		if strings.HasPrefix(p, filepath.ToSlash(home)+"/go/bin/") {
			return InstallMethodGo
		}
	}

	return InstallMethodUnknown
}

// UpgradeCommand returns the recommended upgrade command for the given install method.
func UpgradeCommand(method InstallMethod) string {
	switch method {
	case InstallMethodHomebrew:
		return "brew upgrade dnsimple"
	case InstallMethodScript:
		return "curl -fsSL https://dnsimple-cli.netlify.app/install.sh | sh"
	case InstallMethodPowerShell:
		return "powershell -ExecutionPolicy Bypass -Command \"irm https://dnsimple-cli.netlify.app/install.ps1 | iex\""
	case InstallMethodGo:
		return "go install github.com/dnsimple/dnsimple-cli/cmd/dnsimple@latest"
	case InstallMethodWinget:
		return "winget upgrade --id DNSimple.DNSimpleCLI"
	default:
		return ""
	}
}
