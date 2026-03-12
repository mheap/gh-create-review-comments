package comments

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

// ReviewComment is the shape of a comment within a review submission.
type ReviewComment struct {
	Body      string `json:"body"`
	Path      string `json:"path"`
	Side      string `json:"side"`
	StartSide string `json:"start_side,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	Line      int    `json:"line"`
}

// Suggestion is a ReviewComment with an associated commit ID.
type Suggestion struct {
	ReviewComment
	CommitID string `json:"commit_id"`
}

// Review is the payload for POST /repos/{owner}/{repo}/pulls/{pr}/reviews.
type Review struct {
	CommitID string          `json:"commit_id"`
	Body     string          `json:"body"`
	Event    string          `json:"event"`
	Comments []ReviewComment `json:"comments"`
}

// ExistingComment represents an existing review comment on a PR (for deduplication).
type ExistingComment struct {
	Body      string `json:"body"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	StartLine int    `json:"start_line"`
}

// pendingGroup tracks a group of consecutive minus and/or plus lines that form
// a single logical change (pure addition, pure deletion, or replacement).
type pendingGroup struct {
	minusCount       int      // number of deleted lines
	plusLines        []string // accumulated added lines
	startOld         int      // old-side line number where the group starts
	startNew         int      // new-side line number where the group starts
	precedingLine    string   // content of the context line immediately before this group
	hasPrecedingLine bool     // true if there was a context line before this group
	followingLine    string   // content of the context line immediately after this group
	hasFollowingLine bool     // true if there was a context line after this group
}

// makeSuggestionBody wraps lines in a suggestion code fence.
func makeSuggestionBody(lines []string) string {
	return fmt.Sprintf("```suggestion\n%s\n```", strings.Join(lines, "\n"))
}

// buildSuggestion creates a Suggestion from a pendingGroup.
//
// All suggestions target the RIGHT side of the PR diff because we're suggesting
// changes to lines that the PR introduced. Line numbers use the old-side of
// our local diff (origin/branch → working tree), which correspond to line
// positions in the file as it exists at the PR's head commit.
//
// It handles three cases:
//   - Replacement (-old +new): target the old lines and suggest new content
//   - Pure deletion (-old): target the old lines and suggest empty content
//   - Pure addition (+new): anchor to the preceding context line and include
//     the anchor line's content plus the new lines in the suggestion body,
//     because the new lines don't yet exist in the pushed PR
func buildSuggestion(commitID, path string, g pendingGroup) Suggestion {
	hasPlus := len(g.plusLines) > 0

	if g.minusCount > 0 && !hasPlus {
		// Pure deletion: suggest removing the lines with an empty suggestion block.
		body := makeSuggestionBody([]string{""})
		oldEnd := g.startOld + g.minusCount - 1
		sug := Suggestion{
			ReviewComment: ReviewComment{
				Body: body,
				Path: path,
				Side: "RIGHT",
				Line: oldEnd,
			},
			CommitID: commitID,
		}
		if g.minusCount > 1 {
			sug.StartSide = "RIGHT"
			sug.StartLine = g.startOld
		}
		return sug
	}

	if g.minusCount == 0 && hasPlus {
		// Pure addition: new lines don't exist in the pushed PR.
		// Anchor the suggestion to an adjacent context line and include
		// its content along with the new lines.
		if g.hasPrecedingLine {
			// Anchor to the preceding line: "replace" it with itself + new content.
			allLines := append([]string{g.precedingLine}, g.plusLines...)
			body := makeSuggestionBody(allLines)
			anchorLine := g.startOld - 1 // the preceding line in old-side numbering
			sug := Suggestion{
				ReviewComment: ReviewComment{
					Body: body,
					Path: path,
					Side: "RIGHT",
					Line: anchorLine,
				},
				CommitID: commitID,
			}
			return sug
		}
		if g.hasFollowingLine {
			// Addition at the very start of the file (no preceding line).
			// Anchor to the following line: "replace" it with new content + itself.
			allLines := append(g.plusLines, g.followingLine)
			body := makeSuggestionBody(allLines)
			anchorLine := g.startOld // the following line in old-side numbering
			sug := Suggestion{
				ReviewComment: ReviewComment{
					Body: body,
					Path: path,
					Side: "RIGHT",
					Line: anchorLine,
				},
				CommitID: commitID,
			}
			return sug
		}
		// No preceding or following line — this is an addition to a brand new
		// empty file. The entire file is new content; target the new lines directly.
		body := makeSuggestionBody(g.plusLines)
		newEnd := g.startNew + len(g.plusLines) - 1
		sug := Suggestion{
			ReviewComment: ReviewComment{
				Body: body,
				Path: path,
				Side: "RIGHT",
				Line: newEnd,
			},
			CommitID: commitID,
		}
		if len(g.plusLines) > 1 {
			sug.StartSide = "RIGHT"
			sug.StartLine = g.startNew
		}
		return sug
	}

	// Replacement (minusCount > 0 && hasPlus): target the old lines directly
	// and suggest the new content.
	body := makeSuggestionBody(g.plusLines)
	oldEnd := g.startOld + g.minusCount - 1
	sug := Suggestion{
		ReviewComment: ReviewComment{
			Body: body,
			Path: path,
			Side: "RIGHT",
			Line: oldEnd,
		},
		CommitID: commitID,
	}
	if g.minusCount > 1 {
		sug.StartSide = "RIGHT"
		sug.StartLine = g.startOld
	}
	return sug
}

