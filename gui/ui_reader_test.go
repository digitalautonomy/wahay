package gui

import (
	"errors"
	"os"
	"strings"

	. "github.com/digitalautonomy/wahay/test"

	"github.com/coyim/gotk3adapter/glibi"
	"github.com/coyim/gotk3adapter/gtk_mock"
	"github.com/coyim/gotk3adapter/gtki"
	. "gopkg.in/check.v1"
)

type WahayGUIUIReaderSuite struct{}

var _ = Suite(&WahayGUIUIReaderSuite{})

func (s *WahayGUIUIReaderSuite) Test_getActualDefsFolder(c *C) {
	wd, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(wd)
	}()

	c.Assert(getActualDefsFolder(), Equals, "definitions")

	_ = os.Chdir("/")
	c.Assert(getActualDefsFolder(), Equals, "gui/definitions")
}

func (s *WahayGUIUIReaderSuite) Test_getDefinitionWithFileFallback_returnsDefinitionIfExists(c *C) {
	ss := getDefinitionWithFileFallback("MainWindow")
	c.Assert(ss, Not(Equals), "")
	c.Assert(strings.Contains(ss, "\"GtkApplicationWindow\""), Equals, true)
}

func (s *WahayGUIUIReaderSuite) Test_getDefinitionWithFileFallback_panicsForNonExistingDefinition(c *C) {
	g1 := CreateGraphics(nil, nil, nil)
	c.Assert(func() { g1.uiBuilderFor("definitionThatDoesntExist") }, PanicMatches,
		"(?ms).*Developer error.*")
}

type testGtkWithBuilder struct {
	gtk_mock.Mock

	builderNewToReturn1 gtki.Builder
	builderNewToReturn2 error
}

type testBuilder struct {
	gtk_mock.MockBuilder

	getObjectArg1      string
	getObjectToReturn1 glibi.Object
	getObjectToReturn2 error

	addFromStringToReturn error
}

func (t *testBuilder) GetObject(v1 string) (glibi.Object, error) {
	t.getObjectArg1 = v1
	return t.getObjectToReturn1, t.getObjectToReturn2
}

func (t *testBuilder) AddFromString(v1 string) error {
	return t.addFromStringToReturn
}

func (t *testGtkWithBuilder) BuilderNew() (gtki.Builder, error) {
	return t.builderNewToReturn1, t.builderNewToReturn2
}

func (s *WahayGUIUIReaderSuite) Test_uiBuilder_get_returnsTheObjectForKnown(c *C) {
	ourGtk := &testGtkWithBuilder{}
	ourBuilder := &testBuilder{}
	ourGtk.builderNewToReturn1 = ourBuilder

	ourBuilder.getObjectToReturn1 = ourBuilder
	ourBuilder.getObjectToReturn2 = nil

	g1 := CreateGraphics(ourGtk, nil, nil)
	ss := g1.uiBuilderFor("MainWindow")
	v := ss.get("something")
	c.Assert(v, Equals, ourBuilder)
}

func (s *WahayGUIUIReaderSuite) Test_uiBuilder_get_forUnknownObjectPanics(c *C) {
	ourGtk := &testGtkWithBuilder{}
	ourBuilder := &testBuilder{}
	ourGtk.builderNewToReturn1 = ourBuilder

	ourBuilder.getObjectToReturn1 = nil
	ourBuilder.getObjectToReturn2 = errors.New("couldn't find it")

	g1 := CreateGraphics(ourGtk, nil, nil)
	ss := g1.uiBuilderFor("MainWindow")
	c.Assert(func() { ss.get("somethingNonExisting") }, PanicMatches, "failing on error: couldn't find it")
}

func (s *WahayGUIUIReaderSuite) Test_uiBuilderFor_panicsOnBadlyFormattedTemplate(c *C) {
	ourGtk := &testGtkWithBuilder{}
	ourBuilder := &testBuilder{}
	ourGtk.builderNewToReturn1 = ourBuilder

	ourBuilder.addFromStringToReturn = errors.New("badly formatted template")

	g1 := CreateGraphics(ourGtk, nil, nil)

	c.Assert(func() { g1.uiBuilderFor("MainWindow") }, PanicMatches,
		"gui: failed load MainWindow: badly formatted template")
}

func (s *WahayGUIUIReaderSuite) Test_uiBuilderFor_panicsIfBuilderCantBeCreated(c *C) {
	ourGtk := &testGtkWithBuilder{}
	ourGtk.builderNewToReturn2 = errors.New("bad GTK error")

	g1 := CreateGraphics(ourGtk, nil, nil)

	c.Assert(func() { g1.uiBuilderFor("MainWindow") }, PanicMatches, "failing on error: bad GTK error")
}

func (s *WahayGUIUIReaderSuite) Test_readFile_failsIfErrorHappens(c *C) {
	c.Assert(func() { readFile("none_existing_file") }, PanicMatches,
		"failing on error: open none_existing_file: no such file or directory")
}

func (s *WahayGUIUIReaderSuite) Test_getConfigFileFor_returnsTheWahayDesktopConfigFile(c *C) {
	val := getConfigFileFor("wahay", ".desktop")

	c.Assert(val, HasLen, 221)
	c.Assert(val, Contains, "Terminal=false")
	c.Assert(val, Contains, "Secure and Decentralized Conference")
}

func (s *WahayGUIUIReaderSuite) Test_getConfigFileFor_panicsWhenAskedForAConfigFileThatDoesntExist(c *C) {
	c.Assert(func() {
		getConfigFileFor("foobar", ".something")
	}, PanicMatches, "(?ms).*Developer error.*")
}

func (s *WahayGUIUIReaderSuite) Test_getImage_returnsAnImageThatExists(c *C) {
	c.Assert(len(getImage("help.svg")), Equals, 1072)
	c.Assert(len(getImage("wahay-192x192.png")), Equals, 26279)
	c.Assert(len(getImage("join-meeting.svg")), Equals, 1206)
}

func (s *WahayGUIUIReaderSuite) Test_getImage_panicsWhenTheImageDoesntExist(c *C) {
	c.Assert(func() {
		getImage("santa.dancing.jpg")
	}, PanicMatches, "(?ms).*Developer error.*")
}

func (s *WahayGUIUIReaderSuite) Test_getCSSFileWithFallback_returnsTheMainCSSFile(c *C) {
	val := getCSSFileWithFallback("gui")

	c.Assert(val, HasLen, 16457)
	c.Assert(val, Contains, "padding: 1px 2px;")
	c.Assert(val, Contains, ".host-meeting-toolbar .message")
}

func (s *WahayGUIUIReaderSuite) Test_getCSSFileWithFallback_panicsWhenAskedForAConfigFileThatDoesntExist(c *C) {
	c.Assert(func() {
		getCSSFileWithFallback("foobar")
	}, PanicMatches, "(?ms).*Developer error.*")
}
