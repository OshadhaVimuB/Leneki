package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupWritesToAFileInTheGivenDirectory(t *testing.T) {
	dir := t.TempDir()
	logger, closeLog, err := Setup(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("paths resolved", "data", "somewhere")
	if err := closeLog(); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(filepath.Join(dir, FileName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "paths resolved") {
		t.Errorf("log file does not contain the message: %q", body)
	}
}

func TestSetupCreatesTheDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "not", "there", "yet")
	_, closeLog, err := Setup(dir, false)
	if err != nil {
		t.Fatalf("Setup should create a missing directory: %v", err)
	}
	defer closeLog()
	if _, err := os.Stat(dir); err != nil {
		t.Error(err)
	}
}

func TestSetupAppendsRatherThanTruncating(t *testing.T) {
	dir := t.TempDir()
	l1, c1, err := Setup(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	l1.Info("first run")
	c1()

	l2, c2, err := Setup(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	l2.Info("second run")
	c2()

	body, _ := os.ReadFile(filepath.Join(dir, FileName))
	if !strings.Contains(string(body), "first run") {
		t.Error("the previous run's log was truncated")
	}
	if !strings.Contains(string(body), "second run") {
		t.Error("the current run was not written")
	}
}

func TestRotateOnlyWhenOverTheCap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	if err := os.WriteFile(path, []byte("small"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := rotateIfLarge(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".old"); err == nil {
		t.Error("a small log should not be rotated")
	}

	if err := os.WriteFile(path, make([]byte, maxLogBytes+1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := rotateIfLarge(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".old"); err != nil {
		t.Errorf("an oversized log should be rotated: %v", err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("the oversized log should have been moved, not copied")
	}
}

func TestRotateOnMissingFileIsNotAnError(t *testing.T) {
	if err := rotateIfLarge(filepath.Join(t.TempDir(), "absent.log")); err != nil {
		t.Errorf("first ever run must not fail: %v", err)
	}
}
