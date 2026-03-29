package mo2

import (
	"path/filepath"
	"testing"
)

func TestListSiblingProfiles(t *testing.T) {
	prof := filepath.Join("testdata", "profiles_parent", "profiles", "Alpha")
	list, err := ListSiblingProfiles(prof)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 profiles, got %d: %+v", len(list), list)
	}
	if list[0].Name != "Alpha" || list[1].Name != "Beta" {
		t.Fatalf("order/names: %+v", list)
	}
	if list[0].ModlistEntries != 1 || list[1].ModlistEntries != 2 {
		t.Fatalf("modlist counts: %+v", list)
	}
}
