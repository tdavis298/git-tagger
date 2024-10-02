package utils

import (
	"fmt"
	"os"
	"strings"
)

// HandleUnparsedArgs handles the unparsed arguments
func HandleUnparsedArgs(args []string) {
	if len(args) > 0 {
		// Write to stderr and check for errors while doing so
		if _, err := fmt.Fprintf(os.Stderr, "Incorrect flag(s) provided: %s\n", strings.Join(args, ", ")); err != nil {
			_, err := fmt.Fprintf(os.Stderr, "Error occurred while writing to stderr: %v\n", err)
			if err != nil {
				return
			}
			os.Exit(2)
		}

		// Exit with a non-zero status indicating incorrect usage
		os.Exit(1)
	}
}
