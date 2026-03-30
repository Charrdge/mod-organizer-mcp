package toolreg

import (
	"context"
	"encoding/json"

	"github.com/charrdge/mod-organizer-mcp/internal/mo2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serverVersion = "0.6.0"

type profileSnapshotArgs struct {
	IncludeMeta            *bool  `json:"include_meta,omitempty" jsonschema:"When false, skips per-mod meta.ini (faster). Default true."`
	IncludePluginLines     *bool  `json:"include_plugin_lines,omitempty" jsonschema:"Include plugins.txt lines in output. Default true."`
	IncludeLoadorderLines  *bool  `json:"include_loadorder_lines,omitempty" jsonschema:"Include loadorder.txt lines. Default true."`
	OnlyEnabled            bool   `json:"only_enabled,omitempty" jsonschema:"Only mods with + in modlist."`
	ModNamePrefix          string `json:"mod_name_prefix,omitempty" jsonschema:"Only mods whose folder name starts with this prefix."`
	IncludeContract        *bool  `json:"include_contract,omitempty" jsonschema:"When false, omits snapshot_contract_version, profile_ini, profile_list_paths, archive_search_roots. Default true."`
	IncludePluginLoadOrder *bool  `json:"include_plugin_load_order,omitempty" jsonschema:"When true, adds plugins_ordered (loadorder.txt + plugins.txt merge). Default false; use mutagen_list_plugins when Mutagen is available."`
}

func mergeAssetConflictOpts(in assetConflictsArgs) mo2.AssetConflictOptions {
	o := mo2.DefaultAssetConflictOptions()
	if in.OnlyEnabled != nil {
		o.OnlyEnabled = *in.OnlyEnabled
	}
	o.PathPrefix = in.PathPrefix
	if in.MaxFilesTotal > 0 {
		o.MaxFilesTotal = in.MaxFilesTotal
	}
	if in.MaxDepth > 0 {
		o.MaxDepth = in.MaxDepth
	}
	if in.StripDataPrefix != nil {
		o.StripDataPrefix = *in.StripDataPrefix
	}
	o.IncludeSingleWinnerPaths = in.IncludeSingleWinnerPaths
	return o
}

func mergeSnapshotOpts(in profileSnapshotArgs) mo2.SnapshotOptions {
	o := mo2.DefaultSnapshotOptions()
	if in.IncludeMeta != nil {
		o.IncludeMeta = *in.IncludeMeta
	}
	if in.IncludePluginLines != nil {
		o.IncludePluginLines = *in.IncludePluginLines
	}
	if in.IncludeLoadorderLines != nil {
		o.IncludeLoadorderLines = *in.IncludeLoadorderLines
	}
	o.OnlyEnabled = in.OnlyEnabled
	o.ModNamePrefix = in.ModNamePrefix
	if in.IncludeContract != nil {
		o.IncludeContract = *in.IncludeContract
	}
	if in.IncludePluginLoadOrder != nil {
		o.IncludePluginLoadOrder = *in.IncludePluginLoadOrder
	}
	return o
}

type machineContractArgs struct {
	OnlyEnabled   bool   `json:"only_enabled,omitempty" jsonschema:"If true, only + mods from modlist for archive_search_roots."`
	ModNamePrefix string `json:"mod_name_prefix,omitempty" jsonschema:"Only mods whose folder name starts with this prefix (same as mo2_profile_snapshot)."`
}

type nexusLocalArgs struct {
	OnlyEnabled bool `json:"only_enabled,omitempty" jsonschema:"If true, only + mods from modlist."`
}

type modLookupArgs struct {
	Name string `json:"name" jsonschema:"Mod folder name as in modlist (exact, case-insensitive, or unique prefix)"`
}

type listModPluginsArgs struct {
	Name     string `json:"name" jsonschema:"Mod folder name under MO2_MODS_DIR"`
	MaxDepth int    `json:"max_depth,omitempty" jsonschema:"Max directory depth from mod root; default 8"`
	MaxFiles int    `json:"max_files,omitempty" jsonschema:"Max plugin files returned; default 200"`
}

type sksePluginsArgs struct {
	OnlyEnabled      *bool  `json:"only_enabled,omitempty" jsonschema:"If true (default), only + mods from modlist"`
	ModNamePrefix    string `json:"mod_name_prefix,omitempty" jsonschema:"Only mods whose folder name starts with this prefix"`
	GameDataDir      string `json:"game_data_dir,omitempty" jsonschema:"Optional absolute path to the game's Data folder; scans <path>/SKSE/Plugins as lowest-priority overlay (below all mods)"`
	IncludeSize      *bool  `json:"include_size,omitempty" jsonschema:"Include winner file size in bytes. Default true."`
	IncludePEVersion bool   `json:"include_pe_version,omitempty" jsonschema:"Parse PE StringFileInfo for FileVersion/ProductVersion (slower). Default false."`
	MaxDLLs          int    `json:"max_dlls,omitempty" jsonschema:"Max unique virtual paths in output; default 500"`
	StripDataPrefix  *bool  `json:"strip_data_prefix,omitempty" jsonschema:"If true (default), strip leading Data/ for virtual paths"`
}

func mergeSKSEOpts(in sksePluginsArgs) mo2.SKSEPluginOptions {
	o := mo2.DefaultSKSEPluginOptions()
	if in.OnlyEnabled != nil {
		o.OnlyEnabled = *in.OnlyEnabled
	}
	o.ModNamePrefix = in.ModNamePrefix
	o.GameDataDir = in.GameDataDir
	if in.IncludeSize != nil {
		o.IncludeSize = *in.IncludeSize
	}
	o.IncludePEVersion = in.IncludePEVersion
	if in.MaxDLLs > 0 {
		o.MaxDLLs = in.MaxDLLs
	}
	if in.StripDataPrefix != nil {
		o.StripDataPrefix = *in.StripDataPrefix
	}
	return o
}

type assetConflictsArgs struct {
	OnlyEnabled              *bool  `json:"only_enabled,omitempty" jsonschema:"If true (default), only + mods from modlist"`
	PathPrefix               string `json:"path_prefix,omitempty" jsonschema:"Only virtual paths with this prefix (forward slashes), e.g. textures/"`
	MaxFilesTotal            int    `json:"max_files_total,omitempty" jsonschema:"Stop after scanning this many files across all mods; default 200000"`
	MaxDepth                 int    `json:"max_depth,omitempty" jsonschema:"Max slash-depth from mod root (0 = unlimited)"`
	StripDataPrefix          *bool  `json:"strip_data_prefix,omitempty" jsonschema:"If true (default), strip leading Data/ segment so Data/textures/ matches textures/"`
	IncludeSingleWinnerPaths bool   `json:"include_single_winner_paths,omitempty" jsonschema:"If true, include paths provided by only one mod (large JSON). Default false."`
}

func jsonText(v any) *mcp.CallToolResult {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return toolErr(err.Error())
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}},
	}
}

