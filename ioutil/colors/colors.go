package colors

const (
	CLEAR  = "\033[H\033[2J"
	BLUE   = "\033[1;34m"
	RED    = "\033[1;31m"
	GREEN  = "\033[1;32m"
	YELLOW = "\033[1;33m"
	GRAY   = "\033[1;37m"
	END    = "\033[0m"

	// DELETE_PREVIOUS_LINE = "\033[1A\033[2K"
	DELETE_PREVIOUS_LINE = "\x1b[1A\x1b[2K"
)
