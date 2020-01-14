//go:generate esc -o definitions.go -modtime 1489449600 -pkg gui -ignore "Makefile" definitions/ styles/

package gui

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/coyim/gotk3adapter/gdki"
	"github.com/coyim/gotk3adapter/glibi"
	"github.com/coyim/gotk3adapter/gtki"
)

const (
	xmlExtension = ".xml"
	cssExtension = ".css"
	imagesFolder = "images/"
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
		fatalf("No definition found for %s", fname)
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

func (b *uiBuilder) getItem(name string, target interface{}) {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr {
		panic("builder.getItem() target argument must be a pointer")
	}
	elem := v.Elem()
	elem.Set(reflect.ValueOf(b.get(name)))
}

//TODO: Why not a map[string]interface{}?
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

func (b *uiBuilder) get(name string) glibi.Object {
	obj, err := b.GetObject(name)
	if err != nil {
		fatal(err)
	}
	return obj
}

func mustGetImageBytes(filename string) []byte {
	bs, err := FSByte(false, imagesFolder+filename)
	if err != nil {
		panic("Developer error: getting the image " + filename + " but it does not exist")
	}
	return bs
}

// getPixbufFromBytes returns the pixbuff from a byte array.
func (u *gtkUI) getPixbufFromBytes(bstream []byte) gdki.Pixbuf {
	pl, err := u.g.gdk.PixbufLoaderNew()
	if err != nil {
		panic("Developer error: setting the image from >>>>>>>>")
	}

	var w sync.WaitGroup
	w.Add(1)
	pl.Connect("area-prepared", w.Done)

	if _, err := pl.Write(bstream); err != nil {
		log.Println(">> WARN - cannot write to PixbufLoader: " + err.Error())
		return nil
	}
	if err := pl.Close(); err != nil {
		log.Println(">> WARN - cannot close PixbufLoader: " + err.Error())
		return nil
	}

	w.Wait() //Waiting for Pixbuf to load before using it

	pb, err := pl.GetPixbuf()
	if err != nil {
		log.Println(">> WARN - cannot write to PixbufLoader: " + err.Error())
		return nil
	}
	return pb
}

// setImageFromFile sets an image from a file name.
func (u *gtkUI) setImageFromFile(i gtki.Image, filename string) {
	pb := u.getPixbufFromBytes(mustGetImageBytes(filename))
	i.SetFromPixbuf(pb)
}
