package pipeline

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var discard = io.Discard

func contains(s, sub string) bool { return strings.Contains(s, sub) }

// execFuncProbe executes a subtemplate calling the given func to verify it
// is registered on pd's template base (html/template has no FuncMap accessor).
func execFuncProbe(t *testing.T, pd *PipeData, fn string) string {
	t.Helper()
	sub := pd.Template.New("probe")
	if _, err := sub.Parse("{{ " + fn + " \"abc\" }}"); err != nil {
		t.Fatalf("parsing probe template: %v", err)
	}
	var b strings.Builder
	if err := sub.Execute(&b, nil); err != nil {
		t.Fatalf("executing probe template: %v", err)
	}
	return b.String()
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func newTestPD() *PipeData {
	return NewHTML(nil).WithLogger(discardLogger())
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

func readOut(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(b)
}
