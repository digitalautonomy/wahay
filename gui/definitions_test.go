package gui

import (
	"os"
	"time"

	. "gopkg.in/check.v1"
)

type WahayGUIDefinitionsSuite struct{}

var _ = Suite(&WahayGUIDefinitionsSuite{})

func (s *WahayGUIDefinitionsSuite) Test_definitions_static_Open(c *C) {
	f := FS(false)
	_, e := f.Open("doesntexist")
	c.Assert(e, ErrorMatches, "file does not exist")

	_, e = f.Open("/definitions/MainWindow.xml")
	c.Assert(e, IsNil)

	_escData["/definitions/EmptyDef.xml"] = &_escFile{
		local:      "definitions/EmptyDef.xml",
		size:       0,
		modtime:    1489449600,
		compressed: ``,
	}

	_, e = f.Open("/definitions/EmptyDef.xml")
	c.Assert(e, IsNil)

	_escData["/definitions/BadDef.xml"] = &_escFile{
		local:   "definitions/BadDef.xml",
		size:    326,
		modtime: 1489449600,
		compressed: `
PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz4KPGludGVyZmFjZT4KICA8b2JqZWN0
IGNsYXNzPSJHdGtBcHBsaWNhdGlvbldpbmRvdyIgaWQ9Im1haW5XaW5kb3ciPgogICAgPHByb3BlcnR5
IG5hbWU9ImNhbl9mb2N1cyI+RmFsc2U8L3=yb3BlcnR5PgogICAgPHByb3BlcnR5IG5hbWU9InRpdGxl
Ij5Ub25pbyE8L3Byb3BlcnR5PgogICAgPHByb3BlcnR5IG5hbWU9ImRlZmF1bHRfd2lkdGgiPjYwMDwv
cHJvcGVydHk+CiAgICA8cHJvcGVydHkgbmFtZT0iZGVmYXVsdF9oZWlnaHQiPjQwMDwvcHJvcGVydHk+
CiAgPC9vYmplY3Q+CjwvaW50ZXJmYWNlP===
`,
	}

	_, e = f.Open("/definitions/BadDef.xml")
	c.Assert(e, ErrorMatches, "illegal base64 data.*")
}

func (s *WahayGUIDefinitionsSuite) Test_definitions_fsFileImplementations(c *C) {
	f := FS(false)
	ff, _ := f.Open("/definitions/MainWindow.xml")

	c.Assert(ff.Close(), IsNil)

	v1, v2 := ff.Readdir(2)
	c.Assert(v1, IsNil)
	c.Assert(v2, IsNil)

	fs, e := ff.Stat()
	c.Assert(e, IsNil)

	c.Assert(fs.Name(), Equals, "MainWindow.xml")

	c.Assert(fs.Size() > 100, Equals, true)

	c.Assert(fs.Mode(), Equals, os.FileMode(0))

	c.Assert(fs.IsDir(), Equals, false)

	c.Assert(fs.ModTime(), Equals, time.Unix(1489449600, 0))

	c.Assert(fs.Sys(), Not(IsNil))
}

func dirExists(d string) bool {
	_, err := os.Stat(d)
	return err == nil
}

func (s *WahayGUIDefinitionsSuite) Test_definitions_static_DirectoryOpen(c *C) {
	d := Dir(false, "/definitions")
	_, e := d.Open("/doesntexist")
	c.Assert(e, ErrorMatches, "file does not exist")

	_, e = d.Open("/MainWindow.xml")
	c.Assert(e, IsNil)
}

func (s *WahayGUIDefinitionsSuite) Test_definitions_local_Open(c *C) {
	if dirExists("definitions") {
		f := FS(true)
		_, e := f.Open("doesntexist")
		c.Assert(e, ErrorMatches, "file does not exist")

		_, e = f.Open("/definitions/MainWindow.xml")
		c.Assert(e, IsNil)
	}
}

func (s *WahayGUIDefinitionsSuite) Test_definitions_local_DirectoryOpen(c *C) {
	d := Dir(true, "/definitions")
	_, e := d.Open("/doesntexist")
	c.Assert(e, ErrorMatches, "file does not exist")

	if dirExists("definitions") {
		_, e = d.Open("/MainWindow.xml")
		c.Assert(e, IsNil)
	}
}

func (s *WahayGUIDefinitionsSuite) Test_definitions_local_FSByte(c *C) {
	_, e := FSByte(true, "doesntexist")
	c.Assert(e, ErrorMatches, "file does not exist")

	if dirExists("definitions") {
		_, e = FSByte(true, "/definitions/MainWindow.xml")
		c.Assert(e, IsNil)
	}
}

func (s *WahayGUIDefinitionsSuite) Test_definitions_local_FSString(c *C) {
	if dirExists("definitions") {
		_, e := FSString(true, "doesntexist")
		c.Assert(e, ErrorMatches, "file does not exist")

		_, e = FSString(true, "/definitions/MainWindow.xml")
		c.Assert(e, IsNil)
	}
}

func (s *WahayGUIDefinitionsSuite) Test_definitions_local_FSMustByte(c *C) {
	if dirExists("definitions") {
		c.Assert(func() { FSMustByte(true, "doesntexist") }, PanicMatches, "file does not exist")

		_ = FSMustByte(true, "/definitions/MainWindow.xml")
	}
}

func (s *WahayGUIDefinitionsSuite) Test_definitions_local_FSMustString(c *C) {
	if dirExists("definitions") {
		c.Assert(func() { FSMustString(true, "doesntexist") }, PanicMatches, "file does not exist")

		_ = FSMustString(true, "/definitions/MainWindow.xml")
	}
}
