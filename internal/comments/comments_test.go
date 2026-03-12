package comments

import (
	"testing"
)

func TestCreateSuggestions_SingleLineAddition(t *testing.T) {
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 line1
 line2
+new line
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	if s.Path != "file.go" {
		t.Errorf("expected path 'file.go', got %q", s.Path)
	}
	// Pure addition anchored to preceding context line "line2" (old-side line 2)
	if s.Line != 2 {
		t.Errorf("expected line 2, got %d", s.Line)
	}
	if s.Side != "RIGHT" {
		t.Errorf("expected side 'RIGHT', got %q", s.Side)
	}
	if s.CommitID != "abc123" {
		t.Errorf("expected commit_id 'abc123', got %q", s.CommitID)
	}
	// Body includes the preceding anchor line's content + new lines
	expectedBody := "```suggestion\nline2\nnew line\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
	// Single anchor line: no start_line
	if s.StartLine != 0 {
		t.Errorf("expected start_line 0 for single anchor line, got %d", s.StartLine)
	}
}

func TestCreateSuggestions_MultiLineAddition(t *testing.T) {
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,2 +1,5 @@
 line1
+added1
+added2
+added3
 line2
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	// Pure addition anchored to preceding context line "line1" (old-side line 1)
	if s.Line != 1 {
		t.Errorf("expected line 1, got %d", s.Line)
	}
	// Single anchor line: no start_line/start_side
	if s.StartLine != 0 {
		t.Errorf("expected start_line 0 for single anchor line, got %d", s.StartLine)
	}
	if s.StartSide != "" {
		t.Errorf("expected start_side empty for single anchor line, got %q", s.StartSide)
	}
	// Body includes the preceding anchor line's content + new lines
	expectedBody := "```suggestion\nline1\nadded1\nadded2\nadded3\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestCreateSuggestions_SingleLineDeletion(t *testing.T) {
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,4 +1,3 @@
 line1
 line2
-deleted line
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	// Deletions target RIGHT side (these are lines the PR introduced)
	if s.Side != "RIGHT" {
		t.Errorf("expected side 'RIGHT', got %q", s.Side)
	}
	if s.Line != 3 {
		t.Errorf("expected line 3, got %d", s.Line)
	}
	// Single line deletion: no start_line
	if s.StartLine != 0 {
		t.Errorf("expected start_line 0 for single line deletion, got %d", s.StartLine)
	}
	expectedBody := "```suggestion\n\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestCreateSuggestions_MultiLineDeletion(t *testing.T) {
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,5 +1,3 @@
 line1
-deleted1
-deleted2
 line2
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	// Deletions target RIGHT side
	if s.Side != "RIGHT" {
		t.Errorf("expected side 'RIGHT', got %q", s.Side)
	}
	if s.StartLine != 2 {
		t.Errorf("expected start_line 2, got %d", s.StartLine)
	}
	if s.Line != 3 {
		t.Errorf("expected line 3, got %d", s.Line)
	}
	if s.StartSide != "RIGHT" {
		t.Errorf("expected start_side 'RIGHT', got %q", s.StartSide)
	}
}

