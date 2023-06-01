package pipeline

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func (pd *PipeData) SetDataYAML(key, path string) *PipeData {
	// Check if nil
	if pd == nil {
		return nil
	}

	// Open file
	dataFilePath := filepath.Join(pd.DataDir, path)
	logger := log.With().Str("file", dataFilePath).Str("key", key).Str("step", "SetDataYAML").Logger()
	f, err := os.Open(dataFilePath)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to open YAML file")
		pd.AddError(&FileError{action: FileActionOpen, filePath: dataFilePath, err: err})
		return pd
	}
	defer func() {
		if err = f.Close(); err != nil {
			logger.Error().Err(err).Msg("Failed to close YAML file")
			pd.AddError(&FileError{action: FileActionClose, filePath: dataFilePath, err: err})
		}
	}()

	// Decode YAML file
	var data any
	if err = yaml.NewDecoder(f).Decode(&data); err != nil {
		logger.Error().Err(err).Msg("Failed to decode YAML file")
		pd.AddError(&FileError{action: FileActionDecodeYAML, filePath: dataFilePath, err: err})
		return pd
	}
	pd.SetData(key, data)
	logger.Debug().Msg("Data set successfully from YAML file")
	return pd
}

func (pd *PipeData) SetData(key string, value any) *PipeData {
	if pd == nil {
		return nil
	}
	pd.Data[key] = value
	log.Debug().Str("key", key).Str("step", "SetData").Msg("Data successfully set")
	return pd
}

func (pd *PipeData) SetExtension(extension string) *PipeData {
	if pd == nil {
		return nil
	}
	pd.Extension = extension
	log.Debug().Str("extension", extension).Str("step", "SetExtension").Msg("Extension successfully set")
	return pd
}
