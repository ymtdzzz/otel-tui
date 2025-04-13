package json

import (
	"testing"

	"gotest.tools/v3/assert"
)

func Test_prettyJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "valid JSON object",
			input: `{"foo":"bar"}`,
			expected: `{
  "foo": "bar"
}`,
		},
		{
			name:     "invalid JSON",
			input:    `{invalid}`,
			expected: `{invalid}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrettyJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
