package utils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// utility functions:

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

// BuildExecutable builds the Go executable for the project.
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
