package config

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
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
	UniqueConfigurationID string
	AutoJoin              bool
	PathTor               string
	PathTorsocks          string
	LogsEnabled           bool
	RawLogFile            string
	PathMumble            string
	PortMumble            string
}

var (
	errInvalidConfigFile = errors.New("failed to parse config file")
)

// New creates a new instance of the application config struct
func New() *ApplicationConfig {
	a := new(ApplicationConfig)
	return a
}

// DetectPersistence initializes the application config
func (a *ApplicationConfig) DetectPersistence() (string, error) {
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

// LoadFromFile loads the content of a specific file and import it
// into the configuration instance.
func (a *ApplicationConfig) LoadFromFile(filename string, k KeySupplier) (invalid bool, repeat bool, err error) {
	if !a.initialized {
		return false, false, errors.New("required configuration-init not executed")
	}

	if a.IsPersistentConfiguration() {
		err = a.loadFromFile(filename, k)
		if err == errorEncryptionBadFile || err == errInvalidConfigFile {
			invalid = true
			return
		}

		repeat = err != nil && (err == errorEncryptionNoPassword ||
			err == errorEncryptionDecryptFailed)
	} else {
		repeat = false
	}

	return
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
	a.LogsEnabled = false
	a.RawLogFile = GetDefaultLogFile()
}

// WhenLoaded will ensure that the function f is not called until the configuration has been loaded
func (a *ApplicationConfig) WhenLoaded(f func(*ApplicationConfig)) {
	a.afterLoad = append(a.afterLoad, f)
}

// OnAfterLoad executes multiple callbacks when the configuration is ready
func (a *ApplicationConfig) OnAfterLoad() {
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
	if !a.IsPersistentConfiguration() {
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
	if FileExists(a.filename) {
		_ = os.Remove(a.filename)
	}
}

// CreateBackup creates a backup of the current configuration file
func (a *ApplicationConfig) CreateBackup() {
	if FileExists(a.filename) {
		data, err := ReadFileOrTemporaryBackup(a.filename)
		if err != nil {
			log.Println("Configuration file backup failed")
			return
		}

		backupFile := filepath.Join(filepath.Dir(a.filename), appConfigFileBackup)
		if FileExists(backupFile) {
			os.Remove(backupFile)
		}

		err = ioutil.WriteFile(backupFile, data, 0600)
		if err != nil {
			log.Println("Configuration file backup failed")
		}
	}
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

// IsPersistentConfiguration returns the setting value to persist the configuration file in the device
func (a *ApplicationConfig) IsPersistentConfiguration() bool {
	return a.persistentMode
}

// SetPersistentConfiguration sets the specified value to persist the configuration file in the device
func (a *ApplicationConfig) SetPersistentConfiguration(v bool) {
	a.persistentMode = v
}

// GetPathTor returns the configured path to Tor binary
func (a *ApplicationConfig) GetPathTor() string {
	return a.PathTor
}

// SetPathTor set the configuration value for the Tor binary path
func (a *ApplicationConfig) SetPathTor(p string) {
	a.PathTor = p
}

// GetPathTorSocks returns the configured path to Torsocks library
func (a *ApplicationConfig) GetPathTorSocks() string {
	return a.PathTorsocks
}

// SetPathTorSocks sets the configuration value for the path of Torsocks library
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

// IsLogsEnabled returns the current configured value for saving logs
func (a *ApplicationConfig) IsLogsEnabled() bool {
	return a.LogsEnabled
}

// EnableLogs sets the value for enabling or disabling logs
func (a *ApplicationConfig) EnableLogs(v bool) {
	a.LogsEnabled = v
}

// GetRawLogFile returns the configured value for the file to write logs
func (a *ApplicationConfig) GetRawLogFile() string {
	return a.RawLogFile
}

// SetCustomLogFile sets the value for the raw log file
func (a *ApplicationConfig) SetCustomLogFile(v string) {
	a.RawLogFile = v
}

// SetPathMumble sets the value for the Mumble binary path
func (a *ApplicationConfig) SetPathMumble(v string) {
	a.PathMumble = v
}

// GetPathMumble returns the custom path to find the Mumble binary
func (a *ApplicationConfig) GetPathMumble() string {
	return a.PathMumble
}

// SetPortMumble sets the value for the port for Mumble
func (a *ApplicationConfig) SetPortMumble(v string) {
	a.PortMumble = v
}

// GetPortMumble returns the custom value of the Mumble port
func (a *ApplicationConfig) GetPortMumble() string {
	return a.PortMumble
}

// GetDefaultLogFile returns the default path for the log file
func GetDefaultLogFile() string {
	return filepath.Join(Dir(), GetDefaultLogFileName())
}

// GetDefaultLogFileName returns the default filename for the log file
func GetDefaultLogFileName() string {
	return appLogFile
}
