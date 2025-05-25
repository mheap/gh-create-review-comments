package comments

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

type Suggestion struct {
	Body      string `json:"body"`
	CommitID  string `json:"commit_id"`
	Path      string `json:"path"`
	Side      string `json:"side"`
	StartSide string `json:"start_side,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	Line      int    `json:"line"`
}

func CreateSuggestions(commitID, diff string) ([]Suggestion, error) {

	var suggestions []Suggestion
	var currentPlusLines []string
	var startPlusLine int
	scanner := bufio.NewScanner(strings.NewReader(diff))

	var currentFile string
	var lastDiffLine int
	var inDiffHunk bool

	for scanner.Scan() {
		line := scanner.Text()

		// Check for a new file in the diff
		if strings.HasPrefix(line, "diff --git") {
			inDiffHunk = false
			// flush any pending plus lines
			if len(currentPlusLines) > 0 {
				body := fmt.Sprintf("```suggestion\n%s\n```", strings.Join(currentPlusLines, "\n"))
				groupLen := len(currentPlusLines)
				if groupLen > 1 {
					suggestions = append(suggestions, Suggestion{
						Body:      body,
						CommitID:  commitID,
						Path:      currentFile,
						Side:      "RIGHT",
						StartSide: "RIGHT",
						StartLine: startPlusLine,
						Line:      startPlusLine + groupLen - 1,
					})
				} else {
					suggestions = append(suggestions, Suggestion{
						Body:     body,
						CommitID: commitID,
						Path:     currentFile,
						Side:     "RIGHT",
						Line:     startPlusLine,
					})
				}
				currentPlusLines = nil
			}
			continue
		}

		// Get the file path
		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
			continue
		}

		// Start of a diff hunk
		if strings.HasPrefix(line, "@@") {
			inDiffHunk = true
			// flush any pending plus lines at hunk boundary
			if len(currentPlusLines) > 0 {
				body := fmt.Sprintf("```suggestion\n%s\n```", strings.Join(currentPlusLines, "\n"))
				groupLen := len(currentPlusLines)
				if groupLen > 1 {
					suggestions = append(suggestions, Suggestion{
						Body:      body,
						CommitID:  commitID,
						Path:      currentFile,
						Side:      "RIGHT",
						StartSide: "RIGHT",
						StartLine: startPlusLine,
						Line:      startPlusLine + groupLen - 1,
					})
				} else {
					suggestions = append(suggestions, Suggestion{
						Body:     body,
						CommitID: commitID,
						Path:     currentFile,
						Side:     "RIGHT",
						Line:     startPlusLine,
					})
				}
				currentPlusLines = nil
			}

			// Extract starting line numbers from the hunk header
			parts := strings.Split(line, " ")
			if len(parts) >= 3 {
				lineInfo := strings.Split(strings.TrimPrefix(parts[2], "+"), ",")
				if len(lineInfo) > 0 {
					lastDiffLine = atoi(lineInfo[0]) - 1
				}
			}
			continue
		}

		if inDiffHunk {
			if strings.HasPrefix(line, "+") {
				// start new group if needed
				if len(currentPlusLines) == 0 {
					startPlusLine = lastDiffLine + 1
				}
				currentPlusLines = append(currentPlusLines, strings.TrimPrefix(line, "+"))
			} else {
				// on non-plus line, flush group if exists
				if len(currentPlusLines) > 0 {
					body := fmt.Sprintf("```suggestion\n%s\n```", strings.Join(currentPlusLines, "\n"))
					groupLen := len(currentPlusLines)
					if groupLen > 1 {
						suggestions = append(suggestions, Suggestion{
							Body:      body,
							CommitID:  commitID,
							Path:      currentFile,
							Side:      "RIGHT",
							StartSide: "RIGHT",
							StartLine: startPlusLine,
							Line:      startPlusLine + groupLen - 1,
						})
					} else {
						suggestions = append(suggestions, Suggestion{
							Body:     body,
							CommitID: commitID,
							Path:     currentFile,
							Side:     "RIGHT",
							Line:     startPlusLine,
						})
					}
					currentPlusLines = nil
				}
			}

			// Update line count: skip deletions
			if !strings.HasPrefix(line, "-") {
				lastDiffLine++
			}
		}
	}

	// flush any remaining grouped suggestions
	if len(currentPlusLines) > 0 {
		body := fmt.Sprintf("```suggestion\n%s\n```", strings.Join(currentPlusLines, "\n"))
		groupLen := len(currentPlusLines)
		if groupLen > 1 {
			suggestions = append(suggestions, Suggestion{
				Body:      body,
				CommitID:  commitID,
				Path:      currentFile,
				Side:      "RIGHT",
				StartSide: "RIGHT",
				StartLine: startPlusLine,
				Line:      startPlusLine + groupLen - 1,
			})
		} else {
			suggestions = append(suggestions, Suggestion{
				Body:     body,
				CommitID: commitID,
				Path:     currentFile,
				Side:     "RIGHT",
				Line:     startPlusLine,
			})
		}
		currentPlusLines = nil
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading diff: %w", err)
	}

	return suggestions, nil
}

type Comment struct {
	Body string `json:"body"`
	Path string `json:"path"`
	Line int    `json:"line"`
}

func getExistingComments(client api.RESTClient, owner, repo string, prNumber int) ([]Comment, error) {
	// Check if this suggestion already exists by its path and line number
	// by calling the GitHub API to list existing comments on the PR.
	checkEndpoint := fmt.Sprintf("repos/%s/%s/pulls/%d/comments", owner, repo, prNumber)
	var existingComments []Comment
	err := client.Get(checkEndpoint, &existingComments)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing comments: %w", err)
	}
	return existingComments, nil
}

func PostSuggestions(client api.RESTClient, owner, repo string, prNumber int, suggestions []Suggestion) error {

	existingComments, err := getExistingComments(client, owner, repo, prNumber)
	if err != nil {
		return fmt.Errorf("failed to get existing comments: %w", err)
	}

	for _, s := range suggestions {
		// Check if the suggestion already exists
		exists := false
		for _, comment := range existingComments {
			if comment.Path == s.Path && comment.Line == s.Line {
				// Check that the body of the existing comment matches the suggestion body
				if comment.Body == "" {
					continue // Skip comments without body
				}
				if comment.Body == s.Body {
					// Existing comment matches the suggestion, skip posting

					exists = true
					break
				}
			}
		}
		if exists {
			continue // Skip posting suggestion if it already exists
		}

		jsonData, err := json.Marshal(s)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		endpoint := fmt.Sprintf("repos/%s/%s/pulls/%d/comments", owner, repo, prNumber)
		var resp map[string]interface{}
		err = client.Post(endpoint, bytes.NewReader(jsonData), &resp)
		if err != nil {
			return fmt.Errorf("failed to post suggestion: %w", err)
		}
	}
	return nil
}

func atoi(s string) int {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0
	}
	return n
}
