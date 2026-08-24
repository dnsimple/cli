package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func isolateConfigHome(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
}

func TestLoadStateMissingFile(t *testing.T) {
	isolateConfigHome(t)

	state, err := LoadState()
	assert.NoError(t, err)
	assert.Empty(t, state.LatestVersion)
	assert.True(t, state.CheckedAt.IsZero())
}

func TestStateSaveAndLoad(t *testing.T) {
	isolateConfigHome(t)

	now := time.Now().Truncate(time.Second)
	state := &State{
		LatestVersion: "1.2.3",
		CheckedAt:     now,
	}
	assert.NoError(t, state.Save())

	loaded, err := LoadState()
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3", loaded.LatestVersion)
	assert.Equal(t, now.UTC(), loaded.CheckedAt.UTC())
}

func TestIsStale(t *testing.T) {
	tests := []struct {
		name      string
		checkedAt time.Time
		interval  time.Duration
		want      bool
	}{
		{"zero time is stale", time.Time{}, 24 * time.Hour, true},
		{"recent is not stale", time.Now().Add(-1 * time.Hour), 24 * time.Hour, false},
		{"old is stale", time.Now().Add(-25 * time.Hour), 24 * time.Hour, true},
		{"exactly at interval is stale", time.Now().Add(-24*time.Hour - time.Second), 24 * time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{CheckedAt: tt.checkedAt}
			assert.Equal(t, tt.want, s.IsStale(tt.interval))
		})
	}
}

func TestFetchLatestVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/releases/latest" {
			http.Redirect(w, r, "/releases/tag/v0.5.0", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	version, err := FetchLatestVersion(context.Background(), server.URL+"/releases")
	assert.NoError(t, err)
	assert.Equal(t, "0.5.0", version)
}

func TestFetchLatestVersionNoRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := FetchLatestVersion(context.Background(), server.URL+"/releases")
	assert.Error(t, err)
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"0.3.0", "0.4.0", true},
		{"0.3.0", "0.3.0", false},
		{"0.4.0", "0.3.0", false},
		{"1.0.0", "0.9.0", false},
		{"0.9.0", "1.0.0", true},
		{"1.2.3", "1.2.4", true},
		{"1.2.4", "1.2.3", false},
		{"1.2.3", "1.3.0", true},
		{"v0.3.0", "v0.4.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.current+"_vs_"+tt.latest, func(t *testing.T) {
			assert.Equal(t, tt.want, CompareVersions(tt.current, tt.latest))
		})
	}
}

func TestReleaseURL(t *testing.T) {
	assert.Equal(t, "https://github.com/dnsimple/homebrew-tap/releases/tag/v0.4.0", ReleaseURL("0.4.0"))
	assert.Equal(t, "https://github.com/dnsimple/homebrew-tap/releases/tag/v0.4.0", ReleaseURL("v0.4.0"))
}
