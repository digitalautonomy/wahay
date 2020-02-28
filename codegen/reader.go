package codegen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

// FSStringProvider is the custom type to define the provider
// that returns the named file from the embedded assets.
type FSStringProvider func(useLocal bool, name string) (string, error)

// GetFileWithFallback returns the generated file (or the local) if available
func GetFileWithFallback(name, directory string, fs FSStringProvider) string {
	filename := filepath.Join("/"+directory, name)
	embeddedName := filepath.Join("/"+filepath.Base(directory), name)

	embeddedFile, err := fs(false, embeddedName)
	if err != nil {
		panic(fmt.Sprintf("No definition found for %s", embeddedName))
	}

	if fileNotFound(filename) {
		return embeddedFile
	}

	return readFile(filename)
}

func fileNotFound(filename string) bool {
	_, err := os.Stat(getRealFileName(filename))
	return os.IsNotExist(err)
}

func readFile(filename string) string {
	data, err := ioutil.ReadFile(getRealFileName(filename))
	if err != nil {
		panic(err)
	}

	return string(data)
}

func getRealFileName(filename string) string {
	var ok bool
	var callerFileName string

	if _, callerFileName, _, ok = runtime.Caller(1); !ok {
		return filename
	}

	return path.Join(path.Dir(callerFileName), "..", filename)
}
