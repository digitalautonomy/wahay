package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

// helper functions

func compareConfigs(c *C, obtained, expected *ApplicationConfig) {
	c.Assert(obtained.UniqueConfigurationID, Equals, expected.UniqueConfigurationID)
	c.Assert(obtained.AsSuperUser, Equals, expected.AsSuperUser)
	c.Assert(obtained.AutoJoin, Equals, expected.AutoJoin)
	c.Assert(obtained.PathTor, Equals, expected.PathTor)
	c.Assert(obtained.PathTorsocks, Equals, expected.PathTorsocks)
	c.Assert(obtained.LogsEnabled, Equals, expected.LogsEnabled)
	c.Assert(obtained.RawLogFile, Equals, expected.RawLogFile)
	c.Assert(obtained.PathMumble, Equals, expected.PathMumble)
	c.Assert(obtained.PortMumble, Equals, expected.PortMumble)
}

// tests

func (cs *ConfigSuite) Test_New_createsNewInstanceOfConfiguration(c *C) {
	ac := New()

	c.Assert(ac, NotNil)
	c.Assert(ac.initialized, Equals, false)
}

func (cs *ConfigSuite) Test_Init_setsInitializedFieldValueToTrue(c *C) {
	ac := New()
	ac.Init()

	c.Assert(ac.initialized, Equals, true)
}

