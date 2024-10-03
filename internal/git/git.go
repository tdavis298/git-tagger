package git

import (
	"bufio"
	"bytes"
	"fmt"
	"git-tagger/internal/utils"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// tagging functions:

// CreateTag creates an annotated Git tag with the given tag name, message, and commit.
// Parameters:
// - tag: The name of the tag to create
// - message: The message to annotate the tag with
// - commit: The commit hash to tag
// Returns:
// - error: An error object if something went wrong, otherwise nil
func CreateTag(tag, message, commit string) error {
	return runGitCommandVoid("tag", "-a", tag, "-m", message, commit)
}

// GetLatestTag retrieves the latest semantic version tag from the git repository.
// Returns:
// - string: The latest tag as a string
// - error: An error object if something went wrong or if no semantic version tags were found
func GetLatestTag() (string, error) {
	// Get all tags
	tags, err := runGitCommand("tag")
	if err != nil {
		return "", utils.WrapErrorf("failed to retrieve tags: %w", err)
	}

	// filter tags that match the semantic versioning format
	var semVerTags []string
	for _, tag := range tags {
		if isSemVer(tag) {
			semVerTags = append(semVerTags, tag)
		}
	}

	if len(semVerTags) == 0 {
		return "", utils.WrapErrorf("no semantic version tags found", err)
	}

	// sort tags to find the latest version
	sort.Slice(semVerTags, func(i, j int) bool {
		return compareSemVer(semVerTags[i], semVerTags[j]) < 0
	})

	// return the highest (latest) version tag
	return semVerTags[len(semVerTags)-1], nil
}

// GetTagsForCommit retrieves tags for a specific commit.
// parameters:
// - commit: the commit hash for which to retrieve associated tags
// returns:
// - []string: a slice of tag names associated with the commit
// - error: an error object if something went wrong, otherwise nil
func GetTagsForCommit(commit string) ([]string, error) {
	tags, err := runGitCommand("tag", "--contains", commit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tags for commit %s: %w", commit, err)
	}
	return tags, nil
}

// commit functions:

// FindUntagged finds true untagged commits in a given branch.
// parameters:
// - branch: the branch from which to find untagged commits
// returns:
// - []string: a slice of commit hashes that are untagged
// - error: an error object if something went wrong, otherwise nil
func FindUntagged(branch string) ([]string, error) {
	// get all commits on the branch
	commits, err := runGitCommand("rev-list", "--reverse", branch)
	if err != nil {
		return nil, err
	}

	// filter out commits that already have tags
	var untaggedCommits []string
	for _, commit := range commits {
		tags, err := GetTagsForCommit(commit)
		if err != nil {
			return nil, err
		}
		if len(tags) == 0 {
			untaggedCommits = append(untaggedCommits, commit)
		}
	}

	return untaggedCommits, nil
}

// GetShortCommitHash retrieves the short form of a given commit hash.
// parameters:
// - commit: the full commit hash to shorten
// returns:
// - string: the short form of the commit hash
// - error: an error object if something went wrong, otherwise nil
func GetShortCommitHash(commit string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", commit)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get short hash: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetCommitMessage retrieves the commit message for a given commit hash.
// parameters:
// - commit: the commit hash for which to retrieve the message
// returns:
// - string: the commit message
// - error: an error object if something went wrong, otherwise nil
func GetCommitMessage(commit string) (string, error) {
	cmd := exec.Command("git", "show", "-s", "--format=%s", commit)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit message: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// branch functions:

// GetBranches retrieves all local branches, trimming any leading '*' character
// Returns:
// - []string: A slice of branch names
// - error: An error object if something went wrong, otherwise nil
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
// parameters:
// - branches: a slice of strings containing the branch names to choose from
// returns:
// - string: the name of the selected branch
// - error: an error object if something went wrong, otherwise nil
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

// utility functions:

// filterEmptyStrings removes empty strings from the provided slice of strings and trims whitespace from each string.
// Parameters:
// - slice: The slice of strings to filter
// Returns:
// - []string: A slice containing non-empty, trimmed strings
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

// filterEmptyStrings removes empty or whitespace-only strings from a slice.
// parameters:
// - slice: a slice of strings to filter
// returns:
// - []string: a slice containing only non-empty strings
func filterEmptyStrings(slice []string) []string {
	var result []string
	for _, str := range slice {
		if trimmed := strings.TrimSpace(str); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// isSemVer checks if a given string is a semantic version
// parameters:
// - version: the version string to check
// returns:
// - bool: true if the string is a semantic version, otherwise false
func isSemVer(tag string) bool {
	return strings.HasPrefix(tag, "v") && len(strings.Split(tag[1:], ".")) == 3
}

// runGitCommand executes a git command and returns the output as a slice of strings
// parameters:
// - args: the arguments for the git command
// returns:
// - []string: the output lines from the git command
// - error: an error object if something went wrong, otherwise nil
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

// runGitCommandVoid executes a git command without requiring the output
// parameters:
// - args: the arguments for the git command
// returns:
// - error: an error object if something went wrong, otherwise nil
func runGitCommandVoid(args ...string) error {
	cmd := exec.Command("git", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run git command: %w", err)
	}
	return nil
}
