// ----------------------------------------------------------------------------
// The code here is about building files from templates, and Aldev config
// ----------------------------------------------------------------------------
package utils

import (
	"fmt"
	"os"
	"text/template"

	core "github.com/aldesgroup/corego"
)

func EnsureFileFromTemplate(filepath, tpl string, params ...any) {
	Debug("Making sure this file exists: %s", filepath)

	content := tpl
	if len(params) > 0 {
		content = fmt.Sprintf(content, params...)
	}

	// Create a new template
	tmpl, errTpl := template.New(filepath).Parse(content)
	core.PanicIfErr(errTpl)

	// Create a new file to write the result
	outputFile, errCreate := os.Create(filepath)
	core.PanicIfErr(errCreate)
	defer outputFile.Close()

	// Execute the template with the data
	core.PanicIfErr(tmpl.Execute(outputFile, Config()))
}
