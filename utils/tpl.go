// ----------------------------------------------------------------------------
// The code here is about building files from templates, and Aldev config
// ----------------------------------------------------------------------------
package utils

import (
	"fmt"
	"os"
	"text/template"
)

func EnsureFileFromTemplate(cfg *AldevConfig, filepath, tpl string, params ...any) {
	Debug("Making sure this file exists: %s", filepath)

	content := tpl
	if len(params) > 0 {
		content = fmt.Sprintf(content, params...)
	}

	// Create a new template
	tmpl, errTpl := template.New(filepath).Parse(content)
	FatalIfErr(errTpl)

	// Create a new file to write the result
	outputFile, errCreate := os.Create(filepath)
	FatalIfErr(errCreate)
	defer outputFile.Close()

	// Execute the template with the data
	FatalIfErr(tmpl.Execute(outputFile, cfg))
}
