package logs

import (
	"compress/gzip"
	"fmt"
	"os"
	"time"
)

type LogFile struct {
	// dirEntry  fs.DirEntry
	Path      string
	Size      int64
	Time      time.Time
	ToolID    string
	PkgctlCmd string
}

func (l LogFile) Open() (*os.File, error) {
	return os.Open(l.Path)
}

type LogWriter struct {
	file   *os.File
	writer *gzip.Writer
	time   time.Time
}

const LOG_FILE_FORMAT = "%v/%v.%v.%v.log.gz"

func newLogPath(pkgctlCmd, tool string, t time.Time) string {
	return fmt.Sprintf(LOG_FILE_FORMAT, LOG_DIR, pkgctlCmd, tool, t.Format(time.RFC3339))
}

func NewLogWriter(command, tool string) (*LogWriter, error) {
	time := time.Now()
	file, err := os.Create(newLogPath(command, tool, time))

	if err != nil {
		return nil, err
	}

	writer := gzip.NewWriter(file)

	return &LogWriter{
		file:   file,
		writer: writer,
		time:   time,
	}, nil
}

func (w *LogWriter) Close() {
	if w.writer != nil {
		w.writer.Close()
	}
	if w.file != nil {
		w.file.Close()
	}
}
