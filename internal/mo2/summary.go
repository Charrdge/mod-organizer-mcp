package mo2

import (
	"os"
	"path/filepath"
)

// ProfileSummary is a lightweight view of the active profile (no per-mod meta.ini).
type ProfileSummary struct {
	ProfileDir            string `json:"profile_dir"`
	ModsDir               string `json:"mods_dir"`
	EnabledMods           int    `json:"enabled_mods"`
	DisabledMods          int    `json:"disabled_mods"`
	TotalModlistEntries   int    `json:"total_modlist_entries"`
	DuplicateModNames     int    `json:"duplicate_mod_names"`
	HasPluginsTxt         bool   `json:"has_plugins_txt"`
	PluginsLineCount      int    `json:"plugins_line_count,omitempty"`
	HasLoadorderTxt       bool   `json:"has_loadorder_txt"`
	LoadorderLineCount    int    `json:"loadorder_line_count,omitempty"`
	ModsMissingFolder     int    `json:"mods_missing_folder"`
	ModsNotDirectory      int    `json:"mods_not_directory"`
}

// BuildProfileSummary counts modlist state and optional list files; optionally checks mod folders exist under modsDir.
func BuildProfileSummary(cfg Config) (*ProfileSummary, error) {
	modlistPath := filepath.Join(cfg.ProfileDir, "modlist.txt")
	entries, err := ParseModlist(modlistPath)
	if err != nil {
		return nil, err
	}
	s := &ProfileSummary{
		ProfileDir:          cfg.ProfileDir,
		ModsDir:             cfg.ModsDir,
		TotalModlistEntries: len(entries),
	}
	seen := make(map[string]int)
	for _, e := range entries {
		if e.Enabled {
			s.EnabledMods++
		} else {
			s.DisabledMods++
		}
		seen[e.Name]++
	}
	for _, n := range seen {
		if n > 1 {
			s.DuplicateModNames += n - 1
		}
	}
	pluginsPath := filepath.Join(cfg.ProfileDir, "plugins.txt")
	if pl, err := ReadTextLines(pluginsPath); err == nil {
		s.HasPluginsTxt = true
		s.PluginsLineCount = len(pl)
	}
	loadPath := filepath.Join(cfg.ProfileDir, "loadorder.txt")
	if ll, err := ReadTextLines(loadPath); err == nil {
		s.HasLoadorderTxt = true
		s.LoadorderLineCount = len(ll)
	}
	for _, e := range entries {
		modPath := filepath.Join(cfg.ModsDir, e.Name)
		st, statErr := os.Stat(modPath)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				s.ModsMissingFolder++
			}
			continue
		}
		if !st.IsDir() {
			s.ModsNotDirectory++
		}
	}
	return s, nil
}
