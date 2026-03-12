package git

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

type RepoInfo struct {
	Owner    string `json:"owner"`
	RepoName string `json:"name"`
}

// GetRepositoryInfo uses the gh CLI to reliably determine the owner and repo name.
func GetRepositoryInfo() (RepoInfo, error) {
	cmd := exec.Command("gh", "repo", "view", "--json", "owner,name")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return RepoInfo{}, fmt.Errorf("error getting repository info: %v - %s", err, stderr.String())
	}

	// gh repo view --json returns: {"name":"repo","owner":{"login":"owner"}}
	var raw struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	}
	if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
		return RepoInfo{}, fmt.Errorf("error parsing repository info: %w", err)
	}

	if raw.Owner.Login == "" || raw.Name == "" {
		return RepoInfo{}, fmt.Errorf("could not determine owner or repo name from gh output")
	}

	return RepoInfo{
		Owner:    raw.Owner.Login,
		RepoName: raw.Name,
	}, nil
}
