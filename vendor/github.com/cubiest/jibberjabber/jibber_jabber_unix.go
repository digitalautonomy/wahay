// +build darwin freebsd linux netbsd openbsd

package jibberjabber

import (
	"errors"
	"os"
	"strings"

	"golang.org/x/text/language"
)

func getLangFromEnv() string {
	for _, env := range []string{"LC_MESSAGES", "LC_ALL", "LANG"} {
		locale := os.Getenv(env)
		if len(locale) > 0 {
			return locale
		}
	}
	return ""
}

func getUnixLocale() (string, error) {
	locale := getLangFromEnv()
	if len(locale) <= 0 {
		return "", errors.New(COULD_NOT_DETECT_PACKAGE_ERROR_MESSAGE)
	}
	return locale, nil
}

// DetectIETF detects and returns the IETF language tag of UNIX systems, like Linux and macOS.
// If a territory is defined, the returned value will be in the format of `[language]-[territory]`,
// e.g. `en-GB`.
func DetectIETF() (string, error) {
	locale, err := getUnixLocale()
	if err != nil {
		return "", err
	}

	language, territory := splitLocale(locale)
	locale = language
	if len(territory) > 0 {
		locale = strings.Join([]string{language, territory}, "-")
	}

	return locale, nil
}

// DetectLanguage detects the IETF language tag of UNIX systems, like Linux and macOS,
// and returns the first half of the string, before the `_`.
func DetectLanguage() (string, error) {
	locale, err := getUnixLocale()
	if err != nil {
		return "", err
	}
	language, _ := splitLocale(locale)
	return language, nil
}

// DetectLanguageTag detects the IETF language tag of UNIX systems, like Linux and macOS,
// and returns a fitting language tag.
func DetectLanguageTag() (language.Tag, error) {
	locale, err := getUnixLocale()
	if err != nil {
		return language.Und, err
	}
	return language.Parse(locale)
}

// DetectTerritory detects the IETF language tag of UNIX systems, like Linux and macOS,
// and returns the second half of the string, after the `_`.
func DetectTerritory() (string, error) {
	locale, err := getUnixLocale()
	if err != nil {
		return "", err
	}
	_, territory := splitLocale(locale)
	return territory, nil
}
