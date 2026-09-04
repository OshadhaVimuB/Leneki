package binaries

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

// ChecksumFile is written by the fetch tool next to the payload, so a launch
// verifies what was extracted without rehashing the embedded copy.
const ChecksumFile = "checksums.json"

type Paths struct {
	Whisper string
	FFmpeg  string
	FFprobe string
}

// ErrNoPayload means the binaries were never fetched into this build.
var ErrNoPayload = fmt.Errorf("no bundled binaries in this build, run: go run ./tools/fetchbinaries")

// Ensure extracts the bundled binaries into binDir if they are missing or
// damaged, then returns their absolute paths.
func Ensure(binDir string) (Paths, error) {
	return ensure(payloadFS, payloadDir, binDir)
}

func ensure(src fs.FS, dir, binDir string) (Paths, error) {
	sums, err := readChecksums(src, dir)
	if err != nil {
		return Paths{}, err
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return Paths{}, err
	}
	for name, want := range sums {
		if err := extractIfNeeded(src, path.Join(dir, name), filepath.Join(binDir, name), want); err != nil {
			return Paths{}, err
		}
	}

	p := Paths{
		Whisper: filepath.Join(binDir, exeName("whisper-cli")),
		FFmpeg:  filepath.Join(binDir, exeName("ffmpeg")),
		FFprobe: filepath.Join(binDir, exeName("ffprobe")),
	}
	for _, f := range []string{p.Whisper, p.FFmpeg, p.FFprobe} {
		if _, err := os.Stat(f); err != nil {
			return Paths{}, fmt.Errorf("%s is missing from the bundled payload: %w", filepath.Base(f), ErrNoPayload)
		}
	}
	return p, nil
}

func readChecksums(src fs.FS, dir string) (map[string]string, error) {
	body, err := fs.ReadFile(src, path.Join(dir, ChecksumFile))
	if err != nil {
		return nil, ErrNoPayload
	}
	var sums map[string]string
	if err := json.Unmarshal(body, &sums); err != nil {
		return nil, fmt.Errorf("reading %s: %w", ChecksumFile, err)
	}
	if len(sums) == 0 {
		return nil, ErrNoPayload
	}
	return sums, nil
}

// A partial or corrupted extract self-heals rather than failing later with a
// confusing error from a child process.
func extractIfNeeded(src fs.FS, from, to, want string) error {
	if got, err := hashFile(to); err == nil && got == want {
		return nil
	}
	body, err := fs.ReadFile(src, from)
	if err != nil {
		return fmt.Errorf("reading embedded %s: %w", path.Base(from), err)
	}
	if got := hashBytes(body); got != want {
		return fmt.Errorf("embedded %s does not match its recorded checksum", path.Base(from))
	}
	if err := os.WriteFile(to, body, 0o755); err != nil {
		return fmt.Errorf("extracting %s: %w", path.Base(to), err)
	}
	// WriteFile respects umask, so set the bit explicitly for Unix targets.
	return os.Chmod(to, 0o755)
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

func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func exeName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}