// hunkHeaderRe matches @@ -oldStart[,oldCount] +newStart[,newCount] @@
var hunkHeaderRe = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

// CreateSuggestions parses a unified diff (with full file context via -U99999)
// and produces Suggestion values for every change: additions, deletions, and replacements.
//
// Rename/copy diffs are handled transparently: git still emits --- and +++ lines
// for renames (e.g. "rename from old.go" / "rename to new.go" appear before the
// standard "--- a/old.go" / "+++ b/new.go" headers), and the parser extracts the
// path from +++ as usual. Any content changes within a rename are parsed normally.
//
// Binary diffs (e.g. "Binary files ... differ") contain no hunk headers or +/- lines,
// so they naturally produce zero suggestions.
func CreateSuggestions(commitID, diff string) ([]Suggestion, error) {
	var suggestions []Suggestion
	var currentFile string
	var oldLine, newLine int
	var inDiffHunk bool
	var isDeletedFile bool
	var group *pendingGroup

	// Track the most recent context line content and its old-side line number
	var lastContextLine string
	var hasLastContextLine bool

	flush := func(followingLine string, hasFollowing bool) {
		if group == nil {
			return
		}
		if group.minusCount == 0 && len(group.plusLines) == 0 {
			group = nil
			return
		}
		// Attach following context line info for pure additions at file start
		group.followingLine = followingLine
		group.hasFollowingLine = hasFollowing
		// Skip suggestions for fully deleted files — the GitHub API cannot
		// anchor review comments on files that no longer exist in the PR.
		if !isDeletedFile {
			suggestions = append(suggestions, buildSuggestion(commitID, currentFile, *group))
		}
		group = nil
	}

	scanner := bufio.NewScanner(strings.NewReader(diff))
	// Increase buffer size to handle long lines (e.g., minified files).
	// Default is 64KB which is too small for full-file-context diffs.
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// New file in the diff
		if strings.HasPrefix(line, "diff --git") {
			flush("", false)
			inDiffHunk = false
			isDeletedFile = false
			lastContextLine = ""
			hasLastContextLine = false
			continue
		}

		// Handle --- and +++ lines only in file headers (before the first hunk)
		if !inDiffHunk {
			// Handle --- line: extract path as fallback for deleted files
			// where +++ is /dev/null
			if strings.HasPrefix(line, "--- ") {
				rest := strings.TrimPrefix(line, "--- ")
				if rest != "/dev/null" {
					currentFile = strings.TrimPrefix(rest, "a/")
				}
				continue
			}

			// Extract file path from the +++ line
			if strings.HasPrefix(line, "+++ ") {
				rest := strings.TrimPrefix(line, "+++ ")
				if rest == "/dev/null" {
					isDeletedFile = true
				} else {
					currentFile = strings.TrimPrefix(rest, "b/")
				}
				continue
			}
		}

		// Start of a diff hunk
		if matches := hunkHeaderRe.FindStringSubmatch(line); matches != nil {
			flush("", false)
			inDiffHunk = true
			oldLine = atoi(matches[1])
			newLine = atoi(matches[2])
			lastContextLine = ""
			hasLastContextLine = false
			continue
		}

		// Skip "\ No newline at end of file" markers
		if strings.HasPrefix(line, `\ `) {
			continue
		}

		if !inDiffHunk {
			continue
		}

		if strings.HasPrefix(line, "+") {
			content := strings.TrimPrefix(line, "+")
			if group == nil {
				group = &pendingGroup{
					startOld:         oldLine,
					startNew:         newLine,
					precedingLine:    lastContextLine,
					hasPrecedingLine: hasLastContextLine,
				}
			}
			group.plusLines = append(group.plusLines, content)
			newLine++
		} else if strings.HasPrefix(line, "-") {
			if group == nil {
				group = &pendingGroup{
					startOld:         oldLine,
					startNew:         newLine,
					precedingLine:    lastContextLine,
					hasPrecedingLine: hasLastContextLine,
				}
			} else if len(group.plusLines) > 0 {
				// We had plus lines and now hit a minus line - this is a new group.
				// Flush the previous group and start a new one.
				flush("", false)
				group = &pendingGroup{
					startOld:         oldLine,
					startNew:         newLine,
					precedingLine:    lastContextLine,
					hasPrecedingLine: hasLastContextLine,
				}
			}
			group.minusCount++
			oldLine++
		} else {
			// Context line (starts with ' ') - flush any pending group.
			// Pass the context line content so pure additions at file start
			// can anchor to the following line.
			contextContent := strings.TrimPrefix(line, " ")
			flush(contextContent, true)
			lastContextLine = contextContent
			hasLastContextLine = true
			oldLine++
			newLine++
		}
	}

	// Flush any remaining group at end of input
	flush("", false)

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading diff: %w", err)
	}

	return suggestions, nil
}

