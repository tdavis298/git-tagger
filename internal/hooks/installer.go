package hooks

import (
	"fmt"
	"git-tagger/internal/utils"
	"os"
	"path/filepath"
	"strings"
)

const (
	hookDest = ".git/hooks/post-commit" // destination of the hook in a Git repository
)

// ---------- Git Hook Functions ----------

// CleanGitHook removes content added by the versioning tool from the post-commit hook.
// returns:
// - error: an error object if something went wrong, otherwise nil
func CleanGitHook() error {
	// Generate the hook content to identify what was added
	hookContent, err := utils.GenerateHookContent()
	if err != nil {
		return err
	}

	// Read the existing file content
	input, err := os.ReadFile(hookDest)
	if err != nil {
		return utils.WrapErrorf("failed to read post-commit hook: %w", err)
	}

	// Split the file content into lines
	lines := strings.Split(string(input), "\n")

	// Open the file for writing (truncate the file first)
	file, err := os.OpenFile(hookDest, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return utils.WrapErrorf("failed to open post-commit hook for cleaning: %w", err)
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

		// Check if the line is part of the hook block and handle the error properly
		contains := false
		for _, hookLine := range hookLines {
			isMatch, err := utils.HookContainsFullContent(trimmedLine, hookLine)
			if err != nil {
				return utils.WrapErrorf("failed to check hook line: %w", err)
			}

			if isMatch {
				contains = true
				break
			}
		}

		if inHookBlock && contains {
			continue
		} else {
			inHookBlock = false
		}

		// Write lines that are not part of the block
		if _, err := file.WriteString(line + "\n"); err != nil {
			return utils.WrapErrorf("failed to write cleaned post-commit hook: %w", err)
		}
	}

	fmt.Println("Successfully cleaned post-commit hook.")
	return nil
}

// InstallGitHook installs or updates the post-commit hook with necessary content.
// returns:
// - error: an error object if something went wrong, otherwise nil
func InstallGitHook(executablePath string) error {

	hookContent, err := utils.GenerateHookContent()
	if err != nil {
		return utils.WrapErrorf("failed to generate hook content: %w", err)
	}

	// Ensure the executablePath is an absolute path
	absPath, err := filepath.Abs(executablePath)
	if err != nil {
		return utils.WrapErrorf("failed to resolve absolute path: %w", err)
	}

	// Convert the path to a WSL-compatible format if needed
	absPath = utils.ConvertToUnixPath(absPath)

	// Check if the hook file already exists
	if _, err := os.Stat(hookDest); err == nil {
		// Hook exists, check for the necessary content
		fmt.Println("A post-commit hook already exists. Checking for required content...")

		// Read the current hook content using hookContainsFullContent
		hookContains, err := utils.HookContainsFullContent(hookDest, hookContent)
		if err != nil {
			return utils.WrapErrorf("failed to check existing hook content: %w", err)
		}

		if hookContains {
			fmt.Println("The post-commit hook already contains the necessary content.")
			return nil
		}

		// Append the required content if not already present
		err = utils.AppendLineToFile(hookDest, hookContent)
		if err != nil {
			return utils.WrapErrorf("failed to append content to existing post-commit hook: %w", err)
		}
		fmt.Println("Appended content to existing post-commit hook.")
		return nil
	}

	// If the hook does not exist, create it and add the necessary content
	fmt.Println("No existing post-commit hook found. Installing new hook.")

	// Write the new hook content
	err = utils.WriteFile(hookDest, hookContent)
	if err != nil {
		return utils.WrapErrorf("failed to write post-commit hook: %w", err)
	}

	fmt.Println("Post-commit hook installed successfully.")
	return nil
}
