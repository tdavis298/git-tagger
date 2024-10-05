package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ---------- File Utility Functions ----------

// appendLineToFile appends a line to a file, ensuring the file is executable.
// parameters:
// - filePath: the path to the file to which the line should be appended
// - line: the line to append to the file
// returns:
// - error: an error object if something went wrong, otherwise nil
func AppendLineToFile(filePath, content string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	// Append the full content to the file, not just a single line
	_, err = file.WriteString("\n" + content + "\n")
	if err != nil {
		return err
	}

	// Make sure the hook file is executable
	return os.Chmod(filePath, 0755)
}

// convertToUnixPath converts a Windows-style file path to a Unix-style file path.
// It is particularly useful for compatibility in environments that require Unix-style paths,
// such as Windows Subsystem for Linux (WSL).
//
// Parameters:
// - path (string): The original Windows-style file path to be converted.
//
// Returns:
// - string: The converted Unix-style file path.
func ConvertToUnixPath(path string) string {
	// Convert drive letter (e.g., D:) to /mnt/d
	re := regexp.MustCompile(`^([a-zA-Z]):\\`)
	path = re.ReplaceAllString(path, `/mnt/$1/`)

	// Replace backslashes with forward slashes
	path = strings.ReplaceAll(path, `\`, `/`)

	// Convert drive letter to lowercase
	return strings.ToLower(path)
}

// hookContainsFullContent checks if a given line is present in a slice of hook lines.
// parameters:
// - line: the line to check for
// - hookLines: the slice of hook lines to check within
// returns:
// - bool: true if the line is found, false otherwise
func HookContainsFullContent(filePath, requiredContent string) (bool, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	return strings.Contains(string(fileContent), strings.TrimSpace(requiredContent)), nil
}

// generateHookContent generates the content for a git post-commit hook.
// returns:
// - string: the generated hook content
// - error: an error object if something went wrong, otherwise nil
func GenerateHookContent() (string, error) {
	// Get the root directory of the Git repository
	gitRootCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := gitRootCmd.Output()
	if err != nil {
		return "", WrapErrorf("failed to get git root directory: %w", err)
	}
	gitRoot := strings.TrimSpace(string(output))

	// Define the relative path to your executable within the repository
	relativeExecPath := "bin/tagger"

	// Join the git root and the relative path to form the full path to the executable
	execPath := filepath.Join(gitRoot, relativeExecPath)

	// Ensure the path is fully resolved if there are any symlinks
	realExecPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", WrapErrorf("failed to resolve symlinks: %w", err)
	}

	// Convert to Unix-style path if running in WSL
	realExecPath = strings.ReplaceAll(realExecPath, "\\", "/")

	// Generate the full script to be added as the post-commit hook
	hookContent := fmt.Sprintf(`# Added by versioning tool

export GIT_POST_COMMIT="true"

# Execute the versioning tool
"%s" -version-tag
`, realExecPath)

	return hookContent, nil
}

/* hookContainsLine checks if a hook file contains the required content in sequence.
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
*/

// writeFile writes content to a file, creating or truncating the file first.
// parameters:
// - filePath: the path to the file to write to
// - content: the content to write to the file
// returns:
// - error: an error object if something went wrong, otherwise nil
func WriteFile(filePath, content string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return WrapErrorf("failed to open or create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = file.WriteString(content)
	if err != nil {
		return WrapErrorf("failed to write content to file: %w", err)
	}

	return nil
}

// ---------- Utility Functions ----------

// CompareSemVer compares two semantic version strings.
// Parameters:
// - a: the first semantic version string to compare (e.g., "v1.2.3").
// - b: the second semantic version string to compare (e.g., "v2.0.1").
// Returns:
//   - int: an integer indicating the result of the comparison:
//     -1 if version a is less than version b
//     1 if version a is greater than version b
//     0 if version a is equal to version b
//
// The function assumes the input versions follow the semantic versioning format (vX.Y.Z).
// It will panic if the version strings do not have exactly three parts or if the parts
// are not valid integers.
func CompareSemVer(a, b string) int {
	// Validate both tags as semantic versions
	if !IsSemVer(a) || !IsSemVer(b) {
		panic("Tags must be in semantic version format (vX.Y.Z)")
	}

	// Remove 'v' prefix from both tags
	aParts := strings.Split(a[1:], ".")
	bParts := strings.Split(b[1:], ".")

	for i := 0; i < 3; i++ {
		aNum, errA := strconv.Atoi(aParts[i])
		bNum, errB := strconv.Atoi(bParts[i])

		// Handle parsing errors
		if errA != nil || errB != nil {
			fmt.Printf("Skipping invalid version parts: %v, %v\n", aParts[i], bParts[i])
			return 0
		}

		if aNum < bNum {
			return -1
		} else if aNum > bNum {
			return 1
		}
	}

	return 0
}

/* BuildExecutable builds the Go executable for the project.
// - projectRoot: the root directory of the project where the `cmd/tagger/main.go` is located
// - outputPath: the path to output the built binary
// returns:
// - error: an error object if something went wrong, otherwise nil
func BuildExecutable(projectRoot string, outputPath string) error {
	// Absolute path to the Go file
	mainGoPath := filepath.Join(projectRoot, "cmd/tagger/main.go")

	// Construct the `go build` command
	cmd := exec.Command("go", "build", "-o", outputPath, mainGoPath)

	// Execute the build command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build executable: %w", err)
	}

	fmt.Println("Executable built successfully at:", outputPath)
	return nil
}
*/

// FilterEmptyStrings removes empty or whitespace-only strings from a slice.
// parameters:
// - slice: a slice of strings to filter
// returns:
// - []string: a slice containing only non-empty strings
func FilterEmptyStrings(slice []string) []string {
	var result []string
	for _, str := range slice {
		if trimmed := strings.TrimSpace(str); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// IsSemVer checks if a given string is a semantic version
// parameters:
// - version: the version string to check
// returns:
// - bool: true if the string is a semantic version, otherwise false
func IsSemVer(tag string) bool {
	// Regex to strictly match semantic versions like vX.Y.Z where X, Y, and Z are integers
	// It allows an optional hash suffix (e.g., v1.2.3-hash)
	semVerPattern := `^v\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`
	match, _ := regexp.MatchString(semVerPattern, tag)
	return match
}

// StripHashSuffix removes the hash suffix from a version tag if present.
// parameters:
// - tag: the version tag with potential hash suffix
// returns:
// - string: the core semantic version without hash
func StripHashSuffix(tag string) string {
	// Split by the dash to separate version and hash
	parts := strings.SplitN(tag, "-", 2)
	return parts[0] // Return the core semantic version without the hash
}
