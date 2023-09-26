package logs

import (
	"fmt"
	"os"
)

var LOG_DIR = fmt.Sprintf("%v/.pkgctl/logs", os.Getenv("HOME"))
