package tuiv2

import "strings"

func MergeKeysToString(keys ...string) string {
	if len(keys) == 0 {
		return ""
	}

	// Replace ctrl+ with ^
	for i, key := range keys {
		if strings.HasPrefix(key, "ctrl+") {
			keys[i] = strings.ReplaceAll(key, "ctrl+", "^")
		}
	}

	return strings.Join(keys, ",")
}
