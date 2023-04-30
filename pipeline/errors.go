package pipeline

import (
	"errors"
	"fmt"
)

var ErrCreateTargetDirectoryFailed = errors.New("failed to create target directory")
var ErrPipeDataNil = errors.New("PipeData is nil")
var ErrRepeatedDataNotFound = errors.New("repeated data key not found in data, must be map[string]any")
var ErrPlaceholderMissingInTargetPathTemplate = errors.New("{{KEY}} placeholder missing in target path template")

type ParseHTMLError struct {
	err error
}

func (e *ParseHTMLError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Failed to parse HTML template: %v", e.err)
}

func (e *ParseHTMLError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

type FileAction string

const (
	FileActionOpen           FileAction = "OPEN"
	FileActionCreateTruncate FileAction = "CREATE_TRUNCATE"
	FileActionDecodeYAML     FileAction = "DECODE_YAML"
	FileActionClose          FileAction = "CLOSE"
)

type FileError struct {
	action   FileAction
	err      error
	filePath string
}

func (e *FileError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Failed to %s file %s: %v", e.action, e.filePath, e.err)
}

func (e *FileError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

type ExecuteTemplateError struct {
	err          error
	templateName string
}

func (e *ExecuteTemplateError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Failed to execute template %s: %v", e.templateName, e.err)
}

func (e *ExecuteTemplateError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}
