# Jibber Jabber [![Build Status](https://travis-ci.org/cubiest/jibberjabber.svg?branch=master)](https://travis-ci.org/cubiest/jibberjabber)
Jibber Jabber is a GoLang Library that can be used to detect an operating system's current language.

### OS Support

UNIX OS like Linux and macOS via the `LC_MESSAGES`, `LC_ALL` and `LANG` environment variables. They are checked in the aforementioned order.  
These variables are used in ALL versions of UNIX for language detection.

Windows via [GetUserDefaultLocaleName](https://docs.microsoft.com/en-gb/windows/win32/api/winnls/nf-winnls-getuserdefaultlocalename) and [GetSystemDefaultLocaleName](https://docs.microsoft.com/en-gb/windows/win32/api/winnls/nf-winnls-getsystemdefaultlocalename) system calls. These calls are supported in Windows Vista and up.

# Usage
Add the following line to your go `import`:

```golang
	"github.com/cubiest/jibberjabber"
```

### DetectIETF
`DetectIETF` will return the current locale as a string. The format of the locale will be the [ISO 639](http://en.wikipedia.org/wiki/ISO_639) two-letter language code, a DASH, then an [ISO 3166](http://en.wikipedia.org/wiki/ISO_3166-1) two-letter country code.

**Example:**
```golang
	userLocale, err := jibberjabber.DetectIETF()
	println("Locale:", userLocale)
```

### DetectLanguage
`DetectLanguage` will return the current language as a string. The format will be the [ISO 639](http://en.wikipedia.org/wiki/ISO_639) two-letter language code.

also import the following packages for parsing the returned locale
```golang
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"

```

**Example:**
```golang
	userLanguage, err := jibberjabber.DetectLanguage()
	println("Language:", userLanguage)
	languageTag, parseErr := language.Parse(userLanguage)
	println("Language:", display.Self.Name(languageTag))
	
```

### DetectLanguageTag
`DetectLanguageTag` will return the current language as a language tag as specified by `"golang.org/x/text/language"`.

**Example:**
```golang
	languageTag, parseErr := jibberjabber.DetectLanguageTag()
	println("Language:", display.Self.Name(languageTag))
	
```

### DetectTerritory
`DetectTerritory` will return the current locale territory as a string. The format will be the [ISO 3166](http://en.wikipedia.org/wiki/ISO_3166-1) two-letter country code.

**Example:**
```golang
	localeTerritory, err := jibberjabber.DetectTerritory()
	println("Territory:", localeTerritory)
```

### Errors
All the Detect commands will return an error if they are unable to read the Locale from the system.

For Windows, additional error information is provided due to the nature of the system call being used.
