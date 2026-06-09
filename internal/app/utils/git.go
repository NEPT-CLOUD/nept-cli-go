package utils

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
)

type GitInfo struct {
	Branch        string
	Commit        string
	CommitMessage string
	FullRepoName  string
}

func gitOut(dir string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

var repoNameRegex = regexp.MustCompile(`github\.com[:/](.+?)(?:\.git)?\/?$`)

func parseRepoName(url string) string {
	m := repoNameRegex.FindStringSubmatch(url)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

func ResolveGitInfo(dir string) GitInfo {
	branch := gitOut(dir, "rev-parse", "--abbrev-ref", "HEAD")
	commit := gitOut(dir, "rev-parse", "HEAD")
	msg := gitOut(dir, "log", "-1", "--pretty=%s")
	remote := gitOut(dir, "config", "--get", "remote.origin.url")

	return GitInfo{
		Branch:        branch,
		Commit:        commit,
		CommitMessage: msg,
		FullRepoName:  parseRepoName(remote),
	}
}