func (cs *ConfigSuite) Test_InitDefault_initializesWithDefaultValues(c *C) {
	ac := New()
	ac.InitDefault()

	c.Assert(ac.AsSuperUser, Equals, true)
	c.Assert(ac.AutoJoin, Equals, true)
	c.Assert(ac.LogsEnabled, Equals, false)
	c.Assert(ac.RawLogFile, NotNil)
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

func (cs *ConfigSuite) Test_LoadFromFile_showErrorWhenPersistentConfigurationIsTrue(c *C) {
	mockKeySupplier := MockKeySupplier{}

	expectedKey := EncryptionResult{
		key:   []byte{0x01, 0x02, 0x03},
		mac:   []byte{0x04, 0x05, 0x06},
		valid: true,
	}
	mockKeySupplier.On("GenerateKey", mock.Anything).Return(expectedKey)

	ac := New()
	ac.Init()

	invalid, repeat, err := ac.LoadFromFile("example.json", &mockKeySupplier)

	c.Assert(ac.initialized, Equals, true)
	c.Assert(invalid, Equals, false)
	c.Assert(repeat, Equals, false)
	c.Assert(err, IsNil)
}

func (cs *ConfigSuite) Test_LoadFromFile_LoadsPersistentConfigFile(c *C) {
	mockKeySupplier := MockKeySupplier{}

	expectedKey := EncryptionResult{
		key:   []byte{0x01, 0x02, 0x03},
		mac:   []byte{0x04, 0x05, 0x06},
		valid: true,
	}
	mockKeySupplier.On("GenerateKey", mock.Anything).Return(expectedKey)

	fakeAppConfig := &ApplicationConfig{
		UniqueConfigurationID: "12345ABC",
		AsSuperUser:           true,
		AutoJoin:              true,
		PathTor:               "path/to/tor",
		PathTorsocks:          "path/to/tor/socks",
		LogsEnabled:           true,
		RawLogFile:            "raw_log_file",
		PathMumble:            "path/to/mumble",
		PortMumble:            "567",
	}

	fakeJsonContent, err := json.Marshal(fakeAppConfig)
	c.Assert(err, IsNil)

	tempDir, err := ioutil.TempDir("", "test")
	c.Assert(err, IsNil)

	tempFile := filepath.Join(tempDir, "example.json")
	err = ioutil.WriteFile(tempFile, []byte(fakeJsonContent), 0600)
	c.Assert(err, IsNil)
	defer os.RemoveAll(tempDir)

	ac := New()
	ac.Init()

	ac.persistentMode = true

	ac.LoadFromFile(tempFile, &mockKeySupplier)

	compareConfigs(c, ac, fakeAppConfig)
}

func (cs *ConfigSuite) Test_genUniqueID_generatesUniqueID(c *C) {
	ac := New()
	ac.genUniqueID()

	c.Assert(ac.UniqueConfigurationID, Not(Equals), "")
	c.Assert(len(ac.UniqueConfigurationID), Equals, 32*2)
}

func (cs *ConfigSuite) Test_GetUniqueID_returnsUniqueID(c *C) {
	ac := New()
	uniqueID := ac.GetUniqueID()

	c.Assert(ac.UniqueConfigurationID, Equals, uniqueID)
}

func (cs *ConfigSuite) Test_onBeforeSave_uniqueConfigIDIsGeneratedCorrectly(c *C) {
	ac := New()
	ac.Init()

	ac.onBeforeSave()

	c.Assert(ac.UniqueConfigurationID, Not(Equals), "")
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

func (cs *ConfigSuite) Test_Save_nonPersistentConfiguration(c *C) {

	ac := New()

	ac.persistentMode = false

	err := ac.Save(nil)

	c.Assert(err, NotNil)
}

func (cs *ConfigSuite) Test_DeleteFileIfExists_fileExists(c *C) {

	filename := "temp_file.txt"
	_, err := os.Create(filename)
	c.Assert(err, IsNil)

	ac := &ApplicationConfig{filename: filename}

	_, err = os.Stat(filename)
	c.Assert(err, IsNil)

	ac.DeleteFileIfExists()

	_, err = os.Stat(filename)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (cs *ConfigSuite) Test_DeleteFileIfExists_fileDoesNotExist(c *C) {
	ac := New()
	ac.filename = ""

	ac.DeleteFileIfExists()

	_, err := os.Stat(ac.filename)
	c.Assert(err, NotNil)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (cs *ConfigSuite) Test_CreateBackup_fileExists(c *C) {

	filename := "temp_file.txt"
	err := ioutil.WriteFile(filename, []byte("test data"), 0644)
	c.Assert(err, IsNil)

	ac := &ApplicationConfig{filename: filename}

	ac.CreateBackup()

	backupFile := filepath.Join(filepath.Dir(filename), appConfigFileBackup)
	_, err = os.Stat(backupFile)
	c.Assert(err, IsNil)

	data, err := ioutil.ReadFile(backupFile)
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "test data")

	defer os.Remove(filename)
	defer os.Remove(backupFile)
}

func (cs *ConfigSuite) Test_CreateBackup_fileDoesNotExist(c *C) {
	ac := &ApplicationConfig{filename: "non_existent_file.txt"}

	ac.CreateBackup()

	backupFile := filepath.Join(filepath.Dir("non_existent_file.txt"), appConfigFileBackup)
	_, err := os.Stat(backupFile)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (cs *ConfigSuite) Test_doAfterSave_addFunctionToList(c *C) {
	ac := New()

	testFunc := func() {}

	ac.doAfterSave(testFunc)

	found := false
	for _, f := range ac.afterSave {
		if reflect.ValueOf(f).Pointer() == reflect.ValueOf(testFunc).Pointer() {
			found = true
			break
		}
	}
	c.Assert(found, Equals, true)
}

type MockKeySupplier struct {
	mock.Mock
}

func (_m *MockKeySupplier) GenerateKey(p EncryptionParameters) EncryptionResult {
	ret := _m.Called(p)

	if len(ret) == 0 {
		panic("no return value specified for GenerateKey")
	}

	var r0 EncryptionResult
	if rf, ok := ret.Get(0).(func(EncryptionParameters) EncryptionResult); ok {
		r0 = rf(p)
	} else {
		r0 = ret.Get(0).(EncryptionResult)
	}

	return r0
}

func (_m *MockKeySupplier) CacheFromResult(r EncryptionResult) error {
	ret := _m.Called(r)

	if len(ret) == 0 {
		panic("no return value specified for CacheFromResult")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(EncryptionResult) error); ok {
		r0 = rf(r)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func (_m *MockKeySupplier) Invalidate() {
	_m.Called()
}

func (_m *MockKeySupplier) LastAttemptFailed() {
	_m.Called()
}

func NewMockKeySupplier(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockKeySupplier {
	mock := &MockKeySupplier{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
