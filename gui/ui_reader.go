//go:generate ./ui_generate.sh

package gui

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/digitalautonomy/wahay/codegen"
	log "github.com/sirupsen/logrus"

	"github.com/coyim/gotk3adapter/gdki"
	"github.com/coyim/gotk3adapter/glibi"
	"github.com/coyim/gotk3adapter/gtki"
)

const (
	xmlExtension = ".xml"
	cssExtension = ".css"
	pngExtension = ".png"
)

const (
	definitionsDir = "definitions"
	cssDir         = "styles"
	imagesDir      = "images"
	configFilesDir = "config_files"
)

var builderMutex sync.Mutex

type uiBuilder struct {
	gtki.Builder
}

func (g *Graphics) cssFor(name string) gtki.CssProvider {
	return g.loadCSSFor(name)
}

func (g *Graphics) uiBuilderFor(name string) *uiBuilder {
	return &uiBuilder{g.builderForDefinition(name)}
}

func readFile(fileName string) string {
	data, err := ioutil.ReadFile(filepath.Clean(fileName))
	if err != nil {
		fatal(err)
	}
	return string(data)
}

func (u *gtkUI) getConfigDesktopFile(fileName string) string {
	return u.getConfigFileFor(fileName, ".desktop")
}

func (u *gtkUI) getConfigFileFor(fileName, extension string) string {
	return codegen.GetFileWithFallback(fileName+extension, filepath.Join("gui", configFilesDir), FSString)
}

func getCSSFileWithFallback(fileName string) string {
	return codegen.GetFileWithFallback(fileName+cssExtension, filepath.Join("gui", cssDir), FSString)
}

func getDefinitionWithFileFallback(uiName string) string {
	return codegen.GetFileWithFallback(uiName+xmlExtension, filepath.Join("gui", definitionsDir), FSString)
}

func getActualDefsFolder() string {
	wd, _ := os.Getwd()
	if strings.HasSuffix(wd, "/gui") {
		return "definitions"
	}
	return "gui/definitions"
}

func isCSSVersionSufficient(gg gtki.Gtk) bool {
	major := gg.GetMajorVersion()
	minor := gg.GetMinorVersion()

	return major > uint(3) || (major == uint(3) && minor > uint(18))
}

func (g *Graphics) loadCSSFor(cssFile string) gtki.CssProvider {
	if isCSSVersionSufficient(g.gtk) {
		cssData := getCSSFileWithFallback(cssFile)

		builderMutex.Lock()
		defer builderMutex.Unlock()

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

	// We just return an empty css provider if we don't have a
	// late enough version of GTK. Later we can fallback
	// to other CSS versions or something like that
	cssProvider, err := g.gtk.CssProviderNew()
	if err != nil {
		fatal(err)
	}

	return cssProvider
}

// This must be called from the UI thread - otherwise bad things will happen sooner or later
func (g *Graphics) builderForDefinition(uiName string) gtki.Builder {
	template := getDefinitionWithFileFallback(uiName)

	builderMutex.Lock()
	defer builderMutex.Unlock()

	builder, err := g.gtk.BuilderNew()
	if err != nil {
		fatal(err)
	}

	// We dont use NewFromString because it doesnt give us an error message
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

func (g *Graphics) getImage(imageName string) []byte {
	return g.getImageBytes(imageName)
}

func (g Graphics) getImageBytes(filename string) []byte {
	image := filepath.Join("/"+imagesDir, filename)
	bs, err := FSByte(false, image)
	if err != nil {
		log.Fatal("Developer error: getting the image " + image + " but it does not exist")
	}
	return bs
}

func (g Graphics) getImagePixbufForSize(imageName string, size int) (gdki.Pixbuf, error) {
	var w sync.WaitGroup

	pl, err := g.gdk.PixbufLoaderNew()
	if err != nil {
		return nil, err
	}

	w.Add(1)

	_, err = pl.Connect("area-prepared", func() {
		defer w.Done()
	})
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "pl.Connect(\"area-prepared\")",
		}).Errorf("getImagePixbufForSize(): %s", err.Error())
	}

	_, err = pl.Connect("size-prepared", func() {
		pl.SetSize(size, size)
	})
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "pl.Connect(\"size-prepared\")",
		}).Errorf("getImagePixbufForSize(): %s", err.Error())
	}

	bytes := g.getImage(imageName)
	if _, err := pl.Write(bytes); err != nil {
		return nil, err
	}

	if err := pl.Close(); err != nil {
		return nil, err
	}

	w.Wait()

	return pl.GetPixbuf()
}
