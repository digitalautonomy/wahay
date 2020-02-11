package gui

//go:generate gotext -srclang=en update -out=catalog/catalog.go -lang=en,es

import (
	"fmt"

	"github.com/coyim/gotk3adapter/glibi"
	"github.com/coyim/gotk3adapter/gtki"

	"github.com/cubiest/jibberjabber"

	// This is necessary because that's how the translation stuff works
	_ "github.com/digitalautonomy/wahay/gui/catalog"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var i18n *message.Printer

func init() {
	// TODO: here we should initialize from the system language
	tag, _ := jibberjabber.DetectLanguageTag()
	if tag == language.Und {
		tag = language.English
	}
	fmt.Printf("Detected language: %v\n", tag)
	i18n = message.NewPrinter(tag)
}

func (b *uiBuilder) i18nLabel(id string) {
	lbl := b.get(id).(gtki.Label)
	lbl.SetLabel(i18n.Sprint(lbl.GetLabel()))
}

func (b *uiBuilder) i18nProperty(id string) {
	lbl := b.get(id).(glibi.Object)
	currVal, e := lbl.GetProperty("label")
	if e != nil {
		// programmer error, so ok to die here
		fatal(e)
	}
	lbl.SetProperty("label", i18n.Sprint(currVal.(string)))
}