func TestCreateSuggestions_Replacement(t *testing.T) {
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,3 +1,3 @@
 line1
-old line
+new line
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	if s.Side != "RIGHT" {
		t.Errorf("expected side 'RIGHT', got %q", s.Side)
	}
	if s.Line != 2 {
		t.Errorf("expected line 2, got %d", s.Line)
	}
	expectedBody := "```suggestion\nnew line\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestCreateSuggestions_MultiLineReplacement(t *testing.T) {
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,4 +1,5 @@
 line1
-old1
-old2
+new1
+new2
+new3
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	if s.Side != "RIGHT" {
		t.Errorf("expected side 'RIGHT', got %q", s.Side)
	}
	// Multi-line replacement: startOld=2, minusCount=2, Line = startOld + minusCount - 1 = 3
	if s.StartLine != 2 {
		t.Errorf("expected start_line 2, got %d", s.StartLine)
	}
	if s.Line != 3 {
		t.Errorf("expected line 3, got %d", s.Line)
	}
	expectedBody := "```suggestion\nnew1\nnew2\nnew3\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestCreateSuggestions_MultipleChangesInFile(t *testing.T) {
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,7 +1,8 @@
 line1
+added at top
 line2
 line3
-removed middle
+replaced middle
 line5
 line6
+added at bottom
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 3 {
		t.Fatalf("expected 3 suggestions, got %d", len(suggestions))
	}

	// First: pure addition after line1, anchored to preceding "line1" (old-side line 1)
	if suggestions[0].Line != 1 {
		t.Errorf("suggestion 0: expected line 1, got %d", suggestions[0].Line)
	}
	if suggestions[0].Side != "RIGHT" {
		t.Errorf("suggestion 0: expected side RIGHT, got %q", suggestions[0].Side)
	}
	expectedBody0 := "```suggestion\nline1\nadded at top\n```"
	if suggestions[0].Body != expectedBody0 {
		t.Errorf("suggestion 0: expected body %q, got %q", expectedBody0, suggestions[0].Body)
	}

	// Second: replacement of "removed middle" with "replaced middle" (old-side line 4)
	if suggestions[1].Line != 4 {
		t.Errorf("suggestion 1: expected line 4, got %d", suggestions[1].Line)
	}
	if suggestions[1].Side != "RIGHT" {
		t.Errorf("suggestion 1: expected side RIGHT, got %q", suggestions[1].Side)
	}

	// Third: pure addition at bottom, anchored to preceding "line6" (old-side line 6)
	if suggestions[2].Line != 6 {
		t.Errorf("suggestion 2: expected line 6, got %d", suggestions[2].Line)
	}
}

func TestCreateSuggestions_MultipleFiles(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -1,3 +1,4 @@
 line1
+added in foo
 line2
diff --git a/bar.go b/bar.go
--- a/bar.go
+++ b/bar.go
@@ -1,3 +1,4 @@
 line1
+added in bar
 line2
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}
	if suggestions[0].Path != "foo.go" {
		t.Errorf("expected path 'foo.go', got %q", suggestions[0].Path)
	}
	if suggestions[1].Path != "bar.go" {
		t.Errorf("expected path 'bar.go', got %q", suggestions[1].Path)
	}
}

func TestCreateSuggestions_EmptyDiff(t *testing.T) {
	suggestions, err := CreateSuggestions("abc123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected 0 suggestions, got %d", len(suggestions))
	}
}

func TestCreateSuggestions_ContextOnlyNoChanges(t *testing.T) {
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,3 +1,3 @@
 line1
 line2
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected 0 suggestions (no changes), got %d", len(suggestions))
	}
}

func TestCreateSuggestions_ConsecutiveDeleteThenAdd(t *testing.T) {
	// When - lines are followed by + lines, they form a replacement (not separate groups)
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,5 +1,5 @@
 line1
-old1
-old2
+new1
+new2
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion (replacement), got %d", len(suggestions))
	}
	s := suggestions[0]
	// Replacement targets the RIGHT side
	if s.Side != "RIGHT" {
		t.Errorf("expected side RIGHT, got %q", s.Side)
	}
	expectedBody := "```suggestion\nnew1\nnew2\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestCreateSuggestions_DeletionFollowedByAdditionSeparately(t *testing.T) {
	// When - lines and + lines are separated by context, they're separate groups
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,5 +1,5 @@
 line1
-deleted
 line2
+added
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions (delete + add), got %d", len(suggestions))
	}
	// Deletion uses RIGHT side
	if suggestions[0].Side != "RIGHT" {
		t.Errorf("suggestion 0: expected side RIGHT (deletion), got %q", suggestions[0].Side)
	}
	if suggestions[1].Side != "RIGHT" {
		t.Errorf("suggestion 1: expected side RIGHT (addition), got %q", suggestions[1].Side)
	}
}