// getExistingComments fetches all existing review comments on a PR for deduplication.
// It paginates through all pages to ensure complete deduplication.
func getExistingComments(client *api.RESTClient, owner, repo string, prNumber int) ([]ExistingComment, error) {
	var allComments []ExistingComment
	page := 1

	for {
		endpoint := fmt.Sprintf("repos/%s/%s/pulls/%d/comments?per_page=100&page=%d", owner, repo, prNumber, page)
		var pageComments []ExistingComment
		err := client.Get(endpoint, &pageComments)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch existing comments (page %d): %w", page, err)
		}

		allComments = append(allComments, pageComments...)

		// If we got fewer than 100, we've reached the last page
		if len(pageComments) < 100 {
			break
		}
		page++
	}

	return allComments, nil
}

// isDuplicate checks if a suggestion matches an existing comment by path, line, start_line, and body.
func isDuplicate(s Suggestion, existing []ExistingComment) bool {
	for _, c := range existing {
		if c.Path == s.Path && c.Line == s.Line && c.StartLine == s.StartLine && c.Body == s.Body {
			return true
		}
	}
	return false
}

// maxCommentsPerReview is the maximum number of comments the GitHub Reviews API
// accepts in a single review. Larger batches are split into multiple reviews.
const maxCommentsPerReview = 50

// PostSuggestions submits suggestions as review comments via the Reviews API.
// It deduplicates against existing comments on the PR before posting, and
// batches comments into groups of maxCommentsPerReview to stay within API limits.
func PostSuggestions(client *api.RESTClient, owner, repo string, prNumber int, suggestions []Suggestion, reviewBody string) error {
	if len(suggestions) == 0 {
		return nil
	}

	existingComments, err := getExistingComments(client, owner, repo, prNumber)
	if err != nil {
		return fmt.Errorf("failed to get existing comments: %w", err)
	}

	commitID := suggestions[0].CommitID

	var reviewComments []ReviewComment
	for _, s := range suggestions {
		if isDuplicate(s, existingComments) {
			continue
		}

		rc := s.ReviewComment
		if s.StartLine == 0 {
			// Omit StartLine/StartSide for single-line comments
			rc.StartLine = 0
			rc.StartSide = ""
		}
		reviewComments = append(reviewComments, rc)
	}

	if len(reviewComments) == 0 {
		return nil // all suggestions already exist
	}

	endpoint := fmt.Sprintf("repos/%s/%s/pulls/%d/reviews", owner, repo, prNumber)

	// Submit comments in batches to stay within API limits.
	for i := 0; i < len(reviewComments); i += maxCommentsPerReview {
		end := i + maxCommentsPerReview
		if end > len(reviewComments) {
			end = len(reviewComments)
		}
		batch := reviewComments[i:end]

		review := Review{
			CommitID: commitID,
			Body:     reviewBody,
			Event:    "COMMENT",
			Comments: batch,
		}

		jsonData, err := json.Marshal(review)
		if err != nil {
			return fmt.Errorf("failed to marshal review payload: %w", err)
		}

		var resp map[string]interface{}
		err = client.Post(endpoint, bytes.NewReader(jsonData), &resp)
		if err != nil {
			return fmt.Errorf("failed to post review (batch %d-%d of %d): %w", i+1, end, len(reviewComments), err)
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
