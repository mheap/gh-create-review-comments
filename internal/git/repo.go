package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type RepoInfo struct {
	RemoteURL string `json:"remote_url"`
	Branch    string `json:"branch"`
	Owner     string `json:"owner"`
	RepoName  string `json:"repo_name"`
}

func GetRepositoryInfo() (RepoInfo, error) {
	var repoInfo RepoInfo
	branch, err := GetCurrentBranch()
	if err != nil {
		return repoInfo, err
	}
	repoInfo.Branch = branch
	// Get remote URL from git config
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return repoInfo, fmt.Errorf("error getting remote URL: %v - %s", err, stderr.String())
	}
	remoteURL := strings.TrimSpace(out.String())
	if remoteURL == "" {
		return repoInfo, fmt.Errorf("no remote URL found")
	}
	repoInfo.RemoteURL = remoteURL
	// Parse remote URL to extract owner and repo name
	parts := strings.Split(remoteURL, "/")
	if len(parts) < 2 {
		return repoInfo, fmt.Errorf("invalid remote URL format: %s", remoteURL)
	}

	repoInfo.Owner = parts[len(parts)-2]
	repoInfo.Owner = strings.Replace(repoInfo.Owner, "git@github.com:", "", -1)
	repoInfo.RepoName = strings.TrimSuffix(parts[len(parts)-1], ".git")

	// Normalize branch name for GitHub API
	repoInfo.Branch = strings.TrimPrefix(repoInfo.Branch, "origin/")
	repoInfo.Branch = strings.TrimPrefix(repoInfo.Branch, "refs/heads/")
	repoInfo.Branch = strings.TrimSpace(repoInfo.Branch)

	if repoInfo.Branch == "" {
		return repoInfo, fmt.Errorf("no branch found")
	}

	// Normalize owner and repo name for GitHub API
	repoInfo.Owner = strings.TrimSpace(repoInfo.Owner)
	repoInfo.RepoName = strings.TrimSpace(repoInfo.RepoName)
	if repoInfo.Owner == "" || repoInfo.RepoName == "" {
		return repoInfo, fmt.Errorf("owner or repo name is empty")
	}

	return repoInfo, nil
}
