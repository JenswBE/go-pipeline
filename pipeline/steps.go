package pipeline

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SetDataYAML loads a YAML file from given path and adds it to pd.Data on given key.
func (pd *PipeData) SetDataYAML(key, path string) *PipeData {
	// Check if nil
	if pd == nil {
		return nil
	}

	// Open file
	dataFilePath := filepath.Join(pd.DataDir, path)
	logger := pd.logger.With("file", dataFilePath, "key", key, "step", "SetDataYAML")
	f, err := os.Open(dataFilePath)
	if err != nil {
		logger.Error("Failed to open YAML file", "error", err)
		pd.AddError(&FileError{action: FileActionOpen, filePath: dataFilePath, err: err})
		return pd
	}
	defer func() {
		if err = f.Close(); err != nil {
			logger.Error("Failed to close YAML file", "error", err)
			pd.AddError(&FileError{action: FileActionClose, filePath: dataFilePath, err: err})
		}
	}()

	// Decode YAML file
	var data any
	if err = yaml.NewDecoder(f).Decode(&data); err != nil {
		logger.Error("Failed to decode YAML file", "error", err)
		pd.AddError(&FileError{action: FileActionDecodeYAML, filePath: dataFilePath, err: err})
		return pd
	}
	pd.SetData(key, data)
	logger.Debug("Data set successfully from YAML file")
	return pd
}

// SetData set provided value to provided data key.
func (pd *PipeData) SetData(key string, value any) *PipeData {
	if pd == nil {
		return nil
	}
	pd.Data[key] = value
	pd.logger.Debug("Data successfully set", "key", key, "step", "SetData")
	return pd
}

// TransformData takes a function which receives pd.Data and should return the modified pd.Data.
// TransformData will panic if the given function returns an error to prevent an incorrectly generated website.
func (pd *PipeData) MustTransformData(transform func(map[string]any) (map[string]any, error)) *PipeData {
	if pd == nil {
		return nil
	}
	var err error
	if pd.Data, err = transform(pd.Data); err != nil {
		pd.logger.Error("Failed to transform data", "error", err)
		os.Exit(1)
		return nil
	}
	pd.logger.Debug("Data successfully transformed")
	return pd
}
