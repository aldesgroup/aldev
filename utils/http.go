// ----------------------------------------------------------------------------
// The code here is about handling HTTP requests
// ----------------------------------------------------------------------------
package utils

import (
	"encoding/json"
	"io"
	"net/http"
)

// getting JSON from an open URL
func getJSON(url string) []byte {
	Debug("Requesting URL: %s", url)

	// Send HTTP GET request
	response, errGet := http.Get(url)
	FatalIfErr(nil, errGet)
	defer response.Body.Close()

	// Read the response body
	body, errRead := io.ReadAll(response.Body)
	FatalIfErr(nil, errRead)

	return body
}

// getting JSON and unmarshaling it right away into a given object
func jsonAsStruct[T any](url string, obj T) T {
	FatalIfErr(nil, json.Unmarshal(getJSON(url), obj))
	return obj
}

// func structAsJSON(obj any) string {
// 	jsonBytes, errMarshal := json.MarshalIndent(obj, "", "  ")
// 	FatalIfErr(nil, errMarshal)
// 	return string(jsonBytes)
// }
