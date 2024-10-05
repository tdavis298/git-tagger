package utils

import (
	"fmt"
	"os"
	"strings"
)

// ---------- Argument Handling Functions ----------

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
func WrapErrorf(format string, err error) error {
	if err != nil {
		// Correctly wrap the error with additional context
		return fmt.Errorf(format+": %w", err)
	}
	return nil
}

// LogAndExit logs the error message along with context and exits the program.
func LogAndExit(context string, err error) {
	// Print a formatted error message including the context and error details
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Error - %s: %v\n", context, err)
		if err != nil {
			return
		}
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Error - %s\n", context)
	}
	os.Exit(1)
}
