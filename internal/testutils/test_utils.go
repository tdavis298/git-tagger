package testutils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// SetupTestRepo initializes a new Git repository for testing.
func SetupTestRepo(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-repo")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	t.Cleanup(func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Fatalf("Failed to clean up temp directory: %v", err)
		}
	})

	// Change to the temporary directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Initialize a new Git repository
	RunGitCommand(t, "init")

	// Create an initial commit
	CreateAndCommitFile(t, "README.md", "Initial commit")
}

// CreateAndCommitFile creates a file with the specified filename and commits it.
func CreateAndCommitFile(t *testing.T, filename, commitMsg string) {
	err := os.WriteFile(filename, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file %s: %v", filename, err)
	}

	RunGitCommand(t, "add", filename)
	RunGitCommand(t, "commit", "-m", commitMsg)
}

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

// FindUntaggedCommits returns a list of commit hashes on the current branch that do not have associated tags.
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

// IsWSL checks if the code is running inside a WSL environment.
func IsWSL() bool {
	// Read the contents of /proc/version
	versionData, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}

	// Check if the contents mention "Microsoft" or "WSL"
	return strings.Contains(string(versionData), "Microsoft") || strings.Contains(string(versionData), "WSL")
}

/* Mock function to wrap errors
func MockWrapErrorf(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}
*/

// RunGitCommand runs a Git command without returning its output.
func RunGitCommand(t *testing.T, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Git command failed: %v", err)
	}
}

// RunGitCommandAndGetOutput runs a Git command and returns its output as a string.
func RunGitCommandAndGetOutput(t *testing.T, args ...string) string {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Git command failed: %v\nOutput: %s", err, output)
	}
	return strings.TrimSpace(string(output))
}

// VerifyCommitMessage checks if the commit with the given hash has the expected commit message.
func VerifyCommitMessage(t *testing.T, commitHash, expectedMessage string) {
	// Get the commit message of the specified commit
	commitMessage := RunGitCommandAndGetOutput(t, "log", "-1", "--pretty=%B", commitHash)

	// Trim whitespace for consistent comparison
	commitMessage = strings.TrimSpace(commitMessage)

	// Compare with the expected message
	if commitMessage != expectedMessage {
		t.Fatalf("Expected commit message '%s', but got '%s'", expectedMessage, commitMessage)
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
