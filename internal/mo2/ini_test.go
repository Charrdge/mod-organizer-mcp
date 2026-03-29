package mo2

import (
	"path/filepath"
	"testing"
)

func TestParseMetaINI(t *testing.T) {
	path := filepath.Join("testdata", "mods", "SkyUI", "meta.ini")
	m, err := ParseMetaINI(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := m["General.modid"]; got != "2014" {
		t.Errorf("General.modid = %q, want 2014", got)
	}
	if got := m["General.version"]; got != "5.2" {
		t.Errorf("General.version = %q, want 5.2", got)
	}
	if got := m["General.installationFile"]; got == "" {
		t.Error("expected installationFile")
	}
}

func TestParseINIBytes(t *testing.T) {
	raw := []byte("[Foo]\nbar = baz \n#skip\n")
	m, err := parseINIBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	if m["Foo.bar"] != "baz" {
		t.Errorf("got %v", m)
	}
}
