package json

import (
	"bytes"
	"encoding/json"
)

func PrettyJSON(s string) string {
	if !json.Valid([]byte(s)) {
		return s
	}
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(s), "", "  "); err != nil {
		return s
	}
	return prettyJSON.String()
}
