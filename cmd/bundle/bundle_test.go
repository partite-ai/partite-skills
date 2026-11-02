package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(t.TempDir(), "source")
	for _, name := range []string{"skills", ".agents", ".claude-plugin", ".codex-plugin", "plugin.json", "package.json", "README.md", "LICENSE.md"} {
		base := filepath.Join(source, name)
		err := filepath.WalkDir(base, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			target := filepath.Join(root, relative)
			if entry.IsDir() {
				return os.MkdirAll(target, 0755)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, 0644)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func write(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func expectError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), message) {
		t.Fatalf("expected error containing %q, got %v", message, err)
	}
}

func TestArchivesPreserveSkillsAndHostBoundaries(t *testing.T) {
	root, output := fixture(t), t.TempDir()
	index, err := build(root, output, "")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		t.Fatal(err)
	}
	var expectedSkills []string
	for _, entry := range entries {
		expectedSkills = append(expectedSkills, entry.Name())
	}
	if !reflect.DeepEqual(index.Skills, expectedSkills) || len(index.Artifacts) != 3 {
		t.Fatalf("expected every source skill in three bundles: %+v", index)
	}
	for _, artifact := range index.Artifacts {
		t.Run(artifact.Format, func(t *testing.T) {
			archivePath := filepath.Join(output, artifact.File)
			data := read(t, archivePath)
			if artifact.SHA256 != fmtHash(data) || artifact.Bytes != int64(len(data)) {
				t.Fatal("release manifest does not describe the archive")
			}
			archive, err := zip.OpenReader(archivePath)
			if err != nil {
				t.Fatal(err)
			}
			defer archive.Close()
			names := make(map[string]bool)
			destination := t.TempDir()
			for _, file := range archive.File {
				if !fs.ValidPath(file.Name) || !file.Mode().IsRegular() || names[file.Name] {
					t.Fatalf("unsafe or duplicate archive entry: %s", file.Name)
				}
				names[file.Name] = true
				for _, prefix := range []string{"cmd/", "tests/", "scripts/", ".github/"} {
					if strings.HasPrefix(file.Name, prefix) {
						t.Fatalf("build-only file leaked into archive: %s", file.Name)
					}
				}
				stream, err := file.Open()
				if err != nil {
					t.Fatal(err)
				}
				contents, err := io.ReadAll(stream)
				closeErr := stream.Close()
				if err != nil || closeErr != nil {
					t.Fatalf("cannot read entry %s: %v / %v", file.Name, err, closeErr)
				}
				if !bytes.Equal(contents, read(t, filepath.Join(root, filepath.FromSlash(file.Name)))) {
					t.Fatalf("archive changed source content: %s", file.Name)
				}
				path := filepath.Join(destination, filepath.FromSlash(file.Name))
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				write(t, path, contents)
			}
			for _, required := range []string{"README.md", "LICENSE.md"} {
				if !names[required] {
					t.Fatalf("missing %s", required)
				}
			}
			var discovered []string
			for _, skill := range index.Skills {
				name, err := validateSkill(filepath.Join(destination, "skills", skill))
				if err != nil {
					t.Fatal(err)
				}
				discovered = append(discovered, name)
			}
			if !reflect.DeepEqual(discovered, index.Skills) {
				t.Fatal("skills did not survive extraction")
			}
			err = filepath.WalkDir(filepath.Join(root, "skills"), func(path string, entry fs.DirEntry, err error) error {
				if err != nil || entry.IsDir() {
					return err
				}
				relative, _ := filepath.Rel(root, path)
				name := filepath.ToSlash(relative)
				want := artifact.Format == "openai" || !strings.HasSuffix(name, "/agents/openai.yaml")
				if names[name] != want {
					t.Errorf("entry %s present=%t, want=%t", name, names[name], want)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if names[".claude-plugin/plugin.json"] != (artifact.Format == "anthropic") ||
				names["plugin.json"] != (artifact.Format == "openai") ||
				names[".codex-plugin/plugin.json"] != (artifact.Format == "openai") ||
				names["package.json"] != (artifact.Format == "agent-skills") {
				t.Fatal("host manifests are missing or leaked into a different bundle")
			}
			if artifact.Format == "agent-skills" {
				packageInfo, err := readJSON(filepath.Join(destination, "package.json"))
				if err != nil || !reflect.DeepEqual(field(packageInfo, "pi", "skills"), []any{"./skills"}) {
					t.Fatalf("Pi cannot discover the standard bundle: %v", err)
				}
			}
		})
	}
}

func fmtHash(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func TestReproducibleBuildsAndChecksums(t *testing.T) {
	root, first, second := fixture(t), t.TempDir(), t.TempDir()
	if _, err := build(root, first, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := build(root, second, ""); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(first)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !bytes.Equal(read(t, filepath.Join(first, entry.Name())), read(t, filepath.Join(second, entry.Name()))) {
			t.Errorf("non-deterministic output: %s", entry.Name())
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(string(read(t, filepath.Join(first, "SHA256SUMS")))), "\n") {
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 || parts[0] != fmtHash(read(t, filepath.Join(first, parts[1]))) {
			t.Fatalf("invalid checksum line: %s", line)
		}
	}
}

func TestReleaseVersionMustMatchManifests(t *testing.T) {
	root := fixture(t)
	_, err := validate(root, "9.9.9")
	expectError(t, err, "differs from package version")
	path := filepath.Join(root, ".claude-plugin/plugin.json")
	data := bytes.Replace(read(t, path), []byte(`"version": "0.1.0"`), []byte(`"version": "9.9.9"`), 1)
	write(t, path, data)
	_, err = validate(root, "")
	expectError(t, err, "version differs")
}

func TestMissingReferenceFailsBeforeWritingArtifacts(t *testing.T) {
	root := fixture(t)
	if err := os.Remove(filepath.Join(root, "skills/partite-get-started/references/first-mesh.md")); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "dist")
	_, err := build(root, output, "")
	expectError(t, err, "missing reference")
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("invalid input created release output")
	}
}

func TestRejectInvalidOrNonPortableSkills(t *testing.T) {
	cases := []struct {
		name, from, to, message string
	}{
		{"cross-skill link", "", "\n[other](../partite-get-started/SKILL.md)\n", "escapes its skill"},
		{"duplicate frontmatter", "name: partite-concepts", "name: partite-concepts\nname: hidden-name", "already defined"},
		{"missing description", "description:", "unsupported:", "unsupported frontmatter"},
		{"bad JSON example", "", "\n```json\n{bad}\n```\n", "invalid json example"},
		{"bad YAML example", "", "\n```yaml\na: 1\na: 2\n```\n", "invalid yaml example"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			root := fixture(t)
			path := filepath.Join(root, "skills/partite-concepts/SKILL.md")
			data := string(read(t, path))
			if test.from == "" {
				data += test.to
			} else {
				data = strings.Replace(data, test.from, test.to, 1)
			}
			write(t, path, []byte(data))
			_, err := validate(root, "")
			expectError(t, err, test.message)
		})
	}
}

func TestRejectSymlinksAndOutputInsideSkills(t *testing.T) {
	root := fixture(t)
	skill := filepath.Join(root, "skills/partite-concepts")
	link := filepath.Join(skill, "linked.md")
	if err := os.Symlink(filepath.Join(skill, "SKILL.md"), link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, err := validate(root, "")
	expectError(t, err, "symlinks")
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "skills", "nested", "dist")
	_, err = build(root, output, "")
	expectError(t, err, "output must not be inside skills")
	link = filepath.Join(root, "output-link")
	if err := os.Symlink(filepath.Join(root, "skills"), link); err != nil {
		t.Fatal(err)
	}
	_, err = build(root, filepath.Join(link, "new", "dist"), "")
	expectError(t, err, "output must not be inside skills")
}

func TestJSONRejectsDuplicateKeysAndTrailingData(t *testing.T) {
	for _, data := range []string{`{"x":1,"x":2}`, `{"x":{"a":1,"a":2}}`, `{} {}`, `{"broken":}`} {
		path := filepath.Join(t.TempDir(), "metadata.json")
		write(t, path, []byte(data))
		if _, err := readJSON(path); err == nil {
			t.Errorf("accepted invalid metadata: %s", data)
		}
	}
}

func TestCLIValidationAndFlagFailures(t *testing.T) {
	root := fixture(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--root", root, "--check"}, &stdout, &stderr); code != 0 {
		t.Fatalf("validation failed with %d: %s", code, &stderr)
	}
	index, err := validate(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), fmt.Sprintf("Validated %d skills", len(index.Skills))) {
		t.Fatal("missing validation result")
	}
	if _, err := os.Stat(filepath.Join(root, "dist")); !os.IsNotExist(err) {
		t.Fatal("--check wrote release output")
	}
	for _, args := range [][]string{{"--unknown"}, {"unexpected"}} {
		if code := run(args, &stdout, &stderr); code != 2 {
			t.Fatalf("invalid arguments returned %d", code)
		}
	}
	if code := run([]string{"--root", root, "--check", "--version", "9.9.9"}, &stdout, &stderr); code != 1 {
		t.Fatalf("invalid release returned %d", code)
	}
	if code := run([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("help returned %d", code)
	}
}

func TestParticleSkillsRetainLicenseInEveryBundle(t *testing.T) {
	root, output := fixture(t), t.TempDir()
	index, err := build(root, output, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range index.Artifacts {
		archive, err := zip.OpenReader(filepath.Join(output, artifact.File))
		if err != nil {
			t.Fatal(err)
		}
		for _, skill := range []string{"write-particle-python", "write-particle-typescript"} {
			licenseName := "skills/" + skill + "/LICENSE"
			file, err := archive.Open(licenseName)
			if err != nil {
				t.Fatalf("%s: missing source license for %s: %v", artifact.Format, skill, err)
			}
			license, err := io.ReadAll(file)
			file.Close()
			if err != nil || !bytes.Equal(license, read(t, filepath.Join(root, filepath.FromSlash(licenseName)))) {
				t.Fatalf("%s: source license differs for %s", artifact.Format, skill)
			}
			if !bytes.Contains(read(t, filepath.Join(root, "skills", skill, "SKILL.md")), []byte("license: GPL-3.0-or-later")) {
				t.Fatalf("%s: missing per-skill license declaration", skill)
			}
		}
		archive.Close()
	}
}

func TestManifestSchemaStaysCompatible(t *testing.T) {
	index, err := build(fixture(t), t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(index)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		t.Fatal(err)
	}
	var keys []string
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if !reflect.DeepEqual(keys, []string{"artifacts", "name", "skills", "version"}) {
		t.Fatalf("release manifest fields changed: %v", keys)
	}
}
