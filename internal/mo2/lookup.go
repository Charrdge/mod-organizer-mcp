package mo2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ModLookupResult is returned by LookupMod for MCP mo2_mod_lookup.
type ModLookupResult struct {
	Match      *ModLookupMatch `json:"match,omitempty"`
	Ambiguous  []string        `json:"ambiguous_candidates,omitempty"`
	NotFound   bool            `json:"not_found"`
}

// ModLookupMatch is one resolved mod from modlist + meta.ini.
type ModLookupMatch struct {
	Name     string            `json:"name"`
	Enabled  bool              `json:"enabled"`
	Order    int               `json:"order"`
	Meta     map[string]string `json:"meta,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
}

// LookupMod resolves a mod folder name from modlist: exact (case-sensitive), then case-insensitive, then unique prefix.
func LookupMod(cfg Config, query string) (*ModLookupResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("name is empty")
	}
	modlistPath := filepath.Join(cfg.ProfileDir, "modlist.txt")
	entries, err := ParseModlist(modlistPath)
	if err != nil {
		return nil, err
	}
	// exact case-sensitive
	for _, e := range entries {
		if e.Name == query {
			return buildMatch(cfg, e)
		}
	}
	// case-insensitive exact
	ql := strings.ToLower(query)
	var ci []ModlistEntry
	for _, e := range entries {
		if strings.ToLower(e.Name) == ql {
			ci = append(ci, e)
		}
	}
	if len(ci) == 1 {
		return buildMatch(cfg, ci[0])
	}
	if len(ci) > 1 {
		names := make([]string, len(ci))
		for i, e := range ci {
			names[i] = e.Name
		}
		return &ModLookupResult{Ambiguous: names}, nil
	}
	// unique prefix (case-insensitive)
	var pref []ModlistEntry
	for _, e := range entries {
		if strings.HasPrefix(strings.ToLower(e.Name), ql) {
			pref = append(pref, e)
		}
	}
	if len(pref) == 1 {
		return buildMatch(cfg, pref[0])
	}
	if len(pref) > 1 {
		names := make([]string, len(pref))
		for i, e := range pref {
			names[i] = e.Name
		}
		return &ModLookupResult{Ambiguous: names}, nil
	}
	return &ModLookupResult{NotFound: true}, nil
}

func buildMatch(cfg Config, e ModlistEntry) (*ModLookupResult, error) {
	m := &ModLookupMatch{Name: e.Name, Enabled: e.Enabled, Order: e.Order}
	modPath := filepath.Join(cfg.ModsDir, e.Name)
	st, statErr := os.Stat(modPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			m.Warnings = append(m.Warnings, "mod folder missing under MO2_MODS_DIR")
		} else {
			m.Warnings = append(m.Warnings, fmt.Sprintf("mod folder: %v", statErr))
		}
		return &ModLookupResult{Match: m}, nil
	}
	if !st.IsDir() {
		m.Warnings = append(m.Warnings, "mods path entry is not a directory")
		return &ModLookupResult{Match: m}, nil
	}
	metaPath := filepath.Join(modPath, "meta.ini")
	meta, err := ParseMetaINI(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.Warnings = append(m.Warnings, "meta.ini missing")
		} else {
			m.Warnings = append(m.Warnings, fmt.Sprintf("meta.ini: %v", err))
		}
	} else if len(meta) > 0 {
		m.Meta = meta
	}
	return &ModLookupResult{Match: m}, nil
}
