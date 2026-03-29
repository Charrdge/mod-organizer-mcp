package mo2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds read-only MO2 paths from the environment.
type Config struct {
	ProfileDir string
	ModsDir    string
}

// ConfigFromEnv loads MO2_PROFILE_DIR and MO2_MODS_DIR (required, absolute paths recommended).
func ConfigFromEnv() (Config, error) {
	prof := strings.TrimSpace(os.Getenv("MO2_PROFILE_DIR"))
	mods := strings.TrimSpace(os.Getenv("MO2_MODS_DIR"))
	if prof == "" {
		return Config{}, fmt.Errorf("MO2_PROFILE_DIR is not set")
	}
	if mods == "" {
		return Config{}, fmt.Errorf("MO2_MODS_DIR is not set")
	}
	prof = filepath.Clean(prof)
	mods = filepath.Clean(mods)
	if err := requireDir("MO2_PROFILE_DIR", prof); err != nil {
		return Config{}, err
	}
	if err := requireDir("MO2_MODS_DIR", mods); err != nil {
		return Config{}, err
	}
	return Config{ProfileDir: prof, ModsDir: mods}, nil
}

func requireDir(envName, path string) error {
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s: directory does not exist: %s", envName, path)
		}
		return fmt.Errorf("%s: %w", envName, err)
	}
	if !st.IsDir() {
		return fmt.Errorf("%s: not a directory: %s", envName, path)
	}
	return nil
}
