package config

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ApplicationConfig contains the configuration for the application.
type ApplicationConfig struct {
	initialized      bool
	filename         string
	ioLock           sync.Mutex
	afterSave        []func()
	afterLoad        []func(*ApplicationConfig)
	persistentMode   bool
	encryptedFile    bool
	encryptionParams *EncryptionParameters

	// The fields to save as the JSON representation of the configuration
	AutoJoin              bool
	UniqueConfigurationID string
	PathTor               string
	PathTorsocks          string
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
func (a *ApplicationConfig) Init() (string, error) {
	filename := a.getRealConfigFile()
	if len(filename) != 0 {
		a.SetPersistentConfiguration(true)
	} else {
		a.InitDefault()
		a.SetPersistentConfiguration(false)
	}

	a.initialized = true

	return filename, nil
}

// Load initialies the application once initialized.
// This function should be called after config.Init function.
func (a *ApplicationConfig) Load(filename string, k KeySupplier) (bool, bool, error) {
	if !a.initialized {
		return false, false, errors.New("required configuration-init not executed")
	}

	var err error
	var repeat bool

	if a.GetPersistentConfiguration() {
		err = a.loadFromFile(filename, k)
		if err == errorEncryptionBadFile || err == errInvalidConfigFile {
			return false, true, err
		}

		// Can we ask again for the key?
		repeat = err != nil && (err == errorEncryptionNoPassword ||
			err == errorEncryptionDecryptFailed)
	} else {
		// We are going to work with the default configuration
		repeat = false
	}

	if err == nil {
		a.onAfterLoad()
	}

	return repeat, false, err
}

func (a *ApplicationConfig) getRealConfigFile() string {
	dir := Dir()
	encryptedFile := filepath.Join(dir, appEncryptedConfigFile)
	if FileExists(encryptedFile) {
		a.SetShouldEncrypt(true)
		return encryptedFile
	}

	nonEncryptedFile := filepath.Join(dir, appConfigFile)
	if FileExists(nonEncryptedFile) {
		return nonEncryptedFile
	}

	return ""
}

// loadFromFile will try to load the configuration from the given configuration file.
// If no file exists or it is malformed, or it could not be decrypted, an error will be returned.
func (a *ApplicationConfig) loadFromFile(configFile string, k KeySupplier) error {
	a.ioLock.Lock()
	defer a.ioLock.Unlock()

	a.filename = configFile

	return a.tryLoad(k)
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
	_ = RandomString(s[:])
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

func (a *ApplicationConfig) tryLoad(k KeySupplier) error {
	var contents []byte
	var err error

	contents, err = ReadFileOrTemporaryBackup(a.filename)
	if err != nil {
		return errInvalidConfigFile
	}

	isEncrypted := isDataEncrypted(contents)
	if isEncrypted {
		a.SetShouldEncrypt(true)
		contents, a.encryptionParams, err = decryptConfigContent(contents, k)
		if err != nil {
			return err
		}
	} else if strings.HasSuffix(a.filename, encrytptedFileExtension) {
		// The file has been corrupted or manually updated
		// TODO: we should remove it?
		return errorEncryptionBadFile
	}

	if err = json.Unmarshal(contents, a); err != nil {
		return errInvalidConfigFile
	}

	return nil
}

// Save will save the application configuration
func (a *ApplicationConfig) Save(k KeySupplier) error {
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

	if a.ShouldEncrypt() {
		if a.encryptionParams == nil {
			p := newEncryptionParameters()
			a.encryptionParams = &p
		} else {
			// We should re-generate the nonce value every time as possible
			a.encryptionParams.regenerateNonce()
		}

		contents, err = encryptConfigContent(string(contents), a.encryptionParams, k)
		if err != nil {
			return err
		}
	}

	return SafeWrite(a.filename, contents, 0600)
}

// EnsureDestination check the destination for copying the configuration file
func (a *ApplicationConfig) EnsureDestination() {
	dir := Dir()
	EnsureDir(dir, 0700)

	if len(a.filename) == 0 {
		if a.ShouldEncrypt() {
			a.filename = filepath.Join(dir, appEncryptedConfigFile)
		} else {
			a.filename = filepath.Join(dir, appConfigFile)
		}
	} else {
		if a.ShouldEncrypt() && !strings.HasSuffix(a.filename, encrytptedFileExtension) {
			a.filename = filepath.Join(dir, appEncryptedConfigFile)
		}
	}
}

// DeleteFileIfExists deletes the config file if exists
func (a *ApplicationConfig) DeleteFileIfExists() {
	_ = os.RemoveAll(Dir())
}

func (a *ApplicationConfig) removeOldFileOnNextSave() {
	oldFilename := a.filename

	a.doAfterSave(func() {
		if FileExists(oldFilename) && a.filename != oldFilename {
			// TODO: Remove the file securely
			os.Remove(oldFilename)
		}
	})
}

func (a *ApplicationConfig) doAfterSave(f func()) {
	a.afterSave = append(a.afterSave, f)
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

func (a *ApplicationConfig) GetPathTor() string {
	return a.PathTor
}

func (a *ApplicationConfig) SetPathTor(p string) {
	a.PathTor = p
}

func (a *ApplicationConfig) GetPathTorSocks() string {
	return a.PathTorsocks
}

func (a *ApplicationConfig) SetPathTorSocks(ps string) {
	a.PathTorsocks = ps
}

// ShouldEncrypt returns a boolean indicating the configuration
// file is encrypted
func (a *ApplicationConfig) ShouldEncrypt() bool {
	return a.encryptedFile
}

// SetShouldEncrypt sets the encryption option to true or false
func (a *ApplicationConfig) SetShouldEncrypt(v bool) {
	if a.encryptedFile == v {
		return
	}

	a.encryptedFile = v
	if a.encryptedFile {
		a.turnOnEncryption()
	} else {
		a.turnOffEncryption()
	}
}
