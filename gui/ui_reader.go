//go:generate esc -o definitions.go -modtime 1489449600 -pkg gui -ignore "Makefile" definitions/

package gui

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/coyim/gotk3adapter/glibi"
	"github.com/coyim/gotk3adapter/gtki"
)

const (
	xmlExtension = ".xml"
)

type uiBuilder struct {
	gtki.Builder
}

func (g *Graphics) uiBuilderFor(name string) *uiBuilder {
	return &uiBuilder{g.builderForDefinition(name)}
}

func getActualDefsFolder() string {
	wd, _ := os.Getwd()
	if strings.HasSuffix(wd, "/gui") {
		return "definitions"
	}
	return "gui/definitions"
}

func fileNotFound(fileName string) bool {
	_, fnf := os.Stat(fileName)
	return os.IsNotExist(fnf)
}

func readFile(fileName string) string {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fatal(err)
	}
	return string(data)
}

func getDefinitionWithFileFallback(uiName string) string {
	fname := path.Join("/definitions", uiName+xmlExtension)

	embeddedFile, err := FSString(false, fname)
	if err != nil {
		fatalf("No definition found for %s", uiName)
	}

	fileName := filepath.Join(getActualDefsFolder(), uiName+xmlExtension)
	if fileNotFound(fileName) {
		return embeddedFile
	}

	log.Printf("Loading definition from local file: %q\n", fileName)
	return readFile(fileName)
}

// This must be called from the UI thread - otherwise bad things will happen sooner or later
func (g *Graphics) builderForDefinition(uiName string) gtki.Builder {
	template := getDefinitionWithFileFallback(uiName)

	builder, err := g.gtk.BuilderNew()
	if err != nil {
		fatal(err)
	}

	//We dont use NewFromString because it doesnt give us an error message
	err = builder.AddFromString(template)
	if err != nil {
		fatalf("gui: failed load %s: %s", uiName, err.Error())
	}

	return builder
}

func (b *uiBuilder) get(name string) glibi.Object {
	obj, err := b.GetObject(name)
	if err != nil {
		fatal(err)
	}
	return obj
}

func (b *uiBuilder) getItems(args ...interface{}) {
	for len(args) >= 2 {
		name, ok := args[0].(string)
		if !ok {
			panic("string argument expected in builder.getItems()")
		}
		b.getItem(name, args[1])
		args = args[2:]
	}
}

func (b *uiBuilder) getItem(name string, target interface{}) {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr {
		panic("builder.getItem() target argument must be a pointer")
	}
	elem := v.Elem()
	elem.Set(reflect.ValueOf(b.get(name)))
}
