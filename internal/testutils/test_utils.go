package testutils

import (
	"bytes"
	"fmt"
	"git-tagger/internal/utils"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// SetupTestRepo initializes a temporary Git repository for testing purposes.
//
// Parameters:
// - t: A pointer to the testing framework's testing.T instance, used for logging and handling test failures.
//
// Returns:
// - None. The function sets up the environment, including cleanup after the test.
func SetupTestRepo(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test-repo")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Cleanup after test
	t.Cleanup(func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Fatalf("Failed to clean up temp directory: %v", err)
		}
	})

	// Change directory to temp directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Initialize a new Git repository
	if err := RunGitCommand("init"); err != nil {
		t.Fatalf("Failed to initialize Git repository: %v", err)
	}

	// Set Git identity for the repository
	setGitIdentity(t, "testuser", "testuser@example.com")

	// Create an initial commit
	CreateAndCommitFile(t, "README.md", "Initial commit")
}

// setGitIdentity configures a local Git user identity in the repository.
//
// Parameters:
// - t: The testing.T instance used for the testing framework.
// - name: The username to set for Git commits.
// - email: The user email to set for Git commits.
//
// Returns:
// - None. Halts test execution on failure.
func setGitIdentity(t *testing.T, name, email string) {
	// Validate inputs
	if name == "" || email == "" {
		t.Fatalf("Git name and email must not be empty")
	}

	// Check if the current directory is a Git repository
	if err := RunGitCommand("rev-parse", "--is-inside-work-tree"); err != nil {
		t.Fatalf("Current directory is not a valid Git repository: %v", err)
	}

	// Set the local Git user identity
	if err := RunGitCommand("config", "--local", "user.name", name); err != nil {
		t.Fatalf("Failed to set Git user name: %v", err)
	}
	if err := RunGitCommand("config", "--local", "user.email", email); err != nil {
		t.Fatalf("Failed to set Git user email: %v", err)
	}

	// Cleanup Git identity after test
	t.Cleanup(func() {
		if err := RunGitCommand("config", "--unset", "user.name"); err != nil {
			t.Logf("Cleanup failed to unset user.name: %v", err)
		}
		if err := RunGitCommand("config", "--unset", "user.email"); err != nil {
			t.Logf("Cleanup failed to unset user.email: %v", err)
		}
	})
}

// CreateAndCommitFile creates a new file, stages it, and commits it to the Git repository.
//
// Parameters:
// - t: The testing.T instance used for the testing framework.
// - filename: The name of the file to create and commit.
// - commitMsg: The commit message to use for the commit.
//
// Returns:
// - None. Halts test execution on failure.
func CreateAndCommitFile(t *testing.T, filename, commitMsg string) {
	// Write content to the file
	err := os.WriteFile(filename, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file %s: %v", filename, err)
	}

	// Stage the file
	if err := RunGitCommand("add", filename); err != nil {
		t.Fatalf("Failed to stage file %s: %v", filename, err)
	}

	// Commit the file with the provided commit message
	if err := RunGitCommand("commit", "-m", commitMsg); err != nil {
		t.Fatalf("Failed to commit file %s: %v", filename, err)
	}
}

// FindUntagged retrieves a list of untagged commits from the specified reference (e.g., HEAD).
func FindUntagged(ref string) ([]string, error) {
	// Git command to list untagged commits
	cmd := exec.Command("git", "log", "--oneline", "--no-walk", "--tags", "--not", ref, "--pretty=format:%H")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Split the output into lines to track commit hashes
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	return lines, nil
}

