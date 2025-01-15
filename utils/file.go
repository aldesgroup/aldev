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

func EnsureDir(ctx CancelableContext, pathElem ...string) string {
	dirname := path.Join(pathElem...)
	FatalIfErr(ctx, os.MkdirAll(dirname, 0o755))

	return dirname
}

func WriteBytesToFile(ctx CancelableContext, filename string, bytes []byte) {
	// creating the intermediate directory if necessary
	if filename != path.Base(filename) {
		EnsureDir(ctx, path.Dir(filename))
	}

	FatalIfErr(ctx, os.WriteFile(filename, bytes, 0o644))
}

func WriteStringToFile(ctx CancelableContext, filename string, content string, params ...any) {
	WriteBytesToFile(ctx, filename, []byte(fmt.Sprintf(content, params...)))
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

func WriteJsonObjToFile(ctx CancelableContext, filename string, obj any) {
	jsonBytes, errMarshal := json.MarshalIndent(obj, "", "\t")
	FatalIfErr(ctx, errMarshal)
	WriteBytesToFile(ctx, filename, jsonBytes)
}

func ReadFile(ctx CancelableContext, filename string, failIfNotExist bool) []byte {
	if _, fileExists := FileExists(filename); !fileExists {
		if failIfNotExist {
			Fatal(ctx, "File '%s' cannot be found!", filename)
		}
		return nil
	}

	fileBytes, errRead := os.ReadFile(filename)
	FatalIfErr(ctx, errRead)
	return fileBytes
}

func ReadFileToJson[T any, Y *T](ctx CancelableContext, filename string, obj Y, failIfNotExist bool) Y {
	if fileBytes := ReadFile(ctx, filename, failIfNotExist); fileBytes != nil {
		FatalIfErr(ctx, json.Unmarshal(fileBytes, obj))
	}
	return obj
}

func ReplaceInFile(ctx CancelableContext, filename string, replacements map[string]string) {
	fileContent := string(ReadFile(ctx, filename, true))
	for replace, by := range replacements {
		fileContent = strings.ReplaceAll(fileContent, replace, by)
	}
	WriteStringToFile(ctx, filename, "%s", fileContent)
}
