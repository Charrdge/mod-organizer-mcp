package toolreg

import (
	"context"
	"encoding/json"

	"github.com/charrdge/mod-organizer-mcp/internal/mo2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serverVersion = "0.2.0"

type profileSnapshotArgs struct {
	IncludeMeta           *bool  `json:"include_meta,omitempty" jsonschema:"When false, skips per-mod meta.ini (faster). Default true."`
	IncludePluginLines    *bool  `json:"include_plugin_lines,omitempty" jsonschema:"Include plugins.txt lines in output. Default true."`
	IncludeLoadorderLines *bool  `json:"include_loadorder_lines,omitempty" jsonschema:"Include loadorder.txt lines. Default true."`
	OnlyEnabled           bool   `json:"only_enabled,omitempty" jsonschema:"Only mods with + in modlist."`
	ModNamePrefix         string `json:"mod_name_prefix,omitempty" jsonschema:"Only mods whose folder name starts with this prefix."`
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
	return o
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
		Description: "Read-only: JSON snapshot from MO2_PROFILE_DIR (modlist.txt, optional plugins.txt/loadorder.txt) and MO2_MODS_DIR (per-mod meta.ini). Optional filters reduce payload size. Paths only from environment.",
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
