package mo2

// SnapshotOptions controls what BuildSnapshotWithOptions includes.
type SnapshotOptions struct {
	IncludeMeta           bool
	IncludePluginLines    bool
	IncludeLoadorderLines bool
	OnlyEnabled           bool
	ModNamePrefix         string
}

// DefaultSnapshotOptions matches historical full snapshot behavior.
func DefaultSnapshotOptions() SnapshotOptions {
	return SnapshotOptions{
		IncludeMeta:           true,
		IncludePluginLines:    true,
		IncludeLoadorderLines: true,
		OnlyEnabled:           false,
		ModNamePrefix:         "",
	}
}
