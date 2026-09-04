package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func unzip(archive, out string, patterns []string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return fmt.Errorf("opening %s: %w", filepath.Base(archive), err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() || !matches(f.Name, patterns) {
			continue
		}
		if err := writeZipEntry(f, filepath.Join(out, filepath.Base(f.Name))); err != nil {
			return err
		}
	}
	return nil
}

func writeZipEntry(f *zip.File, to string) error {
	src, err := f.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(to, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("writing %s: %w", filepath.Base(to), err)
	}
	return nil
}
