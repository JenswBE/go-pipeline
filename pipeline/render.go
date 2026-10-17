package pipeline

import (
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func (pd *PipeData) render(logger *slog.Logger, templateName, targetPath string, data map[string]any) *PipeData {
	// Check if PipeData is set
	if pd == nil {
		return nil
	}

	// Ensure target folder exists
	targetFolder, _ := filepath.Split(targetPath)
	if err := os.MkdirAll(targetFolder, 0o755); err != nil {
		logger.Error("Failed to create target directory", "error", err)
		return pd.AddError(ErrCreateTargetDirectoryFailed)
	}

	// Open target file
	f, err := os.Create(targetPath)
	if err != nil {
		logger.Error("Failed to create/truncate target file", "error", err)
		return pd.AddError(&FileError{action: FileActionCreateTruncate, filePath: targetPath, err: err})
	}
	defer func() {
		if fileError := f.Close(); fileError != nil {
			logger.Error("Failed to close target file", "error", fileError)
			pd.AddError(&FileError{action: FileActionClose, filePath: targetPath, err: fileError})
		}
	}()

	// Execute template
	if err = pd.Template.ExecuteTemplate(f, templateName, data); err != nil {
		logger.Error("Failed to render template", "error", err, "data", pd.Data)
		return pd.AddError(&ExecuteTemplateError{templateName: templateName, err: err})
	}
	logger.Debug("Single template successfully rendered")
	return pd
}

func (pd *PipeData) RenderSingle(templateName, targetPath string) *PipeData {
	outputPath := filepath.Join(pd.OutputDir, targetPath)
	logger := pd.logger.With("template_name", templateName, "target_path", targetPath, "output_path", outputPath, "step", "RenderSingle")
	return pd.render(logger, templateName, outputPath, map[string]any{"Data": pd.Data})
}

// LoadRenderSingle clones the PipeData, loads specified template and executes it.
// Original PipeData is returned. Make sure to check Clone() on effects of cloning!
func (pd *PipeData) LoadRenderSingle(templatePath, targetPath string) *PipeData {
	// Check if PipeData is set
	if pd == nil {
		return nil
	}

	cloned := pd.Clone()
	_, file := path.Split(templatePath)
	cloned.LoadGlob(templatePath).RenderSingle(file, targetPath)
	return pd
}

// RenderRepeated renders multiple pages.
// "repeatedDataKey" should point to a value of type map[string]any.
// "targetPathTemplate" should contain "{{KEY}}" which is replaced by the key of provided map.
func (pd *PipeData) RenderRepeated(templateName, repeatedDataKey, targetPathTemplate string) *PipeData {
	// Check if PipeData is set
	if pd == nil {
		return nil
	}

	// Setup logger
	outputPath := filepath.Join(pd.OutputDir, targetPathTemplate)
	logger := pd.logger.With("template_name", templateName, "repeated_data_key", repeatedDataKey, "target_path_template", targetPathTemplate, "output_path_template", outputPath, "step", "RenderRepeated")

	// Check if repeatedDataKey points to map[string]any
	if pd.Data == nil {
		logger.Error("RenderRepeated requires Data to be set")
		return pd.AddError(ErrRepeatedDataNotFound)
	}
	rawRepeatedData, ok := pd.Data[repeatedDataKey]
	if !ok {
		logger.Error("Repeated data key not found in data", "data", pd.Data)
		return pd.AddError(ErrRepeatedDataNotFound)
	}
	repeatedData, ok := rawRepeatedData.(map[string]any)
	if !ok {
		logger.Error("Repeated data key points to type %T, but type map[string]any is required.", "type", rawRepeatedData)
		return pd.AddError(ErrRepeatedDataNotFound)
	}

	// Check if targetPathTemplate contains placeholder
	if !strings.Contains(outputPath, "{{KEY}}") {
		logger.Error("{{KEY}} placeholder missing in target path template")
		return pd.AddError(ErrPlaceholderMissingInTargetPathTemplate)
	}

	// Render pages
	for k, v := range repeatedData {
		targetPath := strings.ReplaceAll(outputPath, "{{KEY}}", k)
		logger = logger.With("target_path", targetPath)
		pd.render(logger, templateName, targetPath, map[string]any{
			"Data":          pd.Data,
			"RepeatedKey":   k,
			"RepeatedValue": v,
		})
	}
	return pd
}

// LoadRenderRepeated clones the PipeData, loads specified template and executes it.
// Original PipeData is returned. Make sure to check Clone() on effects of cloning!
func (pd *PipeData) LoadRenderRepeated(templatePath, repeatedDataKey, targetPathTemplate string) *PipeData {
	// Check if PipeData is set
	if pd == nil {
		return nil
	}

	cloned := pd.Clone()
	_, file := path.Split(templatePath)
	cloned.LoadGlob(templatePath).RenderRepeated(file, repeatedDataKey, targetPathTemplate)
	return pd
}
