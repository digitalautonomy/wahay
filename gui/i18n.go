package gui

//go:generate gotext -srclang=en update -out=catalog/catalog.go -lang=en,es,sv,ar,fr

import (
	"sync"

	"github.com/coyim/gotk3adapter/glibi"

	"github.com/digitalautonomy/wahay/config"
	// This is necessary because that's how the translation stuff works
	_ "github.com/digitalautonomy/wahay/gui/catalog"

	log "github.com/sirupsen/logrus"

	"golang.org/x/text/message"
)

var i18n = func() func() *message.Printer {
	var o sync.Once
	var p *message.Printer

	initI18n := func() {
		tag := config.DetectLanguage()

		log.Infof("Detected language: %v\n", tag)
		p = message.NewPrinter(tag, message.Catalog(message.DefaultCatalog))
	}

	return func() *message.Printer {
		o.Do(initI18n)
		return p
	}
}()

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

func (b *uiBuilder) i18nProperty(id, property string) {
	obj := b.get(id).(glibi.Object)
	switch property {
	case "placeholder":
		property = "placeholder_text"
	case "button", "checkbox":
		property = "label"
	case "tooltip":
		property = "tooltip_text"
	}

	currentVal, e := obj.GetProperty(property)
	if e != nil {
		// programmer error, so ok to die here
		log.Errorf("Error getting property '%s' from object with id %s (%v): %v", property, id, obj, e)
		fatal(e)
	}

	e = obj.SetProperty(property, i18n().Sprintf(currentVal))
	if e != nil {
		log.Errorf("Error setting property '%s' from object with id %s (%v): %v", property, id, obj, e)
		fatal(e)
	}
}
