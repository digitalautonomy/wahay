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
	tag, _ := jibberjabber.DetectLanguageTag()
	if tag == language.Und {
		tag = language.English
	}
	fmt.Printf("Detected language: %v\n", tag)
	i18n = message.NewPrinter(tag)
}

func (b *uiBuilder) i18nProperties(objs ...string) {
	if len(objs)%2 == 1 {
		panic("programmer error, uneven amount of arguments")
	}

	i := 0
	for i < len(objs) {
		b.i18nProperty(objs[i+1], objs[i])
		i += 2
	}
}

func (b *uiBuilder) i18nLabel(id string) {
	lbl := b.get(id).(gtki.Label)
	lbl.SetLabel(i18n.Sprint(lbl.GetLabel()))
}

func (b *uiBuilder) i18nProperty(id, property string) {
	obj := b.get(id).(glibi.Object)
	switch property {
	case "placeholder":
		property = "placeholder_text"
	case "button":
	case "checkbox":
		property = "label"
	case "tooltip":
		property = "tooltip_text"
	}

	currentVal, e := obj.GetProperty(property)
	if e != nil {
		// programmer error, so ok to die here
		fatal(e)
	}
	_ = obj.SetProperty(property, i18n.Sprint(currentVal))
}
