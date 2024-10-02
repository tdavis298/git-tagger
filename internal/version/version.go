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
		log.Printf("No tags found on the branch. Starting from v0.0.0.")
		latestTag = "v0.0.0" // Assume initial tag as v0.0.0 to start from v0.0.1
	}

	// Start versioning from the latest tag
	nextTag, err := IncrementVersion(latestTag, "patch")
	if err != nil {
		return fmt.Errorf("failed to increment version: %w", err)
	}

	// Tag each untagged commit
	for _, commit := range untaggedCommits {
		// Get the commit message and determine the increment level
		commitMessage, err := git.GetCommitMessage(commit)
		if err != nil {
			return fmt.Errorf("failed to get commit message for %s: %w", commit, err)
		}

		incrementLevel := determineIncrementLevel(commitMessage)
		nextTag, err = IncrementVersion(nextTag, incrementLevel)
		if err != nil {
			return fmt.Errorf("failed to increment version: %w", err)
		}

		// Get the first 7 characters of the commit hash
		commitHash, err := git.GetCommitHash(commit)
		if err != nil {
			return fmt.Errorf("failed to get commit hash for %s: %w", commit, err)
		}
		shortHash := commitHash[:7] // Get the first 7 characters of the hash

		// Append the hash to the tag
		versionedTag := fmt.Sprintf("%s-%s", nextTag, shortHash)
		fmt.Printf("Tagging commit %s with %s\n", commit, versionedTag)

		// Create a tag for the untagged commit
		err = git.CreateTag(versionedTag, fmt.Sprintf("Automated tagging for commit %s", commit))
		if err != nil {
			return fmt.Errorf("failed to create tag %s for commit %s: %w", versionedTag, commit, err)
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
