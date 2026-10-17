package pipeline

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	template "html/template"
)

// TestMust covers the in-process safe path only. The error path calls
// os.Exit(1) and cannot be tested without spawning a subprocess.
func TestMustNoErrors(t *testing.T) {
	pd := newTestPD()
	pd.Must() // must not exit
}

func TestClone(t *testing.T) {
	tmplDir := t.TempDir()
	writeFile(t, tmplDir+"/a.html", "a")
	writeFile(t, tmplDir+"/b.html", "b")
	out := t.TempDir()
	pd := NewHTML(nil).WithLogger(discardLogger()).WithTemplatesDir(tmplDir).WithOutputDir(out)
	pd.AddError(errors.New("seed"))
	pd.SetData("orig", 1)

	clone := pd.Clone()
	if clone == pd {
		t.Fatal("Clone should return a new *PipeData")
	}
	if clone.Template == pd.Template {
		t.Error("expected a cloned template, not the same pointer")
	}
	if clone.logger != pd.logger {
		t.Error("expected logger to be shared")
	}
	if clone.TemplatesDir != tmplDir || clone.OutputDir != out {
		t.Errorf("expected dirs to be copied, got %q %q", clone.TemplatesDir, clone.OutputDir)
	}

	if reflect.ValueOf(clone.Data).Pointer() != reflect.ValueOf(pd.Data).Pointer() {
		t.Error("expected Data to share the same underlying map")
	}
	clone.Data["from_clone"] = 2
	if pd.Data["from_clone"] != 2 {
		t.Error("mutations to clone Data must be visible on the original")
	}

	if len(clone.Errors) != 1 {
		t.Fatalf("expected shared Errors with 1 entry, got %v", clone.Errors)
	}
	pd.Errors[0] = errors.New("replaced")
	if clone.Errors[0].Error() != "replaced" {
		t.Error("expected shared Errors slice backing")
	}

	// Parsing into the clone must not affect the original template.
	clone.LoadGlob("b.html")
	if clone.Template.Lookup("b.html") == nil {
		t.Error("expected b.html to be defined on the clone")
	}
	if pd.Template.Lookup("b.html") != nil {
		t.Error("expected b.html to be absent on the original template")
	}
}

func TestClonePreservesTemplateFuncs(t *testing.T) {
	pd := NewHTML(template.FuncMap{"upper": strings.ToUpper}).WithLogger(discardLogger())
	clone := pd.Clone()
	if got := execFuncProbe(t, clone, "upper"); got != "ABC" {
		t.Errorf("expected funcs preserved on clone, got %q", got)
	}
}

func TestMustWithClones(t *testing.T) {
	pd := newTestPD()
	calls := 0
	ret := pd.MustWithClones([]func(d *PipeData) *PipeData{
		func(d *PipeData) *PipeData {
			calls++
			d.SetData("from_clone", 1) // Data is shared between clone and original
			return d
		},
		func(d *PipeData) *PipeData {
			calls++
			return d
		},
	})
	if ret != pd {
		t.Fatal("MustWithClones should return the original *PipeData")
	}
	if calls != 2 {
		t.Fatalf("expected 2 pipeline invocations, got %d", calls)
	}
	if len(pd.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", pd.Errors)
	}
	if pd.Data["from_clone"] != 1 {
		t.Error("expected Data shared with clones to be updated")
	}
}

func TestToMapStringAny(t *testing.T) {
	t.Run("converts typed map", func(t *testing.T) {
		got := ToMapStringAny(map[string]int{"a": 1, "b": 2})
		want := map[string]any{"a": 1, "b": 2}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, want %#v", got, want)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		got := ToMapStringAny[string](map[string]string{})
		if got == nil || len(got) != 0 {
			t.Errorf("expected non-nil empty map, got %#v", got)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		got := ToMapStringAny[string](nil)
		if got == nil || len(got) != 0 {
			t.Errorf("expected non-nil empty map, got %#v", got)
		}
	})
}
