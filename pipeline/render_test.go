package pipeline

import (
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadTemplate(t *testing.T, name, content string) (pd *PipeData, out string) {
	t.Helper()
	tmplDir := t.TempDir()
	writeFile(t, filepath.Join(tmplDir, name), content)
	out = t.TempDir()
	pd = newTestPD().WithTemplatesDir(tmplDir).WithOutputDir(out)
	pd.LoadGlob(name)
	if len(pd.Errors) != 0 {
		t.Fatalf("LoadGlob failed: %v", pd.Errors)
	}
	return pd, out
}

func TestRenderSingle(t *testing.T) {
	t.Run("renders to nested path", func(t *testing.T) {
		pd, out := loadTemplate(t, "t.html", "Hello {{ .Data.Name }}")
		pd.SetData("Name", "World")
		got := pd.RenderSingle("t.html", "deep/nested/out.html")
		if got != pd {
			t.Fatal("RenderSingle should return the same *PipeData")
		}
		if len(pd.Errors) != 0 {
			t.Fatalf("expected no errors, got %v", pd.Errors)
		}
		if got := readOut(t, filepath.Join(out, "deep/nested/out.html")); got != "Hello World" {
			t.Errorf("rendered: got %q, want %q", got, "Hello World")
		}
	})

	t.Run("undefined template name", func(t *testing.T) {
		pd, _ := loadTemplate(t, "t.html", "x")
		pd.RenderSingle("missing.html", "out.html")
		if len(pd.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(pd.Errors), pd.Errors)
		}
		var e *ExecuteTemplateError
		if !errors.As(pd.Errors[0], &e) {
			t.Fatalf("expected *ExecuteTemplateError, got %T", pd.Errors[0])
		}
		if e.templateName != "missing.html" {
			t.Errorf("templateName: got %q", e.templateName)
		}
	})

	t.Run("template function error", func(t *testing.T) {
		tmplDir := t.TempDir()
		writeFile(t, filepath.Join(tmplDir, "t.html"), "{{ boom }}")
		out := t.TempDir()
		pd := NewHTML(template.FuncMap{"boom": func() (string, error) { return "", errors.New("boom!") }}).
			WithLogger(discardLogger()).
			WithTemplatesDir(tmplDir).
			WithOutputDir(out)
		pd.LoadGlob("t.html")
		pd.RenderSingle("t.html", "out.html")
		if len(pd.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(pd.Errors), pd.Errors)
		}
		var e *ExecuteTemplateError
		if !errors.As(pd.Errors[0], &e) {
			t.Fatalf("expected *ExecuteTemplateError, got %T", pd.Errors[0])
		}
		if !errors.Is(e, errors.New("boom!")) && !strings.Contains(e.Error(), "boom!") {
			t.Errorf("expected wrapped fn error, got %q", e.Error())
		}
	})

	t.Run("target directory blocked by file", func(t *testing.T) {
		pd, out := loadTemplate(t, "t.html", "x")
		writeFile(t, filepath.Join(out, "block"), "not a dir")
		pd.RenderSingle("t.html", "block/inner.html")
		if len(pd.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(pd.Errors), pd.Errors)
		}
		if !errors.Is(pd.Errors[0], ErrCreateTargetDirectoryFailed) {
			t.Errorf("expected ErrCreateTargetDirectoryFailed, got %v", pd.Errors[0])
		}
	})

	t.Run("target path is existing directory", func(t *testing.T) {
		pd, out := loadTemplate(t, "t.html", "x")
		if err := os.MkdirAll(filepath.Join(out, "isdir"), 0o755); err != nil {
			t.Fatal(err)
		}
		pd.RenderSingle("t.html", "isdir")
		if len(pd.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(pd.Errors), pd.Errors)
		}
		var fe *FileError
		if !errors.As(pd.Errors[0], &fe) {
			t.Fatalf("expected *FileError, got %T", pd.Errors[0])
		}
		if fe.action != FileActionCreateTruncate {
			t.Errorf("expected action CREATE_TRUNCATE, got %q", fe.action)
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		// RenderSingle has no nil guard; the guard lives in the
		// unprefixed render helper, so test that directly.
		var pd *PipeData
		if got := pd.render(nil, "t.html", "out.html", map[string]any{}); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}

func TestLoadRenderSingle(t *testing.T) {
	tmplDir := t.TempDir()
	writeFile(t, filepath.Join(tmplDir, "base.html"), "<h1>base</h1>")
	writeFile(t, filepath.Join(tmplDir, "extra.html"), "<h1>extra {{ .Data.Name }}</h1>")
	out := t.TempDir()
	pd := newTestPD().WithTemplatesDir(tmplDir).WithOutputDir(out)
	pd.SetData("Name", "World")

	t.Run("renders on clone and returns original", func(t *testing.T) {
		ret := pd.LoadRenderSingle("extra.html", "extraout.html")
		if ret != pd {
			t.Fatal("LoadRenderSingle should return the original *PipeData")
		}
		if len(pd.Errors) != 0 {
			t.Fatalf("expected no errors, got %v", pd.Errors)
		}
		if got := readOut(t, filepath.Join(out, "extraout.html")); got != "<h1>extra World</h1>" {
			t.Errorf("rendered: got %q", got)
		}
		if pd.Template.Lookup("extra.html") != nil {
			t.Error("original template must not be polluted by clone's LoadGlob")
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var pd *PipeData
		if got := pd.LoadRenderSingle("x.html", "y.html"); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}

func TestRenderRepeated(t *testing.T) {
	t.Run("renders one file per entry incl. nested paths", func(t *testing.T) {
		pd, out := loadTemplate(t, "page.html", "{{ .RepeatedKey }}={{ .RepeatedValue }}")
		pd.SetData("pages", map[string]any{"a": "A", "b": "B"})
		got := pd.RenderRepeated("page.html", "pages", "pages/{{KEY}}/{{KEY}}.html")
		if got != pd {
			t.Fatal("RenderRepeated should return the same *PipeData")
		}
		if len(pd.Errors) != 0 {
			t.Fatalf("expected no errors, got %v", pd.Errors)
		}
		if got := readOut(t, filepath.Join(out, "pages/a/a.html")); got != "a=A" {
			t.Errorf("a.html: got %q, want %q", got, "a=A")
		}
		if got := readOut(t, filepath.Join(out, "pages/b/b.html")); got != "b=B" {
			t.Errorf("b.html: got %q, want %q", got, "b=B")
		}
	})

	t.Run("Data map is nil", func(t *testing.T) {
		pd, _ := loadTemplate(t, "page.html", "x")
		pd.Data = nil
		pd.RenderRepeated("page.html", "pages", "pages/{{KEY}}.html")
		if len(pd.Errors) != 1 || !errors.Is(pd.Errors[0], ErrRepeatedDataNotFound) {
			t.Fatalf("expected ErrRepeatedDataNotFound, got %v", pd.Errors)
		}
	})

	t.Run("repeated key not found", func(t *testing.T) {
		pd, _ := loadTemplate(t, "page.html", "x")
		pd.Data = map[string]any{}
		pd.RenderRepeated("page.html", "pages", "pages/{{KEY}}.html")
		if len(pd.Errors) != 1 || !errors.Is(pd.Errors[0], ErrRepeatedDataNotFound) {
			t.Fatalf("expected ErrRepeatedDataNotFound, got %v", pd.Errors)
		}
	})

	t.Run("repeated value has wrong type", func(t *testing.T) {
		pd, _ := loadTemplate(t, "page.html", "x")
		pd.SetData("pages", "not a map")
		pd.RenderRepeated("page.html", "pages", "pages/{{KEY}}.html")
		if len(pd.Errors) != 1 || !errors.Is(pd.Errors[0], ErrRepeatedDataNotFound) {
			t.Fatalf("expected ErrRepeatedDataNotFound, got %v", pd.Errors)
		}
	})

	t.Run("target template missing KEY placeholder", func(t *testing.T) {
		pd, _ := loadTemplate(t, "page.html", "x")
		pd.SetData("pages", map[string]any{"a": "A"})
		pd.RenderRepeated("page.html", "pages", "static.html")
		if len(pd.Errors) != 1 || !errors.Is(pd.Errors[0], ErrPlaceholderMissingInTargetPathTemplate) {
			t.Fatalf("expected ErrPlaceholderMissingInTargetPathTemplate, got %v", pd.Errors)
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var pd *PipeData
		if got := pd.RenderRepeated("page.html", "pages", "p/{{KEY}}.html"); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}

func TestLoadRenderRepeated(t *testing.T) {
	t.Run("renders on clone and returns original", func(t *testing.T) {
		tmplDir := t.TempDir()
		writeFile(t, filepath.Join(tmplDir, "page.html"), "{{ .RepeatedKey }}")
		out := t.TempDir()
		pd := newTestPD().WithTemplatesDir(tmplDir).WithOutputDir(out)
		pd.SetData("pages", map[string]any{"a": "A"})

		ret := pd.LoadRenderRepeated("page.html", "pages", "pages/{{KEY}}.html")
		if ret != pd {
			t.Fatal("LoadRenderRepeated should return the original *PipeData")
		}
		if len(pd.Errors) != 0 {
			t.Fatalf("expected no errors, got %v", pd.Errors)
		}
		if got := readOut(t, filepath.Join(out, "pages/a.html")); got != "a" {
			t.Errorf("rendered: got %q, want %q", got, "a")
		}
		if pd.Template.Lookup("page.html") != nil {
			t.Error("original template must not be polluted by clone's LoadGlob")
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var pd *PipeData
		if got := pd.LoadRenderRepeated("p.html", "k", "x/{{KEY}}.html"); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}
