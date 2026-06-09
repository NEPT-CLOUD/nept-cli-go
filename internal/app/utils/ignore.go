package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var DefaultIgnores = []string{
	"node_modules", ".git", ".next", ".output", ".nuxt", ".svelte-kit",
	"dist", "build", "out", ".cache", ".turbo", "coverage", ".DS_Store",
	".env", ".env.*", "*.log", ".vscode", ".idea", "__pycache__",
	".venv", "venv", "target", ".gradle", "vendor", ".nept",
}

type IgnoreRule struct {
	Re *regexp.Regexp
}

func CompileIgnore(pattern string) IgnoreRule {
	p := strings.TrimSpace(pattern)
	p = strings.Trim(p, "/")
	anchored := strings.HasPrefix(pattern, "/")

	var sb strings.Builder
	segments := strings.Split(p, "/")
	for i, seg := range segments {
		if i > 0 {
			sb.WriteString("/")
		}
		escaped := regexp.QuoteMeta(seg)
		escaped = strings.ReplaceAll(escaped, `\*`, `[^/]*`)
		escaped = strings.ReplaceAll(escaped, `\?`, `[^/]`)
		sb.WriteString(escaped)
	}

	prefix := "(^|/)"
	if anchored {
		prefix = "^"
	}
	patternRegex := prefix + sb.String() + "($|/)"
	re := regexp.MustCompile(patternRegex)
	return IgnoreRule{Re: re}
}

func LoadIgnoreRules(dir string) ([]IgnoreRule, error) {
	patterns := append([]string{}, DefaultIgnores...)

	ignoreFiles := []string{".gitignore", ".neptignore"}
	for _, name := range ignoreFiles {
		filePath := filepath.Join(dir, name)
		file, err := os.Open(filePath)
		if err != nil {
			continue // Optional files
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
				continue
			}
			patterns = append(patterns, line)
		}
		file.Close()
	}

	rules := make([]IgnoreRule, 0, len(patterns))
	for _, pat := range patterns {
		rules = append(rules, CompileIgnore(pat))
	}
	return rules, nil
}

func IsIgnored(relPath string, rules []IgnoreRule) bool {
	normalized := strings.ReplaceAll(relPath, "\\", "/")
	for _, r := range rules {
		if r.Re.MatchString(normalized) {
			return true
		}
	}
	return false
}
