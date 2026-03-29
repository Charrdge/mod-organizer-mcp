package mo2

import (
	"os"
	"path/filepath"
	"strings"
)

// PluginOrderedEntry is one plugin in MO2 load order (optional snapshot contract block).
type PluginOrderedEntry struct {
	Index  int    `json:"index"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// BuildPluginsOrdered merges loadorder.txt order with plugins.txt active flags. Does not read plugin binaries.
func BuildPluginsOrdered(absProfileDir string) (entries []PluginOrderedEntry, warnings []string) {
	pluginsPath := filepath.Join(absProfileDir, "plugins.txt")
	pluginRaw, errPl := os.ReadFile(pluginsPath)
	var pluginEntries []PluginEntry
	if errPl == nil {
		pluginEntries = ParsePluginsFromBytes(pluginRaw)
	} else if !os.IsNotExist(errPl) {
		warnings = append(warnings, "plugins.txt: "+errPl.Error())
	}

	activeByLower := make(map[string]bool)
	inPlugins := make(map[string]struct{})
	for _, e := range pluginEntries {
		low := strings.ToLower(e.Name)
		activeByLower[low] = e.Active
		inPlugins[low] = struct{}{}
	}

	loadPath := filepath.Join(absProfileDir, "loadorder.txt")
	loLines, errLo := ReadTextLines(loadPath)
	if errLo != nil && !os.IsNotExist(errLo) {
		warnings = append(warnings, "loadorder.txt: "+errLo.Error())
	}

	var orderNames []string
	if len(loLines) > 0 {
		for _, line := range loLines {
			n := pluginNameFromListLine(line)
			if n != "" {
				orderNames = append(orderNames, n)
			}
		}
	}

	if len(orderNames) == 0 && len(pluginEntries) > 0 {
		for _, e := range pluginEntries {
			orderNames = append(orderNames, e.Name)
		}
		warnings = append(warnings, "load_order_source=plugins_txt_fallback")
	}

	seen := make(map[string]struct{})
	for _, name := range orderNames {
		low := strings.ToLower(name)
		if _, dup := seen[low]; dup {
			warnings = append(warnings, "load_order_duplicate_plugin:"+name)
			continue
		}
		seen[low] = struct{}{}

		active := true
		if a, ok := activeByLower[low]; ok {
			active = a
		} else {
			warnings = append(warnings, "plugin_in_order_missing_from_plugins_txt:"+name)
		}

		entries = append(entries, PluginOrderedEntry{
			Index:  len(entries),
			Name:   name,
			Active: active,
		})
	}

	for low := range inPlugins {
		if _, ok := seen[low]; !ok {
			var orig string
			for _, e := range pluginEntries {
				if strings.ToLower(e.Name) == low {
					orig = e.Name
					break
				}
			}
			if orig != "" {
				warnings = append(warnings, "plugin_in_plugins_txt_missing_from_load_order:"+orig)
			}
		}
	}

	return entries, warnings
}

func pluginNameFromListLine(line string) string {
	s := strings.TrimSpace(line)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "*") {
		s = strings.TrimSpace(s[1:])
	}
	return s
}
