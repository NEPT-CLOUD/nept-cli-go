package utls

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ZipResult struct {
	Base64    string
	FileCount int
	Bytes     int
}

func ZipDirectory(dir string) (ZipResult, error) {
	rules, err := LoadIgnoreRules(dir)
	if err != nil {
		return ZipResult{}, err
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	fileCount := 0

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		// Check ignore rules
		if IsIgnored(rel, rules) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		fileCount++

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		zipFileHeader, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		zipFileHeader.Name = filepath.ToSlash(rel)
		zipFileHeader.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(zipFileHeader)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return ZipResult{}, err
	}

	err = zipWriter.Close()
	if err != nil {
		return ZipResult{}, err
	}

	if fileCount == 0 {
		return ZipResult{}, fmt.Errorf("no files to deploy after applying ignore rules")
	}

	zipBytes := buf.Bytes()
	b64 := base64.StdEncoding.EncodeToString(zipBytes)

	return ZipResult{
		Base64:    b64,
		FileCount: fileCount,
		Bytes:     len(zipBytes),
	}, nil
}

func HumanBytes(n int) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	if n < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
}
