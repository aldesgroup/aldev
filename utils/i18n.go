// ----------------------------------------------------------------------------
// Downloading translations from a Google sheet into a local, structured file
// ----------------------------------------------------------------------------
package utils

import (
	"encoding/json"
	"path"
	"strings"
	"time"
)

type googleTranslations struct {
	Values [][]string `json:"values"`
}

type translation struct {
	Route  string   `json:"r"`
	Part   string   `json:"p"`
	Key    string   `json:"k"`
	Values []string `json:"v"`
}

func downloadTranslationsFromGoogle(ctx CancelableContext, cfg *AldevConfig) {
	// making sure we recover any big crashing error
	defer Recover(ctx, "downloading the external resources")

	if cfg.IsLibrary() {
		return
	}

	if cfg.API.I18n == nil || cfg.API.I18n.Link == "" {
		Error("Empty or incomplete i18n configuration!")
		ctx.CancelAll()
	}

	start := time.Now()

	// fetching the google translations
	gooTranslations := &googleTranslations{}
	jsonAsStruct(cfg.API.I18n.Link, gooTranslations)

	// controlling the header row
	header := gooTranslations.Values[0]
	routeCol := 0
	partCol := 1
	keyCol := 2
	valueColStart := 2
	if header[keyCol] != "EN" {
		// the english translation is part of the key
		Error("The third column should be the EN translation column")
		ctx.CancelAll()
	}

	// getting the outbounds
	lastRow := len(gooTranslations.Values) - 1
	lastCol := len(header) - 1

	// initialising the output object
	output := make([]*translation, lastRow)

	// going through the data and building the output
	var row []string
	for rowIdx := 1; rowIdx <= lastRow; rowIdx++ {
		// getting the current row
		row = gooTranslations.Values[rowIdx]

		// building the translation object
		translation := &translation{
			Route: trim(row[routeCol]),
			Part:  trim(row[partCol]),
			Key:   keyFromEnglishTranslation(trim(row[keyCol]), cfg.API.I18n.KeySize),
		}

		for colIdx := valueColStart; colIdx <= lastCol; colIdx++ {
			translation.Values = append(translation.Values, strings.ToLower(trim(header[colIdx]))+": "+trim(row[colIdx]))
		}

		// adding to the output
		output[rowIdx-1] = translation
	}

	// jsonification
	jsonOutput, errJson := json.MarshalIndent(output, "", "  ")
	FatalIfErr(errJson)

	// writing out to a file
	WriteBytesToFile(path.Join(cfg.API.DataDir, cfg.API.I18n.File), jsonOutput)
	Info("Done downloading the translation in %s", time.Since(start))
}

// shortcut to trim the spaces
func trim(s string) string {
	return strings.TrimSpace(s)
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
