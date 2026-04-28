package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dnsimple/cli/internal/config"
)

const (
	// DefaultReleaseURL is the base URL for release artifacts.
	DefaultReleaseURL = "https://github.com/dnsimple/homebrew-tap/releases"

	// DefaultCheckInterval is how often the CLI checks for updates.
	DefaultCheckInterval = 24 * time.Hour

	stateFileName = "state.json"
)

// State persists update check metadata between CLI invocations.
type State struct {
	LatestVersion string    `json:"latest_version"`
	CheckedAt     time.Time `json:"checked_at"`
}

// CheckResult holds the outcome of a version check.
type CheckResult struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	InstallMethod   InstallMethod
}

var versionRe = regexp.MustCompile(`/v?(\d+\.\d+\.\d+)$`)

func statePath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, stateFileName), nil
}

// LoadState reads the persisted update check state.
// Returns a zero State if the file does not exist.
func LoadState() (*State, error) {
	path, err := statePath()
	if err != nil {
		return &State{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		// Corrupted state file — treat as empty.
		return &State{}, nil
	}
	return &s, nil
}

// Save writes the state to disk.
func (s *State) Save() error {
	path, err := statePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	// Atomic write: write to temp file, then rename.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// IsStale returns true if the last check was longer ago than the given interval.
func (s *State) IsStale(interval time.Duration) bool {
	if s.CheckedAt.IsZero() {
		return true
	}
	return time.Since(s.CheckedAt) > interval
}

// FetchLatestVersion queries the release URL and extracts the latest version
// by following the /latest redirect.
func FetchLatestVersion(ctx context.Context, releaseURL string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Stop following redirects — we just need the final Location.
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, releaseURL+"/latest", nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	location := resp.Header.Get("Location")
	if location == "" {
		// Some servers may not redirect but respond with 200.
		// Fall back to the request URL itself.
		location = resp.Request.URL.String()
	}

	matches := versionRe.FindStringSubmatch(location)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not extract version from URL: %s", location)
	}
	return matches[1], nil
}

// CompareVersions returns true if latest is newer than current.
// Both must be in MAJOR.MINOR.PATCH format (no leading "v").
func CompareVersions(current, latest string) bool {
	parse := func(v string) [3]int {
		v = strings.TrimPrefix(v, "v")
		parts := strings.SplitN(v, ".", 3)
		var nums [3]int
		for i := 0; i < 3 && i < len(parts); i++ {
			n, _ := strconv.Atoi(parts[i])
			nums[i] = n
		}
		return nums
	}

	c := parse(current)
	l := parse(latest)

	for i := 0; i < 3; i++ {
		if l[i] > c[i] {
			return true
		}
		if l[i] < c[i] {
			return false
		}
	}
	return false
}
