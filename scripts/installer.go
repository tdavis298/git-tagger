package scripts

import (
	"bufio"
	"fmt"
	"git-tagger/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	hookDest = ".git/hooks/post-commit" // destination of the hook in a Git repository
)

// git hook functions:

// CleanGitHook removes content added by the versioning tool from the post-commit hook.
// returns:
// - error: an error object if something went wrong, otherwise nil
func CleanGitHook() error {
	// Generate the hook content to identify what was added
	hookContent, err := generateHookContent()
	if err != nil {
		return err
	}

	// Read the existing file content
	input, err := os.ReadFile(hookDest)
	if err != nil {
		return utils.WrapErrorf("failed to read post-commit hook (%s): %w", err, hookDest)
	}

	// Split the file content into lines
	lines := strings.Split(string(input), "\n")

	// Open the file for writing (truncate the file first)
	file, err := os.OpenFile(hookDest, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return utils.WrapErrorf("failed to open post-commit hook for cleaning (%s): %w", err, hookDest)
	}
	defer func() {
		_ = file.Close()
	}()

	// Prepare to identify lines added by the versioning tool
	hookLines := strings.Split(hookContent, "\n")
	var inHookBlock bool

	// Re-write only lines that are not part of the added content
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Identify the start of the added block
		if strings.Contains(trimmedLine, "# Added by versioning tool") {
			inHookBlock = true
			continue
		}

		// Skip lines within the block until the block ends
		if inHookBlock && containsHookLine(trimmedLine, hookLines) {
			continue
		} else {
			inHookBlock = false
		}

		// Write lines that are not part of the block
		if _, err := file.WriteString(line + "\n"); err != nil {
			return utils.WrapErrorf("failed to write cleaned post-commit hook (%s): %w", err, hookDest)
		}
	}

	fmt.Println("Successfully cleaned post-commit hook.")
	return nil
}

// InstallGitHook installs or updates the post-commit hook with necessary content.
// returns:
// - error: an error object if something went wrong, otherwise nil
func InstallGitHook(executablePath string) error {
	// Ensure the executablePath is an absolute path
	absPath, err := filepath.Abs(executablePath)
	if err != nil {
		return utils.WrapErrorf("failed to resolve absolute path: %w", err)
	}

	// Generate the hook content
	hookContent := fmt.Sprintf("# Added by versioning tool\n%s -version-tag", absPath)

	// Check if the hook file already exists
	if _, err := os.Stat(hookDest); err == nil {
		// Hook exists, check for the necessary content
		fmt.Println("A post-commit hook already exists. Checking for required content...")

		// Read the current hook content
		hookContains, err := hookContainsLine(hookDest, hookContent)
		if err != nil {
			return utils.WrapErrorf("failed to check existing hook content: %w", err)
		}

		if hookContains {
			fmt.Println("The post-commit hook already contains the necessary content.")
			return nil
		}

		// Append the required content if not already present
		err = appendLineToFile(hookDest, hookContent)
		if err != nil {
			return utils.WrapErrorf("failed to append content to existing post-commit hook: %w", err)
		}
		fmt.Println("Appended content to existing post-commit hook.")
		return nil
	}

	// If the hook does not exist, create it and add the necessary content
	fmt.Println("No existing post-commit hook found. Installing new hook.")

	// Write the new hook content
	err = writeFile(hookDest, hookContent)
	if err != nil {
		return utils.WrapErrorf("failed to write post-commit hook: %w", err)
	}

	fmt.Println("Post-commit hook installed successfully.")
	return nil
}

// file utility functions:

// appendLineToFile appends a line to a file, ensuring the file is executable.
// parameters:
// - filePath: the path to the file to which the line should be appended
// - line: the line to append to the file
// returns:
// - error: an error object if something went wrong, otherwise nil
func appendLineToFile(filePath, line string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = file.WriteString("\n" + line + "\n")
	if err != nil {
		return err
	}

	// Make sure the hook file is executable
	return os.Chmod(filePath, 0755)
}

// containsHookLine checks if a given line is present in a slice of hook lines.
// parameters:
// - line: the line to check for
// - hookLines: the slice of hook lines to check within
// returns:
// - bool: true if the line is found, false otherwise
func containsHookLine(line string, hookLines []string) bool {
	trimmedLine := strings.TrimSpace(line)
	for _, hookLine := range hookLines {
		if strings.TrimSpace(hookLine) == trimmedLine {
			return true
		}
	}
	return false
}

// generateHookContent generates the content for a git post-commit hook.
// returns:
// - string: the generated hook content
// - error: an error object if something went wrong, otherwise nil
func generateHookContent() (string, error) {
	// Get the root directory of the Git repository
	gitRootCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := gitRootCmd.Output()
	if err != nil {
		return "", utils.WrapErrorf("failed to get git root directory: %w", err)
	}
	gitRoot := strings.TrimSpace(string(output))

	// Define the relative path to your executable within the repository
	relativeExecPath := "./"

	// Join the git root and the relative path to form the full path to the executable
	execPath := filepath.Join(gitRoot, relativeExecPath)

	// Ensure the path is fully resolved if there are any symlinks
	realExecPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", utils.WrapErrorf("failed to evaluate symlinks for executable path: %w", err)
	}

	// Generate the full script to be added as the post-commit hook
	hookContent := fmt.Sprintf(`# Added by versioning tool
# This script triggers the versioning logic after every commit.

# Navigate to the Git project root
cd "$(git rev-parse --show-toplevel)" || exit

# Path to your Go program (assuming you have built it already)
GIT_TAGGER_PATH="%s"

# Check if the versioning executable exists
if [ ! -f "$GIT_TAGGER_PATH" ]; then
  echo "Versioning executable not found at $GIT_TAGGER_PATH"
  exit 1
fi

# Run the Go program with the necessary flag for version tagging
"$GIT_TAGGER_PATH" -version-tag

# Check if the tagging was successful
if [ $? -eq 0 ]; then
  echo "Version tag updated successfully."
else
  echo "Failed to update version tag."
  exit 1
fi
`, realExecPath)

	return hookContent, nil
}

// hookContainsLine checks if a hook file contains the required content in sequence.
// parameters:
// - filePath: the path to the hook file to check
// - requiredContent: the content to check for within the hook file
// returns:
// - bool: true if the required content is found in sequence, false otherwise
// - error: an error object if something went wrong, otherwise nil
func hookContainsLine(filePath, requiredContent string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = file.Close()
	}()

	// Split the required content into lines for comparison
	requiredLines := strings.Split(strings.TrimSpace(requiredContent), "\n")

	// Read the file line by line and check for the required content in sequence
	scanner := bufio.NewScanner(file)
	var currentIndex int

	for scanner.Scan() {
		trimmedLine := strings.TrimSpace(scanner.Text())
		if trimmedLine == strings.TrimSpace(requiredLines[currentIndex]) {
			currentIndex++
			// If all required lines have been found in sequence
			if currentIndex == len(requiredLines) {
				return true, nil
			}
		} else {
			// Reset the index if sequence breaks
			currentIndex = 0
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// writeFile writes content to a file, creating or truncating the file first.
// parameters:
// - filePath: the path to the file to write to
// - content: the content to write to the file
// returns:
// - error: an error object if something went wrong, otherwise nil
func writeFile(filePath, content string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return utils.WrapErrorf("failed to open or create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = file.WriteString(content)
	if err != nil {
		return utils.WrapErrorf("failed to write content to file: %w", err)
	}

	return nil
}