func TestCreateSuggestions_FullContextDiff(t *testing.T) {
	// Simulates a full-context diff (-U99999) where all file lines are visible
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,10 +1,11 @@
 package main
 
 import "fmt"
 
 func main() {
-	fmt.Println("hello")
+	fmt.Println("hello world")
+	fmt.Println("goodbye")
 }
 
 func helper() {
 }
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	// This is a replacement: 1 minus line at old-side line 6, so Line=6 with no StartLine
	if s.Side != "RIGHT" {
		t.Errorf("expected side RIGHT, got %q", s.Side)
	}
	if s.StartLine != 0 {
		t.Errorf("expected start_line 0 (single minus line replacement), got %d", s.StartLine)
	}
	if s.Line != 6 {
		t.Errorf("expected line 6, got %d", s.Line)
	}
}

func TestBuildSuggestion_PureAddition(t *testing.T) {
	g := pendingGroup{
		plusLines: []string{"new content"},
		startNew:  5,
	}
	s := buildSuggestion("abc", "file.go", g)
	if s.Side != "RIGHT" {
		t.Errorf("expected RIGHT, got %q", s.Side)
	}
	if s.Line != 5 {
		t.Errorf("expected line 5, got %d", s.Line)
	}
}

func TestBuildSuggestion_PureDeletion(t *testing.T) {
	g := pendingGroup{
		minusCount: 1,
		startOld:   3,
	}
	s := buildSuggestion("abc", "file.go", g)
	// Deletions target RIGHT side
	if s.Side != "RIGHT" {
		t.Errorf("expected RIGHT, got %q", s.Side)
	}
	if s.Line != 3 {
		t.Errorf("expected line 3, got %d", s.Line)
	}
	expectedBody := "```suggestion\n\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestBuildSuggestion_Replacement(t *testing.T) {
	g := pendingGroup{
		minusCount: 1,
		plusLines:  []string{"new"},
		startOld:   3,
		startNew:   3,
	}
	s := buildSuggestion("abc", "file.go", g)
	if s.Side != "RIGHT" {
		t.Errorf("expected RIGHT, got %q", s.Side)
	}
	if s.Line != 3 {
		t.Errorf("expected line 3, got %d", s.Line)
	}
	expectedBody := "```suggestion\nnew\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestIsDuplicate(t *testing.T) {
	existing := []ExistingComment{
		{Body: "```suggestion\nfoo\n```", Path: "file.go", Line: 5, StartLine: 0},
		{Body: "```suggestion\nbar\n```", Path: "file.go", Line: 10, StartLine: 8},
	}

	// Exact match
	s1 := Suggestion{ReviewComment: ReviewComment{Body: "```suggestion\nfoo\n```", Path: "file.go", Line: 5}}
	if !isDuplicate(s1, existing) {
		t.Error("expected duplicate for exact match")
	}

	// Different body
	s2 := Suggestion{ReviewComment: ReviewComment{Body: "```suggestion\ndifferent\n```", Path: "file.go", Line: 5}}
	if isDuplicate(s2, existing) {
		t.Error("expected no duplicate for different body")
	}

	// Different path
	s3 := Suggestion{ReviewComment: ReviewComment{Body: "```suggestion\nfoo\n```", Path: "other.go", Line: 5}}
	if isDuplicate(s3, existing) {
		t.Error("expected no duplicate for different path")
	}

	// Multi-line match
	s4 := Suggestion{ReviewComment: ReviewComment{Body: "```suggestion\nbar\n```", Path: "file.go", Line: 10, StartLine: 8}}
	if !isDuplicate(s4, existing) {
		t.Error("expected duplicate for multi-line match")
	}

	// Same line but different start_line
	s5 := Suggestion{ReviewComment: ReviewComment{Body: "```suggestion\nbar\n```", Path: "file.go", Line: 10, StartLine: 9}}
	if isDuplicate(s5, existing) {
		t.Error("expected no duplicate when start_line differs")
	}
}

