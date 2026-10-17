package pipeline

import (
	"path/filepath"
	"testing"
)

func TestLoadGlob(t *testing.T) {
	t.Run("parses matching templates", func(t *testing.T) {
		tmplDir := t.TempDir()
		writeFile(t, filepath.Join(tmplDir, "a.html"), "a")
		writeFile(t, filepath.Join(tmplDir, "sub", "b.html"), "b")
		writeFile(t, filepath.Join(tmplDir, "skip.md"), "nope")

		pd := newTestPD().WithTemplatesDir(tmplDir)
		pd.LoadGlob("*.html").LoadGlob("*/*.html")
		if len(pd.Errors) != 0 {
			t.Fatalf("expected no errors, got %v", pd.Errors)
		}
		defines := pd.Template.DefinedTemplates()
		if !contains(defines, "a.html") {
			t.Errorf("expected template a.html to be defined, got %q", defines)
		}
		if !contains(defines, "b.html") {
			t.Errorf("expected template b.html to be defined, got %q", defines)
		}
	})

	t.Run("no matches appends error", func(t *testing.T) {
		tmplDir := t.TempDir()
		writeFile(t, filepath.Join(tmplDir, "a.html"), "a")

		pd := newTestPD().WithTemplatesDir(tmplDir)
		pd.LoadGlob("*.md")
		if len(pd.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(pd.Errors))
		}
		if def := pd.Template.DefinedTemplates(); def != "" {
			t.Errorf("expected no defined templates, got %q", def)
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var pd *PipeData
		if got := pd.LoadGlob("*.html"); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}
