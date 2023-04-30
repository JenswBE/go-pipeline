package pipeline

import (
	"github.com/rs/zerolog/log"
)

// Must calls log.Fatal in case errors are provided.
func (pd *PipeData) Must() {
	if pd == nil {
		log.Fatal().Msg("Must called, but PipeData is nil")
		return // For correct linter detection of log.Fatal
	}
	if len(pd.Errors) > 0 {
		log.Fatal().Errs("errors", pd.Errors).Msg("Must called, but received errors")
	}
}

// Clone clones the PipeData for reuse of common steps.
// Note: For simplicity, Data and Errors are not cloned and keep the original pointer.
// So, updating Data and Errors in a cloned PipeData will also affect the original one.
func (pd *PipeData) Clone() *PipeData {
	// Clone template
	template, err := pd.Template.Clone()
	if err != nil {
		log.Error().Err(err).Msg("Trying to clone an executed template")
		pd.AddError(err)
	}

	// Clone PipeData
	return &PipeData{
		TemplatesDir: pd.TemplatesDir,
		OutputDir:    pd.OutputDir,
		Extension:    pd.Extension,
		Template:     template,
		Data:         pd.Data,
		Errors:       pd.Errors,
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
