# mod-organizer-mcp

Read-only [MCP](https://modelcontextprotocol.io/) server for **Mod Organizer 2**: builds a JSON snapshot from `modlist.txt`, optional `plugins.txt` / `loadorder.txt`, and per-mod `meta.ini` under the `mods` folder. **No writes** to your MO2 or game directories.

## Requirements

- Go 1.26+
- Environment variables (absolute paths). On WSL, a Windows drive is usually `/mnt/<letter>/...` (e.g. `S:` → `/mnt/s/...`).

| Variable | Meaning |
|----------|---------|
| `MO2_PROFILE_DIR` | Profile directory containing `modlist.txt` |
| `MO2_MODS_DIR` | MO2 `mods` directory (each mod is a subfolder, often with `meta.ini`) |

## Build

```bash
go build -o mod-organizer-mcp ./cmd/server
```

## Run (stdio, for Cursor)

```bash
export MO2_PROFILE_DIR="/path/to/MO2_instance/MO2_data/profiles/MyProfile"
export MO2_MODS_DIR="/path/to/MO2_instance/MO2_data/mods"
./mod-organizer-mcp
```

Optional HTTP transport (same pattern as other Go MCP servers):

```bash
export MCP_TRANSPORT=http
export MCP_HTTP_ADDR=":8080"
./mod-organizer-mcp
```

## Cursor MCP config example

Adjust paths to your layout. Paths with spaces must be quoted in JSON.

```json
{
  "mcpServers": {
    "mod-organizer-mcp": {
      "command": "/path/to/mod-organizer-mcp",
      "env": {
        "MO2_PROFILE_DIR": "/path/to/MO2_instance/MO2_data/profiles/MyProfile",
        "MO2_MODS_DIR": "/path/to/MO2_instance/MO2_data/mods"
      }
    }
  }
}
```

If the binary lives elsewhere, use `go build` output path or `command": "go"` with `args`: `["run", "./cmd/server"]` and `cwd` set to this repo (slower startup).

## Tools

| Tool | Purpose |
|------|---------|
| **`mo2_server_info`** | Version and resolved env paths, or `config_error`. |
| **`mo2_list_profiles`** | Sibling profiles under the parent of `MO2_PROFILE_DIR` (each folder with `modlist.txt`): name, path, modlist entry count. |
| **`mo2_profile_summary`** | Counts (`+`/`-` mods, duplicates in modlist), `plugins.txt` / `loadorder.txt` presence and line counts, missing mod folders. Does **not** read `meta.ini`. |
| **`mo2_profile_snapshot`** | Full JSON snapshot (`mods[]` with optional `meta`, `plugin_lines`, `loadorder_lines`). Optional arguments (all optional): `include_meta`, `include_plugin_lines`, `include_loadorder_lines` (default **true** each), `only_enabled`, `mod_name_prefix`. Omit arguments for legacy full snapshot. |
| **`mo2_nexus_local_index`** | Compact Nexus-oriented fields from each mod’s `meta.ini` (`nexus_mod_id`, `nexus_file_id`, version, game, etc.). Root fields: `source` = `"meta_ini"`, `live_nexus_api` = **false** — this is MO’s last-known disk metadata, not the live Nexus API. Optional: `only_enabled`. |
| **`mo2_mod_lookup`** | Argument `name`: one mod (exact → case-insensitive → unique prefix). Returns `match` or `ambiguous_candidates` / `not_found`. |
| **`mo2_list_plugins`** | Structured `plugins.txt`: `name` + `active` (leading `*` → inactive). |
| **`mo2_list_mod_plugins`** | Arguments: `name`, optional `max_depth` (default 8), `max_files` (default 200). Lists `.esp` / `.esm` / `.esl` under that mod folder (paths relative to `MO2_MODS_DIR`). |

## Safety and policy

- This server **only reads** files. It does not modify `modlist.txt`, `meta.ini`, or anything else.
- Prefer a **copy** of the profile when Mod Organizer is running and may lock files.
- Treat paths under your Skyrim/MO2 tree as sensitive: only grant write access to tools or agents when **you** explicitly allow it in context (this MCP does not write).

## WSL: Windows drive not visible

If tools fail with errors like `Transport endpoint is not connected` or `no such file` under `/mnt/<letter>`:

1. Ensure the drive is mounted in WSL (`/mnt/s` or `/mnt/c` exists and lists the folder where MO2 lives).
2. From Windows, run `wsl --mount` or open the distro after accessing `S:` once, depending on your WSL version and `wsl.conf` automount settings.
3. Use `mo2_server_info` after fixing mounts to confirm paths resolve.

## Nexus Mods

- **This MCP** exposes **`mo2_nexus_local_index`**: whatever MO2 stored in `meta.ini` (ids, version, file name). Treat it as a **local cache**; it can be stale.
- **Live Nexus** (current file version, descriptions, dependencies): use a **separate** MCP that talks to the Nexus Mods API (any implementation you trust). Typical flow: `mo2_nexus_local_index` or `mo2_mod_lookup` → then that MCP with the mod/file id from disk.

### Backlog (not implemented here)

MCP **resources** (`mo2://…` URIs), optional `MO2_INSTANCE_DIR` + `modorganizer.ini`, and parsing MO2 **`webcache`** are left for future work.

## Agent notes

- Prefer **`mo2_profile_summary`** or **`mo2_nexus_local_index`** when you need small JSON; use **`mo2_profile_snapshot`** with filters or only when you need full `meta` / line lists.
- After context compaction, re-call **`mo2_nexus_local_index`** instead of repeating the same Nexus API requests for ids already on disk.

## Public repo / forks

- **`.cursor/`** is gitignored: IDE plans often embed absolute local paths; keep them out of git.
- **`go.mod` module path** (`github.com/.../mod-organizer-mcp`) should match the Git URL you publish under; after forking, replace the module path and all `import` strings (or use a multi-step rename as in standard Go fork guides).

## License

No license file is included unless you add one.
