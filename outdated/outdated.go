package outdated

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/pkgctl/pkgctl/ioutil/colors"
	"github.com/pkgctl/pkgctl/tools"
)

// Report is what the last check found.
type Report struct {
	// CheckedAt is when the check ran, or zero if no check has run.
	CheckedAt time.Time
	// Packages maps each tool's name to its installed packages that have a newer version available.
	Packages map[string][]tools.Package
}

// Summary counts the report's outdated packages in one line, or returns "" if there are none.
func (r Report) Summary() string {
	total := 0
	var counts []string
	for _, tool := range tools.List() {
		packages := r.Packages[tool.Name()]
		if len(packages) == 0 {
			continue
		}
		total += len(packages)
		counts = append(counts, fmt.Sprintf("%v %v", tool.Name(), len(packages)))
	}

	if total == 0 {
		return ""
	}

	noun := "packages"
	if total == 1 {
		noun = "package"
	}
	return fmt.Sprintf("%v (%v). Run `pkgctl update`.", colors.Stdout.Yellow(fmt.Sprintf("%v outdated %v", total, noun)), strings.Join(counts, ", "))
}

// Describe lists tool's outdated packages, one per line, then the command that updates them.
func Describe(tool tools.Tool, packages []tools.Package) string {
	var description strings.Builder
	for _, outdatedPackage := range packages {
		fmt.Fprintf(&description, "  %v", colors.Stdout.Yellow(outdatedPackage.Name))
		if outdatedPackage.Latest != "" {
			fmt.Fprintf(&description, " %v → %v", outdatedPackage.Current, outdatedPackage.Latest)
		}
		description.WriteString("\n")
	}
	fmt.Fprintf(&description, "%v %v outdated. Run `%v`.\n", colors.Stdout.Yellow("↑"), len(packages), tool.UpdateCommand())
	return description.String()
}

// Load returns the report saved by the last check, or an empty report if no check has run.
func Load() (Report, error) {
	path, err := reportPath()
	if err != nil {
		return Report{}, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Report{}, nil
	}
	if err != nil {
		return Report{}, err
	}

	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return Report{}, err
	}
	return report, nil
}

// Save replaces the saved report with report.
func Save(report Report) error {
	path, err := reportPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	// Write a temporary file and rename it over the report, so a reader never sees a half-written report.
	file, err := os.CreateTemp(filepath.Dir(path), "outdated-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

// ErrCheckRunning means another `pkgctl check` holds the check lock.
var ErrCheckRunning = errors.New("another check is already running")

// LockCheck takes the lock that keeps two checks from running at once, or returns ErrCheckRunning.
// The lock is released when the returned file is closed or the process exits.
func LockCheck() (*os.File, error) {
	dir, err := pkgctlCacheDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(dir, "check.lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, ErrCheckRunning
		}
		return nil, err
	}
	return file, nil
}

func reportPath() (string, error) {
	dir, err := pkgctlCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "outdated.json"), nil
}

func pkgctlCacheDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "pkgctl"), nil
}
