package config

import (
	"reflect"
	"testing"

	. "gopkg.in/check.v1"
)

type ConfigSuite struct {}

var _ = Suite(&ConfigSuite{})

func TestConfig(t *testing.T) { TestingT(t) }

func (cs *ConfigSuite) Test_New_createsNewInstanceOfConfiguration(c *C) {
	ac := New()

	c.Assert(ac, Not(Equals), nil)
	c.Assert(ac.initialized, Equals, false)
}

func (cs *ConfigSuite) Test_Init_setsInitializedFieldValueToTrue(c *C) {
	ac := New()
	ac.Init()

	c.Assert(ac.initialized, Equals, true)
}

func (cs *ConfigSuite) Test_DetectPersistance_configurationInitializedCorrectly(c *C) {
	ac := New()
	filename, err := ac.DetectPersistence()

	c.Assert(err, IsNil)
	c.Assert(filename, Equals, "")
	c.Assert(ac.persistentMode, Equals, false)

}

func (cs *ConfigSuite) Test_LoadFromFile_showErrorWhenConfigurationInitIsNotExecuted(c *C) {
	ac := New()

	invalid, repeat, err := ac.LoadFromFile("example.json", nil)

	c.Assert(ac.initialized, Equals, false)
	c.Assert(invalid, Equals, false)
	c.Assert(repeat, Equals, false)
	c.Assert(err, ErrorMatches, "required configuration-init not executed")

}

func (cs *ConfigSuite) Test_LoadFromFile_showErrorWhenPersistentConfigurationIsFalse(c *C) {
	ac := New()
	ac.Init()

	invalid, repeat, err := ac.LoadFromFile("example.json", nil)

	c.Assert(ac.initialized, Equals, true)
	c.Assert(invalid, Equals, false)
	c.Assert(repeat, Equals, false)
	c.Assert(err, IsNil)

}

func (cs *ConfigSuite) Test_GenUniqueID_generatesUniqueID(c *C) {
    ac := New()
    ac.genUniqueID()

    c.Assert(ac.UniqueConfigurationID, Not(Equals), "")
    c.Assert(len(ac.UniqueConfigurationID), Equals, 32*2)
}

func (cs *ConfigSuite) Test_WhenLoaded_addsFunctionToList(c *C) {

	ac := New()
    var testFunc = func(a *ApplicationConfig) {}
    ac.WhenLoaded(testFunc)

	var found bool
    for _, f := range ac.afterLoad {
        if reflect.ValueOf(f).Pointer() == reflect.ValueOf(testFunc).Pointer() {
            found = true
            break
        }
    }
    c.Assert(found, Equals, true)
}


func (cs *ConfigSuite) Test_OnAfterLoad_executesCallbacksCorrectly(c *C) {

    ac := New()
	var testFunc1 = func(a *ApplicationConfig) {
        a.LogsEnabled = true
    }
	var testFunc2 = func(a *ApplicationConfig) {
        a.AutoJoin = true
    }

    ac.afterLoad = []func(*ApplicationConfig){testFunc1, testFunc2}
    ac.OnAfterLoad()

    c.Assert(ac.LogsEnabled, Equals, true)
    c.Assert(ac.AutoJoin, Equals, true)
    c.Assert(len(ac.afterLoad), Equals, 0)
}

func (cs *ConfigSuite) Test_OnAfterSave_executesCallbacksCorrectly(c *C) {
    ac := New()
    flag := false
    var testFunc = func() {
        flag = true
    }
    ac.afterSave = []func(){testFunc}

    ac.onAfterSave()

  
    c.Assert(flag, Equals, true)
    c.Assert(len(ac.afterSave), Equals, 0)
}
