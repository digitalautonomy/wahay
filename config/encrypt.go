package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/scrypt"
)

// EncryptionResult is a representation of a result provided
// by the encryption functionality for the configuration system
type EncryptionResult struct {
	key   []byte
	mac   []byte
	valid bool
}

func (r *EncryptionResult) isValid() bool {
	return r.valid
}

func (r *EncryptionResult) getKey() []byte {
	return r.key
}

func (r *EncryptionResult) getMacKey() []byte {
	return r.mac
}

func (r *EncryptionResult) setKey(key []byte) {
	r.key = key
}

func (r *EncryptionResult) setMacKey(mac []byte) {
	r.mac = mac
}

func (r *EncryptionResult) setValid(v bool) {
	r.valid = v
}

// KeySupplier is a function that can be used to get key data from the user
type KeySupplier interface {
	GenerateKey(p EncryptionParameters) EncryptionResult
	CacheFromResult(r EncryptionResult) error
	Invalidate()
	LastAttemptFailed()
}

type keySupplierWrap struct {
	sync.Mutex
	haveKeys          bool
	key, mac          []byte
	getKeys           func(p EncryptionParameters, lastAttemptFailed bool) EncryptionResult
	lastAttemptFailed bool
}

// CreateKeySupplier wraps a function for returning the encryption key
func CreateKeySupplier(getKeys func(p EncryptionParameters, lastAttemptFailed bool) EncryptionResult) KeySupplier {
	return &keySupplierWrap{
		getKeys: getKeys,
	}
}

func (k *keySupplierWrap) GenerateKey(p EncryptionParameters) EncryptionResult {
	result := EncryptionResult{}

	k.Lock()
	defer k.Unlock()

	if !k.haveKeys {
		r := k.getKeys(p, k.lastAttemptFailed)
		if !r.isValid() {
			return result
		}

		k.setKey(r.getKey())
		k.setMacKey(r.getMacKey())
		k.haveKeys = true
	}

	result.setKey(k.getKey())
	result.setMacKey(k.getMacKey())
	result.setValid(true)

	return result
}

func (k *keySupplierWrap) Invalidate() {
	k.Lock()
	defer k.Unlock()

	k.haveKeys = false
	k.key = []byte{}
	k.mac = []byte{}
}

func (k *keySupplierWrap) LastAttemptFailed() {
	k.lastAttemptFailed = true
}

func (k *keySupplierWrap) CacheFromResult(r EncryptionResult) error {
	if !r.isValid() {
		return errors.New("invalid encryption result source")
	}

	k.setKey(r.getKey())
	k.setMacKey(r.getMacKey())
	k.haveKeys = true

	return nil
}

func (k *keySupplierWrap) getKey() []byte {
	return k.key
}

func (k *keySupplierWrap) getMacKey() []byte {
	return k.mac
}

func (k *keySupplierWrap) setKey(key []byte) {
	k.key = key
}

func (k *keySupplierWrap) setMacKey(mac []byte) {
	k.mac = mac
}

// EncryptionParameters contains the parameters used for scrypting
// the password and encrypting the configuration file
type EncryptionParameters struct {
	Nonce string
	Salt  string
	N     int
	R     int
	P     int

	// Similarly to ApplicationConfig, EncryptionParameters should
	// be just a JSON representation of whatever we use internally
	// to represent application configuration.
	nonceInternal []byte
	saltInternal  []byte
}

// TODO: Similarly to ApplicationConfig, this should be where we generate a new JSON representation and serialize it.
func (p *EncryptionParameters) serialize() {
	p.Nonce = hex.EncodeToString(p.nonceInternal)
	p.Salt = hex.EncodeToString(p.saltInternal)
}

func (p *EncryptionParameters) unserialize() (err error) {
	p.nonceInternal, err = hex.DecodeString(p.Nonce)
	if err != nil {
		return
	}

	p.saltInternal, err = hex.DecodeString(p.Salt)
	if err != nil {
		return
	}

	if len(p.nonceInternal) == 0 || len(p.saltInternal) == 0 {
		return errors.New("decryption params are empty")
	}

	return nil
}

func (p *EncryptionParameters) check() bool {
	err := p.unserialize()
	return err == nil
}

type encryptedData struct {
	Params EncryptionParameters
	Data   string
}

const (
	aesKeyLen = 32
	macKeyLen = 16
	nonceLen  = 12
	saltLen   = 16
)

// IsFileEncrypted returns a boolean indicating if the configuration
// file is apparently encrypted
func (a *ApplicationConfig) IsFileEncrypted() bool {
	return strings.HasSuffix(a.filename, encrytptedFileExtension)
}

