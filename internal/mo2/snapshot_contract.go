package mo2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const snapshotContractVersion = "1"

// ProfileListFileRef is an absolute path to a profile list file and whether it exists.
type ProfileListFileRef struct {
	Path    string `json:"path"`
	Present bool   `json:"present"`
}

// ProfileListPaths points to plugins.txt and loadorder.txt under the profile directory.
type ProfileListPaths struct {
	PluginsTxt   ProfileListFileRef `json:"plugins_txt"`
	LoadorderTxt ProfileListFileRef `json:"loadorder_txt"`
}

// ArchiveSearchRoot is a filesystem root for scanning archives (.bsa/.ba2) under a mod.
type ArchiveSearchRoot struct {
	Path     string `json:"path"`
	ModName  string `json:"mod_name"`
	ModOrder int    `json:"mod_order"`
	Enabled  bool   `json:"enabled"`
	Kind     string `json:"kind"` // mod_root | data_subdir
}

func buildProfileListPaths(absProfile string) ProfileListPaths {
	out := ProfileListPaths{
		PluginsTxt:   fileRef(filepath.Join(absProfile, "plugins.txt")),
		LoadorderTxt: fileRef(filepath.Join(absProfile, "loadorder.txt")),
	}
	return out
}

func fileRef(path string) ProfileListFileRef {
	st, err := os.Stat(path)
	return ProfileListFileRef{
		Path:    path,
		Present: err == nil && !st.IsDir(),
	}
}

func buildArchiveSearchRoots(absModsDir string, mods []ModRow) []ArchiveSearchRoot {
	var roots []ArchiveSearchRoot
	for _, m := range mods {
		if invalidModFolderName(m.Name) {
			continue
		}
		modRoot := filepath.Join(absModsDir, m.Name)
		absMod, err := filepath.Abs(modRoot)
		if err != nil {
			continue
		}
		roots = append(roots, ArchiveSearchRoot{
			Path:     absMod,
			ModName:  m.Name,
			ModOrder: m.Order,
			Enabled:  m.Enabled,
			Kind:     "mod_root",
		})
		dataPath := filepath.Join(absMod, "Data")
		if st, err := os.Stat(dataPath); err == nil && st.IsDir() {
			roots = append(roots, ArchiveSearchRoot{
				Path:     dataPath,
				ModName:  m.Name,
				ModOrder: m.Order,
				Enabled:  m.Enabled,
				Kind:     "data_subdir",
			})
		}
	}
	return roots
}

func invalidModFolderName(name string) bool {
	return name == "" || name == "." || strings.Contains(name, "..") || filepath.IsAbs(name)
}

// MachineContractPayload is the dedicated tool response for mo2_profile_machine_contract (no mods[] / meta).
type MachineContractPayload struct {
	ProfileDir              string              `json:"profile_dir"`
	ModsDir                 string              `json:"mods_dir"`
	GeneratedAt             string              `json:"generated_at"`
	SnapshotContractVersion string              `json:"snapshot_contract_version"`
	ProfileIni              []ProfileIniEntry   `json:"profile_ini"`
	ProfileListPaths        ProfileListPaths    `json:"profile_list_paths"`
	ArchiveSearchRoots      []ArchiveSearchRoot `json:"archive_search_roots"`
	Warnings                []string            `json:"warnings,omitempty"`
}

// BuildMachineContractPayload returns only the machine-readable contract (profile INI whitelist, list paths, archive roots). Does not read meta.ini or embed plugin line arrays.
func BuildMachineContractPayload(cfg Config, onlyEnabled bool, modNamePrefix string) (*MachineContractPayload, error) {
	opts := DefaultSnapshotOptions()
	opts.IncludeMeta = false
	opts.IncludePluginLines = false
	opts.IncludeLoadorderLines = false
	opts.IncludeContract = true
	opts.IncludePluginLoadOrder = false
	opts.OnlyEnabled = onlyEnabled
	opts.ModNamePrefix = modNamePrefix
	snap, err := BuildSnapshotWithOptions(cfg, opts)
	if err != nil {
		return nil, err
	}
	if snap.ProfileListPaths == nil {
		return nil, fmt.Errorf("internal: contract missing profile_list_paths")
	}
	return &MachineContractPayload{
		ProfileDir:              snap.ProfileDir,
		ModsDir:                 snap.ModsDir,
		GeneratedAt:             snap.GeneratedAt,
		SnapshotContractVersion: snap.SnapshotContractVersion,
		ProfileIni:              snap.ProfileIni,
		ProfileListPaths:        *snap.ProfileListPaths,
		ArchiveSearchRoots:      snap.ArchiveSearchRoots,
		Warnings:                snap.Warnings,
	}, nil
}

// PluginLoadOrderPayload is the dedicated tool response for mo2_profile_plugin_load_order.
type PluginLoadOrderPayload struct {
	ProfileDir       string               `json:"profile_dir"`
	ProfileListPaths ProfileListPaths     `json:"profile_list_paths"`
	PluginsOrdered   []PluginOrderedEntry `json:"plugins_ordered"`
	Warnings         []string             `json:"warnings,omitempty"`
}

// BuildPluginLoadOrderPayload merges loadorder.txt with plugins.txt active flags; includes profile_list_paths for convenience.
func BuildPluginLoadOrderPayload(cfg Config) (*PluginLoadOrderPayload, error) {
	profAbs, err := filepath.Abs(cfg.ProfileDir)
	if err != nil {
		return nil, fmt.Errorf("profile dir abs: %w", err)
	}
	pl := buildProfileListPaths(profAbs)
	ordered, w := BuildPluginsOrdered(profAbs)
	return &PluginLoadOrderPayload{
		ProfileDir:       cfg.ProfileDir,
		ProfileListPaths: pl,
		PluginsOrdered:   ordered,
		Warnings:         w,
	}, nil
}

func applySnapshotContract(snap *Snapshot, cfg Config, opts SnapshotOptions) error {
	if !opts.IncludeContract && !opts.IncludePluginLoadOrder {
		return nil
	}
	profAbs, err := filepath.Abs(cfg.ProfileDir)
	if err != nil {
		return fmt.Errorf("profile dir abs: %w", err)
	}
	modsAbs, err := filepath.Abs(cfg.ModsDir)
	if err != nil {
		return fmt.Errorf("mods dir abs: %w", err)
	}

	if opts.IncludeContract {
		snap.SnapshotContractVersion = snapshotContractVersion
		snap.ProfileIni = DiscoverProfileINIs(profAbs)
		pl := buildProfileListPaths(profAbs)
		snap.ProfileListPaths = &pl
		snap.ArchiveSearchRoots = buildArchiveSearchRoots(modsAbs, snap.Mods)
	}
	if opts.IncludePluginLoadOrder {
		ordered, w := BuildPluginsOrdered(profAbs)
		snap.PluginsOrdered = ordered
		snap.Warnings = append(snap.Warnings, w...)
	}
	return nil
}
