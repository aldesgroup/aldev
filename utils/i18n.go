// ----------------------------------------------------------------------------
// Downloading translations from a Google sheet into a local, structured file
// ----------------------------------------------------------------------------
package utils

import (
	"encoding/json"
	"path"
	"sort"
	"strings"
	"time"
)

// this struct is used to read a Google spreadsheet containing translations
type googleTranslations struct {
	Values [][]string `json:"values"`
}

// this is how we keep in a JSON file the translations for 1 namespace/key couple
type translation struct {
	Namespace string   `json:"n"`
	Key       string   `json:"k"`
	Values    []string `json:"v"` // each value has the form: "lg: bla bla bla", lg being the language code (fr, en, de, etc.)

	// technical properties
	order int
}

// Downloading all the translations, i.e. all the translation files configured, for all the applications that need them
func downloadAllTranslationsFromGoogle(ctx CancelableContext) {
	// making sure we recover any big crashing error
	defer Recover(ctx, "downloading the external resources")

	// no translation for a library
	if IsDevLibrary() {
		return
	}

	// if there's an API, we need translations for it
	if IsDevAPI() && Config().API != nil {
		downloadTranslationForApp(ctx, Config().API.I18n, Config().API.DataDir)
	} //

	// if there's a native app, we need translations for it
	if IsDevNative() && Config().Native != nil {
		downloadTranslationForApp(ctx, Config().Native.I18n, Config().Native.DataDir)
	}
}

// Downloading all the translations files configured for a given application (the API, or native app)
func downloadTranslationForApp(ctx CancelableContext, i18nCfg *I18nConfig, destDir string) {
	if i18nCfg == nil || len(i18nCfg.Links) == 0 {
		Error("Empty or incomplete i18n configuration!")
		ctx.CancelAll()
		return
	}

	// tracking the time spent here
	start := time.Now()

	// gathering all the translations in a map, allowing overrides
	translations := map[string]*translation{}
	translationCount := 0

	// treating each google spreadsheet that we must download
	for _, link := range i18nCfg.Links {
		// fetching the google translations
		gooTranslations := jsonAsStruct(link, &googleTranslations{})

		// controlling the header row
		header := gooTranslations.Values[0]
		namespaceCol := 0
		keyCol := 1
		// hint := 2 // not using it here
		valueColStart := 3

		// getting the outbounds
		lastRow := len(gooTranslations.Values) - 1
		lastCol := len(header) - 1

		// going through the data and building the output - keeping the source file's order
		var row []string
		for rowIdx := 1; rowIdx <= lastRow; rowIdx++ {
			// getting the current row
			row = gooTranslations.Values[rowIdx]

			// building the translation object
			translationCount++
			translation := &translation{
				Namespace: trim(row[namespaceCol]),
				Key:       keyFromColumn(trim(row[keyCol]), i18nCfg.KeySize),
				order:     translationCount,
			}

			for colIdx := valueColStart; colIdx <= lastCol; colIdx++ {
				translation.Values = append(translation.Values, strings.ToLower(trim(header[colIdx]))+": "+trim(row[colIdx]))
			}

			// adding to the output
			translations[translation.Namespace+translation.Key] = translation
		}
	}

	// initialising the output object
	output := []*translation{}

	// pouring the translations we've gathered so far into the output
	for _, translationObj := range translations {
		output = append(output, translationObj)
	}

	// sorting the output
	sort.Slice(output, func(i, j int) bool {
		return output[i].order < output[j].order
	})

	// jsonification
	jsonOutput, errJson := json.MarshalIndent(output, "", "  ")
	FatalIfErr(ctx, errJson)

	// writing out to a file
	filename := path.Join(destDir, i18nCfg.File)
	WriteBytesToFile(ctx, filename, jsonOutput)
	Info("Done downloading the translations into '%s' in %s", filename, time.Since(start))
}

// shortcut to trim the spaces
func trim(s string) string {
	return strings.TrimSpace(s)
}

// keyFromColumn builds a key from the EN translation
func keyFromColumn(keyColumnValue string, max int) string {
	// If the string is already shorter than the maximum size, return it as is
	// if utf8.RuneCountInString(keyColumnValue) <= max {
	if len(keyColumnValue) <= max {
		return keyColumnValue
	}

	// Finding the place where to cut the string, starting from the end
	var cutIndex int
	for i := max; i > 0; i-- {
		if keyColumnValue[i] == ' ' {
			cutIndex = i
			break
		}
	}

	// If no space is found, cut at the maximum size
	if cutIndex == 0 {
		cutIndex = max
	}

	// Return the substring up to the cut index
	return keyColumnValue[:cutIndex]
}
