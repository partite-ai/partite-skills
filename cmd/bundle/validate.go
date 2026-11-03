package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

const pluginName = "partite-ai"

const portableSchema = "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json"

var (
	versionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`)
	namePattern    = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	frontmatter    = regexp.MustCompile(`(?s)^---\n(.*?)\n---(?:\n|$)`)
	markdownLink   = regexp.MustCompile(`\]\(([^)]+)\)`)
	codeBlock      = regexp.MustCompile("(?s)```(json|yaml)\\n(.*?)```")
	unfinished     = regexp.MustCompile(`\[(TODO|FIXME):`)
)

func text(value any, label string, limit int) error {
	s, ok := value.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return fmt.Errorf("%s: expected non-empty text", label)
	}
	if limit > 0 && utf8.RuneCountInString(s) > limit {
		return fmt.Errorf("%s: exceeds %d characters", label, limit)
	}
	return nil
}

// JSON's default decoder silently replaces duplicate object keys. Release
// metadata must instead reject them, including keys in nested objects.
func jsonValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch token {
	case json.Delim('{'):
		object := make(map[string]any)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil, fmt.Errorf("expected a JSON object key")
			}
			if _, exists := object[key]; exists {
				return nil, fmt.Errorf("duplicate JSON key: %s", key)
			}
			value, err := jsonValue(decoder)
			if err != nil {
				return nil, err
			}
			object[key] = value
		}
		_, err := decoder.Token()
		return object, err
	case json.Delim('['):
		values := []any{}
		for decoder.More() {
			value, err := jsonValue(decoder)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		_, err := decoder.Token()
		return values, err
	default:
		if _, delimiter := token.(json.Delim); delimiter {
			return nil, fmt.Errorf("unexpected JSON delimiter %v", token)
		}
		return token, nil
	}
}

func readJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	value, err := jsonValue(decoder)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("%s: expected one JSON object", path)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: expected a JSON object", path)
	}
	return object, nil
}

func field(object map[string]any, path ...string) any {
	var current any = object
	for _, key := range path {
		mapping, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = mapping[key]
	}
	return current
}

// Resolve existing ancestors as well as the final path, so a new output
// directory beneath a symlink cannot bypass containment checks.
func resolved(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var suffix []string
	for {
		canonical, err := filepath.EvalSymlinks(abs)
		if err == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				canonical = filepath.Join(canonical, suffix[i])
			}
			return canonical, nil
		}
		if !os.IsNotExist(err) || filepath.Dir(abs) == abs {
			return "", err
		}
		suffix = append(suffix, filepath.Base(abs))
		abs = filepath.Dir(abs)
	}
}

func inside(path, parent string) bool {
	path, err := resolved(path)
	if err != nil {
		return false
	}
	parent, err = resolved(parent)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(parent, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func yamlObject(data []byte) (map[string]any, error) {
	var object map[string]any
	if err := yaml.Unmarshal(data, &object); err != nil {
		return nil, err
	}
	if object == nil {
		return nil, fmt.Errorf("expected a YAML mapping")
	}
	return object, nil
}

func validateSkill(skill string) (string, error) {
	entry := filepath.Join(skill, "SKILL.md")
	data, err := os.ReadFile(entry)
	if err != nil {
		return "", err
	}
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	match := frontmatter.FindSubmatchIndex(data)
	if match == nil {
		return "", fmt.Errorf("%s: missing YAML frontmatter", entry)
	}
	metadata, err := yamlObject(data[match[2]:match[3]])
	if err != nil {
		return "", fmt.Errorf("%s: %w", entry, err)
	}
	allowed := map[string]bool{"name": true, "description": true, "license": true, "compatibility": true, "metadata": true, "allowed-tools": true}
	for key := range metadata {
		if !allowed[key] {
			return "", fmt.Errorf("%s: unsupported frontmatter field %s", entry, key)
		}
	}
	for key, limit := range map[string]int{"name": 64, "description": 1024, "license": 0, "compatibility": 500, "allowed-tools": 0} {
		value, exists := metadata[key]
		if exists || key == "name" || key == "description" {
			if err := text(value, entry+": "+key, limit); err != nil {
				return "", err
			}
		}
	}
	name := metadata["name"].(string)
	if !namePattern.MatchString(name) || name != filepath.Base(skill) {
		return "", fmt.Errorf("%s: name must be lowercase kebab-case and match its folder", entry)
	}
	if value, exists := metadata["metadata"]; exists {
		mapping, ok := value.(map[string]any)
		if !ok {
			return "", fmt.Errorf("%s: metadata must be a string mapping", entry)
		}
		for _, value := range mapping {
			if _, ok := value.(string); !ok {
				return "", fmt.Errorf("%s: metadata values must be strings", entry)
			}
		}
	}
	if len(bytes.TrimSpace(data[match[1]:])) == 0 {
		return "", fmt.Errorf("%s: missing instructions", entry)
	}
	err = filepath.WalkDir(skill, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s: symlinks are not portable", path)
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("%s: unsupported file type", path)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !utf8.Valid(content) || bytes.ContainsRune(content, 0) {
			return fmt.Errorf("%s: binary content is not supported", path)
		}
		if filepath.Ext(path) == ".md" {
			return validateMarkdown(path, skill, content)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	uiPath := filepath.Join(skill, "agents", "openai.yaml")
	if data, err := os.ReadFile(uiPath); err == nil {
		ui, err := yamlObject(data)
		if err != nil {
			return "", fmt.Errorf("%s: %w", uiPath, err)
		}
		if err := text(field(ui, "interface", "display_name"), uiPath+": display_name", 0); err != nil {
			return "", err
		}
		short := field(ui, "interface", "short_description")
		if err := text(short, uiPath+": short_description", 64); err != nil {
			return "", err
		}
		if utf8.RuneCountInString(short.(string)) < 25 {
			return "", fmt.Errorf("%s: short_description is too short", uiPath)
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}
	return name, nil
}

func validateMarkdown(path, skill string, content []byte) error {
	if unfinished.Match(content) {
		return fmt.Errorf("%s: unfinished scaffold", path)
	}
	for _, match := range markdownLink.FindAllSubmatch(content, -1) {
		link, err := url.Parse(strings.Trim(string(match[1]), "<>"))
		if err != nil {
			return fmt.Errorf("%s: invalid reference: %w", path, err)
		}
		if link.Scheme != "" || link.Path == "" {
			continue
		}
		linked := filepath.Join(filepath.Dir(path), filepath.FromSlash(link.Path))
		if filepath.IsAbs(link.Path) || link.Host != "" || !inside(linked, skill) {
			return fmt.Errorf("%s: reference escapes its skill: %s", path, match[1])
		}
		if info, err := os.Stat(linked); err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("%s: missing reference: %s", path, match[1])
		}
	}
	for _, match := range codeBlock.FindAllSubmatch(content, -1) {
		var err error
		if string(match[1]) == "json" {
			if !json.Valid(match[2]) {
				err = fmt.Errorf("invalid JSON")
			}
		} else {
			var value any
			err = yaml.Unmarshal(match[2], &value)
		}
		if err != nil {
			return fmt.Errorf("%s: invalid %s example: %w", path, match[1], err)
		}
	}
	return nil
}

func validate(root, expectedVersion string) (manifest, error) {
	packageInfo, err := readJSON(filepath.Join(root, "package.json"))
	if err != nil {
		return manifest{}, err
	}
	if err := text(packageInfo["version"], "package version", 0); err != nil {
		return manifest{}, err
	}
	version := packageInfo["version"].(string)
	if !versionPattern.MatchString(version) {
		return manifest{}, fmt.Errorf("version must be MAJOR.MINOR.PATCH with an optional prerelease")
	}
	if expectedVersion != "" && expectedVersion != version {
		return manifest{}, fmt.Errorf("release version %s differs from package version %s", expectedVersion, version)
	}
	if packageInfo["private"] != true {
		return manifest{}, fmt.Errorf("package must remain private; publishing uses GitHub releases")
	}
	if !reflect.DeepEqual(packageInfo["pi"], map[string]any{"skills": []any{"./skills"}}) {
		return manifest{}, fmt.Errorf("Pi must load only ./skills")
	}
	if scripts := packageInfo["scripts"]; scripts != nil && !reflect.DeepEqual(scripts, map[string]any{}) {
		return manifest{}, fmt.Errorf("the installable package must not run lifecycle scripts")
	}
	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		return manifest{}, err
	}
	index := manifest{Name: pluginName, Version: version}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			return manifest{}, fmt.Errorf("%s: expected a skill directory", entry.Name())
		}
		name, err := validateSkill(filepath.Join(root, "skills", entry.Name()))
		if err != nil {
			return manifest{}, err
		}
		index.Skills = append(index.Skills, name)
	}
	if len(index.Skills) == 0 {
		return manifest{}, fmt.Errorf("no skills found")
	}
	if err := validateManifests(root, version, packageInfo); err != nil {
		return manifest{}, err
	}
	for _, name := range []string{"README.md", "LICENSE.md"} {
		if info, err := os.Stat(filepath.Join(root, name)); err != nil || !info.Mode().IsRegular() {
			return manifest{}, fmt.Errorf("missing %s", name)
		}
	}
	return index, nil
}

func validateManifests(root, version string, packageInfo map[string]any) error {
	portable, err := readJSON(filepath.Join(root, "plugin.json"))
	if err != nil {
		return err
	}
	allowed := map[string]bool{"$schema": true, "name": true, "version": true, "description": true, "author": true,
		"homepage": true, "repository": true, "license": true, "keywords": true, "extensions": true}
	for key := range portable {
		if !allowed[key] {
			return fmt.Errorf("portable plugin has unsupported field %s", key)
		}
	}
	if portable["$schema"] != portableSchema {
		return fmt.Errorf("unexpected Agent Plugins schema")
	}
	for _, filename := range []string{"plugin.json", ".claude-plugin/plugin.json", ".codex-plugin/plugin.json"} {
		plugin, err := readJSON(filepath.Join(root, filename))
		if err != nil {
			return err
		}
		if plugin["name"] != pluginName || plugin["version"] != version {
			return fmt.Errorf("%s: plugin name or version differs from package.json", filename)
		}
		if err := text(plugin["description"], filename+": description", 0); err != nil {
			return err
		}
		repository, _ := field(packageInfo, "repository", "url").(string)
		if field(plugin, "author", "name") != "Partite.ai" || plugin["license"] != packageInfo["license"] || plugin["repository"] != strings.TrimSuffix(repository, ".git") {
			return fmt.Errorf("%s: inconsistent author, license, or repository", filename)
		}
		for _, key := range []string{"mcpServers", "hooks", "apps", "agents", "commands"} {
			if _, exists := plugin[key]; exists {
				return fmt.Errorf("%s: this is a skills-only plugin", filename)
			}
		}
		if filename == ".codex-plugin/plugin.json" && plugin["skills"] != "./skills/" {
			return fmt.Errorf("Codex skills path differs")
		}
	}
	onboarding := "./skills/partite-get-started/SKILL.md"
	if field(portable, "extensions", "com.openai", "onboardingSkill") != onboarding {
		return fmt.Errorf("OpenAI onboarding must reference the getting-started skill")
	}
	if info, err := os.Stat(filepath.Join(root, onboarding)); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("onboarding skill is absent")
	}
	for _, filename := range []string{".claude-plugin/marketplace.json", ".agents/plugins/marketplace.json"} {
		catalog, err := readJSON(filepath.Join(root, filename))
		if err != nil {
			return err
		}
		entries, ok := catalog["plugins"].([]any)
		if catalog["name"] != "partite-skills" || !ok || len(entries) != 1 {
			return fmt.Errorf("%s: expected the partite-skills marketplace with one plugin", filename)
		}
		entry, ok := entries[0].(map[string]any)
		if !ok || entry["name"] != pluginName {
			return fmt.Errorf("%s: plugin name differs", filename)
		}
		if strings.HasPrefix(filename, ".claude") {
			if entry["source"] != "./" || field(catalog, "owner", "name") != "Partite.ai" {
				return fmt.Errorf("%s: invalid marketplace owner or root source", filename)
			}
		} else {
			if !reflect.DeepEqual(entry["source"], map[string]any{"source": "local", "path": "./"}) ||
				!reflect.DeepEqual(entry["policy"], map[string]any{"installation": "AVAILABLE", "authentication": "ON_USE"}) {
				return fmt.Errorf("%s: invalid root source or install policy", filename)
			}
			if err := text(entry["category"], "OpenAI marketplace category", 0); err != nil {
				return err
			}
		}
	}
	return nil
}
