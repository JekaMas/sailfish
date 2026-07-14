package sailfish

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionHasNoPanicCalls(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		raw, readErr := os.ReadFile(filepath.Clean(name))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if bytes.Contains(raw, []byte("panic(")) {
			t.Errorf("%s contains a production panic call", name)
		}
	}
}
