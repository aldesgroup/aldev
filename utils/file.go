// ----------------------------------------------------------------------------
// The code here is about file manipulations
// ----------------------------------------------------------------------------
package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

func EnsureDir(pathElem ...string) string {
	dirname := path.Join(pathElem...)
	FatalIfErr(os.MkdirAll(dirname, 0o755))

	return dirname
}

func WriteBytesToFile(filename string, bytes []byte) {
	// creating the intermediate directory if necessary
	if filename != path.Base(filename) {
		EnsureDir(path.Dir(filename))
	}

	FatalIfErr(os.WriteFile(filename, bytes, 0o644))
}

func WriteStringToFile(filename string, content string, params ...any) {
	WriteBytesToFile(filename, []byte(fmt.Sprintf(content, params...)))
}

func DirExists(dirPaths ...string) (string, bool) {
	fullpath := path.Join(dirPaths...)
	info, err := os.Stat(fullpath)
	return fullpath, !os.IsNotExist(err) && info.IsDir()
}

func FileExists(dirPaths ...string) (string, bool) {
	fullpath := path.Join(dirPaths...)
	info, err := os.Stat(fullpath)
	return fullpath, !os.IsNotExist(err) && !info.IsDir()
}

func WriteJsonObjToFile(filename string, obj any) {
	jsonBytes, errMarshal := json.MarshalIndent(obj, "", "\t")
	FatalIfErr(errMarshal)
	WriteBytesToFile(filename, jsonBytes)
}

func ReadFile(filename string, failIfNotExist bool) []byte {
	if _, fileExists := FileExists(filename); !fileExists {
		if failIfNotExist {
			Fatal("File '%s' cannot be found!", filename)
		}
		return nil
	}

	fileBytes, errRead := os.ReadFile(filename)
	FatalIfErr(errRead)
	return fileBytes
}

func ReadFileToJson[T any, Y *T](filename string, obj Y, failIfNotExist bool) Y {
	if fileBytes := ReadFile(filename, failIfNotExist); fileBytes != nil {
		FatalIfErr(json.Unmarshal(fileBytes, obj))
	}
	return nil
}

func ReplaceInFile(filename string, replacements map[string]string) {
	fileContent := string(ReadFile(filename, true))
	for replace, by := range replacements {
		fileContent = strings.ReplaceAll(fileContent, replace, by)
	}
	WriteStringToFile(filename, fileContent)
}
