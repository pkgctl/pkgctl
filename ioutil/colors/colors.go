// Package colors styles text for a terminal, and leaves it plain for anything else, such as a pipe, a log file,
// or a dumb terminal, or when NO_COLOR is set (https://no-color.org).
package colors

import (
	"os"

	"github.com/pkgctl/pkgctl/ioutil/terminal"
)

// A Style styles text written to one output.
type Style struct {
	enabled bool
}

// Stdout and Stderr style text written to standard output and standard error.
var (
	Stdout = styleFor(os.Stdout)
	Stderr = styleFor(os.Stderr)
)

func styleFor(file *os.File) Style {
	if os.Getenv("NO_COLOR") != "" {
		return Style{}
	}
	return Style{enabled: terminal.SupportsEscapes(file)}
}

func (s Style) Bold(text string) string   { return s.wrap("1", text) }
func (s Style) Cyan(text string) string   { return s.wrap("36", text) }
func (s Style) Green(text string) string  { return s.wrap("32", text) }
func (s Style) Red(text string) string    { return s.wrap("31", text) }
func (s Style) Yellow(text string) string { return s.wrap("33", text) }

// Header styles a tool's name above its output.
func (s Style) Header(text string) string { return s.wrap("1;34", text) }

func (s Style) wrap(code, text string) string {
	if !s.enabled {
		return text
	}
	return "\033[" + code + "m" + text + "\033[0m"
}
