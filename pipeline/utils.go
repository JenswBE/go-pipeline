package pipeline

import (
	"log/slog"
	"os"
)

// Must calls log.Fatal in case errors are provided.
func (pd *PipeData) Must() {
	if pd == nil {
		slog.Error("Must called, but PipeData is nil")
		os.Exit(1)
		return // For correct linter detection of log.Fatal
	}
	if len(pd.Errors) > 0 {
		pd.logger.Error("Must called, but received errors", "errors", pd.Errors)
		os.Exit(1)
	}
}

// Clone clones the PipeData for reuse of common steps.
// Note: For simplicity, Data and Errors are not cloned and keep the original pointer.
// So, updating Data and Errors in a cloned PipeData will also affect the original one.
func (pd *PipeData) Clone() *PipeData {
	// Clone template
	template, err := pd.Template.Clone()
	if err != nil {
		pd.logger.Error("Trying to clone an executed template", "error", err)
		pd.AddError(err)
	}

	// Clone PipeData
	return &PipeData{
		TemplatesDir: pd.TemplatesDir,
		OutputDir:    pd.OutputDir,
		Template:     template,
		Data:         pd.Data,
		Errors:       pd.Errors,
		logger:       pd.logger,
	}
}

// MustWithClones is a convenience wrapper which takes a slice of functions.
// Each function receives a cloned version of the PipeData.
// Must is called on the return value of each function.
// Note: For simplicity, Data and Errors are not cloned and keep the original pointer.
// So, updating Data and Errors in a cloned PipeData will also affect the original one.
func (pd *PipeData) MustWithClones(pipelines []func(pd *PipeData) *PipeData) *PipeData {
	for _, p := range pipelines {
		p(pd.Clone()).Must()
	}
	return pd
}

func ToMapStringAny[T any](input map[string]T) map[string]any {
	output := make(map[string]any, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}
