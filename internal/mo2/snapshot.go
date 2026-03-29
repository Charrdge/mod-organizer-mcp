package mo2

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ModRow is one mod in the snapshot JSON.
type ModRow struct {
	Name     string            `json:"name"`
	Enabled  bool              `json:"enabled"`
	Order    int               `json:"order"`
	Meta     map[string]string `json:"meta,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
}

// Snapshot is the mo2_profile_snapshot JSON payload.
type Snapshot struct {
	ProfileDir              string               `json:"profile_dir"`
	ModsDir                 string               `json:"mods_dir"`
	GeneratedAt             string               `json:"generated_at"`
	Mods                    []ModRow             `json:"mods"`
	PluginLines             []string             `json:"plugin_lines,omitempty"`
	LoadorderLines          []string             `json:"loadorder_lines,omitempty"`
	SnapshotContractVersion string               `json:"snapshot_contract_version,omitempty"`
	ProfileIni              []ProfileIniEntry    `json:"profile_ini,omitempty"`
	ProfileListPaths        *ProfileListPaths    `json:"profile_list_paths,omitempty"`
	ArchiveSearchRoots      []ArchiveSearchRoot  `json:"archive_search_roots,omitempty"`
	PluginsOrdered          []PluginOrderedEntry `json:"plugins_ordered,omitempty"`
	Warnings                []string             `json:"warnings"`
}

// BuildSnapshot reads modlist, optional plugins/loadorder, and meta.ini per mod folder (full snapshot).
func BuildSnapshot(cfg Config) (*Snapshot, error) {
	return BuildSnapshotWithOptions(cfg, DefaultSnapshotOptions())
}

// BuildSnapshotWithOptions reads modlist and optional plugins/loadorder/meta per flags.
func BuildSnapshotWithOptions(cfg Config, opts SnapshotOptions) (*Snapshot, error) {
	modlistPath := filepath.Join(cfg.ProfileDir, "modlist.txt")
	entries, err := ParseModlist(modlistPath)
	if err != nil {
		return nil, err
	}

	snap := &Snapshot{
		ProfileDir:  cfg.ProfileDir,
		ModsDir:     cfg.ModsDir,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Warnings:    nil,
	}

	if opts.IncludePluginLines {
		pluginsPath := filepath.Join(cfg.ProfileDir, "plugins.txt")
		if pl, err := ReadTextLines(pluginsPath); err == nil {
			snap.PluginLines = pl
		}
	}
	if opts.IncludeLoadorderLines {
		loadPath := filepath.Join(cfg.ProfileDir, "loadorder.txt")
		if ll, err := ReadTextLines(loadPath); err == nil {
			snap.LoadorderLines = ll
		}
	}

	prefix := opts.ModNamePrefix
	seenCount := make(map[string]int)
	for _, e := range entries {
		if opts.OnlyEnabled && !e.Enabled {
			continue
		}
		if prefix != "" {
			if len(e.Name) < len(prefix) || e.Name[:len(prefix)] != prefix {
				continue
			}
		}
		seenCount[e.Name]++
		row := ModRow{Name: e.Name, Enabled: e.Enabled, Order: e.Order}
		if seenCount[e.Name] > 1 {
			row.Warnings = append(row.Warnings, "duplicate name earlier in modlist.txt")
		}
		modPath := filepath.Join(cfg.ModsDir, e.Name)
		st, statErr := os.Stat(modPath)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				row.Warnings = append(row.Warnings, "mod folder missing under MO2_MODS_DIR")
			} else {
				row.Warnings = append(row.Warnings, fmt.Sprintf("mod folder: %v", statErr))
			}
			snap.Mods = append(snap.Mods, row)
			continue
		}
		if !st.IsDir() {
			row.Warnings = append(row.Warnings, "mods path entry is not a directory")
			snap.Mods = append(snap.Mods, row)
			continue
		}

		if opts.IncludeMeta {
			metaPath := filepath.Join(modPath, "meta.ini")
			meta, err := ParseMetaINI(metaPath)
			if err != nil {
				if os.IsNotExist(err) {
					row.Warnings = append(row.Warnings, "meta.ini missing")
				} else {
					row.Warnings = append(row.Warnings, fmt.Sprintf("meta.ini: %v", err))
				}
			} else if len(meta) > 0 {
				row.Meta = meta
			}
		}
		snap.Mods = append(snap.Mods, row)
	}

	if err := applySnapshotContract(snap, cfg, opts); err != nil {
		return nil, err
	}

	return snap, nil
}