// Register wires Mod Organizer 2 read-only tools onto the MCP server.
func Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "mo2_profile_snapshot",
		Description: "Read-only: JSON snapshot from MO2_PROFILE_DIR and MO2_MODS_DIR. By default includes machine-readable contract: snapshot_contract_version, profile_ini (whitelist), profile_list_paths (abs plugins.txt/loadorder.txt), archive_search_roots (per-mod mod_root and optional data_subdir for .bsa/.ba2 scanning). Set include_plugin_load_order true for plugins_ordered (off by default; prefer Mutagen for plugin lists). Optional filters: include_meta, plugin/loadorder lines, only_enabled, mod_name_prefix, include_contract.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in profileSnapshotArgs) (*mcp.CallToolResult, any, error) {
		cfg, err := mo2.ConfigFromEnv()
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		snap, err := mo2.BuildSnapshotWithOptions(cfg, mergeSnapshotOpts(in))
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		return jsonText(snap), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mo2_profile_machine_contract",
		Description: "Read-only: machine-readable next-step contract only (snapshot_contract_version, profile_ini whitelist, profile_list_paths, archive_search_roots). No mods[] or meta.ini. Same path env as snapshot. Optional only_enabled and mod_name_prefix filter roots.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in machineContractArgs) (*mcp.CallToolResult, any, error) {
		cfg, err := mo2.ConfigFromEnv()
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		out, err := mo2.BuildMachineContractPayload(cfg, in.OnlyEnabled, in.ModNamePrefix)
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		return jsonText(out), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mo2_profile_plugin_load_order",
		Description: "Read-only: plugins_ordered from loadorder.txt merged with plugins.txt active flags, plus profile_list_paths. For MO2-only workflows without Mutagen; prefer mutagen_list_plugins when Mutagen is configured.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		cfg, err := mo2.ConfigFromEnv()
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		out, err := mo2.BuildPluginLoadOrderPayload(cfg)
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		return jsonText(out), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mo2_list_profiles",
		Description: "List sibling MO2 profiles: subfolders of the parent of MO2_PROFILE_DIR that contain modlist.txt (name, path, modlist entry count).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		cfg, err := mo2.ConfigFromEnv()
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		list, err := mo2.ListSiblingProfiles(cfg.ProfileDir)
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		return jsonText(list), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mo2_profile_summary",
		Description: "Lightweight profile stats: enabled/disabled mod counts, plugins.txt/loadorder presence and line counts, duplicate names in modlist, missing mod folders. Does not read meta.ini.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		cfg, err := mo2.ConfigFromEnv()
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		s, err := mo2.BuildProfileSummary(cfg)
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		return jsonText(s), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mo2_nexus_local_index",
		Description: "Read-only index of Nexus-related fields from each mod's meta.ini (mod id, file id, version, etc.). source=meta_ini; not live Nexus API — use nexusmods-mcp for current site data.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in nexusLocalArgs) (*mcp.CallToolResult, any, error) {
		cfg, err := mo2.ConfigFromEnv()
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		idx, err := mo2.BuildNexusLocalIndex(cfg, in.OnlyEnabled)
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		return jsonText(idx), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mo2_mod_lookup",
		Description: "Look up one mod by folder name: exact match, then case-insensitive, then unique prefix. Returns modlist state, meta.ini map, and warnings. ambiguous_candidates if multiple matches.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in modLookupArgs) (*mcp.CallToolResult, any, error) {
		cfg, err := mo2.ConfigFromEnv()
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		out, err := mo2.LookupMod(cfg, in.Name)
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		return jsonText(out), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mo2_list_plugins",
		Description: "Structured plugins.txt for the active profile: plugin name and active flag (leading * treated as inactive).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		cfg, err := mo2.ConfigFromEnv()
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		ents, err := mo2.ParsePluginsFile(cfg.ProfileDir)
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		return jsonText(ents), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mo2_skse_plugins",
		Description: "Read-only: list .dll under virtual Data/SKSE/Plugins (loose files in enabled mods: Data/SKSE/Plugins and SKSE/Plugins). Resolves overlay winner by modlist order (later mod wins). Optional game_data_dir = game's Data folder path adds SKSE/Plugins from disk as lowest priority (GameData). Optional include_pe_version for PE FileVersion/ProductVersion. See priority_note in JSON.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in sksePluginsArgs) (*mcp.CallToolResult, any, error) {
		cfg, err := mo2.ConfigFromEnv()
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		report, err := mo2.BuildSKSEPluginInventory(cfg, mergeSKSEOpts(in))
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		return jsonText(report), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mo2_asset_conflicts",
		Description: "Loose files only (not BSA/BA2): scan enabled mods in modlist order and list virtual paths where multiple mods provide the same relative game path. Later mod in modlist (higher order) wins. Use path_prefix and max_files_total to limit scope. See priority_note in JSON.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in assetConflictsArgs) (*mcp.CallToolResult, any, error) {
		cfg, err := mo2.ConfigFromEnv()
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		report, err := mo2.BuildAssetConflicts(cfg, mergeAssetConflictOpts(in))
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		return jsonText(report), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mo2_list_mod_plugins",
		Description: "List .esp/.esm/.esl files under MO2_MODS_DIR/<name> with max_depth (default 8) and max_files (default 200). Paths relative to mods dir.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listModPluginsArgs) (*mcp.CallToolResult, any, error) {
		cfg, err := mo2.ConfigFromEnv()
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		rel, warnings, err := mo2.ListModPluginArchives(cfg.ModsDir, in.Name, in.MaxDepth, in.MaxFiles)
		if err != nil {
			return toolErr(err.Error()), nil, nil
		}
		payload := struct {
			ModName  string   `json:"mod_name"`
			Plugins  []string `json:"plugins"`
			Warnings []string `json:"warnings,omitempty"`
		}{ModName: in.Name, Plugins: rel, Warnings: warnings}
		return jsonText(payload), nil, nil
	})

	type serverInfo struct {
		Version     string `json:"version"`
		ProfileDir  string `json:"profile_dir,omitempty"`
		ModsDir     string `json:"mods_dir,omitempty"`
		ConfigError string `json:"config_error,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "mo2_server_info",
		Description: "Server version and resolved MO2_PROFILE_DIR / MO2_MODS_DIR from the environment (or config_error if invalid).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		info := serverInfo{Version: serverVersion}
		cfg, err := mo2.ConfigFromEnv()
		if err != nil {
			info.ConfigError = err.Error()
		} else {
			info.ProfileDir = cfg.ProfileDir
			info.ModsDir = cfg.ModsDir
		}
		return jsonText(info), nil, nil
	})
}

func toolErr(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}
