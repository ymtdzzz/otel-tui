package test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/require"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)

	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("project root (with go.mod) not found")
		}
		dir = parent
	}
}

func LoadTestdata(t *testing.T, name string) string {
	t.Helper()
	root := projectRoot(t)
	path := filepath.Join(root, "testdata", name)
	data, err := os.ReadFile(path) //gosec:disable G304 -- This is a test code
	require.NoError(t, err, "failed to read testdata %s", name)

	return string(data)
}

func GetScreenContent(t *testing.T, screen tcell.SimulationScreen) bytes.Buffer {
	t.Helper()
	content, w, _ := screen.GetContents()
	var got bytes.Buffer
	for n, v := range content {
		var err error
		ch := v.Runes[0]
		// Replace SimulationScreen's default fill character ('X') with space.
		// Since tcell v2.12.0, the Put API changes (PR #846, #848) caused the fill
		// character to become visible in test environments when tview components don't
		// fully occupy their allocated space. This doesn't occur in actual terminals,
		// so we normalize the test environment to match real-world behavior.
		// See https://github.com/gdamore/tcell/pull/848

		// To avoid replacing intentional 'X' characters (e.g., in text content or
		// control characters like Ctrl-X), we only replace 'X' when it appears as
		// part of a consecutive fill pattern.
		if ch == 'X' {
			prevIsX := n > 0 && content[n-1].Runes[0] == 'X'
			nextIsX := n < len(content)-1 && content[n+1].Runes[0] == 'X'
			if prevIsX || nextIsX {
				ch = ' '
			}
		}
		if n%w == w-1 {
			_, err = fmt.Fprintf(&got, "%c\n", ch)
		} else {
			_, err = fmt.Fprintf(&got, "%c", ch)
		}
		if err != nil {
			t.Error(err)
		}
	}

	return got
}
