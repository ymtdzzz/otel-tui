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
		if n%w == w-1 {
			_, err = fmt.Fprintf(&got, "%c\n", v.Runes[0])
		} else {
			_, err = fmt.Fprintf(&got, "%c", v.Runes[0])
		}
		if err != nil {
			t.Error(err)
		}
	}

	return got
}
