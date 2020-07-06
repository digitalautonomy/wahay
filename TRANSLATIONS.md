# Translations

New translations in Wahay can be done in the following way:

Fork wahay under a directory with the structure:

```console
$GOPATH/src/github.com/digitalautonomy/wahay
```

Another important thing is add **$GOPATH/bin to $PATH**

In this point the following commands must run without issues inside $GOPATH/src/github.com/digitalautonomy/wahay:

```console
$ make deps
$ make default
```

To add a new language translation add the required folder under locales for example fr:

```console
src/github.com/digitalautonomy/wahay/gui/locales/fr
```

After that copy one of the messages.gotext.json files of en or es folders to use as a reference. In this file also is required set the code of language to be used, for example: "language": "fr" and add all the translations in translation tag:
```json
{
"id": "Welcome",
"message": "Welcome",
"translation": "Bienvenue"
}
```

After that modify i18.go (//go:generate gotext -srclang=en update -out=catalog/catalog.go -lang=en,es,sv,ar) to add the new language to be translated:
//go:generate gotext -srclang=en update -out=catalog/catalog.go -lang=en,es,sv,ar,fr

Now its possible execute: 
```console
make gen-ui-locale under src/github.com/digitalautonomy/wahay, 
```

finally
```console
make default
```

To check the new language its possible use: **export LANG="fr_FR.utf8"**
