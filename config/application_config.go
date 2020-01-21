package config

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
)

// ApplicationConfig contains the configuration for the application.
type ApplicationConfig struct {
	filename  string
	ioLock    sync.Mutex
	afterSave []func()

	AutoJoin              bool
	PersistConfigFile     bool
	UniqueConfigurationID string
}

var loadEntries []func(*ApplicationConfig)
var loadEntryLock = sync.Mutex{}

// CreateDefaultConfig initializes a basic application configuration
// with default values for each entry
func CreateDefaultConfig() *ApplicationConfig {
	c := &ApplicationConfig{
		AutoJoin:          true,
		PersistConfigFile: false,
	}

	return c
}

// WhenLoaded will ensure that the function f is not called until the configuration has been loaded
func (a *ApplicationConfig) WhenLoaded(f func(*ApplicationConfig)) {
	if a != nil {
		f(a)
		return
	}
	loadEntryLock.Lock()
	defer loadEntryLock.Unlock()

	loadEntries = append(loadEntries, f)
}

// LoadOrCreate will try to load the configuration from the given configuration file
// or from the standard configuration file. If no file exists or it is malformed,
// or it could not be decrypted, an error will be returned.
// However, the returned Accounts instance will always be usable
func LoadOrCreate(configFile string) (a *ApplicationConfig, e error) {
	a = new(ApplicationConfig)
	a.ioLock.Lock()
	defer a.ioLock.Unlock()

	a.filename = findConfigFile(configFile)
	e = a.tryLoad()

	return a, e
}

var (
	errInvalidConfigFile = errors.New("failed to parse config file")
)

func (a *ApplicationConfig) onAfterSave() {
	afterSaves := a.afterSave
	a.afterSave = nil
	for _, f := range afterSaves {
		f()
	}
}

// genUniqueID will generate and set a new unique ID fro this application config
func (a *ApplicationConfig) genUniqueID() {
	s := [32]byte{}
	_ = randomString(s[:])
	a.UniqueConfigurationID = hex.EncodeToString(s[:])
}

// GetUniqueID returns a unique id for this application config
func (a *ApplicationConfig) GetUniqueID() string {
	if a.UniqueConfigurationID == "" {
		a.genUniqueID()
	}
	return a.UniqueConfigurationID
}

func (a *ApplicationConfig) onBeforeSave() {
	if a.UniqueConfigurationID == "" {
		a.genUniqueID()
	}
}

func (a *ApplicationConfig) tryLoad() error {
	var contents []byte
	var err error

	contents, err = ReadFileOrTemporaryBackup(a.filename)
	if err != nil {
		return errInvalidConfigFile
	}

	if err = json.Unmarshal(contents, a); err != nil {
		return errInvalidConfigFile
	}

	return nil
}

// Save will save the application configuration
func (a *ApplicationConfig) Save() error {
	a.ioLock.Lock()
	defer a.ioLock.Unlock()
	a.onBeforeSave()
	defer a.onAfterSave()

	if len(a.filename) == 0 || !FileExists(a.filename) {
		a.filename = findConfigFile(a.filename)
	}

	contents, err := a.serialize()
	if err != nil {
		return err
	}

	return SafeWrite(a.filename, contents, 0600)
}

// DeleteFileIfExists delete the config file if exists
func (a *ApplicationConfig) DeleteFileIfExists() {
	dirs := []string{
		Dir(),
	}

	// Remove all possible directories
	for index := range dirs {
		if FileExists(dirs[index]) {
			_ = RemoveAll(dirs[index])
		}
	}
}

//TODO: This is where we generate a new JSON representation and serialize it.
//We are currently serializing our internal representation (ApplicationConfig) directly.
func (a *ApplicationConfig) serialize() ([]byte, error) {
	return json.MarshalIndent(a, "", "\t")
}

// GetAutoJoin returns the setting value to autojoin
func (a *ApplicationConfig) GetAutoJoin() bool {
	return a.AutoJoin
}

// SetAutoJoin sets the specified value to autojoin
func (a *ApplicationConfig) SetAutoJoin(v bool) {
	a.AutoJoin = v
}

// GetPersistentConfiguration returns the setting value to persist the configuration file in the device
func (a *ApplicationConfig) GetPersistentConfiguration() bool {
	return a.PersistConfigFile
}

// SetPersistentConfiguration sets the specified value to persist the configuration file in the device
func (a *ApplicationConfig) SetPersistentConfiguration(v bool) {
	a.PersistConfigFile = v
}
