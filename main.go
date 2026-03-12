package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/mheap/gh-create-review-comments/internal/comments"
	"github.com/mheap/gh-create-review-comments/internal/git"

	"github.com/cli/go-gh/v2/pkg/api"
)

func main() {

	dryRun := flag.Bool("dry-run", false, "Show suggestions without posting them to GitHub")
	flag.Parse()

	debug := os.Getenv("GH_EXTENSION_DEBUG") == "1"

	// Get repository information from git remote
	repoInfo, err := git.GetRepositoryInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting repository info: %v\n", err)
		os.Exit(1)
	}
	if debug {
		fmt.Printf("Repository Info: %+v\n", repoInfo)
	}

	// Auto-detect the PR number for the current branch
	prNumber, err := git.GetPRNumber()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error detecting PR number: %v\n", err)
		os.Exit(1)
	}
	if debug {
		fmt.Printf("PR Number: %d\n", prNumber)
	}

	// Detect branch once and reuse for diff and commit ID
	branch, err := git.GetCurrentBranch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error detecting current branch: %v\n", err)
		os.Exit(1)
	}
	originBranch := "origin/" + branch

	// Generate a diff to apply as a PR comment (full file context)
	options := git.DiffOptions{
		CompareTo:          originBranch,
		IncludeUncommitted: true,
	}
	output, err := git.FetchDiff(options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching diff: %v\n", err)
		os.Exit(1)
	}

	if output == "" {
		fmt.Println("no changes found")
		return
	}

	// Convert the diff to a list of suggestions
	commitID, err := git.GetCommitId(originBranch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting commit ID: %v\n", err)
		os.Exit(1)
	}

	suggestions, err := comments.CreateSuggestions(commitID, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating suggestions: %v\n", err)
		os.Exit(1)
	}

	if len(suggestions) == 0 {
		fmt.Println("no suggestions to post")
		return
	}

	if debug {
		fmt.Printf("Created %d suggestions\n", len(suggestions))
		for _, suggestion := range suggestions {
			suggestionJSON, err := json.Marshal(suggestion)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error marshalling suggestion: %v\n", err)
				continue
			}
			fmt.Println(string(suggestionJSON))
		}
	}

	if *dryRun {
		fmt.Printf("Dry run: %d suggestions would be posted to PR #%d\n", len(suggestions), prNumber)
		for _, suggestion := range suggestions {
			suggestionJSON, err := json.Marshal(suggestion)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error marshalling suggestion: %v\n", err)
				continue
			}
			fmt.Println(string(suggestionJSON))
		}
		return
	}

	// Send the review to GitHub
	client, err := api.DefaultRESTClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating API client: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Posting %d suggestions to PR #%d...\n", len(suggestions), prNumber)

	reviewBody := ""
	err = comments.PostSuggestions(client, repoInfo.Owner, repoInfo.RepoName, prNumber, suggestions, reviewBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error posting suggestions: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done!")
}