func TestCreateSuggestions_NewFile(t *testing.T) {
	// New file: --- /dev/null, all lines are additions
	diff := `diff --git a/newfile.go b/newfile.go
--- /dev/null
+++ b/newfile.go
@@ -0,0 +1,3 @@
+line1
+line2
+line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	if s.Path != "newfile.go" {
		t.Errorf("expected path 'newfile.go', got %q", s.Path)
	}
	if s.Side != "RIGHT" {
		t.Errorf("expected side RIGHT, got %q", s.Side)
	}
	if s.StartLine != 1 {
		t.Errorf("expected start_line 1, got %d", s.StartLine)
	}
	if s.Line != 3 {
		t.Errorf("expected line 3, got %d", s.Line)
	}
}

func TestCreateSuggestions_DeletedFile(t *testing.T) {
	// Deleted file: +++ /dev/null — suggestions are skipped because
	// the GitHub API cannot anchor comments on files that no longer exist.
	diff := `diff --git a/old.go b/old.go
--- a/old.go
+++ /dev/null
@@ -1,3 +0,0 @@
-line1
-line2
-line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected 0 suggestions for deleted file, got %d", len(suggestions))
	}
}

func TestCreateSuggestions_NoNewlineAtEOF(t *testing.T) {
	// Diff with "\ No newline at end of file" marker
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,3 +1,3 @@
 line1
 line2
-old
\ No newline at end of file
+new
\ No newline at end of file
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	if s.Line != 3 {
		t.Errorf("expected line 3, got %d", s.Line)
	}
	if s.Side != "RIGHT" {
		t.Errorf("expected side RIGHT, got %q", s.Side)
	}
	expectedBody := "```suggestion\nnew\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestCreateSuggestions_AdditionAtEndOfFile(t *testing.T) {
	// Addition at the very end of a file with no trailing context
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 line1
 line2
 line3
+line4
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	// Pure addition anchored to preceding "line3" (old-side line 3)
	if s.Line != 3 {
		t.Errorf("expected line 3, got %d", s.Line)
	}
	if s.Side != "RIGHT" {
		t.Errorf("expected side RIGHT, got %q", s.Side)
	}
	expectedBody := "```suggestion\nline3\nline4\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestCreateSuggestions_BinaryFile(t *testing.T) {
	// Binary files produce no hunk headers and no +/- lines, so 0 suggestions
	diff := `diff --git a/image.png b/image.png
new file mode 100644
index 0000000..abcdef1
Binary files /dev/null and b/image.png differ
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected 0 suggestions for binary file, got %d", len(suggestions))
	}
}

func TestCreateSuggestions_BinaryFileChange(t *testing.T) {
	// Binary file modification also produces 0 suggestions
	diff := `diff --git a/image.png b/image.png
index abcdef1..1234567 100644
Binary files a/image.png and b/image.png differ
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected 0 suggestions for binary file change, got %d", len(suggestions))
	}
}

func TestCreateSuggestions_FilePathWithSpaces(t *testing.T) {
	diff := `diff --git a/path with spaces/file.go b/path with spaces/file.go
--- a/path with spaces/file.go
+++ b/path with spaces/file.go
@@ -1,3 +1,4 @@
 line1
 line2
+new line
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	if s.Path != "path with spaces/file.go" {
		t.Errorf("expected path 'path with spaces/file.go', got %q", s.Path)
	}
}

func TestCreateSuggestions_FilePathWithABPrefix(t *testing.T) {
	// File path that starts with "a/" or "b/" as part of the actual path name.
	// The parser strips the leading "a/" and "b/" prefixes from --- and +++ lines,
	// so a file literally named "b/config.go" would be parsed as "config.go".
	// This is a known limitation that matches standard git behavior.
	diff := `diff --git a/a/nested/file.go b/a/nested/file.go
--- a/a/nested/file.go
+++ b/a/nested/file.go
@@ -1,3 +1,4 @@
 line1
 line2
+new line
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	// "a/a/nested/file.go" → strip "a/" prefix → "a/nested/file.go"
	if s.Path != "a/nested/file.go" {
		t.Errorf("expected path 'a/nested/file.go', got %q", s.Path)
	}
}

