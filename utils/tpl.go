// ----------------------------------------------------------------------------
// The code here is about building files from templates, and Aldev config
// ----------------------------------------------------------------------------
package utils

import (
	"os"
	"text/template"
)

func EnsureFileFromTemplate(cfg *aldevConfig, filepath, tpl string) {
	Debug("Making sure this file exists: %s", filepath)

	// Create a new template
	tmpl, errTpl := template.New(filepath).Parse(tpl)
	FatalIfErr(errTpl)

	// Create a new file to write the result
	outputFile, errCreate := os.Create(filepath)
	FatalIfErr(errCreate)
	defer outputFile.Close()

	// Execute the template with the data
	FatalIfErr(tmpl.Execute(outputFile, cfg))
}
