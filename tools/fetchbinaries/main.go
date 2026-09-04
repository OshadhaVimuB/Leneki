// Command fetchbinaries downloads the pinned third party binaries, verifies
// them, and lays them out where go:embed expects. Run it before wails build.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

type source struct {
	Name    string   `json:"name"`
	URL     string   `json:"url"`
	SHA256  string   `json:"sha256"`
	Extract []string `json:"extract"`
}

const payloadRoot = "internal/binaries/payload"

func main() {
	target := flag.String("target", runtime.GOOS+"-"+runtime.GOARCH, "platform to fetch, for example linux-amd64")
	cache := flag.String("cache", filepath.Join(os.TempDir(), "leneki-binaries"), "where downloads are kept between runs")
	flag.Parse()

	if err := run(*target, *cache); err != nil {
		fmt.Fprintln(os.Stderr, "fetchbinaries:", err)
		os.Exit(1)
	}
}

func run(target, cache string) error {
	raw, err := os.ReadFile(filepath.Join("internal", "binaries", "sources.json"))
	if err != nil {
		return fmt.Errorf("reading sources.json, run this from the repository root: %w", err)
	}
	var all map[string][]source
	if err := json.Unmarshal(raw, &all); err != nil {
		return err
	}
	sources, ok := all[target]
	if !ok || len(sources) == 0 {
		return fmt.Errorf("no sources defined for %s, see sources.json", target)
	}

	out := filepath.Join(payloadRoot, target)
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return err
	}

	for _, s := range sources {
		archive, err := download(s, cache)
		if err != nil {
			return err
		}
		fmt.Printf("  extracting %s\n", s.Name)
		if err := extract(archive, out, s.Extract); err != nil {
			return err
		}
	}
	return writeChecksums(out)
}

func download(s source, cache string) (string, error) {
	dst := filepath.Join(cache, path.Base(s.URL))
	if got, err := hashFile(dst); err == nil && got == s.SHA256 {
		fmt.Printf("  %s already downloaded\n", s.Name)
		return dst, nil
	}

	fmt.Printf("  downloading %s\n", s.Name)
	resp, err := http.Get(s.URL)
	if err != nil {
		return "", fmt.Errorf("downloading %s: %w", s.Name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading %s: %s", s.Name, resp.Status)
	}

	part := dst + ".part"
	f, err := os.Create(part)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(f, h), resp.Body); err != nil {
		f.Close()
		return "", err
	}
	f.Close()

	if got := hex.EncodeToString(h.Sum(nil)); got != s.SHA256 {
		os.Remove(part)
		return "", fmt.Errorf("%s checksum mismatch:\n  expected %s\n  got      %s", s.Name, s.SHA256, got)
	}
	return dst, os.Rename(part, dst)
}

func writeChecksums(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	sums := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || e.Name() == ".gitkeep" || e.Name() == "checksums.json" {
			continue
		}
		sum, err := hashFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		sums[e.Name()] = sum
	}
	if len(sums) == 0 {
		return fmt.Errorf("nothing was extracted into %s, check the extract patterns", dir)
	}
	body, err := json.MarshalIndent(sums, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("  wrote %d files and their checksums to %s\n", len(sums), dir)
	return os.WriteFile(filepath.Join(dir, "checksums.json"), append(body, '\n'), 0o644)
}

func hashFile(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func matches(name string, patterns []string) bool {
	for _, p := range patterns {
		if ok, _ := path.Match(p, name); ok {
			return true
		}
	}
	return false
}

// tar handles both gzip and xz on every platform we build on, and shelling out
// to it avoids pulling an xz decoder into the module for one tool.
func untar(archive, out string, patterns []string) error {
	tmp, err := os.MkdirTemp("", "leneki-untar")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	cmd := exec.Command("tar", "-xf", archive, "-C", tmp)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("extracting %s: %v: %s", filepath.Base(archive), err, output)
	}
	return filepath.Walk(tmp, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(tmp, p)
		if err != nil {
			return err
		}
		if !matches(filepath.ToSlash(rel), patterns) {
			return nil
		}
		return copyFile(p, filepath.Join(out, filepath.Base(p)))
	})
}

func copyFile(from, to string) error {
	body, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	return os.WriteFile(to, body, 0o755)
}

func extract(archive, out string, patterns []string) error {
	if strings.HasSuffix(archive, ".zip") {
		return unzip(archive, out, patterns)
	}
	return untar(archive, out, patterns)
}
