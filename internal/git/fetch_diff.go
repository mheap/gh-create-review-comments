package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type DiffOptions struct {
	CompareTo          string
	IncludeUncommitted bool
}

// GetPRNumber uses the gh CLI to find the PR number for the current branch.
func GetPRNumber() (int, error) {
	cmd := exec.Command("gh", "pr", "view", "--json", "number", "-q", ".number")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("no pull request found for the current branch: %v - %s", err, stderr.String())
	}

	numStr := strings.TrimSpace(out.String())
	if numStr == "" {
		return 0, fmt.Errorf("no pull request found for the current branch")
	}

	var n int
	_, err := fmt.Sscanf(numStr, "%d", &n)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PR number %q: %w", numStr, err)
	}

	return n, nil
}

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error getting current branch: %v - %s", err, stderr.String())
	}

	branch := strings.TrimSpace(out.String())
	if branch == "" || branch == "HEAD" {
		return "", fmt.Errorf("no current branch found")
	}

	return branch, nil
}

func GetCommitId(branch string) (string, error) {
	if branch == "" {
		b, err := GetCurrentBranch()
		if err != nil {
			return "", fmt.Errorf("error determining branch for commit ID: %w", err)
		}
		branch = "origin/" + b
	}

	cmd := exec.Command("git", "rev-parse", branch)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error getting commit ID for branch %s: %v - %s", branch, err, stderr.String())
	}

	commitID := strings.TrimSpace(out.String())
	if commitID == "" {
		return "", fmt.Errorf("no commit ID found for branch %s", branch)
	}

	return commitID, nil
}

func FetchDiff(options DiffOptions) (string, error) {
	var compareTo string
	if options.CompareTo == "" {
		branch, err := GetCurrentBranch()
		if err != nil {
			return "", fmt.Errorf("error determining branch for diff: %w", err)
		}
		compareTo = "origin/" + branch
	} else {
		compareTo = options.CompareTo
	}

	var cmdArgs []string
	// Force standard a/ and b/ prefixes regardless of user config (e.g. diff.mnemonicPrefix).
	// Without this, prefixes like w/ (working tree), c/ (commit), i/ (index) would
	// cause the diff parser to produce incorrect file paths.
	baseArgs := []string{"diff", "--no-color", "--src-prefix=a/", "--dst-prefix=b/", "-U99999"}
	if options.IncludeUncommitted {
		// Compare working tree (including uncommitted changes) against origin
		cmdArgs = append(baseArgs, compareTo)
	} else {
		// Compare only committed changes against origin (exclude working tree)
		cmdArgs = append(baseArgs, compareTo, "HEAD")
	}

	cmd := exec.Command("git", cmdArgs...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error running git diff: %v - %s", err, stderr.String())
	}

	return out.String(), nil
}
