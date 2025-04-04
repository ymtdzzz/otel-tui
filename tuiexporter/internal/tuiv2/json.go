package tuiv2

import (
	"bytes"
	"encoding/json"
)

func prettyJSON(s string) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(s), "", "  "); err != nil {
		return s
	}
	return prettyJSON.String()
}

// PrettyJSONFromKVFormat takes a string in the format "key: value"
// and returns it with the value formatted as pretty JSON if valid.
func PrettyJSONFromKVFormat(s string) string {
	parts := bytes.Split([]byte(s), []byte(": "))
	if len(parts) >= 2 {
		value := string(parts[1])
		if json.Valid([]byte(value)) {
			value = prettyJSON(value)
		}
		return string(parts[0]) + ": " + value
	}
	return s
}
