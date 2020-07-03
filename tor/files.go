package tor

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		f.data, err = ioutil.ReadAll(b64)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/files/torrc": {
		local:   "files/torrc",
		size:    558,
		modtime: 1489449600,
		compressed: `
IyMgQ29uZmlndXJhdGlvbiBmaWxlIGZvciBhIHR5cGljYWwgVG9yIHVzZXIKCiMjIFRlbGwgVG9yIHRv
IG9wZW4gYSBTT0NLUyBwcm94eSBvbiBwb3J0IF9fUE9SVF9fClNPQ0tTUG9ydCBfX1BPUlRfXwoKIyMg
VGhlIHBvcnQgb24gd2hpY2ggVG9yIHdpbGwgbGlzdGVuIGZvciBsb2NhbCBjb25uZWN0aW9ucyBmcm9t
IFRvcgojIyBjb250cm9sbGVyIGFwcGxpY2F0aW9ucywgYXMgZG9jdW1lbnRlZCBpbiBjb250cm9sLXNw
ZWMudHh0LgpDb250cm9sUG9ydCBfX0NPTlRST0xQT1JUX18KCiMjIFRoZSBkaXJlY3RvcnkgZm9yIGtl
ZXBpbmcgYWxsIHRoZSBrZXlzL2V0Yy4KRGF0YURpcmVjdG9yeSBfX0RBVEFESVJfXwoKIyBBbGxvdyBj
b25uZWN0aW9ucyBvbiB0aGUgY29udHJvbCBwb3J0IHdoZW4gdGhlIGNvbm5lY3RpbmcgcHJvY2Vzcwoj
IGtub3dzIHRoZSBjb250ZW50cyBvZiBhIGZpbGUgbmFtZWQgImNvbnRyb2xfYXV0aF9jb29raWUiLCB3
aGljaCBUb3IKIyB3aWxsIGNyZWF0ZSBpbiBpdHMgZGF0YSBkaXJlY3RvcnkuCkNvb2tpZUF1dGhlbnRp
Y2F0aW9uIF9fQ09PS0lFX18K
`,
	},

	"/files/torrc-logs": {
		local:   "files/torrc-logs",
		size:    177,
		modtime: 1489449600,
		compressed: `
IyMgU2VuZCBhbGwgbWVzc2FnZXMgb2YgbGV2ZWwgJ25vdGljZScgb3IgaGlnaGVyIHRvIF9fTE9HTk9U
SUNFX18KIyBMb2cgbm90aWNlIGZpbGUgX19MT0dOT1RJQ0VfXwoKIyMgU2VuZCBldmVyeSBwb3NzaWJs
ZSBtZXNzYWdlIHRvIF9fTE9HTk9USUNFX18KIyBMb2cgZGVidWcgZmlsZSBfX0xPR0RFQlVHX18K
`,
	},

	"/": {
		isDir: true,
		local: "",
	},

	"/files": {
		isDir: true,
		local: "files",
	},
}
