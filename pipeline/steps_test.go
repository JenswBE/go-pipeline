package pipeline

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDataOps(t *testing.T) {
	t.Run("SetData stores value", func(t *testing.T) {
		pd := newTestPD()
		if got := pd.SetData("key", 42); got != pd {
			t.Fatal("SetData should return the same *PipeData")
		}
		if pd.Data["key"] != 42 {
			t.Errorf("expected Data[key]=42, got %v", pd.Data["key"])
		}
	})

	t.Run("SetData nil receiver", func(t *testing.T) {
		var pd *PipeData
		if got := pd.SetData("k", 1); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})

	t.Run("MustTransformData mutates map in place", func(t *testing.T) {
		pd := newTestPD().SetData("a", 1)
		got := pd.MustTransformData(func(d map[string]any) (map[string]any, error) {
			d["b"] = 2
			return d, nil
		})
		if got != pd {
			t.Fatal("MustTransformData should return the same *PipeData")
		}
		if pd.Data["b"] != 2 {
			t.Errorf("expected Data[b]=2, got %v", pd.Data["b"])
		}
	})

	t.Run("MustTransformData replaces map", func(t *testing.T) {
		pd := newTestPD().SetData("a", 1)
		pd.MustTransformData(func(map[string]any) (map[string]any, error) {
			return map[string]any{"fresh": true}, nil
		})
		if !reflect.DeepEqual(pd.Data, map[string]any{"fresh": true}) {
			t.Errorf("expected replaced Data, got %v", pd.Data)
		}
	})

	t.Run("MustTransformData nil receiver", func(t *testing.T) {
		var pd *PipeData
		if got := pd.MustTransformData(func(d map[string]any) (map[string]any, error) { return d, nil }); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}

func TestSetDataYAML(t *testing.T) {
	t.Run("decodes yaml file into data", func(t *testing.T) {
		dataDir := t.TempDir()
		writeFile(t, filepath.Join(dataDir, "site.yaml"), "title: My Site\nitems:\n  - one\n  - two\n")

		pd := newTestPD().WithDataDir(dataDir)
		pd.SetDataYAML("site", "site.yaml")
		if len(pd.Errors) != 0 {
			t.Fatalf("expected no errors, got %v", pd.Errors)
		}
		want := map[string]any{
			"title": "My Site",
			"items": []any{"one", "two"},
		}
		if !reflect.DeepEqual(pd.Data["site"], want) {
			t.Errorf("Data[site]: got %#v, want %#v", pd.Data["site"], want)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		pd := newTestPD().WithDataDir(t.TempDir())
		pd.SetDataYAML("site", "missing.yaml")
		if len(pd.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(pd.Errors), pd.Errors)
		}
		var fe *FileError
		if !errors.As(pd.Errors[0], &fe) {
			t.Fatalf("expected *FileError, got %T", pd.Errors[0])
		}
		if fe.action != FileActionOpen {
			t.Errorf("expected action OPEN, got %q", fe.action)
		}
		if !contains(fe.Error(), filepath.Join(pd.DataDir, "missing.yaml")) {
			t.Errorf("error should mention data file path: %q", fe.Error())
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		dataDir := t.TempDir()
		writeFile(t, filepath.Join(dataDir, "broken.yaml"), "[unclosed")

		pd := newTestPD().WithDataDir(dataDir)
		pd.SetDataYAML("site", "broken.yaml")
		if len(pd.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(pd.Errors), pd.Errors)
		}
		var fe *FileError
		if !errors.As(pd.Errors[0], &fe) {
			t.Fatalf("expected *FileError, got %T", pd.Errors[0])
		}
		if fe.action != FileActionDecodeYAML {
			t.Errorf("expected action DECODE_YAML, got %q", fe.action)
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var pd *PipeData
		if got := pd.SetDataYAML("k", "f.yaml"); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}
