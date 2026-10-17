package pipeline

import (
	"errors"
	"strings"
	"testing"

	template "html/template"
)

func TestNewHTML(t *testing.T) {
	t.Run("without functions", func(t *testing.T) {
		pd := NewHTML(nil)
		if pd == nil {
			t.Fatal("expected non-nil *PipeData")
		}
		if pd.Template == nil {
			t.Fatal("expected non-nil Template")
		}
		if pd.Template.Name() != "empty" {
			t.Errorf("template name: got %q, want %q", pd.Template.Name(), "empty")
		}
		if pd.Data == nil {
			t.Error("expected non-nil Data map")
		}
		if len(pd.Data) != 0 {
			t.Errorf("expected empty Data map, got %v", pd.Data)
		}
		if pd.Errors != nil {
			t.Errorf("expected nil Errors, got %v", pd.Errors)
		}
		if pd.logger == nil {
			t.Error("expected non-nil logger (slog default)")
		}
		if pd.TemplatesDir != "" || pd.DataDir != "" || pd.OutputDir != "" {
			t.Errorf("expected empty dirs, got %q %q %q", pd.TemplatesDir, pd.DataDir, pd.OutputDir)
		}
	})

	t.Run("with functions", func(t *testing.T) {
		pd := NewHTML(template.FuncMap{"upper": strings.ToUpper})
		if got := execFuncProbe(t, pd, "upper"); got != "ABC" {
			t.Errorf("func probe: got %q, want %q", got, "ABC")
		}
	})
}

func TestWithBuilders(t *testing.T) {
	pd := NewHTML(nil)
	got := pd.WithTemplatesDir("tpl").WithDataDir("data").WithOutputDir("out").WithLogger(discardLogger())
	if got != pd {
		t.Fatal("builders should return the same *PipeData")
	}
	if pd.TemplatesDir != "tpl" {
		t.Errorf("TemplatesDir: got %q", pd.TemplatesDir)
	}
	if pd.DataDir != "data" {
		t.Errorf("DataDir: got %q", pd.DataDir)
	}
	if pd.OutputDir != "out" {
		t.Errorf("OutputDir: got %q", pd.OutputDir)
	}
	if pd.logger == nil {
		t.Error("expected logger to be set")
	}
}

func TestAddError(t *testing.T) {
	t.Run("appends errors", func(t *testing.T) {
		pd := NewHTML(nil)
		got := pd.AddError(errors.New("first")).AddError(errors.New("second"))
		if got != pd {
			t.Fatal("AddError should return the same *PipeData")
		}
		if len(pd.Errors) != 2 {
			t.Fatalf("expected 2 errors, got %d", len(pd.Errors))
		}
		if pd.Errors[0].Error() != "first" || pd.Errors[1].Error() != "second" {
			t.Errorf("unexpected errors: %v", pd.Errors)
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var pd *PipeData
		if got := pd.AddError(errors.New("x")); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}

func TestPipeDataRenderFuncs(t *testing.T) {
	tmplDir := t.TempDir()
	writeFile(t, tmplDir+"/t.html", "Hello {{ upper .Data.Name }}")

	pd := NewHTML(template.FuncMap{"upper": strings.ToUpper}).
		WithLogger(discardLogger()).
		WithTemplatesDir(tmplDir).
		WithOutputDir(t.TempDir())

	if got := pd.LoadGlob("t.html"); len(got.Errors) != 0 {
		t.Fatalf("LoadGlob failed: %v", got.Errors)
	}
	pd.SetData("Name", "World")
	pd.RenderSingle("t.html", "out.html")
	if len(pd.Errors) != 0 {
		t.Fatalf("render failed: %v", pd.Errors)
	}
	if got := readOut(t, pd.OutputDir+"/out.html"); got != "Hello WORLD" {
		t.Errorf("rendered: got %q, want %q", got, "Hello WORLD")
	}
}
