package mo2

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const sksePluginPriorityNote = "MO2 mod priority: a mod later in modlist.txt (higher order index) wins over earlier mods for the same virtual path under SKSE/Plugins. Optional game_data_dir is treated as lowest priority (below all mods). Loose files only, not BSA/BA2."

const gameDataProviderName = "GameData"

// SKSEPluginOptions controls BuildSKSEPluginInventory.
type SKSEPluginOptions struct {
	OnlyEnabled      bool
	ModNamePrefix    string
	GameDataDir      string
	IncludeSize      bool
	IncludePEVersion bool
	MaxDLLs          int
	StripDataPrefix  bool
}

// DefaultSKSEPluginOptions returns defaults aligned with the plan.
func DefaultSKSEPluginOptions() SKSEPluginOptions {
	return SKSEPluginOptions{
		OnlyEnabled:      true,
		IncludeSize:      true,
		IncludePEVersion: false,
		MaxDLLs:          500,
		StripDataPrefix:  true,
	}
}

// SKSEPluginProvider is one source for a virtual SKSE plugin path.
type SKSEPluginProvider struct {
	Name    string `json:"name"`
	Order   int    `json:"order"`
	AbsPath string `json:"abs_path"`
}

// SKSEPluginRow is one virtual DLL path and overlay resolution.
type SKSEPluginRow struct {
	VirtualPath    string               `json:"virtual_path"`
	Winner         SKSEPluginProvider   `json:"winner"`
	Providers      []SKSEPluginProvider `json:"providers"`
	Size           *int64               `json:"size,omitempty"`
	FileVersion    string               `json:"file_version,omitempty"`
	ProductVersion string               `json:"product_version,omitempty"`
}

// SKSEPluginReport is the JSON payload for mo2_skse_plugins.
type SKSEPluginReport struct {
	ProfileDir   string          `json:"profile_dir"`
	ModsDir      string          `json:"mods_dir"`
	GeneratedAt  string          `json:"generated_at"`
	PriorityNote string          `json:"priority_note"`
	Warnings     []string        `json:"warnings,omitempty"`
	Plugins      []SKSEPluginRow `json:"plugins"`
	Truncated    bool            `json:"truncated,omitempty"`
}

type skseContrib struct {
	name    string
	order   int
	absPath string
}

