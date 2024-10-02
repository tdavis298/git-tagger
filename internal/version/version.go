package version

import (
	"fmt"
	"git-tagger/internal/git"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// IncrementVersion increments a semantic version based on a single level.
func IncrementVersion(latestTag string, level string) (string, error) {
	parts := strings.Split(strings.TrimPrefix(latestTag, "v"), ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid version format: %s", latestTag)
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	// Increment the version based on the level determined
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

	// Start with the latest tag found on the branch (if any)
	latestTag, err := git.GetLatestTag()
	if err != nil {
		// No tags found; start from v0.0.1 directly
		log.Printf("No tags found on the branch. Starting from v0.0.1.")
		latestTag = "v0.0.1"
	}

	// Track the current tag
	currentTag := latestTag

	// Tag each untagged commit sequentially
	for i, commit := range untaggedCommits {
		// Get the commit message
		message, err := git.GetCommitMessage(commit)
		if err != nil {
			return fmt.Errorf("failed to retrieve commit message for %s: %w", commit, err)
		}

		// Determine increment level for the current commit message
		incrementLevel, versionFromMessage := determineIncrementLevel(message)

		// If a version tag is found in the message, update the currentTag to that version
		if versionFromMessage != "" {
			currentTag = versionFromMessage
		} else {
			// Increment the tag version based on the commit message
			if i == 0 && currentTag == "v0.0.1" {
				// Use currentTag as is for the first commit if no tag exists
				fmt.Printf("Starting tagging from %s\n", currentTag)
			} else {
				currentTag, err = IncrementVersion(currentTag, incrementLevel)
				if err != nil {
					return fmt.Errorf("failed to increment version: %w", err)
				}
			}
		}

		// Get the short hash of the commit
		shortHash, err := git.GetShortCommitHash(commit)
		if err != nil {
			return fmt.Errorf("failed to get short hash for commit %s: %w", commit, err)
		}

		// Append the hash to the tag
		tagWithHash := fmt.Sprintf("%s-%s", currentTag, shortHash)

		fmt.Printf("Tagging commit %s with %s\n", commit, tagWithHash)

		// Create a tag for the specific commit
		err = git.CreateTag(tagWithHash, fmt.Sprintf("Automated tagging for commit %s", commit), commit)
		if err != nil {
			return fmt.Errorf("failed to create tag %s for commit %s: %w", tagWithHash, commit, err)
		}
	}

	fmt.Println("Successfully tagged all untagged commits.")
	return nil
}

// determineIncrementLevel determines the level of version increment based on commit message.
// It returns the increment level and any version found within the message.
func determineIncrementLevel(commitMessage string) (string, string) {
	if strings.Contains(commitMessage, "BREAKING CHANGE") {
		return "major", ""
	} else if strings.HasPrefix(commitMessage, "feat") {
		return "minor", ""
	} else if strings.HasPrefix(commitMessage, "fix") {
		return "patch", ""
	}

	// Check if the commit message contains a version tag
	versionTag := extractVersionTag(commitMessage)
	if versionTag != "" {
		fmt.Printf("Version tag \"%s\" found in commit message. Using this as the base version.\n", versionTag)
		return "", versionTag
	}

	// Notify user if commit message is unrecognized
	fmt.Printf("Unrecognized commit message: \"%s\". Defaulting to patch update.\n", commitMessage)
	return "patch", "" // Default to patch if none of the keywords match
}

// extractVersionTag extracts a version tag (vX.Y.Z format) from the commit message if present.
func extractVersionTag(commitMessage string) string {
	versionPattern := regexp.MustCompile(`v\d+\.\d+\.\d+`)
	versionTag := versionPattern.FindString(commitMessage)
	return versionTag
}
