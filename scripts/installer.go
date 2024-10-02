package scripts

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	hookSource = "hooks/post-commit"      // Source file for the post-commit hook template
	hookDest   = ".git/hooks/post-commit" // Destination of the hook in a Git repository
)

// InstallGitHook sets up the post-commit hook by copying or modifying it
func InstallGitHook() error {
	// Generate the required line to be appended
	requiredLine, err := generateRequiredLine()
	if err != nil {
		return err
	}

	// Check if the hook file already exists
	if _, err := os.Stat(hookDest); err == nil {
		fmt.Println("A post-commit hook already exists. Checking for required line...")

		// Check if the required line is present in the existing hook
		contains, err := hookContainsLine(hookDest, requiredLine)
		if err != nil {
			return fmt.Errorf("failed to check post-commit hook contents in %s: %w", hookDest, err)
		}

		if contains {
			fmt.Println("The post-commit hook already contains the required line. No action needed.")
			return nil
		}

		// Append the required line if it's missing
		fmt.Println("Appending required line to existing post-commit hook.")
		return appendLineToFile(hookDest, requiredLine)
	}

	// If the hook does not exist, copy the hook file and make it executable
	fmt.Println("No existing post-commit hook found. Installing new hook.")
	if err := copyFile(hookSource, hookDest); err != nil {
		return fmt.Errorf("failed to copy post-commit hook from %s to %s: %w", hookSource, hookDest, err)
	}

	// Append the required line
	if err := appendLineToFile(hookDest, requiredLine); err != nil {
		return err
	}

	// Make sure the new hook file is executable
	if err := os.Chmod(hookDest, 0755); err != nil {
		return fmt.Errorf("failed to make hook file executable: %w", err)
	}

	return nil
}

// CleanGitHook removes the lines added by the installer from the post-commit hook
func CleanGitHook() error {
	// Generate the required line to identify it
	requiredLine, err := generateRequiredLine()
	if err != nil {
		return err
	}

	// Read the existing file content
	input, err := os.ReadFile(hookDest)
	if err != nil {
		return fmt.Errorf("failed to read post-commit hook (%s): %w", hookDest, err)
	}

	// Split the file content into lines
	lines := strings.Split(string(input), "\n")

	// Open the file for writing (truncate the file first)
	file, err := os.OpenFile(hookDest, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open post-commit hook for cleaning (%s): %w", hookDest, err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// Re-write only the lines that do not contain the identifier added by git-tagger
	for _, line := range lines {
		if !strings.Contains(line, "# Added by git-tagger") && !strings.Contains(line, requiredLine) {
			if _, err := file.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("failed to write cleaned post-commit hook (%s): %w", hookDest, err)
			}
		}
	}

	fmt.Println("Successfully cleaned post-commit hook.")
	return nil
}

// hookContainsLine checks if all lines in the required content exist together in sequence in a file.
func hookContainsLine(filePath, requiredContent string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

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

// appendLineToFile appends a line to a file and makes it executable
func appendLineToFile(filePath, line string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	_, err = file.WriteString("\n" + line + "\n")
	if err != nil {
		return err
	}

	// Make sure the hook file is executable
	return os.Chmod(filePath, 0755)
}

// copyFile copies a file from src to dst and ensures the destination file is executable
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {

		}
	}(srcFile)

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func(dstFile *os.File) {
		err := dstFile.Close()
		if err != nil {

		}
	}(dstFile)

	scanner := bufio.NewScanner(srcFile)
	for scanner.Scan() {
		_, err = dstFile.WriteString(scanner.Text() + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to destination file: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading source file: %w", err)
	}

	// Make sure the new hook file is executable
	return os.Chmod(dst, 0755)
}

// generateRequiredLine generates the required line to be appended to the hook file.
func generateRequiredLine() (string, error) {
	// Get the root directory of the Git repository
	gitRootCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := gitRootCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git root directory: %w", err)
	}
	gitRoot := strings.TrimSpace(string(output))

	// Define the relative path to your executable within the repository
	relativeExecPath := "cmd/tagger"

	// Join the git root and the relative path to form the full path to the executable
	execPath := filepath.Join(gitRoot, relativeExecPath)

	// Ensure the path is fully resolved if there are any symlinks
	realExecPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate symlinks for executable path: %w", err)
	}

	// Generate the line to be appended with the correct path
	line := fmt.Sprintf("# Added by git-tagger\n%s -version-tag", realExecPath)

	// Debugging output
	fmt.Println("Generated line:", line)

	return line, nil
}