// BuildSKSEPluginInventory lists .dll under Data/SKSE/Plugins and SKSE/Plugins for enabled mods (by default),
// optionally game_data_dir/SKSE/Plugins, and resolves winners by modlist order.
func BuildSKSEPluginInventory(cfg Config, opts SKSEPluginOptions) (*SKSEPluginReport, error) {
	if opts.MaxDLLs <= 0 {
		opts.MaxDLLs = DefaultSKSEPluginOptions().MaxDLLs
	}

	modlistPath := filepath.Join(cfg.ProfileDir, "modlist.txt")
	entries, err := ParseModlist(modlistPath)
	if err != nil {
		return nil, err
	}

	modsDirAbs, err := filepath.Abs(filepath.Clean(cfg.ModsDir))
	if err != nil {
		return nil, fmt.Errorf("mods dir abs: %w", err)
	}

	var warnings []string
	byPath := make(map[string][]skseContrib)

	gameDir := strings.TrimSpace(opts.GameDataDir)
	if gameDir != "" {
		gameDir = filepath.Clean(gameDir)
		st, statErr := os.Stat(gameDir)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				warnings = append(warnings, fmt.Sprintf("game_data_dir does not exist: %s", gameDir))
			} else {
				warnings = append(warnings, fmt.Sprintf("game_data_dir: %v", statErr))
			}
		} else if !st.IsDir() {
			warnings = append(warnings, fmt.Sprintf("game_data_dir is not a directory: %s", gameDir))
		} else {
			pluginsDir := filepath.Join(gameDir, "SKSE", "Plugins")
			collectSKSEPluginsFromDir(pluginsDir, skseContrib{name: gameDataProviderName, order: -1},
				"SKSE/Plugins", opts.StripDataPrefix, byPath, &warnings)
		}
	}

	prefix := opts.ModNamePrefix
	for _, e := range entries {
		if opts.OnlyEnabled && !e.Enabled {
			continue
		}
		if prefix != "" {
			if len(e.Name) < len(prefix) || e.Name[:len(prefix)] != prefix {
				continue
			}
		}
		if strings.Contains(e.Name, "..") || filepath.IsAbs(e.Name) || e.Name == "" || e.Name == "." {
			warnings = append(warnings, fmt.Sprintf("skipped invalid mod name in modlist: %q", e.Name))
			continue
		}
		modRoot := filepath.Join(cfg.ModsDir, e.Name)
		st, statErr := os.Stat(modRoot)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				warnings = append(warnings, fmt.Sprintf("mod folder missing: %s", e.Name))
			} else {
				warnings = append(warnings, fmt.Sprintf("mod folder %s: %v", e.Name, statErr))
			}
			continue
		}
		if !st.IsDir() {
			warnings = append(warnings, fmt.Sprintf("mod path is not a directory: %s", e.Name))
			continue
		}

		c := skseContrib{name: e.Name, order: e.Order}
		collectSKSEPluginsFromDir(filepath.Join(modRoot, "Data", "SKSE", "Plugins"), c,
			filepath.Join("Data", "SKSE", "Plugins"), opts.StripDataPrefix, byPath, &warnings)
		collectSKSEPluginsFromDir(filepath.Join(modRoot, "SKSE", "Plugins"), c,
			filepath.Join("SKSE", "Plugins"), opts.StripDataPrefix, byPath, &warnings)
	}

	for vpath, list := range byPath {
		filtered := list[:0]
		for _, x := range list {
			if x.name == gameDataProviderName {
				filtered = append(filtered, x)
				continue
			}
			absPath, err := filepath.Abs(x.absPath)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("skipped %s: abs path: %v", vpath, err))
				continue
			}
			relMods, err := filepath.Rel(modsDirAbs, absPath)
			if err != nil || strings.HasPrefix(relMods, "..") {
				warnings = append(warnings, fmt.Sprintf("skipped path outside MO2_MODS_DIR: %s", x.absPath))
				continue
			}
			filtered = append(filtered, x)
		}
		if len(filtered) == 0 {
			delete(byPath, vpath)
		} else {
			byPath[vpath] = filtered
		}
	}

	vpaths := make([]string, 0, len(byPath))
	for p := range byPath {
		vpaths = append(vpaths, p)
	}
	sort.Strings(vpaths)

	truncated := false
	if len(vpaths) > opts.MaxDLLs {
		truncated = true
		warnings = append(warnings, fmt.Sprintf("truncated: output limited to max_dlls=%d (%d unique paths found)", opts.MaxDLLs, len(vpaths)))
		vpaths = vpaths[:opts.MaxDLLs]
	}

	rows := make([]SKSEPluginRow, 0, len(vpaths))
	for _, vpath := range vpaths {
		contribs := byPath[vpath]
		if len(contribs) == 0 {
			continue
		}
		sorted := append([]skseContrib(nil), contribs...)
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].order != sorted[j].order {
				return sorted[i].order < sorted[j].order
			}
			return sorted[i].name < sorted[j].name
		})
		winner := sorted[0]
		for _, c := range sorted[1:] {
			if c.order > winner.order || (c.order == winner.order && c.name > winner.name) {
				winner = c
			}
		}
		prov := make([]SKSEPluginProvider, len(sorted))
		for i, x := range sorted {
			prov[i] = SKSEPluginProvider{Name: x.name, Order: x.order, AbsPath: x.absPath}
		}
		row := SKSEPluginRow{
			VirtualPath: vpath,
			Winner:      SKSEPluginProvider{Name: winner.name, Order: winner.order, AbsPath: winner.absPath},
			Providers:   prov,
		}
		if opts.IncludeSize {
			if st, err := os.Stat(winner.absPath); err == nil && !st.IsDir() {
				sz := st.Size()
				row.Size = &sz
			}
		}
		if opts.IncludePEVersion {
			fv, pv, err := PEVersionStrings(winner.absPath)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("pe version %s: %v", vpath, err))
			}
			row.FileVersion = fv
			row.ProductVersion = pv
		}
		rows = append(rows, row)
	}

	return &SKSEPluginReport{
		ProfileDir:   cfg.ProfileDir,
		ModsDir:      cfg.ModsDir,
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		PriorityNote: sksePluginPriorityNote,
		Warnings:     warnings,
		Plugins:      rows,
		Truncated:    truncated,
	}, nil
}

// relPrefix uses forward slashes (e.g. Data/SKSE/Plugins or SKSE/Plugins).
func collectSKSEPluginsFromDir(absPluginsDir string, c skseContrib, relPrefix string, stripData bool, byPath map[string][]skseContrib, warnings *[]string) {
	entries, err := os.ReadDir(absPluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		*warnings = append(*warnings, fmt.Sprintf("read %s: %v", absPluginsDir, err))
		return
	}
	relPrefix = filepath.ToSlash(relPrefix)
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(ent.Name()), ".dll") {
			continue
		}
		full := filepath.Join(absPluginsDir, ent.Name())
		rel := relPrefix + "/" + ent.Name()
		vpath := NormalizeVirtualPath(rel, stripData)
		if vpath == "" || vpath == "." {
			continue
		}
		addSKSEContrib(byPath, vpath, c, full)
	}
}

func addSKSEContrib(byPath map[string][]skseContrib, vpath string, c skseContrib, abs string) {
	list := byPath[vpath]
	for _, x := range list {
		if x.name == c.name && x.order == c.order {
			return
		}
	}
	nc := c
	nc.absPath = abs
	byPath[vpath] = append(list, nc)
}