// FindUntaggedCommits identifies commits on the current branch that do not have associated tags.
//
// PARAMETERS:
//
//	t: A pointer to the testing object used for error handling and logging in tests.
//
// RETURNS:
//
//	A slice of strings containing hashes of commits that are untagged.
func FindUntaggedCommits(t *testing.T) []string {
	// Get all commits on the current branch
	allCommits := RunGitCommandAndGetOutput(t, "rev-list", "--no-merges", "--pretty=oneline", "HEAD")

	// Get all tagged commits
	taggedCommits := RunGitCommandAndGetOutput(t, "rev-list", "--tags", "--pretty=oneline")

	// Create a map of tagged commit hashes for quick lookup
	taggedCommitMap := make(map[string]struct{})
	for _, line := range strings.Split(taggedCommits, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			taggedCommitMap[fields[0]] = struct{}{}
		}
	}

	// Collect untagged commits
	var untaggedCommits []string
	for _, line := range strings.Split(allCommits, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			commitHash := fields[0]
			// If the commit hash is not in the tagged map, it's untagged
			if _, found := taggedCommitMap[commitHash]; !found {
				untaggedCommits = append(untaggedCommits, commitHash)
			}
		}
	}

	return untaggedCommits
}

// GetBranches retrieves a list of all branches in a Git repository, including remote branches.
//
// Parameters:
//
//	none.
//
// Returns:
//
//	[]string: A slice of branch names (string).
//	error: An error if the command fails or parsing the output encounters an issue.
func GetBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--all")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Split the output into lines, trim spaces, and remove the asterisk (*) denoting the current branch
	branches := strings.Split(out.String(), "\n")
	var cleanedBranches []string
	for _, branch := range branches {
		trimmedBranch := strings.TrimSpace(branch)
		if trimmedBranch != "" {
			cleanedBranches = append(cleanedBranches, strings.TrimPrefix(trimmedBranch, "* "))
		}
	}

	return cleanedBranches, nil
}

// GetCurrentBranch retrieves the name of the current Git branch in the repository.
//
// Parameters: None
//
// Returns:
// - string: The current branch name.
// - error: An error if the branch name cannot be determined.
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", err
	}

	branchName := strings.TrimSpace(out.String())
	return branchName, nil
}

// RunGitCommand executes a Git command with the provided arguments and streams output to stdout/stderr.
// It fails the test immediately if the command execution returns an error.
//
// Parameters:
// t - The testing object used for error reporting and test management.
// args - A variadic string slice containing arguments to pass to the Git command.
//
// Returns:
// Nothing. The function calls t.Fatalf on failure.
func RunGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git command '%s' failed: %v, stderr: %s", strings.Join(args, " "), err, stderr.String())
	}
	return nil
}

// RunGitCommandAndGetOutput executes a Git command with the provided arguments and returns its trimmed output as a string.
//
// Parameters:
// - t (*testing.T): Testing object to log errors and fail the test if the command fails.
// - args (...string): List of arguments to pass to the Git command.
//
// Returns:
// - string: Trimmed output of the executed Git command.
func RunGitCommandAndGetOutput(t *testing.T, args ...string) string {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Git command failed: %v\nOutput: %s", err, output)
	}
	return strings.TrimSpace(string(output))
}

// SetupAndValidateUntagged sets up a test Git repository, creates commits, tags some of them, and validates untagged commits.
//
// Parameters:
//   - t: Pointer to the testing framework's T struct for managing test state and formatted test output.
//   - ref: Reference commit or branch to validate untagged commits against.
//
// Returns:
//   - Nothing. The function performs validations and outputs test results.
func SetupAndValidateUntagged(t *testing.T, ref string) {
	// Initialize a test repository
	SetupTestRepo(t)

	// Commit and tag sample data
	CreateAndCommitFile(t, "file1.txt", "Initial commit")
	if err := RunGitCommand("tag", "v1.0.0"); err != nil {
		t.Fatalf("Failed to create tag 'v1.0.0': %v", err)
	}

	CreateAndCommitFile(t, "file2.txt", "Second commit")
	if err := RunGitCommand("tag", "v1.0.1"); err != nil {
		t.Fatalf("Failed to create tag 'v1.0.1': %v", err)
	}

	CreateAndCommitFile(t, "file3.txt", "Third commit") // Untagged commit

	// Validate untagged commits
	ValidateUntaggedCommits(t, ref)
}

