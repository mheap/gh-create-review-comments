package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mheap/gh-extension-create-review-comments/internal/comments"
	"github.com/mheap/gh-extension-create-review-comments/internal/git"

	"github.com/cli/go-gh/v2/pkg/api"
)

func main() {

	debug := os.Getenv("GH_EXTENSION_DEBUG") == "1"

	client, err := api.DefaultRESTClient()
	if err != nil {
		fmt.Printf("error creating API client: %v\n", err)
		return
	}

	// Get repository information from git remote
	repoInfo, err := git.GetRepositoryInfo()
	if err != nil {
		fmt.Printf("error getting repository info: %v\n", err)
		return
	}
	if debug {
		fmt.Printf("Repository Info: %+v\n", repoInfo)
	}

	// Generate a diff to apply as a PR comment
	options := git.DiffOptions{
		IncludeUncommitted: true,
	}
	output, err := git.FetchDiff(options)
	if err != nil {
		fmt.Printf("error fetching diff: %v\n", err)
		return
	}

	// Convert a Diff to a list of suggestions
	commit_id, err := git.GetCommitId("")
	if err != nil {
		fmt.Printf("error getting commit ID: %v\n", err)
		return
	}

	x, err := comments.CreateSuggestions(commit_id, output)
	if err != nil {
		fmt.Printf("error creating suggestions: %v\n", err)
		return
	}

	if debug {
		fmt.Printf("Created %d suggestions\n", len(x))
		for _, suggestion := range x {
			suggestionJSON, err := json.Marshal(suggestion)
			if err != nil {
				fmt.Printf("error marshalling suggestion: %v\n", err)
				continue
			}
			fmt.Println(string(suggestionJSON))
		}
	}

	// Send the comments to GitHub

	// Get PR number from environment or argument (placeholder - you'll need to implement this)
	prNumber := 2 // This should be dynamically determined

	err = comments.PostSuggestions(*client, repoInfo.Owner, repoInfo.RepoName, prNumber, x)
	if err != nil {
		fmt.Printf("error posting suggestions: %v\n", err)
		return
	}
}
