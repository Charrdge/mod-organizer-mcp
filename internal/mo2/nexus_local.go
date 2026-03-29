package mo2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NexusLocalFields are Nexus-related values MO2 typically stores in meta.ini (not live API).
type NexusLocalFields struct {
	ModID            string `json:"nexus_mod_id,omitempty"`
	FileID           string `json:"nexus_file_id,omitempty"`
	Version          string `json:"version,omitempty"`
	InstallationFile string `json:"installation_file,omitempty"`
	GameName         string `json:"game_name,omitempty"`
}

// NexusLocalIndexRow is one mod row for mo2_nexus_local_index.
type NexusLocalIndexRow struct {
	FolderName string           `json:"folder_name"`
	Enabled    bool             `json:"enabled"`
	Order      int              `json:"order"`
	Nexus      NexusLocalFields `json:"nexus"`
	Warnings   []string         `json:"warnings,omitempty"`
}

// NexusLocalIndex is the full tool payload (disk snapshot, not Nexus API).
type NexusLocalIndex struct {
	Source       string               `json:"source"`
	LiveNexusAPI bool                 `json:"live_nexus_api"`
	ProfileDir   string               `json:"profile_dir"`
	ModsDir      string               `json:"mods_dir"`
	GeneratedAt  string               `json:"generated_at"`
	Mods         []NexusLocalIndexRow `json:"mods"`
}

func metaPick(meta map[string]string, keys ...string) string {
	for _, k := range keys {
		if v, ok := meta[k]; ok {
			v = strings.TrimSpace(v)
			if v != "" {
				return v
			}
		}
	}
	return ""
}

// ExtractNexusLocal pulls known Nexus-related keys from a flat meta.ini map (section.key).
func ExtractNexusLocal(meta map[string]string) (NexusLocalFields, []string) {
	var w []string
	if len(meta) == 0 {
		return NexusLocalFields{}, []string{"empty or unreadable meta.ini"}
	}
	n := NexusLocalFields{
		ModID:            metaPick(meta, "General.modid", "General.modID", "General.modId"),
		FileID:           metaPick(meta, "General.fileid", "General.fileID", "General.fileId", "General.nexusFileId"),
		Version:          metaPick(meta, "General.version", "General.Version"),
		InstallationFile: metaPick(meta, "General.installationFile", "General.installationfile"),
		GameName:         metaPick(meta, "General.gameName", "General.gamename", "General.game"),
	}
	if n.ModID == "" && n.FileID == "" {
		w = append(w, "no nexus mod id or file id in meta.ini")
	}
	return n, w
}

// BuildNexusLocalIndex walks modlist and reads only meta.ini per mod folder.
func BuildNexusLocalIndex(cfg Config, onlyEnabled bool) (*NexusLocalIndex, error) {
	modlistPath := filepath.Join(cfg.ProfileDir, "modlist.txt")
	entries, err := ParseModlist(modlistPath)
	if err != nil {
		return nil, err
	}
	idx := &NexusLocalIndex{
		Source:       "meta_ini",
		LiveNexusAPI: false,
		ProfileDir:   cfg.ProfileDir,
		ModsDir:      cfg.ModsDir,
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		Mods:         nil,
	}
	for _, e := range entries {
		if onlyEnabled && !e.Enabled {
			continue
		}
		row := NexusLocalIndexRow{FolderName: e.Name, Enabled: e.Enabled, Order: e.Order}
		modPath := filepath.Join(cfg.ModsDir, e.Name)
		st, statErr := os.Stat(modPath)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				row.Warnings = append(row.Warnings, "mod folder missing under MO2_MODS_DIR")
			} else {
				row.Warnings = append(row.Warnings, fmt.Sprintf("mod folder: %v", statErr))
			}
			idx.Mods = append(idx.Mods, row)
			continue
		}
		if !st.IsDir() {
			row.Warnings = append(row.Warnings, "mods path entry is not a directory")
			idx.Mods = append(idx.Mods, row)
			continue
		}
		metaPath := filepath.Join(modPath, "meta.ini")
		meta, err := ParseMetaINI(metaPath)
		if err != nil {
			if os.IsNotExist(err) {
				row.Warnings = append(row.Warnings, "meta.ini missing")
			} else {
				row.Warnings = append(row.Warnings, fmt.Sprintf("meta.ini: %v", err))
			}
			idx.Mods = append(idx.Mods, row)
			continue
		}
		nx, nw := ExtractNexusLocal(meta)
		row.Nexus = nx
		row.Warnings = append(row.Warnings, nw...)
		idx.Mods = append(idx.Mods, row)
	}
	return idx, nil
}
