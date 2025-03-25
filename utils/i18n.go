// ----------------------------------------------------------------------------
// Downloading translations from a Google sheet into a local, structured file
// ----------------------------------------------------------------------------
package utils

import (
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/aldesgroup/aldev/templates"
)

// this struct is used to read a Google spreadsheet containing translations
type googleTranslations struct {
	Values [][]string `json:"values"`
}

// some useful types in here
type translationLanguage string
type translationNamespace string
type translationKey string
type translationValue string

// Downloading all the translations, i.e. all the translation files configured, for all the applications that need them
func downloadAllTranslationsFromGoogle(ctx CancelableContext) {
	// making sure we recover any big crashing error
	defer Recover(ctx, "downloading the external resources")

	// no translation for a library
	if IsDevLibrary() {
		return
	}

	// if there's an API - which maybe serves the translation for a web app, we need translations for it
	if IsDevAPI() && Config().API != nil {
		downloadTranslationsForApp(ctx, Config().API.DataDir, "in the API", Config().API.I18n, false)
	}

	// if there's a native app, we need translations for it - and the app should contain them all from the get-go
	if IsDevNative() && Config().Native != nil {
		languages, namespaces := downloadTranslationsForApp(ctx, Config().Native.SrcDir, "in the native app", Config().Native.I18n, true)

		// making the i18n files available in the native app
		generateI18nFile(ctx, languages, namespaces)
	}
}

// Downloading all the translations files configured for a given application (the API, or native app)
func downloadTranslationsForApp(ctx CancelableContext, destDir string, part string, i18nCfg *I18nConfig,
	doReturn bool) (languages []translationLanguage, namespaces []translationNamespace) {
	// a bit of control first
	if i18nCfg == nil || len(i18nCfg.Links) == 0 {
		ErrorAndCancel(ctx, "Empty or incomplete i18n configuration %s!", part)
		return nil, nil
	}

	// tracking the time spent here
	start := time.Now()

	// the destination folder
	destFolder := path.Join(destDir, i18nCfg.Folder)

	// removing the old one, if it exists
	EnsureNoDir(ctx, destFolder)

	// gathering all the translations in a map, allowing overrides
	allTranslations := map[translationLanguage]map[translationNamespace]map[translationKey]translationValue{}

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
		lastCol := len(header) - 1
		lastRow := len(gooTranslations.Values) - 1

		// going through the data and building the output - keeping the source file's order
		var row []string

		// these are the languages required by the Aldev config
		requiredLanguages := MapFn(strings.Split(Config().Languages, ","), func(str string) translationLanguage { return translationLanguage(trim(string(str))) })
		checkedLanguages := []translationLanguage{}

		// adding all the translations, dealing with 1 language at a time
		for colIdx := valueColStart; colIdx <= lastCol; colIdx++ {
			// getting the language
			language := translationLanguage(strings.ToLower(trim(header[colIdx])))

			// only considering the required languages
			if InSlice(requiredLanguages, language) {
				checkedLanguages = append(checkedLanguages, language)

				// init of the translations for this language, if needed
				if allTranslations[language] == nil {
					allTranslations[language] = map[translationNamespace]map[translationKey]translationValue{}
				}

				// dealing with each row now
				for rowIdx := 1; rowIdx <= lastRow; rowIdx++ {
					// getting the current row
					row = gooTranslations.Values[rowIdx]

					// retrieving the namespace & key
					namespace := translationNamespace(trim(row[namespaceCol]))
					key := keyFromColumn(trim(row[keyCol]), i18nCfg.KeySize)

					// inits, if needed
					if allTranslations[language][namespace] == nil {
						allTranslations[language][namespace] = map[translationKey]translationValue{}
					}

					// setting the translation value, and this is where overriding can happen
					allTranslations[language][namespace][key] = translationValue(trim(row[colIdx]))
				}
			}
		}

		// checking that every required language is present
		for _, requiredLanguage := range requiredLanguages {
			if !InSlice(checkedLanguages, requiredLanguage) {
				Fatal(ctx, "Language '%s' is required, but not present in the current translation file", requiredLanguage)
			}
		}
	}

	// Now, we'll generate all the files in parallel to go fast
	waitGroup := new(sync.WaitGroup)

	// We'll generate 1 file per language, per namespace
	for language, translationsForLanguage := range allTranslations {
		for namespace, translationsForNamespace := range translationsForLanguage {
			// one more worker to wait for!
			waitGroup.Add(1)

			// here's a little worker to handler 1 couple (language x namespace)
			go createFile(ctx, waitGroup, destFolder, language, namespace, translationsForNamespace)
		}
	}

	// Let's wait here for everyone to be finished
	waitGroup.Wait()

	// writing out to a file
	Info("Done downloading the translations into '%s' in %s", destFolder, time.Since(start))

	// preparing a return, if required
	if doReturn {
		languages = slices.Collect(maps.Keys(allTranslations))
		namespaces = slices.Collect(maps.Keys(allTranslations[languages[0]]))
	}

	return
}

func createFile(ctx CancelableContext, wg *sync.WaitGroup, destFolder string, lg translationLanguage,
	ns translationNamespace, translations map[translationKey]translationValue) {
	// whatever happens, this worker will end up being DONE
	defer wg.Done()

	// let's get the keys, and sort them
	keys := slices.Collect(maps.Keys(translations))
	slices.Sort(keys)

	// now let's build a list of all the translations we have here
	content := make([]string, len(keys))
	for i, key := range keys {
		content[i] = fmt.Sprintf("\t\"%s\": \"%s\"", key, strings.TrimSpace(string(translations[key])))
	}

	// let's format this a bit
	contentString := fmt.Sprintf("{\n%s\n}", strings.Join(content, ",\n"))

	// let's build the filename
	filename := fmt.Sprintf("%s.json", path.Base(string(ns)))
	filepath := path.Join(destFolder, string(lg), filename)

	// ok for writing it all out
	WriteStringToFile(ctx, filepath, "%s", contentString)

	Debug("Wrote: %s", filepath)
}

// shortcut to trim the spaces
func trim(s string) string {
	return strings.TrimSpace(s)
}

// keyFromColumn builds a key from the EN translation
func keyFromColumn(keyColumnValue string, max int) translationKey {
	// If the string is already shorter than the maximum size, return it as is
	// if utf8.RuneCountInString(keyColumnValue) <= max {
	if len(keyColumnValue) <= max {
		return translationKey(keyColumnValue)
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
	return translationKey(keyColumnValue[:cutIndex])
}

// Creates the i18n-files.ts file in the native app
func generateI18nFile(ctx CancelableContext, languages []translationLanguage, namespaces []translationNamespace) {
	// making sure this will all be sorted
	slices.Sort(languages)
	slices.Sort(namespaces)

	// generating 1 block of requires for each namespace, for each language
	languageBlocks := make([]string, len(languages))
	for i, language := range languages {
		languageBlock := fmt.Sprintf("\t%s: {\n", language)
		for _, ns := range namespaces {
			nsBase := path.Base(string(ns))
			languageBlock += fmt.Sprintf("\t\t'%s': require('../../../locales/%s/%s.json'),\n", nsBase, language, nsBase)
		}
		languageBlock += "\t}"
		languageBlocks[i] = languageBlock
	}

	// outputting
	filepath := path.Join(Config().Native.SrcDir, "src", "components", "utils", "i18n-files.ts")
	EnsureFileFromTemplate(filepath, templates.I18NFiles, strings.Join(languageBlocks, ",\n"))
}
