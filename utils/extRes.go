// ----------------------------------------------------------------------------
// The code here is about downloading external resources
// ----------------------------------------------------------------------------
package utils

import (
	"fmt"
	"strings"
	"time"
)

// Downloading external resources, like for i18n
func DownloadExternalResources(ctx CancelableContext, cfg *AldevConfig) {
	// making sure we recover any big crashing error
	defer Recover(ctx, "downloading the external resources")

	// downloading the translations
	downloadTranslationsFromGoogle(ctx, cfg)
}

type googleTranslations struct {
	Values [][]string `json:"values"`
}

func downloadTranslationsFromGoogle(ctx CancelableContext, cfg *AldevConfig) {
	if cfg.API.I18n.Link == "" {
		return
	}

	start := time.Now()

	// fetching the google translations
	translations := &googleTranslations{}
	jsonAsStruct(cfg.API.I18n.Link, translations)

	// controlling the header row
	header := translations.Values[0]
	routeCol := 0
	partCol := 1
	keyCol := 2
	if header[keyCol] != "EN" {
		// the english translation is part of the key
		Error("The third column should be the EN translation column")
		ctx.CancelAll()
	}

	// starting the output string
	output := "{"

	// writing out for each language
	lastRow := len(translations.Values) - 1
	lastCol := len(header) - 1
	for colIdx := keyCol; colIdx <= lastCol; colIdx++ {
		// current language
		lang := strings.ToLower(header[colIdx])

		// opening the translation object
		output += fmt.Sprintf("\n  \"%s\": {", lang)
		output += "\n    \"translation\": {\n"

		// iterating over each row
		var row []string
		for rowIdx := 1; rowIdx <= lastRow; rowIdx++ {

			// getting the current row
			row = translations.Values[rowIdx]

			// outputting the current row for the current language
			output += fmt.Sprintf("      \"%s.%s.%s\": \"%s\"",
				row[routeCol], row[partCol], keyFromEnglishTranslation(row[keyCol], cfg.API.I18n.KeySize), // that's the whole translation UID here
				row[colIdx], // and that's the corresponding translation
			)

			// handling the row end
			if rowIdx < lastRow {
				output += ",\n"
			} else {
				output += "\n    }"
			}
		}

		// closing the transaction object
		if colIdx < lastCol {
			output += "\n  },"
		} else {
			output += "\n  }"
		}
	}

	// finishing the output
	output += "\n}"

	// writing out to a file
	WriteToFile(cfg.API.I18n.File, output)
	Info("Done downloading the translation in %s", time.Since(start))
}

// keyFromEnglishTranslation builds a key from the EN translation
func keyFromEnglishTranslation(englishTranslation string, max int) string {
	// If the string is already shorter than the maximum size, return it as is
	// if utf8.RuneCountInString(englishTranslation) <= max {
	if len(englishTranslation) <= max {
		return englishTranslation
	}

	// Finding the place where to cut the string, starting from the end
	var cutIndex int
	for i := max; i > 0; i-- {
		if englishTranslation[i] == ' ' {
			cutIndex = i
			break
		}
	}

	// If no space is found, cut at the maximum size
	if cutIndex == 0 {
		cutIndex = max
	}

	// Return the substring up to the cut index
	return englishTranslation[:cutIndex]
}
