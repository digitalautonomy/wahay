# Translations

To translate Wahay you will need to build it with the new translated messages.

To do that, please read the below instructions carefully and follow the steps.

➀ Clone `Wahay` into a directory with the following structure:

```bash
$GOPATH/src/github.com/digitalautonomy/wahay
```

Another important thing to do is to add `$GOPATH/bin` to `$PATH`.

At this point you must run the below commands in Wahay's directory:

```bash
$ cd $GOPATH/src/github.com/digitalautonomy/wahay
$ make deps
$ make default
```

➁ Create new language folder

To add a new translation you should create the language folder under `locales` directory:

```bash
$ cd $GOPATH/src/github.com/digitalautonomy/wahay/gui/locales

# Create a new folder for French language
$ mkdir fr
```

➂ Copy the `messages.gotext.json` file from __English__ (`en`) or __Spanish__ (`es`) folders as a reference, into the created language folder:

```bash
$ cd $GOPATH/src/github.com/digitalautonomy/wahay/gui/locales
$ cp ./en/messages.gotext.json ./fr
```

The `messages.gotext.json` must have the following structure that contains the language code and all translated messages:

```json
{
  "language": "fr",
  "messages": [
    {
      ...
    },
    {
      "id": "Welcome",
      "message": "Welcome",
      "translation": "Bienvenue"
    }
    {
      ...
    }
  ]
}
```

➃ Modify the `i18.go` file to run the new language source:

```bash
$ cd $GOPATH/src/github.com/digitalautonomy/wahay
$ nano gui/i18.go
```

And add the language code to the list of languages 

```go
//go:generate gotext -srclang=en update-out=catalog/catalog.go -lang=en,es,sv,ar,→fr
```

Now its possible execute the command:

```bash
# cd $GOPATH/src/github.com/digitalautonomy/wahay
$ make gen-ui-locale
```

And finally run:

```bash
$ make default
```

➄ Check the new language:

```bash
$ export LANG="fr_FR.utf8"
$ cd $GOPATH/src/github.com/digitalautonomy/wahay
$ bin/wahay
```
