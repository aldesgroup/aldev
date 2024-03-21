// ----------------------------------------------------------------------------
// The code here is about file manipulations
// ----------------------------------------------------------------------------
package utils

import (
	"fmt"
	"os"
)

func EnsureDir(dirname string) {
	FatalIfErr(os.MkdirAll(dirname, 0o755))
}

func WriteToFile(filename string, content string, params ...any) {
	FatalIfErr(os.WriteFile(filename, []byte(fmt.Sprintf(content, params...)), 0o644))
}
