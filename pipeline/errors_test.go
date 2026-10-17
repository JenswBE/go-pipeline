package pipeline

import (
	"errors"
	"os"
	"testing"
)

func TestFileError(t *testing.T) {
	wrapped := &FileError{action: FileActionOpen, filePath: "/tmp/x.yaml", err: os.ErrNotExist}
	if got := wrapped.Error(); got != "Failed to OPEN file /tmp/x.yaml: file does not exist" {
		t.Errorf("Error(): got %q", got)
	}
	if !errors.Is(wrapped, os.ErrNotExist) {
		t.Error("errors.Is should find wrapped sentinel via Unwrap")
	}

	var nilErr *FileError
	if got := nilErr.Error(); got != "<nil>" {
		t.Errorf("nil receiver Error(): got %q, want %q", got, "<nil>")
	}
	if got := nilErr.Unwrap(); got != nil {
		t.Errorf("nil receiver Unwrap(): got %v, want nil", got)
	}
}

func TestParseHTMLError(t *testing.T) {
	inner := errors.New("unclosed tag")
	wrapped := &ParseHTMLError{err: inner}
	if got := wrapped.Error(); got != "Failed to parse HTML template: unclosed tag" {
		t.Errorf("Error(): got %q", got)
	}
	if !errors.Is(wrapped, inner) {
		t.Error("errors.Is should find wrapped error via Unwrap")
	}
	if !errors.As(wrapped, &wrapped) {
		t.Fatal("errors.As should match *ParseHTMLError")
	}

	var nilErr *ParseHTMLError
	if got := nilErr.Error(); got != "<nil>" {
		t.Errorf("nil receiver Error(): got %q, want %q", got, "<nil>")
	}
	if got := nilErr.Unwrap(); got != nil {
		t.Errorf("nil receiver Unwrap(): got %v, want nil", got)
	}
}

func TestExecuteTemplateError(t *testing.T) {
	inner := errors.New("boom")
	wrapped := &ExecuteTemplateError{templateName: "page.html", err: inner}
	if got := wrapped.Error(); got != "Failed to execute template page.html: boom" {
		t.Errorf("Error(): got %q", got)
	}
	if !errors.Is(wrapped, inner) {
		t.Error("errors.Is should find wrapped error via Unwrap")
	}

	var nilErr *ExecuteTemplateError
	if got := nilErr.Error(); got != "<nil>" {
		t.Errorf("nil receiver Error(): got %q, want %q", got, "<nil>")
	}
	if got := nilErr.Unwrap(); got != nil {
		t.Errorf("nil receiver Unwrap(): got %v, want nil", got)
	}
}
