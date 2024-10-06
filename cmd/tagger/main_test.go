package main

import (
	"git-tagger/internal/git"
	"git-tagger/internal/testutils"
	"os"
	"testing"
)

// TestNonInteractiveMode tests the behavior of the code in a non-interactive Git environment.
// Parameters:
// - t: The testing framework's object used to manage test state and support formatted test logs.
// Returns:
// - This function does not return any values. It logs relevant information
// and fails the test using t.Fatalf if any step encounters an error.
func TestNonInteractiveMode(t *testing.T) {
	// Save the original environment variable value to restore later
	originalValue := os.Getenv("GIT_POST_COMMIT")

	// Set the environment variable to simulate non-interactive mode
	err := os.Setenv("GIT_POST_COMMIT", "1")
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer func(key, value string) {
		err := os.Setenv(key, value)
		if err != nil {
			t.Fatalf("Failed to restore environment variable: %v", err)
		}
	}("GIT_POST_COMMIT", originalValue) // Restore the original value after the test

	// Set up a temporary Git repository
	testutils.SetupTestRepo(t)

	// Create and commit a file to trigger any non-interactive behavior
	testutils.CreateAndCommitFile(t, "test-file.txt", "Test commit in non-interactive mode")

	// Verify the commit exists
	latestCommit, err := git.GetShortCommitHash("HEAD")
	if err != nil {
		t.Fatalf("Failed to retrieve latest commit: %v", err)
	}
	t.Logf("Latest commit in non-interactive mode: %s", latestCommit)
}

func TestVersionTag(t *testing.T) {
	// Future improvements for version tagging tests
}
