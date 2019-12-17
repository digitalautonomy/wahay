package gui

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

	"/definitions/InviteCodeWindow.xml": {
		local:   "definitions/InviteCodeWindow.xml",
		size:    3294,
		modtime: 1489449600,
		compressed: `
PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPCEtLSBHZW5lcmF0ZWQgd2l0aCBn
bGFkZSAzLjIyLjEgLS0+CjxpbnRlcmZhY2U+CiAgPHJlcXVpcmVzIGxpYj0iZ3RrKyIgdmVyc2lvbj0i
My4xMiIvPgogIDxvYmplY3QgY2xhc3M9Ikd0a0FwcGxpY2F0aW9uV2luZG93IiBpZD0iaW52aXRlV2lu
ZG93Ij4KICAgIDxwcm9wZXJ0eSBuYW1lPSJjYW5fZm9jdXMiPkZhbHNlPC9wcm9wZXJ0eT4KICAgIDxw
cm9wZXJ0eSBuYW1lPSJ3aW5kb3dfcG9zaXRpb24iPmNlbnRlci1hbHdheXM8L3Byb3BlcnR5PgogICAg
PHByb3BlcnR5IG5hbWU9ImRlZmF1bHRfd2lkdGgiPjMwMDwvcHJvcGVydHk+CiAgICA8cHJvcGVydHkg
bmFtZT0iZGVmYXVsdF9oZWlnaHQiPjIwMDwvcHJvcGVydHk+CiAgICA8Y2hpbGQgdHlwZT0idGl0bGVi
YXIiPgogICAgICA8cGxhY2Vob2xkZXIvPgogICAgPC9jaGlsZD4KICAgIDxjaGlsZD4KICAgICAgPG9i
amVjdCBjbGFzcz0iR3RrQm94Ij4KICAgICAgICA8cHJvcGVydHkgbmFtZT0idmlzaWJsZSI+VHJ1ZTwv
cHJvcGVydHk+CiAgICAgICAgPHByb3BlcnR5IG5hbWU9ImNhbl9mb2N1cyI+RmFsc2U8L3Byb3BlcnR5
PgogICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJvcmllbnRhdGlvbiI+dmVydGljYWw8L3Byb3BlcnR5Pgog
ICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJob21vZ2VuZW91cyI+VHJ1ZTwvcHJvcGVydHk+CiAgICAgICAg
PGNoaWxkPgogICAgICAgICAgPG9iamVjdCBjbGFzcz0iR3RrQm94Ij4KICAgICAgICAgICAgPHByb3Bl
cnR5IG5hbWU9InZpc2libGUiPlRydWU8L3Byb3BlcnR5PgogICAgICAgICAgICA8cHJvcGVydHkgbmFt
ZT0iY2FuX2ZvY3VzIj5GYWxzZTwvcHJvcGVydHk+CiAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJo
YWxpZ24iPmNlbnRlcjwvcHJvcGVydHk+CiAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJ2YWxpZ24i
PmNlbnRlcjwvcHJvcGVydHk+CiAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJ2ZXhwYW5kIj5UcnVl
PC9wcm9wZXJ0eT4KICAgICAgICAgICAgPHByb3BlcnR5IG5hbWU9Im9yaWVudGF0aW9uIj52ZXJ0aWNh
bDwvcHJvcGVydHk+CiAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJob21vZ2VuZW91cyI+VHJ1ZTwv
cHJvcGVydHk+CiAgICAgICAgICAgIDxjaGlsZD4KICAgICAgICAgICAgICA8cGxhY2Vob2xkZXIvPgog
ICAgICAgICAgICA8L2NoaWxkPgogICAgICAgICAgICA8Y2hpbGQ+CiAgICAgICAgICAgICAgPG9iamVj
dCBjbGFzcz0iR3RrQm94Ij4KICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJ2aXNpYmxlIj5U
cnVlPC9wcm9wZXJ0eT4KICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJjYW5fZm9jdXMiPkZh
bHNlPC9wcm9wZXJ0eT4KICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJob21vZ2VuZW91cyI+
VHJ1ZTwvcHJvcGVydHk+CiAgICAgICAgICAgICAgICA8Y2hpbGQ+CiAgICAgICAgICAgICAgICAgIDxv
YmplY3QgY2xhc3M9Ikd0a0xhYmVsIiBpZD0ibGJsTWVldGluZ0lEIj4KICAgICAgICAgICAgICAgICAg
ICA8cHJvcGVydHkgbmFtZT0idmlzaWJsZSI+VHJ1ZTwvcHJvcGVydHk+CiAgICAgICAgICAgICAgICAg
ICAgPHByb3BlcnR5IG5hbWU9ImNhbl9mb2N1cyI+RmFsc2U8L3Byb3BlcnR5PgogICAgICAgICAgICAg
ICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJsYWJlbCIgdHJhbnNsYXRhYmxlPSJ5ZXMiPk1lZXRpbmcgSUQ6
PC9wcm9wZXJ0eT4KICAgICAgICAgICAgICAgICAgPC9vYmplY3Q+CiAgICAgICAgICAgICAgICAgIDxw
YWNraW5nPgogICAgICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJleHBhbmQiPkZhbHNlPC9w
cm9wZXJ0eT4KICAgICAgICAgICAgICAgICAgICA8cHJvcGVydHkgbmFtZT0iZmlsbCI+VHJ1ZTwvcHJv
cGVydHk+CiAgICAgICAgICAgICAgICAgICAgPHByb3BlcnR5IG5hbWU9InBvc2l0aW9uIj4wPC9wcm9w
ZXJ0eT4KICAgICAgICAgICAgICAgICAgPC9wYWNraW5nPgogICAgICAgICAgICAgICAgPC9jaGlsZD4K
ICAgICAgICAgICAgICAgIDxjaGlsZD4KICAgICAgICAgICAgICAgICAgPG9iamVjdCBjbGFzcz0iR3Rr
RW50cnkiIGlkPSJlbnRNZWV0aW5nSUQiPgogICAgICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1l
PSJ2aXNpYmxlIj5UcnVlPC9wcm9wZXJ0eT4KICAgICAgICAgICAgICAgICAgICA8cHJvcGVydHkgbmFt
ZT0iY2FuX2ZvY3VzIj5UcnVlPC9wcm9wZXJ0eT4KICAgICAgICAgICAgICAgICAgPC9vYmplY3Q+CiAg
ICAgICAgICAgICAgICAgIDxwYWNraW5nPgogICAgICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1l
PSJleHBhbmQiPkZhbHNlPC9wcm9wZXJ0eT4KICAgICAgICAgICAgICAgICAgICA8cHJvcGVydHkgbmFt
ZT0iZmlsbCI+VHJ1ZTwvcHJvcGVydHk+CiAgICAgICAgICAgICAgICAgICAgPHByb3BlcnR5IG5hbWU9
InBvc2l0aW9uIj4xPC9wcm9wZXJ0eT4KICAgICAgICAgICAgICAgICAgPC9wYWNraW5nPgogICAgICAg
ICAgICAgICAgPC9jaGlsZD4KICAgICAgICAgICAgICA8L29iamVjdD4KICAgICAgICAgICAgICA8cGFj
a2luZz4KICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJleHBhbmQiPkZhbHNlPC9wcm9wZXJ0
eT4KICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJmaWxsIj5UcnVlPC9wcm9wZXJ0eT4KICAg
ICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJwb3NpdGlvbiI+MTwvcHJvcGVydHk+CiAgICAgICAg
ICAgICAgPC9wYWNraW5nPgogICAgICAgICAgICA8L2NoaWxkPgogICAgICAgICAgICA8Y2hpbGQ+CiAg
ICAgICAgICAgICAgPHBsYWNlaG9sZGVyLz4KICAgICAgICAgICAgPC9jaGlsZD4KICAgICAgICAgICAg
PGNoaWxkPgogICAgICAgICAgICAgIDxwbGFjZWhvbGRlci8+CiAgICAgICAgICAgIDwvY2hpbGQ+CiAg
ICAgICAgICAgIDxjaGlsZD4KICAgICAgICAgICAgICA8cGxhY2Vob2xkZXIvPgogICAgICAgICAgICA8
L2NoaWxkPgogICAgICAgICAgPC9vYmplY3Q+CiAgICAgICAgICA8cGFja2luZz4KICAgICAgICAgICAg
PHByb3BlcnR5IG5hbWU9ImV4cGFuZCI+RmFsc2U8L3Byb3BlcnR5PgogICAgICAgICAgICA8cHJvcGVy
dHkgbmFtZT0iZmlsbCI+RmFsc2U8L3Byb3BlcnR5PgogICAgICAgICAgICA8cHJvcGVydHkgbmFtZT0i
cG9zaXRpb24iPjA8L3Byb3BlcnR5PgogICAgICAgICAgPC9wYWNraW5nPgogICAgICAgIDwvY2hpbGQ+
CiAgICAgIDwvb2JqZWN0PgogICAgPC9jaGlsZD4KICA8L29iamVjdD4KPC9pbnRlcmZhY2U+
`,
	},

	"/definitions/MainWindow.xml": {
		local:   "definitions/MainWindow.xml",
		size:    3871,
		modtime: 1489449600,
		compressed: `
PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPCEtLSBHZW5lcmF0ZWQgd2l0aCBn
bGFkZSAzLjIyLjEgLS0+CjxpbnRlcmZhY2U+CiAgPHJlcXVpcmVzIGxpYj0iZ3RrKyIgdmVyc2lvbj0i
My4xMiIvPgogIDxvYmplY3QgY2xhc3M9Ikd0a0FwcGxpY2F0aW9uV2luZG93IiBpZD0ibWFpbldpbmRv
dyI+CiAgICA8cHJvcGVydHkgbmFtZT0iY2FuX2ZvY3VzIj5GYWxzZTwvcHJvcGVydHk+CiAgICA8cHJv
cGVydHkgbmFtZT0id2luZG93X3Bvc2l0aW9uIj5jZW50ZXItYWx3YXlzPC9wcm9wZXJ0eT4KICAgIDxw
cm9wZXJ0eSBuYW1lPSJkZWZhdWx0X3dpZHRoIj4zMDA8L3Byb3BlcnR5PgogICAgPHByb3BlcnR5IG5h
bWU9ImRlZmF1bHRfaGVpZ2h0Ij4yMDA8L3Byb3BlcnR5PgogICAgPGNoaWxkIHR5cGU9InRpdGxlYmFy
Ij4KICAgICAgPHBsYWNlaG9sZGVyLz4KICAgIDwvY2hpbGQ+CiAgICA8Y2hpbGQ+CiAgICAgIDxvYmpl
Y3QgY2xhc3M9Ikd0a0JveCI+CiAgICAgICAgPHByb3BlcnR5IG5hbWU9InZpc2libGUiPlRydWU8L3By
b3BlcnR5PgogICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJjYW5fZm9jdXMiPkZhbHNlPC9wcm9wZXJ0eT4K
ICAgICAgICA8cHJvcGVydHkgbmFtZT0ib3JpZW50YXRpb24iPnZlcnRpY2FsPC9wcm9wZXJ0eT4KICAg
ICAgICA8cHJvcGVydHkgbmFtZT0ic3BhY2luZyI+MjA8L3Byb3BlcnR5PgogICAgICAgIDxjaGlsZD4K
ICAgICAgICAgIDxvYmplY3QgY2xhc3M9Ikd0a0JveCI+CiAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1l
PSJ2aXNpYmxlIj5UcnVlPC9wcm9wZXJ0eT4KICAgICAgICAgICAgPHByb3BlcnR5IG5hbWU9ImNhbl9m
b2N1cyI+RmFsc2U8L3Byb3BlcnR5PgogICAgICAgICAgICA8cHJvcGVydHkgbmFtZT0iaGFsaWduIj5j
ZW50ZXI8L3Byb3BlcnR5PgogICAgICAgICAgICA8cHJvcGVydHkgbmFtZT0idmFsaWduIj5jZW50ZXI8
L3Byb3BlcnR5PgogICAgICAgICAgICA8cHJvcGVydHkgbmFtZT0idmV4cGFuZCI+VHJ1ZTwvcHJvcGVy
dHk+CiAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJob21vZ2VuZW91cyI+VHJ1ZTwvcHJvcGVydHk+
CiAgICAgICAgICAgIDxjaGlsZD4KICAgICAgICAgICAgICA8cGxhY2Vob2xkZXIvPgogICAgICAgICAg
ICA8L2NoaWxkPgogICAgICAgICAgICA8Y2hpbGQ+CiAgICAgICAgICAgICAgPG9iamVjdCBjbGFzcz0i
R3RrQm94Ij4KICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJ2aXNpYmxlIj5UcnVlPC9wcm9w
ZXJ0eT4KICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJjYW5fZm9jdXMiPkZhbHNlPC9wcm9w
ZXJ0eT4KICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJ2YWxpZ24iPmVuZDwvcHJvcGVydHk+
CiAgICAgICAgICAgICAgICA8cHJvcGVydHkgbmFtZT0ib3JpZW50YXRpb24iPnZlcnRpY2FsPC9wcm9w
ZXJ0eT4KICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJzcGFjaW5nIj4xMDwvcHJvcGVydHk+
CiAgICAgICAgICAgICAgICA8Y2hpbGQ+CiAgICAgICAgICAgICAgICAgIDxwbGFjZWhvbGRlci8+CiAg
ICAgICAgICAgICAgICA8L2NoaWxkPgogICAgICAgICAgICAgICAgPGNoaWxkPgogICAgICAgICAgICAg
ICAgICA8b2JqZWN0IGNsYXNzPSJHdGtCdXR0b24iPgogICAgICAgICAgICAgICAgICAgIDxwcm9wZXJ0
eSBuYW1lPSJsYWJlbCIgdHJhbnNsYXRhYmxlPSJ5ZXMiPkhvc3QgbWV0dGluZzwvcHJvcGVydHk+CiAg
ICAgICAgICAgICAgICAgICAgPHByb3BlcnR5IG5hbWU9IndpZHRoX3JlcXVlc3QiPjEwMDwvcHJvcGVy
dHk+CiAgICAgICAgICAgICAgICAgICAgPHByb3BlcnR5IG5hbWU9InZpc2libGUiPlRydWU8L3Byb3Bl
cnR5PgogICAgICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJjYW5fZm9jdXMiPlRydWU8L3By
b3BlcnR5PgogICAgICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJyZWNlaXZlc19kZWZhdWx0
Ij5UcnVlPC9wcm9wZXJ0eT4KICAgICAgICAgICAgICAgICAgICA8c2lnbmFsIG5hbWU9ImNsaWNrZWQi
IGhhbmRsZXI9Im9uX2hvc3RfbWVldGluZyIgc3dhcHBlZD0ibm8iLz4KICAgICAgICAgICAgICAgICAg
PC9vYmplY3Q+CiAgICAgICAgICAgICAgICAgIDxwYWNraW5nPgogICAgICAgICAgICAgICAgICAgIDxw
cm9wZXJ0eSBuYW1lPSJleHBhbmQiPkZhbHNlPC9wcm9wZXJ0eT4KICAgICAgICAgICAgICAgICAgICA8
cHJvcGVydHkgbmFtZT0iZmlsbCI+VHJ1ZTwvcHJvcGVydHk+CiAgICAgICAgICAgICAgICAgICAgPHBy
b3BlcnR5IG5hbWU9InBvc2l0aW9uIj4xPC9wcm9wZXJ0eT4KICAgICAgICAgICAgICAgICAgPC9wYWNr
aW5nPgogICAgICAgICAgICAgICAgPC9jaGlsZD4KICAgICAgICAgICAgICAgIDxjaGlsZD4KICAgICAg
ICAgICAgICAgICAgPHBsYWNlaG9sZGVyLz4KICAgICAgICAgICAgICAgIDwvY2hpbGQ+CiAgICAgICAg
ICAgICAgICA8Y2hpbGQ+CiAgICAgICAgICAgICAgICAgIDxvYmplY3QgY2xhc3M9Ikd0a0J1dHRvbiI+
CiAgICAgICAgICAgICAgICAgICAgPHByb3BlcnR5IG5hbWU9ImxhYmVsIiB0cmFuc2xhdGFibGU9Inll
cyI+Sm9pbiBtZWV0aW5nPC9wcm9wZXJ0eT4KICAgICAgICAgICAgICAgICAgICA8cHJvcGVydHkgbmFt
ZT0idmlzaWJsZSI+VHJ1ZTwvcHJvcGVydHk+CiAgICAgICAgICAgICAgICAgICAgPHByb3BlcnR5IG5h
bWU9ImNhbl9mb2N1cyI+VHJ1ZTwvcHJvcGVydHk+CiAgICAgICAgICAgICAgICAgICAgPHByb3BlcnR5
IG5hbWU9InJlY2VpdmVzX2RlZmF1bHQiPlRydWU8L3Byb3BlcnR5PgogICAgICAgICAgICAgICAgICAg
IDxzaWduYWwgbmFtZT0iY2xpY2tlZCIgaGFuZGxlcj0ib25fam9pbl9tZWV0aW5nIiBzd2FwcGVkPSJu
byIvPgogICAgICAgICAgICAgICAgICA8L29iamVjdD4KICAgICAgICAgICAgICAgICAgPHBhY2tpbmc+
CiAgICAgICAgICAgICAgICAgICAgPHByb3BlcnR5IG5hbWU9ImV4cGFuZCI+RmFsc2U8L3Byb3BlcnR5
PgogICAgICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJmaWxsIj5UcnVlPC9wcm9wZXJ0eT4K
ICAgICAgICAgICAgICAgICAgICA8cHJvcGVydHkgbmFtZT0icG9zaXRpb24iPjM8L3Byb3BlcnR5Pgog
ICAgICAgICAgICAgICAgICA8L3BhY2tpbmc+CiAgICAgICAgICAgICAgICA8L2NoaWxkPgogICAgICAg
ICAgICAgICAgPGNoaWxkPgogICAgICAgICAgICAgICAgICA8cGxhY2Vob2xkZXIvPgogICAgICAgICAg
ICAgICAgPC9jaGlsZD4KICAgICAgICAgICAgICA8L29iamVjdD4KICAgICAgICAgICAgICA8cGFja2lu
Zz4KICAgICAgICAgICAgICAgIDxwcm9wZXJ0eSBuYW1lPSJleHBhbmQiPlRydWU8L3Byb3BlcnR5Pgog
ICAgICAgICAgICAgICAgPHByb3BlcnR5IG5hbWU9ImZpbGwiPlRydWU8L3Byb3BlcnR5PgogICAgICAg
ICAgICAgICAgPHByb3BlcnR5IG5hbWU9InBvc2l0aW9uIj4xPC9wcm9wZXJ0eT4KICAgICAgICAgICAg
ICA8L3BhY2tpbmc+CiAgICAgICAgICAgIDwvY2hpbGQ+CiAgICAgICAgICAgIDxjaGlsZD4KICAgICAg
ICAgICAgICA8cGxhY2Vob2xkZXIvPgogICAgICAgICAgICA8L2NoaWxkPgogICAgICAgICAgPC9vYmpl
Y3Q+CiAgICAgICAgICA8cGFja2luZz4KICAgICAgICAgICAgPHByb3BlcnR5IG5hbWU9ImV4cGFuZCI+
RmFsc2U8L3Byb3BlcnR5PgogICAgICAgICAgICA8cHJvcGVydHkgbmFtZT0iZmlsbCI+RmFsc2U8L3By
b3BlcnR5PgogICAgICAgICAgICA8cHJvcGVydHkgbmFtZT0icG9zaXRpb24iPjA8L3Byb3BlcnR5Pgog
ICAgICAgICAgPC9wYWNraW5nPgogICAgICAgIDwvY2hpbGQ+CiAgICAgIDwvb2JqZWN0PgogICAgPC9j
aGlsZD4KICA8L29iamVjdD4KPC9pbnRlcmZhY2U+Cg==
`,
	},

	"/": {
		isDir: true,
		local: "",
	},

	"/definitions": {
		isDir: true,
		local: "definitions",
	},
}
