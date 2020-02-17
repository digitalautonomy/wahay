//go:generate ./ui_generate.sh

package gui

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

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
	imagesDir      = "/images"
	configFilesDir = "/config_files"
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

func (u *gtkUI) getConfigDesktopFile(fileName string) string {
	return u.getConfigFileFor(fileName, ".desktop")
}

func (u *gtkUI) getConfigIniFile(fileName string) string {
	return u.getConfigFileFor(fileName, ".ini")
}

func (u *gtkUI) getConfigFileFor(fileName, extension string) string {
	return getFileWithFallback(fileName, extension, configFilesDir)
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

	log.Debugf("Loading content from local file: %q\n", file)
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

// This must be called from the UI thread - otherwise bad things will happen sooner or later
func (g *Graphics) builderForDefinition(uiName string) gtki.Builder {
	template := getDefinitionWithFileFallback(uiName)

	builderMutex.Lock()
	defer builderMutex.Unlock()

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
	image := filepath.Join(imagesDir, filename)
	bs, err := FSByte(false, image)
	if err != nil {
		log.Fatal("Developer error: getting the image " + image + " but it does not exist")
	}
	return bs
}

func (g Graphics) getImagePixbufForSize(imageName string) (gdki.Pixbuf, error) {
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
		pl.SetSize(100, 100)
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
