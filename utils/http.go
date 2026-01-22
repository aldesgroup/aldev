// ----------------------------------------------------------------------------
// The code here is about handling HTTP requests
// ----------------------------------------------------------------------------
package utils

import (
	"encoding/json"
	"io"
	"net/http"

	core "github.com/aldesgroup/corego"
)

// getting JSON from an open URL
func getJSON(url string) []byte {
	Debug("Requesting URL: %s", url)

	// Send HTTP GET request
	response, errGet := http.Get(url)
	core.PanicIfErr(errGet)
	defer response.Body.Close()

	// Read the response body
	body, errRead := io.ReadAll(response.Body)
	core.PanicIfErr(errRead)

	return body
}

// getting JSON and unmarshaling it right away into a given object
func jsonAsStruct[T any](url string, obj T) T {
	core.PanicIfErr(json.Unmarshal(getJSON(url), obj))
	return obj
}
