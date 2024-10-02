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
	versionTagFlag := flag.Bool("version-tag", false, "Tag untagged Git commits with version numbers")
	installFlag := flag.Bool("install", false, "Install the Git post-commit hook")
	uninstallFlag := flag.Bool("clean", false, "Remove the Git post-commit hook")

	flag.Parse()

	// handle unparsed flags
	utils.HandleUnparsedArgs(flag.Args())

	// Handle the version tagging logic
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

			selectedBranch, err := git.SelectBranch(branches)
			if err != nil {
				fmt.Println("Error selecting branch:", err)
				os.Exit(1)
			}

			branch = selectedBranch
		}

		// Update untagged commits for the selected branch
		err := version.UpdateUntaggedCommits(branch)
		if err != nil {
			fmt.Println("Failed to update untagged commits:", err)
			os.Exit(1)
		}

		fmt.Println("Version-tagged untagged files successfully.")
		return
	}

	// if the install flag is provided, install hook
	if *installFlag {
		err := scripts.InstallGitHook()
		if err != nil {
			fmt.Println("Failed to install Git hook:", err)
			os.Exit(1)
		}
		fmt.Println("Git post-commit hook installed successfully.")
		return
	}

	// if the uninstall flag is provided, clean the Git hook
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
