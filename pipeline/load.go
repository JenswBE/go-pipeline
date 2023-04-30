package pipeline

import (
	"github.com/rs/zerolog/log"
)

func (pd *PipeData) LoadGlob(glob string) *PipeData {
	// Check if nil
	if pd == nil {
		return nil
	}

	// Parse templates
	logger := log.With().Str("step", "LoadGlob").Str("glob", glob).Logger()
	tmpl, err := pd.Template.ParseGlob(glob)
	if err != nil {
		logger.Debug().Err(err).Msg("Failed to parse HTML templates")
		return pd.AddError(err)
	}

	// Return result
	logger.Debug().Str("defined_templates", tmpl.DefinedTemplates()).Msg("Templates loaded successfully")
	return pd
}
