package mo2

import (
	"os"
	"path/filepath"
	"strings"
)

// Known profile-level INI basenames MO2 copies per game (lowercase keys for lookup).
var profileIniWhitelist = []string{
	"skyrim.ini",
	"skyrimprefs.ini",
	"skyrimcustom.ini",
	"fallout4.ini",
	"fallout4prefs.ini",
	"fallout4custom.ini",
	"falloutnv.ini",
	"fallout.ini",
	"falloutprefs.ini",
	"oblivion.ini",
	"oblivionprefs.ini",
	"morrowind.ini",
	"starfield.ini",
	"starfieldprefs.ini",
	"starfieldcustom.ini",
}

// ProfileIniEntry is one profile-directory INI (not mod meta.ini).
type ProfileIniEntry struct {
	Basename string `json:"basename"`
	Path     string `json:"path"`
	Present  bool   `json:"present"`
}

// DiscoverProfileINIs lists whitelist names with absolute paths and whether each file exists (case-insensitive match on disk).
func DiscoverProfileINIs(absProfileDir string) []ProfileIniEntry {
	byLower := profileDirFilesByLowerName(absProfileDir)
	var out []ProfileIniEntry
	for _, want := range profileIniWhitelist {
		low := strings.ToLower(want)
		actual, ok := byLower[low]
		path := filepath.Join(absProfileDir, want)
		present := false
		if ok {
			path = filepath.Join(absProfileDir, actual)
			present = true
		}
		out = append(out, ProfileIniEntry{
			Basename: actualNameForEntry(want, actual, ok),
			Path:     path,
			Present:  present,
		})
	}
	return out
}

func actualNameForEntry(want, actual string, present bool) string {
	if present {
		return actual
	}
	return want
}

func profileDirFilesByLowerName(absProfileDir string) map[string]string {
	m := make(map[string]string)
	ents, err := os.ReadDir(absProfileDir)
	if err != nil {
		return m
	}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.EqualFold(filepath.Ext(name), ".ini") {
			continue
		}
		low := strings.ToLower(name)
		if _, ok := m[low]; !ok {
			m[low] = name
		}
	}
	return m
}
