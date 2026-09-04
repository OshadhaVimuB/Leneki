package binaries

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
)

const dir = "payload/test"

func fakePayload(t *testing.T) fstest.MapFS {
	t.Helper()
	contents := map[string][]byte{
		exeName("whisper-cli"): []byte("whisper binary"),
		exeName("ffmpeg"):      []byte("ffmpeg binary"),
		exeName("ffprobe"):     []byte("ffprobe binary"),
	}
	sums := map[string]string{}
	fsys := fstest.MapFS{}
	for name, body := range contents {
		fsys[dir+"/"+name] = &fstest.MapFile{Data: body}
		sums[name] = hashBytes(body)
	}
	encoded, err := json.Marshal(sums)
	if err != nil {
		t.Fatal(err)
	}
	fsys[dir+"/"+ChecksumFile] = &fstest.MapFile{Data: encoded}
	return fsys
}

func TestEnsureExtractsEverythingAndReturnsPaths(t *testing.T) {
	binDir := filepath.Join(t.TempDir(), "bin")
	p, err := ensure(fakePayload(t), dir, binDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{p.Whisper, p.FFmpeg, p.FFprobe} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("%s was not extracted: %v", f, err)
		}
	}
	body, _ := os.ReadFile(p.FFmpeg)
	if string(body) != "ffmpeg binary" {
		t.Errorf("ffmpeg content = %q", body)
	}
}

func TestEnsureCreatesTheDirectory(t *testing.T) {
	binDir := filepath.Join(t.TempDir(), "deep", "not", "there")
	if _, err := ensure(fakePayload(t), dir, binDir); err != nil {
		t.Fatalf("Ensure should create a missing bin directory: %v", err)
	}
}

func TestCorruptedExtractSelfHeals(t *testing.T) {
	binDir := filepath.Join(t.TempDir(), "bin")
	payload := fakePayload(t)
	p, err := ensure(payload, dir, binDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p.FFmpeg, []byte("truncated"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := ensure(payload, dir, binDir); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(p.FFmpeg)
	if string(body) != "ffmpeg binary" {
		t.Errorf("a damaged binary was not repaired, got %q", body)
	}
}

func TestEnsureLeavesGoodFilesAlone(t *testing.T) {
	binDir := filepath.Join(t.TempDir(), "bin")
	payload := fakePayload(t)
	p, err := ensure(payload, dir, binDir)
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(p.Whisper)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ensure(payload, dir, binDir); err != nil {
		t.Fatal(err)
	}
	after, _ := os.Stat(p.Whisper)
	if !before.ModTime().Equal(after.ModTime()) {
		t.Error("an intact binary was rewritten, so every launch pays the extraction cost")
	}
}

func TestMissingPayloadIsReportedClearly(t *testing.T) {
	_, err := ensure(fstest.MapFS{}, dir, t.TempDir())
	if !errors.Is(err, ErrNoPayload) {
		t.Fatalf("err = %v, want ErrNoPayload", err)
	}
	if !strings.Contains(err.Error(), "fetchbinaries") {
		t.Errorf("the error should say how to fix it, got: %v", err)
	}
}

func TestEmbeddedFileThatDoesNotMatchItsChecksumIsRefused(t *testing.T) {
	payload := fakePayload(t)
	payload[dir+"/"+exeName("ffmpeg")] = &fstest.MapFile{Data: []byte("tampered")}

	_, err := ensure(payload, dir, t.TempDir())
	if err == nil {
		t.Fatal("a payload that disagrees with its checksums should not be extracted")
	}
}

func TestExtractedBinariesAreExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not meaningful on Windows")
	}
	binDir := filepath.Join(t.TempDir(), "bin")
	p, err := ensure(fakePayload(t), dir, binDir)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p.Whisper)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("mode = %v, want the executable bit set", info.Mode())
	}
}
