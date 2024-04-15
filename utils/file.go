// ----------------------------------------------------------------------------
// The code here is about file manipulations
// ----------------------------------------------------------------------------
package utils

import (
	"fmt"
	"os"
	"path"
)

func EnsureDir(pathElem ...string) string {
	dirname := path.Join(pathElem...)

	FatalIfErr(os.MkdirAll(dirname, 0o755))

	return dirname
}

func WriteToFileB(filename string, bytes []byte) {
	// creating the intermediate directory if necessary
	if filename != path.Base(filename) {
		EnsureDir(path.Dir(filename))
	}

	FatalIfErr(os.WriteFile(filename, bytes, 0o644))
}

func WriteToFile(filename string, content string, params ...any) {
	WriteToFileB(filename, []byte(fmt.Sprintf(content, params...)))
}
