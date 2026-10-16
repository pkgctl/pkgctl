// Package terminal tells a terminal that takes escape codes apart from other output,
// such as a pipe, a log file, or a dumb terminal.
package terminal

import "os"

// SupportsEscapes reports whether file is a terminal that takes escape codes, such as for color or moving the cursor.
func SupportsEscapes(file *os.File) bool {
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
