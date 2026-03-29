package mo2

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ProfileListEntry is one MO2 profile folder under the profiles parent directory.
type ProfileListEntry struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	ModlistEntries int    `json:"modlist_entries"`
}

// ListSiblingProfiles scans the parent of profileDir for subdirectories that contain modlist.txt.
func ListSiblingProfiles(profileDir string) ([]ProfileListEntry, error) {
	profileDir = filepath.Clean(profileDir)
	parent := filepath.Dir(profileDir)
	st, err := os.Stat(parent)
	if err != nil {
		return nil, fmt.Errorf("profiles parent: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("profiles parent is not a directory: %s", parent)
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil, fmt.Errorf("read profiles parent: %w", err)
	}
	var out []ProfileListEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		p := filepath.Join(parent, name)
		modlist := filepath.Join(p, "modlist.txt")
		if _, err := os.Stat(modlist); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("stat modlist %s: %w", modlist, err)
		}
		ents, err := ParseModlist(modlist)
		n := 0
		if err == nil {
			n = len(ents)
		}
		out = append(out, ProfileListEntry{Name: name, Path: p, ModlistEntries: n})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
