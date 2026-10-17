package pipeline

import (
	"path/filepath"
)

func (pd *PipeData) LoadGlob(glob string) *PipeData {
	// Check if nil
	if pd == nil {
		return nil
	}

	// Parse templates
	templateDirGlob := filepath.Join(pd.TemplatesDir, glob)
	logger := pd.logger.With("step", "LoadGlob", "template_dir_glob", templateDirGlob, "glob", glob)
	tmpl, err := pd.Template.ParseGlob(templateDirGlob)
	if err != nil {
		logger.Debug("Failed to parse HTML templates", "error", err)
		return pd.AddError(err)
	}

	// Return result
	logger.Debug("Templates loaded successfully", "defined_templates", tmpl.DefinedTemplates())
	return pd
}