func (a *ApplicationConfig) turnOnEncryption() {
	a.ioLock.Lock()
	defer a.ioLock.Unlock()

	if !strings.HasSuffix(a.filename, encrytptedFileExtension) {
		a.removeOldFileOnNextSave()
		a.filename = filepath.Join(Dir(), appEncryptedConfigFile)
	}
}

func (a *ApplicationConfig) turnOffEncryption() {
	a.ioLock.Lock()
	defer a.ioLock.Unlock()

	a.removeOldFileOnNextSave()
	a.filename = filepath.Join(Dir(), appConfigFile)
}

// Helper function for creating a default params for encrypt the
// configuration file
func newEncryptionParameters() EncryptionParameters {
	res := EncryptionParameters{
		N: 262144, // 2 ** 18
		R: 8,
		P: 1,
	}
	res.regenerateNonce()
	res.saltInternal = genRand(saltLen)
	return res
}

func (p *EncryptionParameters) regenerateNonce() {
	p.nonceInternal = genRand(nonceLen)
}

func genRand(size int) []byte {
	buf := make([]byte, size)
	if _, err := rand.Reader.Read(buf); err != nil {
		panic("Failed to read random bytes: " + err.Error())
	}
	return buf
}

func encryptConfigContent(content string, p *EncryptionParameters, k KeySupplier) ([]byte, error) {
	r := k.GenerateKey(*p)

	if !r.isValid() {
		return nil, errors.New("invalid password, aborting")
	}

	cipherText := encryptData(r.getKey(), r.getMacKey(), p.nonceInternal, content)

	p.serialize()

	cipherContent := encryptedData{
		Params: *p,
		Data:   hex.EncodeToString(cipherText),
	}

	return json.MarshalIndent(cipherContent, "", "\t")
}

var (
	errorEncryptionDecryptFailed = errors.New("decryption failed")
	errorEncryptionBadFile       = errors.New("invalid or corrupted file")
	errorEncryptionNoEncrypted   = errors.New("the configuration file data is not encrypted")
	errorEncryptionNoPassword    = errors.New("no password supplied to decrypt the config file")
)

func encryptData(key, macKey, nonce []byte, plain string) []byte {
	c, _ := aes.NewCipher(key)
	block, _ := cipher.NewGCM(c)
	return block.Seal(nil, nonce, []byte(plain), macKey)
}

func decryptConfigContent(content []byte, k KeySupplier) ([]byte, *EncryptionParameters, error) {
	data, err := parseEncryptedData(content)
	if err != nil {
		return nil, nil, errorEncryptionNoEncrypted
	}

	r := k.GenerateKey(data.Params)
	if !r.isValid() {
		return nil, nil, errorEncryptionNoPassword
	}

	cypherText, err := hex.DecodeString(data.Data)
	if err != nil {
		return nil, nil, errorEncryptionDecryptFailed
	}

	res, err := decryptData(r.getKey(), r.getMacKey(), data.Params.nonceInternal, cypherText)

	return res, &data.Params, err
}

func decryptData(key, macKey, nonce, cipherText []byte) ([]byte, error) {
	c, _ := aes.NewCipher(key)
	block, _ := cipher.NewGCM(c)
	res, err := block.Open(nil, nonce, cipherText, macKey)
	if err != nil {
		return nil, errorEncryptionDecryptFailed
	}

	return res, nil
}

func isDataEncrypted(content []byte) bool {
	data := new(encryptedData)

	err := json.Unmarshal(content, data)
	if err != nil {
		return false
	}

	return data.Params.check()
}

func parseEncryptedData(content []byte) (d *encryptedData, err error) {
	if !isDataEncrypted(content) {
		err = errorEncryptionNoEncrypted
		return
	}

	d = new(encryptedData)
	err = json.Unmarshal(content, d)
	if err != nil {
		return
	}

	err = d.Params.unserialize()

	return
}

// GenerateKeysBasedOnPassword takes a password and encryption parameters and
// generates an AES key and a MAC key using SCrypt
func GenerateKeysBasedOnPassword(password string, params EncryptionParameters) EncryptionResult {
	r := EncryptionResult{valid: true}
	res, err := scrypt.Key([]byte(password), params.saltInternal, params.N, params.R, params.P, aesKeyLen+macKeyLen)
	if err != nil {
		r.valid = false
		return r
	}

	r.key = res[0:aesKeyLen]
	r.mac = res[aesKeyLen:]

	return r
}
