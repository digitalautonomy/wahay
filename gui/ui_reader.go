//go:generate esc -o definitions.go -modtime 1489449600 -pkg gui -ignore "Makefile" definitions/ styles/

package gui

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/coyim/gotk3adapter/glibi"
	"github.com/coyim/gotk3adapter/gtki"
)

const (
	xmlExtension = ".xml"
	cssExtension = ".css"
)

const (
	definitionsDir = "definitions"
	cssDir         = "styles"
)

type uiBuilder struct {
	gtki.Builder
}

func (g *Graphics) cssFor(name string) gtki.CssProvider {
	return g.loadCSSFor(name)
}

func (g *Graphics) uiBuilderFor(name string) *uiBuilder {
	return &uiBuilder{g.builderForDefinition(name)}
}

func getActualFolder(directory string) string {
	wd, _ := os.Getwd()
	if strings.HasSuffix(wd, "/gui") {
		return directory
	}
	return path.Join("gui/", directory)
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

func getFileWithFallback(fileName string, fileExtension string, directory string) string {
	fname := path.Join("/"+directory, fileName+fileExtension)

	embeddedFile, err := FSString(false, fname)
	if err != nil {
		fatalf("No existing file found for %s", fname)
	}

	file := filepath.Join(getActualFolder(directory), fileName+fileExtension)
	if fileNotFound(file) {
		return embeddedFile
	}

	log.Printf("Loading content from local file: %q\n", file)
	return readFile(file)
}

func getCSSFileWithFallback(fileName string) string {
	return getFileWithFallback(fileName, cssExtension, cssDir)
}

func getDefinitionWithFileFallback(uiName string) string {
	return getFileWithFallback(uiName, xmlExtension, definitionsDir)
}

func (g *Graphics) loadCSSFor(cssFile string) gtki.CssProvider {
	cssData := getCSSFileWithFallback(cssFile)

	cssProvider, err := g.gtk.CssProviderNew()
	if err != nil {
		fatal(err)
	}

	err = cssProvider.LoadFromData(cssData)
	if err != nil {
		fatalf("gui: failed load %s: %s", cssFile, err.Error())
	}

	return cssProvider
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
