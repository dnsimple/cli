package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectInstallMethod(t *testing.T) {
	tests := []struct {
		name string
		path string
		want InstallMethod
	}{
		{"homebrew cellar", "/opt/homebrew/Cellar/dnsimple/0.3.0/bin/dnsimple", InstallMethodHomebrew},
		{"homebrew linux", "/home/linuxbrew/.linuxbrew/homebrew/bin/dnsimple", InstallMethodHomebrew},
		{"homebrew usr local", "/usr/local/Cellar/dnsimple/0.3.0/bin/dnsimple", InstallMethodHomebrew},
		{"install script", "/Users/me/.dnsimple/bin/dnsimple", InstallMethodScript},
		{"install script linux", "/home/user/.dnsimple/bin/dnsimple", InstallMethodScript},
		{"unknown path", "/usr/local/bin/dnsimple", InstallMethodUnknown},
		{"empty path", "", InstallMethodUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, DetectInstallMethod(tt.path))
		})
	}
}

func TestDetectInstallMethodGoPath(t *testing.T) {
	t.Setenv("GOPATH", "/Users/me/go")
	assert.Equal(t, InstallMethodGo, DetectInstallMethod("/Users/me/go/bin/dnsimple"))
}

func TestDetectInstallMethodGoDefault(t *testing.T) {
	t.Setenv("GOPATH", "")
	home := t.TempDir()
	t.Setenv("HOME", home)

	path := home + "/go/bin/dnsimple"
	assert.Equal(t, InstallMethodGo, DetectInstallMethod(path))
}

func TestUpgradeCommand(t *testing.T) {
	tests := []struct {
		method InstallMethod
		want   string
	}{
		{InstallMethodHomebrew, "brew upgrade dnsimple"},
		{InstallMethodScript, "curl -fsSL https://dnsimple-cli.netlify.app/install.sh | sh"},
		{InstallMethodGo, "go install github.com/dnsimple/dnsimple-cli/cmd/dnsimple@latest"},
		{InstallMethodUnknown, ""},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, UpgradeCommand(tt.method))
		})
	}
}
