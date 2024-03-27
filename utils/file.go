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

func WriteToFile(filename string, content string, params ...any) {
	FatalIfErr(os.WriteFile(filename, []byte(fmt.Sprintf(content, params...)), 0o644))
}
