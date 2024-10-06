//go:build windows

package testutils

/*func isSharingViolation(err error) bool {
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return errors.Is(pathErr.Err, windows.ERROR_SHARING_VIOLATION)
	}
	return false
}
*/
