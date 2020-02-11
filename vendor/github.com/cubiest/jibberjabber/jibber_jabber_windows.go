// +build windows

package jibberjabber

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/text/language"
)

// TODO: read LOCALE_NAME_MAX_LENGTH from Windows, instead of hard-coding '85'
const LOCALE_NAME_MAX_LENGTH uint32 = 85

func getWindowsLocaleFrom(sysCall string) (string, error) {
	buffer := make([]uint16, LOCALE_NAME_MAX_LENGTH)

	dll, err := windows.LoadDLL("kernel32")
	if err != nil {
		return "", errors.New("could not find kernel32 dll: " + err.Error())
	}

	proc, err := dll.FindProc(sysCall)
	if err != nil {
		return "", err
	}

	r, _, dllError := proc.Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(LOCALE_NAME_MAX_LENGTH))
	if r == 0 {
		return "", errors.New(COULD_NOT_DETECT_PACKAGE_ERROR_MESSAGE + ":\n" + dllError.Error())
	}

	return windows.UTF16ToString(buffer), nil
}

func getWindowsLocale() (string, error) {
	dll, err := windows.LoadDLL("kernel32")
	if err != nil {
		return "", errors.New("could not find kernel32 dll: " + err.Error())
	}

	locale, err := getWindowsLocaleFrom("GetUserDefaultLocaleName")
	if err != nil {
		locale, err = getWindowsLocaleFrom("GetSystemDefaultLocaleName")
	}
	return locale, err
}

// DetectIETF detects and returns the IETF language tag of Windows.
func DetectIETF() (string, error) {
	return getWindowsLocale()
}

// DetectLanguage detects the IETF language tag of Windows
// and returns the first half of the string, before the `_`.
func DetectLanguage() (string, error) {
	locale, err := getWindowsLocale()
	if err != nil {
		return "", err
	}
	language, _ := splitLocale(locale)
	return language, nil
}

// DetectLanguageTag detects the IETF language tag of Windows
// and returns a fitting language tag.
func DetectLanguageTag() (language.Tag, error) {
	locale, err := getWindowsLocale()
	if err != nil {
		return language.Und, err
	}
	return language.Parse(locale)
}

// DetectTerritory detects the IETF language tag of Windows
// and returns the second half of the string, after the `_`.
func DetectTerritory() (string, error) {
	locale, err := getWindowsLocale()
	if err != nil {
		return "", nil
	}
	_, territory := splitLocale(locale)
	return territory, nil
}
