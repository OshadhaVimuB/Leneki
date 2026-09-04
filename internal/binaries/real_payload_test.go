package binaries

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Skipped unless the payload has been fetched, so the check job stays green on
// a clean checkout while the build jobs exercise the real binaries.
func TestEnsureProducesRunnableBinaries(t *testing.T) {
	binDir := filepath.Join(t.TempDir(), "bin")

	p, err := Ensure(binDir)
	if errors.Is(err, ErrNoPayload) {
		t.Skip("no payload fetched, run: go run ./tools/fetchbinaries")
	}
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		path string
		args []string
		want string
	}{
		{"ffmpeg", p.FFmpeg, []string{"-version"}, "ffmpeg version"},
		{"ffprobe", p.FFprobe, []string{"-version"}, "ffprobe version"},
		{"whisper-cli", p.Whisper, []string{"--help"}, "usage:"},
	} {
		out, err := exec.Command(tc.path, tc.args...).CombinedOutput()
		if err != nil && len(out) == 0 {
			t.Errorf("%s did not run: %v", tc.name, err)
			continue
		}
		if !strings.Contains(strings.ToLower(string(out)), tc.want) {
			t.Errorf("%s output did not contain %q:\n%s", tc.name, tc.want, firstLines(string(out), 3))
		}
	}
}

func firstLines(s string, n int) string {
	lines := strings.SplitN(s, "\n", n+1)
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}
