package logs

import (
	"io/fs"
	"os"
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
		if logFile, ok := Match(file); ok {
			logsFiles = append(logsFiles, logFile)
		}
	}

	return logsFiles, nil
}

func Match(file fs.DirEntry) (LogFile, bool) {
	re := regexp.MustCompile(`(\w+)\.(\w+)\.(.*?)\.log\.gz`)

	matches := re.FindStringSubmatch(file.Name())

	if len(matches) != 4 {
		return LogFile{}, false
	}

	ts, err := time.Parse(time.RFC3339, matches[3])

	if err != nil {
		return LogFile{}, false
	}

	fileInfo, err := file.Info()

	if err != nil {
		return LogFile{}, false
	}
	fileInfo.Size()

	return LogFile{
		Path:      LOG_DIR + string(os.PathSeparator) + file.Name(),
		Size:      fileInfo.Size(),
		PkgctlCmd: matches[1],
		ToolID:    matches[2],
		Time:      ts,
	}, true
}
