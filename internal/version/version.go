package version

import (
	"fmt"
	"git-tagger/internal/git"
	"log"
	"strconv"
	"strings"
)

// IncrementVersion increments a semantic version (e.g., major, minor, patch).
func IncrementVersion(latestTag, level string) (string, error) {
	parts := strings.Split(strings.TrimPrefix(latestTag, "v"), ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid version format: %s", latestTag)
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	switch level {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", fmt.Errorf("unknown version increment level: %s", level)
	}

	newTag := fmt.Sprintf("v%d.%d.%d", major, minor, patch)
	return newTag, nil
}

// UpdateUntaggedCommits finds untagged commits on a branch and tags them in order based on the nearest tagged commit.
func UpdateUntaggedCommits(branch string) error {
	// Find all untagged commits
	untaggedCommits, err := git.FindUntagged(branch)
	if err != nil {
		return fmt.Errorf("failed to find untagged commits: %w", err)
	}

	if len(untaggedCommits) == 0 {
		fmt.Println("No untagged commits found.")
		return nil
	}

	// Find the latest tag on the branch (if any)
	latestTag, err := git.GetLatestTag()
	if err != nil {
		// No tags found; start from v0.0.1 directly
		log.Printf("No tags found on the branch. Starting from v0.0.1.")
		latestTag = "v0.0.0" // Initialize latestTag as "v0.0.0"
	}

	// Start versioning from v0.0.1 if no tags exist
	nextTag, err := IncrementVersion(latestTag, "patch")
	if err != nil {
		return fmt.Errorf("failed to increment version: %w", err)
	}

	// Tag each untagged commit
	for i, commit := range untaggedCommits {
		// Only increment for subsequent commits
		if i > 0 {
			nextTag, err = IncrementVersion(nextTag, "patch")
			if err != nil {
				return fmt.Errorf("failed to increment version: %w", err)
			}
		}

		// Get the short hash of the commit
		shortHash, err := git.GetShortCommitHash(commit)
		if err != nil {
			return fmt.Errorf("failed to get short hash for commit %s: %w", commit, err)
		}

		// Append the hash to the tag
		tagWithHash := fmt.Sprintf("%s-%s", nextTag, shortHash)

		fmt.Printf("Tagging commit %s with %s\n", commit, tagWithHash)

		// Create a tag for the untagged commit
		err = git.CreateTag(tagWithHash, fmt.Sprintf("Automated tagging for commit %s", commit))
		if err != nil {
			return fmt.Errorf("failed to create tag %s for commit %s: %w", tagWithHash, commit, err)
		}
	}

	fmt.Println("Successfully tagged all untagged commits.")
	return nil
}

func determineIncrementLevel(commitMessage string) string {
	if strings.Contains(commitMessage, "BREAKING CHANGE") {
		return "major"
	} else if strings.HasPrefix(commitMessage, "feat") {
		return "minor"
	} else if strings.HasPrefix(commitMessage, "fix") {
		return "patch"
	}

	// Notify user if commit message is unrecognized
	fmt.Printf("Unrecognized commit message: \"%s\". Defaulting to patch update.\n", commitMessage)
	return "patch" // Default to patch if none of the keywords match
}