func TestCreateSuggestions_HunkHeaderWithFunctionContext(t *testing.T) {
	// Hunk headers can include function context after the @@, e.g. @@ -1,5 +1,6 @@ func main()
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,5 +1,6 @@ func main() {
 line1
 line2
+new line
 line3
 line4
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	// Pure addition anchored to preceding "line2" (old-side line 2)
	if s.Line != 2 {
		t.Errorf("expected line 2, got %d", s.Line)
	}
	if s.Side != "RIGHT" {
		t.Errorf("expected side RIGHT, got %q", s.Side)
	}
	expectedBody := "```suggestion\nline2\nnew line\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestCreateSuggestions_MultipleHunksInOneFile(t *testing.T) {
	// A single file with multiple hunk headers (non-full-context diff)
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,5 +1,6 @@
 line1
 line2
+added in first hunk
 line3
 line4
@@ -20,5 +21,6 @@
 line20
 line21
+added in second hunk
 line22
 line23
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}

	// First hunk: pure addition anchored to preceding "line2" (old-side line 2)
	if suggestions[0].Line != 2 {
		t.Errorf("suggestion 0: expected line 2, got %d", suggestions[0].Line)
	}
	expectedBody0 := "```suggestion\nline2\nadded in first hunk\n```"
	if suggestions[0].Body != expectedBody0 {
		t.Errorf("suggestion 0: expected body %q, got %q", expectedBody0, suggestions[0].Body)
	}

	// Second hunk: pure addition anchored to preceding "line21" (old-side line 21)
	if suggestions[1].Line != 21 {
		t.Errorf("suggestion 1: expected line 21, got %d", suggestions[1].Line)
	}
	expectedBody1 := "```suggestion\nline21\nadded in second hunk\n```"
	if suggestions[1].Body != expectedBody1 {
		t.Errorf("suggestion 1: expected body %q, got %q", expectedBody1, suggestions[1].Body)
	}
}

func TestCreateSuggestions_AdditionAtStartOfFile(t *testing.T) {
	// Addition at the very start of an existing file — exercises the
	// hasFollowingLine branch in buildSuggestion (no preceding context line).
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,3 +1,5 @@
+prepended1
+prepended2
 original first
 original second
 original third
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	if s.Path != "file.go" {
		t.Errorf("expected path 'file.go', got %q", s.Path)
	}
	// Anchored to following context line "original first" (old-side line 1)
	if s.Line != 1 {
		t.Errorf("expected line 1, got %d", s.Line)
	}
	if s.Side != "RIGHT" {
		t.Errorf("expected side RIGHT, got %q", s.Side)
	}
	// No start_line — single anchor line
	if s.StartLine != 0 {
		t.Errorf("expected start_line 0, got %d", s.StartLine)
	}
	// Body includes new lines + the following anchor line content
	expectedBody := "```suggestion\nprepended1\nprepended2\noriginal first\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestCreateSuggestions_PlusFollowedByMinus(t *testing.T) {
	// When + lines are followed by - lines (without intervening context),
	// the group is flushed and a new group is started for the minus lines.
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,4 +1,4 @@
 line1
+added
-removed
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}

	// First: pure addition anchored to preceding "line1" (old-side line 1)
	if suggestions[0].Side != "RIGHT" {
		t.Errorf("suggestion 0: expected side RIGHT, got %q", suggestions[0].Side)
	}
	if suggestions[0].Line != 1 {
		t.Errorf("suggestion 0: expected line 1, got %d", suggestions[0].Line)
	}
	expectedBody0 := "```suggestion\nline1\nadded\n```"
	if suggestions[0].Body != expectedBody0 {
		t.Errorf("suggestion 0: expected body %q, got %q", expectedBody0, suggestions[0].Body)
	}

	// Second: pure deletion of "removed" (old-side line 2)
	if suggestions[1].Side != "RIGHT" {
		t.Errorf("suggestion 1: expected side RIGHT, got %q", suggestions[1].Side)
	}
	if suggestions[1].Line != 2 {
		t.Errorf("suggestion 1: expected line 2, got %d", suggestions[1].Line)
	}
	expectedBody1 := "```suggestion\n\n```"
	if suggestions[1].Body != expectedBody1 {
		t.Errorf("suggestion 1: expected body %q, got %q", expectedBody1, suggestions[1].Body)
	}
}

