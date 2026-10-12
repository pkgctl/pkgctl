package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Config holds the settings in ~/.config/pkgctl/config.json.
type Config struct {
	// CheckInterval is how often pkgctl checks for outdated packages in the background.
	CheckInterval time.Duration
}

const defaultCheckInterval = 24 * time.Hour

// Load reads the settings file, using the default for each setting it leaves out, or for all of them if there is no file.
func Load() (Config, error) {
	config := Config{CheckInterval: defaultCheckInterval}

	path, err := configPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return config, nil
	}
	if err != nil {
		return Config{}, err
	}

	var file struct {
		CheckInterval string `json:"check_interval"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		return Config{}, fmt.Errorf("%v: %w", path, err)
	}

	if file.CheckInterval != "" {
		config.CheckInterval, err = time.ParseDuration(file.CheckInterval)
		if err != nil {
			return Config{}, fmt.Errorf("%v: check_interval: %w", path, err)
		}
	}
	return config, nil
}

// CheckDue reports whether a check is due again, given when the last one ran.
func (c Config) CheckDue(lastCheck time.Time) bool {
	return time.Since(lastCheck) >= c.CheckInterval
}

// configPath is ~/.config/pkgctl/config.json even when XDG_CONFIG_HOME is set,
// because the brew service runs without the shell's environment and must read the same file.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "pkgctl", "config.json"), nil
}
