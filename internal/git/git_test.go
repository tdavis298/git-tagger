package git

import (
	"git-tagger/internal/testutils" // Import shared test helper functions
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setup(t *testing.T) {
	testutils.SetupTestRepo(t)
}

func TestFindUntagged(t *testing.T) {
	setup(t)

	// Create and commit test files
	testutils.CreateAndCommitFile(t, "file1.txt", "Add file1.txt")
	testutils.CreateAndCommitFile(t, "file2.txt", "Add file2.txt")

	// Explicitly tag the first commit (HEAD~1 should be the correct reference)
	t.Log("Tagging the first commit:")
	testutils.RunGitCommand(t, "tag", "v1.0", "HEAD~1")

	// Debugging step: list all commits and tags
	t.Log("Git log and tags:")
	logOutput := testutils.RunGitCommandAndGetOutput(t, "log", "--oneline", "--decorate", "--graph", "--all")
	t.Log(logOutput)

	tagOutput := testutils.RunGitCommandAndGetOutput(t, "tag", "--list", "-n1")
	t.Log("Tags found: ", tagOutput)

	// Find untagged commits on the current branch
	untaggedCommits := testutils.FindUntaggedCommits(t)

	// Expect only the second commit to be untagged
	if len(untaggedCommits) != 1 {
		t.Fatalf("Expected 1 untagged commit, but found %d", len(untaggedCommits))
	}

	// Verify the untagged commit message
	testutils.VerifyCommitMessage(t, untaggedCommits[0], "Add file2.txt")
}

func TestGetCurrentBranch(t *testing.T) {
	setup(t)

	// Check that the branch name matches `git rev-parse` output
	branch := testutils.RunGitCommandAndGetOutput(t, "rev-parse", "--abbrev-ref", "HEAD")
	expectedBranch := testutils.RunGitCommandAndGetOutput(t, "rev-parse", "--abbrev-ref", "HEAD")

	t.Logf("Current branch: %s", branch)

	if branch != expectedBranch {
		t.Fatalf("Expected branch %q, but got %q", expectedBranch, branch)
	}
}

func TestPostCommitHook(t *testing.T) {
	if testutils.IsWSL() {
		t.Log("Running inside WSL.")
	} else {
		t.Fatal("Running outside WSL.")
	}

	setup(t)

	// Create the post-commit hook
	hookScript := `#!/bin/sh
    echo "Post-commit hook triggered" >> hook.log
    `
	testutils.CreatePostCommitHook(t, hookScript)

	// Perform a commit to trigger the hook
	testutils.CreateAndCommitFile(t, "testfile.txt", "Test commit")

	// Verify that the hook was triggered by checking the hook.log file
	hookLogPath := filepath.Join("hook.log")
	data, err := os.ReadFile(hookLogPath)
	if err != nil {
		t.Fatalf("Failed to read hook log: %v", err)
	}
	if !strings.Contains(string(data), "Post-commit hook triggered") {
		t.Fatalf("Post-commit hook did not trigger as expected")
	}
}

func TestPostCommitTagging(t *testing.T) {
	// Set up a temporary Git repository for testing
	testutils.SetupTestRepo(t)

	// Define the post-commit hook script content
	hookScript := `#!/bin/sh

    export GIT_POST_COMMIT="true"

    /mnt/d/Projects/git-tagger/bin/tagger -version-tag
    `

	// Create the post-commit hook
	testutils.CreatePostCommitHook(t, hookScript)

	// Create and commit a new file to trigger the post-commit hook
	testutils.CreateAndCommitFile(t, "trigger.txt", "Trigger post-commit hook")

	// Simulate the post-commit hook running by making a new commit
	testutils.RunGitCommand(t, "commit", "--allow-empty", "-m", "Empty commit to trigger hook")

	gitLog := testutils.RunGitCommandAndGetOutput(t, "log", "--oneline", "--decorate", "--graph", "--all")
	t.Log("Git log after post-commit hook:", gitLog)

	tags := testutils.RunGitCommandAndGetOutput(t, "tag", "--list")
	if tags == "" {
		t.Fatalf("Expected tag to be created by post-commit hook, but none found.")
	}
}

// TestCreateTag verifies that CreateTag correctly runs the Git command to create a tag.
func TestCreateTag(t *testing.T) {
	tag := "v1.0.0"
	message := "Initial release"
	commit := "abc123"

	err := CreateTag(tag, message, commit)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// TestGetLatestTag verifies that GetLatestTag correctly retrieves the latest semantic version tag.
func TestGetLatestTag(t *testing.T) {
	setup(t)
	commits := []string{"v1.0.0", "v1.1.0", "v2.0.0"}

	for _, tag := range commits {
		err := CreateTag(tag, "Test tag "+tag, testutils.RunGitCommandAndGetOutput(t, "rev-parse", "HEAD"))
		if err != nil {
			t.Fatalf("Failed to create tag %s: %v", tag, err)
		}
	}

	tag, err := GetLatestTag()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := "v2.0.0"
	if tag != expected {
		t.Errorf("Expected %s, got %s", expected, tag)
	}
}

func TestGetTagsForCommit(t *testing.T) {
	setup(t)
	commit := testutils.RunGitCommandAndGetOutput(t, "rev-parse", "HEAD")
	expectedTags := []string{"v1.0.0", "v1.1.0"}

	for _, tag := range expectedTags {
		err := CreateTag(tag, "Test tag "+tag, commit)
		if err != nil {
			t.Fatalf("Failed to create tag %s: %v", tag, err)
		}
	}

	tags, err := GetTagsForCommit(commit)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tags) != len(expectedTags) {
		t.Fatalf("Expected tags %v, got %v", expectedTags, tags)
	}
	for i, tag := range tags {
		if tag != expectedTags[i] {
			t.Errorf("Expected %s, got %s", expectedTags[i], tag)
		}
	}
}

// TestGetShortCommitHash verifies that GetShortCommitHash correctly retrieves the short form of a commit hash.
func TestGetShortCommitHash(t *testing.T) {
	setup(t)
	fullHash := testutils.RunGitCommandAndGetOutput(t, "rev-parse", "HEAD")

	shortHash, err := GetShortCommitHash(fullHash)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	minLength := 7
	if len(shortHash) < minLength {
		t.Errorf("Expected short hash to have at least %d characters, got %d", minLength, len(shortHash))
	}
}

// TestGetCommitMessage verifies that GetCommitMessage correctly retrieves the commit message.
func TestGetCommitMessage(t *testing.T) {
	setup(t)
	expectedMessage := "Test commit message"
	testutils.CreateAndCommitFile(t, "test-file.txt", expectedMessage)
	commit := testutils.RunGitCommandAndGetOutput(t, "rev-parse", "HEAD")

	message, err := GetCommitMessage(commit)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if message != expectedMessage {
		t.Errorf("Expected %s, got %s", expectedMessage, message)
	}
}
