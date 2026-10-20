package git

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
)

// StatusFile represents a changed file detected via git status.
type StatusFile struct {
	StagingCode  byte
	WorktreeCode byte
	Path         string
}

// GetStatus returns porcelain status entries for the given repo root.
func GetStatus(repoRoot string) ([]StatusFile, error) {
	cmd := exec.Command("git", "-C", repoRoot, "status", "--porcelain")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(stdout.String(), "\n")
	var results []StatusFile
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		st := line[0]
		wt := line[1]
		rest := strings.TrimSpace(line[3:])
		// Handle rename format: "old -> new"
		if idx := strings.Index(rest, "->"); idx != -1 {
			rest = strings.TrimSpace(rest[idx+2:])
		}
		results = append(results, StatusFile{
			StagingCode:  st,
			WorktreeCode: wt,
			Path:         filepath.ToSlash(rest),
		})
	}
	return results, nil
}

// GetStagedFiles returns staged file paths matching diff-filter=ACMR.
func GetStagedFiles(repoRoot string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoRoot, "diff", "--cached", "--name-only", "--diff-filter=ACMR")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(stdout.String(), "\n")
	var results []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			results = append(results, filepath.ToSlash(trimmed))
		}
	}
	return results, nil
}

// CommitExists checks if a commit hash exists in git.
func CommitExists(repoRoot, commitHash string) bool {
	cmd := exec.Command("git", "-C", repoRoot, "cat-file", "-e", commitHash+"^{commit}")
	return cmd.Run() == nil
}

// GetNewerCommits returns log entries for commits newer than commitHash for targetPath.
func GetNewerCommits(repoRoot, commitHash, targetRelPath string) []string {
	cmd := exec.Command("git", "-C", repoRoot, "log", "--format=%h %s", commitHash+"..HEAD", "--", targetRelPath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil
	}

	lines := strings.Split(stdout.String(), "\n")
	var results []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			results = append(results, trimmed)
		}
	}
	return results
}
