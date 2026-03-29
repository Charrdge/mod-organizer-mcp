package mo2

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeVirtualPath(t *testing.T) {
	cases := []struct {
		rel    string
		strip  bool
		expect string
	}{
		{`textures\foo.dds`, true, "textures/foo.dds"},
		{"./meshes/x.nif", true, "meshes/x.nif"},
		{"Data/textures/x.dds", true, "textures/x.dds"},
		{"data/textures/x.dds", true, "textures/x.dds"},
		{"Data/textures/x.dds", false, "Data/textures/x.dds"},
		{"Data", true, ""},
	}
	for _, c := range cases {
		got := NormalizeVirtualPath(c.rel, c.strip)
		if got != c.expect {
			t.Errorf("NormalizeVirtualPath(%q, %v) = %q, want %q", c.rel, c.strip, got, c.expect)
		}
	}
}

func TestBuildAssetConflicts_winnerByModlistOrder(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile_conflicts"),
		ModsDir:    filepath.Join("testdata", "mods_conflicts"),
	}
	opts := DefaultAssetConflictOptions()
	opts.PathPrefix = "textures/"
	report, err := BuildAssetConflicts(cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts) != 1 {
		t.Fatalf("conflicts: got %d, want 1: %+v", len(report.Conflicts), report.Conflicts)
	}
	row := report.Conflicts[0]
	if row.VirtualPath != "textures/a.dds" {
		t.Errorf("virtual_path: %q", row.VirtualPath)
	}
	if row.Winner.Name != "ModHigh" || row.Winner.Order != 1 {
		t.Errorf("winner: %+v (want ModHigh order 1)", row.Winner)
	}
	if len(row.Contributors) != 2 || row.Contributors[0].Name != "ModLow" || row.Contributors[1].Name != "ModHigh" {
		t.Errorf("contributors: %+v", row.Contributors)
	}
}

func TestBuildAssetConflicts_stripDataPrefixUnifiesPaths(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile_conflicts_data"),
		ModsDir:    filepath.Join("testdata", "mods_conflicts"),
	}
	opts := DefaultAssetConflictOptions()
	opts.StripDataPrefix = true
	opts.PathPrefix = "textures/"
	report, err := BuildAssetConflicts(cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts) != 1 || report.Conflicts[0].VirtualPath != "textures/b.dds" {
		t.Fatalf("got %+v", report.Conflicts)
	}
	if report.Conflicts[0].Winner.Name != "ModWrapped" {
		t.Errorf("winner: %+v", report.Conflicts[0].Winner)
	}
}

func TestBuildAssetConflicts_maxFilesTotalTruncation(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile_trunc"),
		ModsDir:    filepath.Join("testdata", "mods_conflicts"),
	}
	opts := DefaultAssetConflictOptions()
	opts.MaxFilesTotal = 2
	opts.IncludeSingleWinnerPaths = true
	report, err := BuildAssetConflicts(cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	if report.ScannedFiles != 2 {
		t.Errorf("scanned: %d want 2", report.ScannedFiles)
	}
	var sawTrunc bool
	for _, w := range report.Warnings {
		if strings.Contains(w, "max_files_total=2") {
			sawTrunc = true
		}
	}
	if !sawTrunc {
		t.Errorf("expected truncation warning, got %v", report.Warnings)
	}
	if len(report.Conflicts) != 2 {
		t.Errorf("with single-winner mode want 2 paths scanned, got %d conflicts", len(report.Conflicts))
	}
}

func TestBuildAssetConflicts_pathPrefixFilters(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile_conflicts"),
		ModsDir:    filepath.Join("testdata", "mods_conflicts"),
	}
	opts := DefaultAssetConflictOptions()
	opts.PathPrefix = "meshes/"
	report, err := BuildAssetConflicts(cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts) != 0 {
		t.Fatalf("expected no meshes conflicts, got %+v", report.Conflicts)
	}
}
