package main

import (
	"flag"
	"fmt"
	"git-tagger/internal/git"
	"git-tagger/internal/hooks"
	"git-tagger/internal/utils"
	"git-tagger/internal/version"
	"os"
)

func main() {
	// define flags for more options
	branchFlag := flag.String("branch", "", "Specify the branch to tag")
	installFlag := flag.Bool("install", false, "Install the Git post-commit hook")
	uninstallFlag := flag.Bool("clean", false, "Remove the Git post-commit hook")
	versionTagFlag := flag.Bool("version-tag", false, "Tag untagged Git commits with version numbers")

	flag.Parse()

	// Detect if running in a non-interactive Git hook environment
	isNonInteractive := os.Getenv("GIT_POST_COMMIT") != ""

	// Handle non-interactive mode immediately if detected
	if isNonInteractive {
		fmt.Println("Running in non-interactive mode...")

		// Verify we're in a Git repository
		if _, err := os.Stat(".git"); os.IsNotExist(err) {
			utils.LogAndExit("No Git repository found in the current directory", nil)
		}

		// Get the currently checked out branch
		currentBranch, err := git.GetCurrentBranch()
		if err != nil {
			utils.LogAndExit("Failed to get the current branch", err)
		}

		// Ensure that the branch is not empty
		if currentBranch == "" {
			utils.LogAndExit("Current branch is not specified or repository is in a detached HEAD state", nil)
		}

		// Update untagged commits for the current branch
		err = version.UpdateUntaggedCommits(currentBranch)
		if err != nil {
			utils.LogAndExit("Failed to update untagged commits", err)
		}

		fmt.Println("Version-tagged untagged commits successfully on branch:", currentBranch)
		return
	}

	// handle unparsed flags
	utils.HandleUnparsedArgs(flag.Args())

	// handle the version tagging logic
	if *versionTagFlag {
		branch := *branchFlag

		// If branch is not specified via the flag, ask the user to select one
		if branch == "" {
			branches, err := git.GetBranches()
			if err != nil {
				utils.LogAndExit("Failed to get branches", err)
			}
			if len(branches) == 0 {
				utils.LogAndExit("No branches found in the repository", nil)
			}
		}

		// update untagged commits for the selected branch
		err := version.UpdateUntaggedCommits(branch)
		if err != nil {
			utils.LogAndExit("Failed to update untagged commits", err)
		}

		fmt.Println("Version-tagged untagged files successfully.")
		os.Exit(0)
		return
	}

	// If the install flag is provided, install hook
	if *installFlag {
		outputPath := "./tagger" // output path for the binary

		// install the Git hook
		err := hooks.InstallGitHook(outputPath)
		if err != nil {
			utils.LogAndExit("Failed to install Git hook", err)
		}
		fmt.Println("Git post-commit hook installed successfully.")
		return
	}

	// If the uninstall flag is provided, clean the Git hook
	if *uninstallFlag {
		err := hooks.CleanGitHook()
		if err != nil {
			utils.LogAndExit("Failed to uninstall Git hook", err)
		}
		fmt.Println("Git post-commit hook uninstalled successfully.")
		return
	}

	// function to list all flags and their descriptions
	if flag.NFlag() == 0 {
		fmt.Println("Available flags:")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("  -%s: %s (default: %s)\n", f.Name, f.Usage, f.DefValue)
		})
	}
}
