package mo2

import (
	"path/filepath"
	"testing"
)

func TestExtractNexusLocal_SkyUIStyle(t *testing.T) {
	meta := map[string]string{
		"General.modid":            "2014",
		"General.version":          "5.2",
		"General.installationFile": "SkyUI_5_2_SE-2014-5-2.7z",
	}
	n, w := ExtractNexusLocal(meta)
	if n.ModID != "2014" || n.Version != "5.2" {
		t.Fatalf("nexus: %+v", n)
	}
	if len(w) != 0 {
		t.Fatalf("unexpected warnings: %v", w)
	}
}

func TestExtractNexusLocal_modIDKeys(t *testing.T) {
	meta := map[string]string{
		"General.modID":    "12345",
		"General.fileID":   "987654",
		"General.gameName": "Skyrim Special Edition",
	}
	n, _ := ExtractNexusLocal(meta)
	if n.ModID != "12345" || n.FileID != "987654" || n.GameName != "Skyrim Special Edition" {
		t.Fatalf("nexus: %+v", n)
	}
}

func TestExtractNexusLocal_noNexus(t *testing.T) {
	meta := map[string]string{"General.notes": "hello"}
	_, w := ExtractNexusLocal(meta)
	if len(w) == 0 {
		t.Fatal("expected warning")
	}
}

func TestBuildNexusLocalIndex(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile_nexus"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	idx, err := BuildNexusLocalIndex(cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if idx.Source != "meta_ini" || idx.LiveNexusAPI != false {
		t.Fatalf("meta: %+v", idx)
	}
	var sawNexus bool
	for _, row := range idx.Mods {
		if row.FolderName == "NexusStyle" {
			sawNexus = true
			if row.Nexus.ModID != "12345" || row.Nexus.FileID != "987654" {
				t.Fatalf("nexus row: %+v", row.Nexus)
			}
		}
		if row.FolderName == "NoMetaMod" {
			if len(row.Warnings) == 0 {
				t.Fatal("expected warnings for NoMetaMod")
			}
		}
	}
	if !sawNexus {
		t.Fatal("missing NexusStyle row")
	}
}
