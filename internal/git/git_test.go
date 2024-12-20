package git

import (
	"git-tagger/internal/testutils"
	"testing"
)

// TestFindUntagged is a unit test that validates the identification of commits without associated tags.
//
// Parameters:
//   - t: Pointer to the testing framework's T struct to manage test state and support formatted logs.
//
// Returns:
//   - Nothing. This function performs test assertions to ensure correctness.
func TestFindUntagged(t *testing.T) {
	testutils.SetupAndValidateUntagged(t, "HEAD")
}

// TestCreateTag tests the creation of an annotated Git tag in a repository.
//
// Parameters:
// - t: A pointer to the testing framework's testing.T instance.
//
// Returns:
// - None.
func TestCreateTag(t *testing.T) {
	testutils.SetupTestRepo(t)
	testutils.CreateAndCommitFile(t, "file1.txt", "Initial commit")

	tagName, message := "v1.0.0", "Version 1.0"
	if err := CreateTag(tagName, message, "HEAD"); err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	testutils.VerifyTagExists(t, tagName)
}

// TestGetCurrentBranch verifies that the current Git branch is correctly identified as "master".
// Parameters:
//   - t: A testing object used to manage test state and support formatted test logs and errors.
//
// Returns:
//   - This function does not return any values but fails the test if validations are unsuccessful.
func TestGetCurrentBranch(t *testing.T) {
	testutils.SetupTestRepo(t)
	testutils.ValidateCurrentBranch(t, "master")
}

// TestGetBranches validates the functionality of listing branches in a Git repository.
// Parameters:
// t - The testing object used for running test cases and assertions.
// Returns:
// Nothing. The function asserts branch existence and reports errors via the testing framework.
func TestGetBranches(t *testing.T) {
	testutils.SetupTestRepo(t)
	if err := testutils.RunGitCommand("checkout", "-b", "feature/test-branch"); err != nil {
		t.Fatalf("Failed to create and switch to branch 'feature/test-branch': %v", err)
	}

	testutils.ValidateBranches(t, []string{"feature/test-branch", "master"})
}
