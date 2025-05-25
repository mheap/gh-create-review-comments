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
	// Implementation to get the commit ID for the given branch
	if branch == "" {
		branch, _ = GetCurrentBranch()
		branch = "origin/" + branch
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
		branch, _ := GetCurrentBranch()
		compareTo = "origin/" + branch
	} else {
		compareTo = options.CompareTo
	}

	var cmdArgs []string
	if options.IncludeUncommitted {
		cmdArgs = append(cmdArgs, "diff", compareTo)
	} else {
		cmdArgs = append(cmdArgs, "diff", "--cached", compareTo)
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
