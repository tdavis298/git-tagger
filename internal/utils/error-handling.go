package utils

import (
	"fmt"
	"os"
	"strings"
)

// argument handling functions:

// HandleUnparsedArgs handles any unparsed command-line arguments by printing an error message and exiting.
// parameters:
// - args: a slice of unparsed command-line arguments
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

// WrapErrorf wraps an error with the given format and arguments if an error occurred.
// Parameters:
// - err: the original error
// - format: the format string for wrapping the error
// - args: additional arguments for formatting the error message
// Returns:
// - an error wrapped with the specified message if it occurred, otherwise nil
func WrapErrorf(format string, err error, args ...interface{}) error {
	if err != nil {
		// Ensure error wrapping is done correctly with fmt.Errorf
		return fmt.Errorf(format+": %w", append(args, err)...)
	}
	return nil
}
