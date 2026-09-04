package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewPutsEverythingUnderTheRightRoot(t *testing.T) {
	p := New(filepath.Join("d", "data"), filepath.Join("c", "cache"))

	underData := []string{p.DB, p.Models, p.Playback}
	for _, path := range underData {
		if !strings.HasPrefix(path, p.Data) {
			t.Errorf("%s should live under the data root %s", path, p.Data)
		}
	}
	underCache := []string{p.Bin, p.Temp}
	for _, path := range underCache {
		if !strings.HasPrefix(path, p.Cache) {
			t.Errorf("%s should live under the cache root %s", path, p.Cache)
		}
	}
}

func TestEnsureCreatesEveryDirectory(t *testing.T) {
	root := t.TempDir()
	p := New(filepath.Join(root, "data"), filepath.Join(root, "cache"))

	if err := p.Ensure(); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{p.Data, p.Cache, p.Models, p.Playback, p.Bin, p.Temp} {
		info, err := os.Stat(d)
		if err != nil {
			t.Errorf("%s was not created: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", d)
		}
	}
}

func TestEnsureIsRepeatable(t *testing.T) {
	root := t.TempDir()
	p := New(filepath.Join(root, "data"), filepath.Join(root, "cache"))
	if err := p.Ensure(); err != nil {
		t.Fatal(err)
	}
	if err := p.Ensure(); err != nil {
		t.Fatalf("second Ensure failed, so startup would fail on every run but the first: %v", err)
	}
}

func TestResolveMatchesTheDocumentedLayout(t *testing.T) {
	p, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(p.Data, appDir) {
		t.Errorf("data dir %q should end in %q", p.Data, appDir)
	}
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(p.Cache, filepath.Join(appDir, "Cache")) {
			t.Errorf("windows cache dir %q should end in %q", p.Cache, filepath.Join(appDir, "Cache"))
		}
	} else if !strings.HasSuffix(p.Cache, appDir) {
		t.Errorf("cache dir %q should end in %q", p.Cache, appDir)
	}
	if p.Data == p.Cache {
		t.Error("data and cache must be separate, they have different lifetimes")
	}
}

func TestLinuxDataDirPrefersXDG(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	t.Setenv("XDG_DATA_HOME", "/tmp/xdg-example")
	d, err := dataDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/xdg-example", appDir)
	if d != want {
		t.Errorf("dataDir() = %q, want %q", d, want)
	}
}

func TestDefaultsPointAtResolvedPaths(t *testing.T) {
	p := New("data", "cache")
	s := Defaults(p)
	if s.ModelDir != p.Models {
		t.Errorf("ModelDir = %q, want %q", s.ModelDir, p.Models)
	}
	if s.TempDir != p.Temp {
		t.Errorf("TempDir = %q, want %q", s.TempDir, p.Temp)
	}
	if s.Threads < 1 {
		t.Errorf("Threads = %d, want at least 1", s.Threads)
	}
}
