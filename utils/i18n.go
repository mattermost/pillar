// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package utils

import (
	"fmt"
	"github.com/mattermost/mattermost-server/v5/utils/fileutils"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/mattermost/go-i18n/i18n"
)

var T i18n.TranslateFunc
var TDefault i18n.TranslateFunc
var locales = make(map[string]string)

const DEFAULT_LOCALE = "en"

// this functions loads translations from filesystem if they are not
// loaded already and assigns english while loading server config
func TranslationsPreInit() error {
	if T != nil {
		return nil
	}

	// Set T even if we fail to load the translations. Lots of shutdown handling code will
	// segfault trying to handle the error, and the untranslated IDs are strictly better.
	T = TfuncWithFallback(DEFAULT_LOCALE)
	TDefault = TfuncWithFallback("en")
	return InitTranslationsWithDir("i18n")
}

func InitTranslationsWithDir(dir string) error {
	i18nDirectory, found := fileutils.FindDir(dir)
	if !found {
		return fmt.Errorf("Unable to find i18n directory")
	}

	files, _ := ioutil.ReadDir(i18nDirectory)
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			filename := f.Name()
			locales[strings.Split(filename, ".")[0]] = filepath.Join(i18nDirectory, filename)

			if err := i18n.LoadTranslationFile(filepath.Join(i18nDirectory, filename)); err != nil {
				return err
			}
		}
	}

	return nil
}

func GetTranslationsAndLocale(w http.ResponseWriter, r *http.Request) (i18n.TranslateFunc, string) {
	// This is for checking against locales like pt_BR or zn_CN
	headerLocaleFull := strings.Split(r.Header.Get("Accept-Language"), ",")[0]
	// This is for checking against locales like en, es
	headerLocale := strings.Split(strings.Split(r.Header.Get("Accept-Language"), ",")[0], "-")[0]
	defaultLocale := DEFAULT_LOCALE
	if locales[headerLocaleFull] != "" {
		translations := TfuncWithFallback(headerLocaleFull)
		return translations, headerLocaleFull
	} else if locales[headerLocale] != "" {
		translations := TfuncWithFallback(headerLocale)
		return translations, headerLocale
	} else if locales[defaultLocale] != "" {
		translations := TfuncWithFallback(defaultLocale)
		return translations, headerLocale
	}

	translations := TfuncWithFallback(DEFAULT_LOCALE)
	return translations, DEFAULT_LOCALE
}

func TfuncWithFallback(pref string) i18n.TranslateFunc {
	t, _ := i18n.Tfunc(pref)
	return func(translationID string, args ...interface{}) string {
		if translated := t(translationID, args...); translated != translationID {
			return translated
		}

		t, _ := i18n.Tfunc(DEFAULT_LOCALE)
		return t(translationID, args...)
	}
}
