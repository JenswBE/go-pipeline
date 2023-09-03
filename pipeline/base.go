package pipeline

import (
	template "html/template"
)

type PipeData struct {
	TemplatesDir string
	DataDir      string
	OutputDir    string
	Template     *template.Template
	Data         map[string]any
	Errors       []error
}

// New creates a new PipeData object with initial values.
// Functions are taken as an argument as these should be populated before you parse any template.
// See https://pkg.go.dev/text/template#Template.Funcs for more info.
func NewHTML(functions template.FuncMap) *PipeData {
	tmpl := template.New("empty")
	if len(functions) > 0 {
		tmpl.Funcs(functions)
	}
	return &PipeData{
		TemplatesDir: "",
		DataDir:      "",
		OutputDir:    "",
		Template:     tmpl,
		Data:         map[string]any{},
		Errors:       nil,
	}
}

// WithTemplatesDir sets the default templates directory.
func (pd *PipeData) WithTemplatesDir(templatesDir string) *PipeData {
	pd.TemplatesDir = templatesDir
	return pd
}

// WithDataDir sets the default data directory.
func (pd *PipeData) WithDataDir(dataDir string) *PipeData {
	pd.DataDir = dataDir
	return pd
}

// WithOutputDir sets the output data directory.
func (pd *PipeData) WithOutputDir(outputDir string) *PipeData {
	pd.OutputDir = outputDir
	return pd
}

// AddError appends an error to the data.
func (pd *PipeData) AddError(err error) *PipeData {
	if pd == nil {
		return nil
	}
	pd.Errors = append(pd.Errors, err)
	return pd
}
