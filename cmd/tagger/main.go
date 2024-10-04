package main

import (
	"flag"
	"fmt"
	"git-tagger/internal/git"
	"git-tagger/internal/utils"
	"git-tagger/internal/version"
	"git-tagger/scripts"
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

		// Get the currently checked out branch
		currentBranch, err := git.GetCurrentBranch()
		if err != nil {
			fmt.Println("Failed to get the current branch:", err)
			os.Exit(1)
		}

		// Ensure that the branch is not empty
		if currentBranch == "" {
			fmt.Println("Error: Current branch is not specified or repository is in a detached HEAD state.")
			os.Exit(1)
		}

		// Update untagged commits for the current branch
		err = version.UpdateUntaggedCommits(currentBranch)
		if err != nil {
			fmt.Println("Failed to update untagged commits:", err)
			os.Exit(1)
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
				fmt.Println("Failed to get branches:", err)
				os.Exit(1)
			}
			if len(branches) == 0 {
				fmt.Println("No branches found in the repository.")
				os.Exit(1)
			}
		}

		// update untagged commits for the selected branch
		err := version.UpdateUntaggedCommits(branch)
		if err != nil {
			fmt.Println("Failed to update untagged commits:", err)
			os.Exit(1)
		}

		fmt.Println("Version-tagged untagged files successfully.")
		return
	}

	// If the install flag is provided, install hook
	if *installFlag {
		// Set project root and output paths
		// projectRoot := "."       // assuming current directory is the project root; adjust if necessary
		outputPath := "./tagger" // output path for the binary

		/* currently unnecessary, but want to remember to come back to it for a different idea
		// check to see if the build is necessary
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			fmt.Println("Building the Go executable...")
			err = utils.BuildExecutable(projectRoot, outputPath)
			if err != nil {
				fmt.Println("Failed to build the executable:", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Executable already exists at:", outputPath)
		}

		// build the Go executable
		err := utils.BuildExecutable(projectRoot, outputPath)
		if err != nil {
			fmt.Println("Failed to build the executable:", err)
			os.Exit(1)
		}

		*/

		// install the Git hook
		err := scripts.InstallGitHook(outputPath)
		if err != nil {
			fmt.Println("Failed to install Git hook:", err)
			os.Exit(1)
		}
		fmt.Println("Git post-commit hook installed successfully.")
		return
	}

	// If the uninstall flag is provided, clean the Git hook
	if *uninstallFlag {
		err := scripts.CleanGitHook()
		if err != nil {
			fmt.Println("Failed to uninstall Git hook:", err)
			os.Exit(1)
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
