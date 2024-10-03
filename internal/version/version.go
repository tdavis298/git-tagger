package version

import (
	"fmt"
	"git-tagger/internal/git"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// version functions

// IncrementVersion increments a semantic version based on a single level.
// parameters:
// - latestTag: the current latest semantic version tag
// - level: the level of version increment (major, minor, patch)
// returns:
// - string: the new incremented version tag
// - error: an error object if something went wrong, otherwise nil
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

// UpdateUntaggedCommits finds untagged commits on a branch, checking tags and messages for version references.
// parameters:
// - branch: the branch from which to find untagged commits
// returns:
// - error: an error object if something went wrong, otherwise nil
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
		// No tags found; start from v0.0.0 directly
		log.Printf("No tags found on the branch. Starting from v0.0.0.")
		latestTag = "v0.0.0"
	}

	// Track the current version for tagging
	currentTag := latestTag

	// Iterate through untagged commits from oldest to newest
	for _, commit := range untaggedCommits {
		// Get the commit message
		message, err := git.GetCommitMessage(commit)
		if err != nil {
			return fmt.Errorf("failed to retrieve commit message for %s: %w", commit, err)
		}

		// Check if the commit message contains a version tag
		incrementLevel, versionFromMessage := determineIncrementLevel(message)

		// If a version tag is found in the message, update the currentTag to that version
		if versionFromMessage != "" {
			currentTag = versionFromMessage
		} else {
			// Otherwise, reference existing tags
			tags, err := git.GetTagsForCommit(commit)
			if err != nil {
				return fmt.Errorf("failed to get tags for commit %s: %w", commit, err)
			}
			if len(tags) > 0 {
				currentTag = tags[0] // Assuming the first tag is the most relevant; adjust based on further logic if needed
			} else {
				// Increment the tag version based on the commit message
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

		// Create a tag for the untagged commit
		err = git.CreateTag(tagWithHash, fmt.Sprintf("Automated tagging for commit %s", commit), commit)

		if err != nil {
			return fmt.Errorf("failed to create tag %s for commit %s: %w", tagWithHash, commit, err)
		}
	}

	fmt.Println("Successfully tagged all untagged commits.")
	return nil
}

// determineIncrementLevel determines the level of version increment based on commit message.
// parameters:
// - commitMessage: the commit message to analyze
// returns:
// - string: the level of version increment (major, minor, patch)
// - string: any version found within the message
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

// utility functions

// extractVersionTag extracts a semantic version tag from a commit message.
// parameters:
// - commitMessage: the commit message from which to extract the version tag
// returns:
// - string: the extracted version tag, or an empty string if none is found
func extractVersionTag(commitMessage string) string {
	versionPattern := regexp.MustCompile(`v\d+\.\d+\.\d+`)
	versionTag := versionPattern.FindString(commitMessage)
	return versionTag
}
