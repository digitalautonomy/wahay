package config

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// ApplicationConfig contains the configuration for the application.
type ApplicationConfig struct {
	filename       string
	ioLock         sync.Mutex
	afterSave      []func()
	afterLoad      []func(*ApplicationConfig)
	persistentMode bool

	AutoJoin              bool
	UniqueConfigurationID string
}

var (
	errInvalidConfigFile = errors.New("failed to parse config file")
)

// New creates a new instance of the application config struct
func New() *ApplicationConfig {
	a := new(ApplicationConfig)
	return a
}

// Init initializes the application config
func (a *ApplicationConfig) Init() error {
	var err error

	defer a.onAfterLoad()

	f := filepath.Join(Dir(), appConfigFile)
	if FileExists(f) {
		err = a.loadFromFile(f)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("the configuration settings can't be loaded: %s", err)
		}
		a.SetPersistentConfiguration(true)
	} else {
		log.Println("Initializing default configuration")
		a.InitDefault()
		a.SetPersistentConfiguration(false)
	}

	return nil
}

// loadFromFile will try to load the configuration from the given configuration file.
// If no file exists or it is malformed, or it could not be decrypted, an error will be returned.
func (a *ApplicationConfig) loadFromFile(configFile string) error {
	a.ioLock.Lock()
	defer a.ioLock.Unlock()

	a.filename = configFile
	err := a.tryLoad()
	if err != nil {
		return err
	}

	return nil
}

// InitDefault initializes a basic application configuration
// with default values for each entry
func (a *ApplicationConfig) InitDefault() {
	a.AutoJoin = true
}

// WhenLoaded will ensure that the function f is not called until the configuration has been loaded
func (a *ApplicationConfig) WhenLoaded(f func(*ApplicationConfig)) {
	a.afterLoad = append(a.afterLoad, f)
}

func (a *ApplicationConfig) onAfterLoad() {
	afterLoads := a.afterLoad
	a.afterLoad = nil
	for _, f := range afterLoads {
		f(a)
	}
}

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
	// Important: Do not save the configuration into a file
	// if we are in a non-persistent config mode
	if !a.GetPersistentConfiguration() {
		return errors.New("the configuration settings can't be saved in a non-persistent mode")
	}

	a.ioLock.Lock()
	defer a.ioLock.Unlock()
	a.onBeforeSave()
	defer a.onAfterSave()

	// Ensure the directory where the configuration file will be saved
	a.EnsureDestination()

	contents, err := a.serialize()
	if err != nil {
		return err
	}

	return SafeWrite(a.filename, contents, 0600)
}

// EnsureDestination check the destination for copying the configuration file
func (a *ApplicationConfig) EnsureDestination() {
	if len(a.filename) == 0 {
		dir := Dir()
		EnsureDir(dir, 0700)
		filename := filepath.Join(dir, appConfigFile)
		a.filename = filename
	}
}

// DeleteFileIfExists deletes the config file if exists
func (a *ApplicationConfig) DeleteFileIfExists() {
	_ = os.RemoveAll(Dir())
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
	return a.persistentMode
}

// SetPersistentConfiguration sets the specified value to persist the configuration file in the device
func (a *ApplicationConfig) SetPersistentConfiguration(v bool) {
	a.persistentMode = v
}
