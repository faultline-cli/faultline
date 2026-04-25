package cmd

import "io"

// isEOF returns true for both io.EOF and io.ErrUnexpectedEOF so that callers
// can treat both as clean end-of-stream without importing the io package.
func isEOF(err error) bool {
	return err == io.EOF || err == io.ErrUnexpectedEOF
}
