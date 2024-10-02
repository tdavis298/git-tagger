package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// GetBranches retrieves all local branches, trimming any leading '*' character.
func GetBranches() ([]string, error) {
	branches, err := runGitCommand("branch")
	if err != nil {
		return nil, err
	}

	// Clean up branch names
	for i, branch := range branches {
		branches[i] = strings.TrimSpace(strings.TrimPrefix(branch, "*"))
	}

	return branches, nil
}

// SelectBranch lets the user choose a branch from the list of branches
func SelectBranch(branches []string) (string, error) {
	fmt.Println("Select a branch to update:")
	for i, branch := range branches {
		// Print the branch options, trimming any extra whitespace
		fmt.Printf("[%d] %s\n", i+1, strings.TrimSpace(branch))
	}

	// Prompt for user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the number of the branch: ")
	choiceStr, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	// Convert the choice to an integer
	choiceStr = strings.TrimSpace(choiceStr)
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(branches) {
		return "", fmt.Errorf("invalid selection: %s", choiceStr)
	}

	// Return the selected branch
	return strings.TrimSpace(branches[choice-1]), nil
}

// CreateTag creates a new Git tag with a message for a specific commit.
func CreateTag(tag, message, commit string) error {
	return runGitCommandVoid("tag", "-a", tag, "-m", message, commit)
}

// FindUntagged finds untagged commits in a given branch
func FindUntagged(branch string) ([]string, error) {
	// Get all commits on the branch
	return runGitCommand("rev-list", "--reverse", branch)
}

// runGitCommand executes a Git command and returns the output as a slice of strings
func runGitCommand(args ...string) ([]string, error) {
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run git command: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	return filterEmptyStrings(lines), nil
}

// runGitCommandVoid executes a Git command without needing to parse the output
func runGitCommandVoid(args ...string) error {
	cmd := exec.Command("git", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run git command: %w", err)
	}
	return nil
}

// isSemVer checks if a tag is a valid semantic version
func isSemVer(tag string) bool {
	return strings.HasPrefix(tag, "v") && len(strings.Split(tag[1:], ".")) == 3
}

// compareSemVer compares two semantic versions
// Returns -1 if a < b, 1 if a > b, 0 if a == b
func compareSemVer(a, b string) int {
	aParts := strings.Split(a[1:], ".")
	bParts := strings.Split(b[1:], ".")

	for i := 0; i < 3; i++ {
		aNum, _ := strconv.Atoi(aParts[i])
		bNum, _ := strconv.Atoi(bParts[i])

		if aNum < bNum {
			return -1
		} else if aNum > bNum {
			return 1
		}
	}
	return 0
}

// filterEmptyStrings removes empty strings from a slice of strings
func filterEmptyStrings(slice []string) []string {
	var result []string
	for _, str := range slice {
		if trimmed := strings.TrimSpace(str); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func GetLatestTag() (string, error) {
	// Get all tags
	tags, err := runGitCommand("tag")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve tags: %w", err)
	}

	// Filter tags that match the semantic versioning format
	var semVerTags []string
	for _, tag := range tags {
		if isSemVer(tag) {
			semVerTags = append(semVerTags, tag)
		}
	}

	if len(semVerTags) == 0 {
		return "", fmt.Errorf("no semantic version tags found")
	}

	// Sort tags to find the latest version
	sort.Slice(semVerTags, func(i, j int) bool {
		return compareSemVer(semVerTags[i], semVerTags[j]) < 0
	})

	// Return the highest (latest) version tag
	return semVerTags[len(semVerTags)-1], nil
}

// GetCommitMessage retrieves the commit message for a given commit hash.
func GetCommitMessage(commit string) (string, error) {
	cmd := exec.Command("git", "show", "-s", "--format=%s", commit)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit message: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetShortCommitHash retrieves the short hash for a given commit.
func GetShortCommitHash(commit string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", commit)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get short hash: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetCurrentBranch retrieves the name of the current Git branch
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
