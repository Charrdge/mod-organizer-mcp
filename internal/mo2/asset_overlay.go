package mo2

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const assetConflictPriorityNote = "MO2 mod priority: a mod later in modlist.txt (higher order index) wins over earlier mods for the same virtual path. This scan covers loose files only, not BSA/BA2 contents."

// AssetConflictOptions controls BuildAssetConflicts scanning and output.
type AssetConflictOptions struct {
	OnlyEnabled              bool
	PathPrefix               string
	MaxFilesTotal            int
	MaxDepth                 int
	StripDataPrefix          bool
	IncludeSingleWinnerPaths bool
}

// DefaultAssetConflictOptions returns defaults: enabled mods only, strip Data/ prefix, cap 200k files, conflicts with 2+ providers only.
func DefaultAssetConflictOptions() AssetConflictOptions {
	return AssetConflictOptions{
		OnlyEnabled:              true,
		MaxFilesTotal:            200_000,
		MaxDepth:                 0,
		StripDataPrefix:          true,
		IncludeSingleWinnerPaths: false,
	}
}

// VirtualContributor is one mod that provides a virtual path.
type VirtualContributor struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// AssetConflictRow is one virtual path and who wins vs all providers.
type AssetConflictRow struct {
	VirtualPath  string               `json:"virtual_path"`
	Winner       VirtualContributor   `json:"winner"`
	Contributors []VirtualContributor `json:"contributors"`
}

// AssetConflictsReport is the JSON payload for asset conflict analysis.
type AssetConflictsReport struct {
	ProfileDir   string             `json:"profile_dir"`
	ModsDir      string             `json:"mods_dir"`
	GeneratedAt  string             `json:"generated_at"`
	PriorityNote string             `json:"priority_note"`
	ScannedFiles int                `json:"scanned_files"`
	Warnings     []string           `json:"warnings,omitempty"`
	Conflicts    []AssetConflictRow `json:"conflicts"`
}

// NormalizeVirtualPath maps a path relative to a mod folder to a game Data-relative key (forward slashes).
// If stripDataPrefix is true, strips a leading "Data" path segment (case-insensitive).
func NormalizeVirtualPath(rel string, stripDataPrefix bool) string {
	s := strings.ReplaceAll(rel, `\`, `/`)
	s = path.Clean(s)
	s = strings.TrimPrefix(s, "./")
	if stripDataPrefix {
		s = stripLeadingDataSegment(s)
	}
	return s
}

func stripLeadingDataSegment(s string) string {
	i := strings.IndexByte(s, '/')
	if i < 0 {
		if strings.EqualFold(s, "data") {
			return ""
		}
		return s
	}
	if strings.EqualFold(s[:i], "data") {
		return s[i+1:]
	}
	return s
}

func depthFromModRel(rel string) int {
	return strings.Count(filepath.ToSlash(rel), "/")
}

// BuildAssetConflicts scans enabled mods in modlist order and reports paths where multiple mods supply the same virtual path (by default).
func BuildAssetConflicts(cfg Config, opts AssetConflictOptions) (*AssetConflictsReport, error) {
	modlistPath := filepath.Join(cfg.ProfileDir, "modlist.txt")
	entries, err := ParseModlist(modlistPath)
	if err != nil {
		return nil, err
	}
	if opts.MaxFilesTotal <= 0 {
		opts.MaxFilesTotal = DefaultAssetConflictOptions().MaxFilesTotal
	}
	prefix := filepath.ToSlash(strings.TrimSpace(opts.PathPrefix))

	modsDirAbs, err := filepath.Abs(filepath.Clean(cfg.ModsDir))
	if err != nil {
		return nil, fmt.Errorf("mods dir abs: %w", err)
	}

	byPath := make(map[string][]VirtualContributor)
	var warnings []string
	scanned := 0
	truncated := false

	for _, e := range entries {
		if opts.OnlyEnabled && !e.Enabled {
			continue
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

		walkErr := filepath.WalkDir(modRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if truncated {
				return filepath.SkipAll
			}
			if d.IsDir() {
				if d.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			if scanned >= opts.MaxFilesTotal {
				truncated = true
				warnings = append(warnings, fmt.Sprintf("truncated: reached max_files_total=%d", opts.MaxFilesTotal))
				return filepath.SkipAll
			}

			rel, err := filepath.Rel(modRoot, path)
			if err != nil || strings.HasPrefix(rel, "..") {
				warnings = append(warnings, "skipped path outside mod root")
				return nil
			}
			relSlash := filepath.ToSlash(rel)
			if relSlash == "meta.ini" {
				return nil
			}
			if opts.MaxDepth > 0 && depthFromModRel(rel) > opts.MaxDepth {
				return nil
			}

			vpath := NormalizeVirtualPath(rel, opts.StripDataPrefix)
			if vpath == "" || vpath == "." {
				return nil
			}
			if prefix != "" && !strings.HasPrefix(vpath, prefix) {
				return nil
			}

			// Ensure contributor is under MO2_MODS_DIR (symlink escape)
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil
			}
			relMods, err := filepath.Rel(modsDirAbs, absPath)
			if err != nil || strings.HasPrefix(relMods, "..") {
				warnings = append(warnings, "skipped path outside MO2_MODS_DIR")
				return nil
			}

			scanned++
			c := VirtualContributor{Name: e.Name, Order: e.Order}
			list := byPath[vpath]
			for _, x := range list {
				if x.Name == e.Name {
					return nil
				}
			}
			byPath[vpath] = append(list, c)
			return nil
		})
		if walkErr != nil {
			warnings = append(warnings, fmt.Sprintf("walk %s: %v", e.Name, walkErr))
		}
		if truncated {
			break
		}
	}

	rows := make([]AssetConflictRow, 0)
	for vpath, contribs := range byPath {
		if len(contribs) == 0 {
			continue
		}
		if !opts.IncludeSingleWinnerPaths && len(contribs) < 2 {
			continue
		}
		sorted := append([]VirtualContributor(nil), contribs...)
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].Order != sorted[j].Order {
				return sorted[i].Order < sorted[j].Order
			}
			return sorted[i].Name < sorted[j].Name
		})
		winner := sorted[0]
		for _, c := range sorted[1:] {
			if c.Order > winner.Order || (c.Order == winner.Order && c.Name > winner.Name) {
				winner = c
			}
		}
		rows = append(rows, AssetConflictRow{
			VirtualPath:  vpath,
			Winner:       winner,
			Contributors: sorted,
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].VirtualPath < rows[j].VirtualPath })

	return &AssetConflictsReport{
		ProfileDir:   cfg.ProfileDir,
		ModsDir:      cfg.ModsDir,
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		PriorityNote: assetConflictPriorityNote,
		ScannedFiles: scanned,
		Warnings:     warnings,
		Conflicts:    rows,
	}, nil
}
