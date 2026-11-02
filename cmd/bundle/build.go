package main

import (
	"archive/zip"
	"compress/flate"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type artifact struct {
	Format string `json:"format"`
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}

type manifest struct {
	Name      string     `json:"name"`
	Version   string     `json:"version"`
	Skills    []string   `json:"skills"`
	Artifacts []artifact `json:"artifacts"`
}

type bundleFormat struct {
	name  string
	files []string
}

var formats = []bundleFormat{
	{"anthropic", []string{".claude-plugin/plugin.json", ".claude-plugin/marketplace.json"}},
	{"openai", []string{"plugin.json", ".codex-plugin/plugin.json", ".agents/plugins/marketplace.json"}},
	{"agent-skills", []string{"package.json"}},
}

func bundleFiles(root string, format bundleFormat) ([]string, error) {
	paths := append([]string{"README.md", "LICENSE.md"}, format.files...)
	err := filepath.WalkDir(filepath.Join(root, "skills"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(relative)
		if format.name != "openai" && strings.HasSuffix(name, "/agents/openai.yaml") {
			return nil
		}
		paths = append(paths, name)
		return nil
	})
	sort.Strings(paths)
	return paths, err
}

func writeArchive(root, destination string, paths []string) (err error) {
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(destination)
		}
	}()
	archive := zip.NewWriter(file)
	archive.RegisterCompressor(zip.Deflate, func(writer io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(writer, flate.BestCompression)
	})
	for _, name := range paths {
		data, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if readErr != nil {
			return readErr
		}
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		header.SetMode(0644)
		header.Modified = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
		writer, writeErr := archive.CreateHeader(header)
		if writeErr != nil {
			return writeErr
		}
		if _, writeErr = writer.Write(data); writeErr != nil {
			return writeErr
		}
	}
	return archive.Close()
}

func build(root, output, expectedVersion string) (manifest, error) {
	index, err := validate(root, expectedVersion)
	if err != nil {
		return manifest{}, err
	}
	if output == "" {
		output = filepath.Join(root, "dist")
	}
	if inside(output, filepath.Join(root, "skills")) {
		return manifest{}, fmt.Errorf("output must not be inside skills")
	}
	if err := os.MkdirAll(output, 0755); err != nil {
		return manifest{}, err
	}
	var sums strings.Builder
	for _, format := range formats {
		paths, err := bundleFiles(root, format)
		if err != nil {
			return manifest{}, err
		}
		filename := fmt.Sprintf("partite-skills-%s-%s.zip", format.name, index.Version)
		path := filepath.Join(output, filename)
		if err := writeArchive(root, path, paths); err != nil {
			return manifest{}, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return manifest{}, err
		}
		digest := fmt.Sprintf("%x", sha256.Sum256(data))
		index.Artifacts = append(index.Artifacts, artifact{format.name, filename, digest, int64(len(data))})
		fmt.Fprintf(&sums, "%s  %s\n", digest, filename)
	}
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return manifest{}, err
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(output, "manifest.json"), data, 0644); err != nil {
		return manifest{}, err
	}
	fmt.Fprintf(&sums, "%x  manifest.json\n", sha256.Sum256(data))
	if err := os.WriteFile(filepath.Join(output, "SHA256SUMS"), []byte(sums.String()), 0644); err != nil {
		return manifest{}, err
	}
	return index, nil
}
