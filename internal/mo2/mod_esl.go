package mo2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var pluginSuffixes = []string{".esp", ".esm", ".esl"}

// ListModPluginArchives finds .esp/.esm/.esl under modsDir/modName up to maxDepth and maxFiles.
func ListModPluginArchives(modsDir, modName string, maxDepth, maxFiles int) ([]string, []string, error) {
	if strings.Contains(modName, "..") || filepath.IsAbs(modName) || modName == "" || modName == "." {
		return nil, nil, fmt.Errorf("invalid mod name")
	}
	if maxDepth <= 0 {
		maxDepth = 8
	}
	if maxFiles <= 0 {
		maxFiles = 200
	}
	root := filepath.Join(modsDir, modName)
	st, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("mod folder does not exist: %s", modName)
		}
		return nil, nil, err
	}
	if !st.IsDir() {
		return nil, nil, fmt.Errorf("not a directory: %s", modName)
	}
	modsDirClean := filepath.Clean(modsDir)
	var out []string
	var warnings []string
	var walk func(dir string, depth int) error
	walk = func(dir string, depth int) error {
		if depth > maxDepth {
			warnings = append(warnings, fmt.Sprintf("truncated at max_depth=%d", maxDepth))
			return nil
		}
		if len(out) >= maxFiles {
			return nil
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if len(out) >= maxFiles {
				warnings = append(warnings, fmt.Sprintf("truncated at max_files=%d", maxFiles))
				return nil
			}
			p := filepath.Join(dir, e.Name())
			rel, err := filepath.Rel(modsDirClean, p)
			if err != nil || strings.HasPrefix(rel, "..") {
				warnings = append(warnings, "skipped path outside mods dir")
				continue
			}
			if e.IsDir() {
				if err := walk(p, depth+1); err != nil {
					return err
				}
				continue
			}
			lower := strings.ToLower(e.Name())
			for _, suf := range pluginSuffixes {
				if strings.HasSuffix(lower, suf) {
					out = append(out, filepath.ToSlash(rel))
					break
				}
			}
		}
		return nil
	}
	if err := walk(root, 0); err != nil {
		return nil, warnings, err
	}
	return out, warnings, nil
}