func TestCreateSuggestions_NewFileSingleLine(t *testing.T) {
	// Brand new file with a single added line — no StartLine/StartSide.
	diff := `diff --git a/single.go b/single.go
--- /dev/null
+++ b/single.go
@@ -0,0 +1 @@
+only line
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	if s.Path != "single.go" {
		t.Errorf("expected path 'single.go', got %q", s.Path)
	}
	if s.Side != "RIGHT" {
		t.Errorf("expected side RIGHT, got %q", s.Side)
	}
	if s.Line != 1 {
		t.Errorf("expected line 1, got %d", s.Line)
	}
	// Single line: no start_line
	if s.StartLine != 0 {
		t.Errorf("expected start_line 0, got %d", s.StartLine)
	}
	if s.StartSide != "" {
		t.Errorf("expected start_side empty, got %q", s.StartSide)
	}
	expectedBody := "```suggestion\nonly line\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestCreateSuggestions_RenameDiff(t *testing.T) {
	// Rename with content changes — path should come from +++ line.
	diff := `diff --git a/old.go b/new.go
similarity index 90%
rename from old.go
rename to new.go
--- a/old.go
+++ b/new.go
@@ -1,3 +1,3 @@
 line1
-old content
+new content
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	if s.Path != "new.go" {
		t.Errorf("expected path 'new.go', got %q", s.Path)
	}
	if s.Line != 2 {
		t.Errorf("expected line 2, got %d", s.Line)
	}
	expectedBody := "```suggestion\nnew content\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}

func TestCreateSuggestions_HunkHeaderResetsContext(t *testing.T) {
	// Verify that a new hunk header resets the preceding context line tracking.
	// The second hunk starts with an addition before any context line, so it
	// should use the hasFollowingLine branch (not the preceding line from hunk 1).
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 line1
 line2
+added after line2
 line3
@@ -10,3 +11,4 @@
+added at hunk start
 line10
 line11
 line12
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}

	// First: addition anchored to preceding "line2" (old-side line 2)
	if suggestions[0].Line != 2 {
		t.Errorf("suggestion 0: expected line 2, got %d", suggestions[0].Line)
	}
	expectedBody0 := "```suggestion\nline2\nadded after line2\n```"
	if suggestions[0].Body != expectedBody0 {
		t.Errorf("suggestion 0: expected body %q, got %q", expectedBody0, suggestions[0].Body)
	}

	// Second: addition at hunk start with no preceding context (hunk header reset it).
	// Anchored to following "line10" (old-side line 10).
	if suggestions[1].Line != 10 {
		t.Errorf("suggestion 1: expected line 10, got %d", suggestions[1].Line)
	}
	expectedBody1 := "```suggestion\nadded at hunk start\nline10\n```"
	if suggestions[1].Body != expectedBody1 {
		t.Errorf("suggestion 1: expected body %q, got %q", expectedBody1, suggestions[1].Body)
	}
}

func TestCreateSuggestions_DeletedFileThenNormalFile(t *testing.T) {
	// A deleted file (+++ /dev/null) followed by a normal file.
	// Verifies that isDeletedFile is reset by the second diff --git.
	diff := `diff --git a/deleted.go b/deleted.go
--- a/deleted.go
+++ /dev/null
@@ -1,2 +0,0 @@
-old1
-old2
diff --git a/normal.go b/normal.go
--- a/normal.go
+++ b/normal.go
@@ -1,3 +1,3 @@
 line1
-old
+new
 line3
`
	suggestions, err := CreateSuggestions("abc123", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Deleted file produces 0 suggestions; normal file produces 1
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	s := suggestions[0]
	if s.Path != "normal.go" {
		t.Errorf("expected path 'normal.go', got %q", s.Path)
	}
	if s.Line != 2 {
		t.Errorf("expected line 2, got %d", s.Line)
	}
	expectedBody := "```suggestion\nnew\n```"
	if s.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, s.Body)
	}
}
