package logs

import (
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func GetAll() ([]LogFile, error) {

	files, err := os.ReadDir(LOG_DIR)

	if err != nil {
		return nil, err
	}

	var logsFiles []LogFile

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if logFile, ok := GetLogFile(filepath.Join(LOG_DIR, file.Name())); ok {
			logsFiles = append(logsFiles, logFile)
		}
	}

	return logsFiles, nil
}

func GetLogFile(path string) (LogFile, bool) {

	fileInfo, err := os.Stat(path)

	if err != nil || fileInfo.IsDir() {
		return LogFile{}, false
	}

	fileNameRe := regexp.MustCompile(`^(\w+)\.(\w+)\.(.*?)\.log\.gz$`)

	matches := fileNameRe.FindStringSubmatch(fileInfo.Name())

	if len(matches) != 4 {
		return LogFile{}, false
	}

	ts, err := time.Parse(time.RFC3339, matches[3])

	if err != nil {
		return LogFile{}, false
	}

	return LogFile{
		Path:      path,
		Size:      fileInfo.Size(),
		PkgctlCmd: matches[1],
		ToolID:    matches[2],
		Time:      ts,
	}, true
}
