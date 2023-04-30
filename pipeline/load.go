package pipeline

import (
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func (pd *PipeData) LoadGlob(glob string) *PipeData {
	// Check if nil
	if pd == nil {
		return nil
	}

	// Parse templates
	templateDirGlob := filepath.Join(pd.TemplatesDir, glob)
	logger := log.With().Str("step", "LoadGlob").Str("template_dir_glob", templateDirGlob).Str("glob", glob).Logger()
	tmpl, err := pd.Template.ParseGlob(templateDirGlob)
	if err != nil {
		logger.Debug().Err(err).Msg("Failed to parse HTML templates")
		return pd.AddError(err)
	}

	// Return result
	logger.Debug().Str("defined_templates", tmpl.DefinedTemplates()).Msg("Templates loaded successfully")
	return pd
}
