package gui

//go:generate gotext -srclang=en update -out=catalog/catalog.go -lang=en,es

import (
	"fmt"

	"github.com/coyim/gotk3adapter/gtki"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	// This is necessary because that's how the translation stuff works
	_ "github.com/digitalautonomy/wahay/gui/catalog"
)

var i18n *message.Printer

func init() {
	// TODO: here we should initialize from the system language
	i18n = message.NewPrinter(language.Spanish)

	labelsGenerated = make(map[string]bool)
}

var labelsGenerated map[string]bool

func generateLocaleDefForString(s string) {
	if labelsGenerated[s] {
		return
	}

	fmt.Printf(`
	{
		"id": "%s",
		"message": "%s",
		"translation": "%s"
	}
`, s, s, s)
	labelsGenerated[s] = true
}

func i18nLabel(b *uiBuilder, id string) {
	lbl := b.get(id).(gtki.Label)
	generateLocaleDefForString(lbl.GetLabel())
	lbl.SetLabel(i18n.Sprint(lbl.GetLabel()))
}