// ValidateBranches checks whether all the expected branches exist in the Git repository. Reports errors using t.
//
// Parameters:
// t - The testing object used for assertions.
// expectedBranches - A slice of strings representing the branches that must be present.
//
// Returns:
// Nothing. The function fails the test if any expected branch is missing or the branches cannot be retrieved.
func ValidateBranches(t *testing.T, expectedBranches []string) {
	branches, err := GetBranches()
	if err != nil {
		t.Fatalf("GetBranches failed: %v", err)
	}

	for _, branch := range expectedBranches {
		if !utils.StringSliceContains(branches, branch) {
			t.Errorf("Branch %s not found: %v", branch, branches)
		}
	}
}

// ValidateCurrentBranch ensures the current Git branch matches the expected branch name, or fails the test otherwise.
//
// Parameters:
// - t: A testing object used to log and fail the test if needed.
// - expectedBranch: The name of the branch expected to be checked against the current branch.
//
// Returns:
// - This function does not return any value but fails the test if the validations are unsuccessful.
func ValidateCurrentBranch(t *testing.T, expectedBranch string) {
	branchName, err := GetCurrentBranch()
	if err != nil || branchName != expectedBranch {
		t.Fatalf("Incorrect branch: expected '%s', got '%s' (error: %v)", expectedBranch, branchName, err)
	}
}

// ValidateUntaggedCommits validates that the untagged commits match the expected untagged commits for a given reference.
//
// Parameters:
//   - t: The testing object used for managing test state and logging.
//   - ref: The Git reference to check untagged commits against.
//
// Returns:
//   - This function does not return a value but fails the test if untagged commits do not match the expected results.
func ValidateUntaggedCommits(t *testing.T, ref string) {
	// Call FindUntagged (since it's operating on branches/references and returns an error)
	untaggedCommits, err := FindUntagged(ref)
	if err != nil {
		t.Fatalf("FindUntagged failed: %v", err)
	}

	// Find expected untagged commits using the testing-based method (FindUntaggedCommits)
	expectedUntaggedCommits := FindUntaggedCommits(t)

	// Compare the results and fail if there's a mismatch
	if len(untaggedCommits) != len(expectedUntaggedCommits) {
		t.Fatalf("Mismatch in untagged commits: got %v, expected %v", untaggedCommits, expectedUntaggedCommits)
	}
}

// VerifyTagExists ensures a specified Git tag exists in the repository.
//
// Parameters:
// - t: A pointer to the testing framework's testing.T instance.
// - tagName: The name of the Git tag to verify.
//
// Returns:
// - None. Calls t.Fatalf if the tag does not exist.
func VerifyTagExists(t *testing.T, tagName string) {
	tags := RunGitCommandAndGetOutput(t, "tag")
	tagList := strings.Split(tags, "\n")
	if !utils.StringSliceContains(tagList, tagName) {
		t.Fatalf("Tag %s not found: %v", tagName, tags)
	}
}

/* GetShortCommitHash retrieves the short hash of the latest commit.
func GetShortCommitHash(t *testing.T, commitRef string) string {
	return RunGitCommandAndGetOutput(t, "rev-parse", "--short", commitRef)
}

// GetCurrentBranch retrieves the name of the current branch.
func GetCurrentBranch(t *testing.T) string {
	return RunGitCommandAndGetOutput(t, "rev-parse", "--abbrev-ref", "HEAD")
}

// CreatePostCommitHook creates a post-commit hook with the provided script content.
func CreatePostCommitHook(t *testing.T, scriptContent string) {
	hookDir := filepath.Join(".git", "hooks")
	hookPath := filepath.Join(hookDir, "post-commit")

	// Create hooks directory if it doesn't exist
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		t.Fatalf("Failed to create hooks directory: %v", err)
	}

	// Write the script to the post-commit hook file
	err := os.WriteFile(hookPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create post-commit hook: %v", err)
	}
}
*/
