package config

import (
	"encoding/hex"
	"encoding/json"

	. "gopkg.in/check.v1"
)

type EncryptSuite struct {
}

var _ = Suite(&EncryptSuite{})

func (e *EncryptSuite) Test_GenerateKey_getsValidKeys(c *C) {
	k := &keySupplierWrap{}
	k.haveKeys = false
	k.lastAttemptFailed = false

	expectedResult := EncryptionResult{
		key: []byte{0x01,0x02,0x03},
		mac: []byte{0x04,0x05,0x06},
		valid: true,
	}

	getFakeKeys := func(p EncryptionParameters, lastAttemptFailed bool) EncryptionResult {
        return expectedResult
    }

	k.getKeys = getFakeKeys

	result := k.GenerateKey(EncryptionParameters{})

	c.Assert(result.isValid(), Equals, true)
    c.Assert(result.getKey(), DeepEquals, expectedResult.getKey())
    c.Assert(result.getMacKey(), DeepEquals, expectedResult.getMacKey())
    c.Assert(k.haveKeys, Equals, true)
}

func (e *EncryptSuite) Test_CacheFromResult_cachesValidResult(c *C) {

	k := &keySupplierWrap{}
    expectedResult := EncryptionResult{
        key:   []byte{0x01, 0x02, 0x03},
        mac:   []byte{0x04, 0x05, 0x06},
        valid: true,
    }

    err := k.CacheFromResult(expectedResult)

    c.Assert(err, IsNil)
    c.Assert(k.haveKeys, Equals, true)
}

func (e *EncryptSuite) Test_CacheFromResult_errorWhenInvalidResult(c *C) {

	k := &keySupplierWrap{}
    expectedResult := EncryptionResult{
        key:   []byte{0x01, 0x02, 0x03},
        mac:   []byte{0x04, 0x05, 0x06},
        valid: false,
    }

    err := k.CacheFromResult(expectedResult)

    c.Assert(err, NotNil)
    c.Assert(k.haveKeys, Equals, false)
}

func (e *EncryptSuite) Test_Invalidate(c *C) {

    k := &keySupplierWrap{}
    k.haveKeys = true
    k.key = []byte{0x01, 0x02, 0x03}
    k.mac = []byte{0x04, 0x05, 0x06}

    k.Invalidate()

    c.Assert(k.haveKeys, Equals, false)
    c.Assert(len(k.key), Equals, 0)
    c.Assert(len(k.mac), Equals, 0)
}

func (e *EncryptSuite) Test_serialize_createsJSONRepresentation(c *C) {

    nonce := []byte{0x01,0x02,0x03}
    salt := []byte{0x04,0x05,0x06}

    params := &EncryptionParameters{
        nonceInternal: nonce,
        saltInternal: salt,
    }

    params.serialize()

    expectedNonce := hex.EncodeToString(nonce)
    expectedSalt := hex.EncodeToString(salt)

    c.Assert(params.Nonce, DeepEquals, expectedNonce)
    c.Assert(params.Salt, DeepEquals, expectedSalt)
}

func (e *EncryptSuite) Test_unserialize_decodingSuccesful(c *C) {

    params := &EncryptionParameters{
        Nonce: "1234567890",
        Salt: "ABCD",
    }

    err := params.unserialize()

    expectedNonceInternal, nonceErr := hex.DecodeString(params.Nonce)
    expectedSaltInternal, saltErr := hex.DecodeString(params.Salt)

    c.Assert(err, IsNil)
    c.Assert(nonceErr, IsNil)
    c.Assert(saltErr, IsNil)

    c.Assert(params.nonceInternal, DeepEquals, expectedNonceInternal)
    c.Assert(params.saltInternal, DeepEquals, expectedSaltInternal)
}

func (e *EncryptSuite) Test_unserialize_(c *C) {

    params := &EncryptionParameters{
        Nonce: "",
        Salt: "",
    }

    err := params.unserialize()

    c.Assert(err, NotNil)
}

func (e *EncryptSuite) Test_newEncryptionParameters_createsDefaultParams(c *C) {

    expectedN := 262144
    expectedR := 8
    expectedP := 1

    params := newEncryptionParameters()

    c.Assert(params.N, Equals, expectedN)
    c.Assert(params.R, Equals, expectedR)
    c.Assert(params.P, Equals, expectedP)

    c.Assert(len(params.nonceInternal), Equals, 12)
    c.Assert(len(params.saltInternal), Equals, 16)
}

func (e *EncryptSuite) Test_encryptConfigContent_verifiesThatContentIsEncrypted(c *C) {

    params := &EncryptionParameters{
        nonceInternal: []byte{
            0x01, 0x02, 0x03, 0x04,
            0x05, 0x06, 0x07, 0x08,
            0x09, 0x0A, 0x0B, 0x0C,
        },
        saltInternal: []byte{
            0x0D, 0x0F, 0x11, 0x12,
            0x13, 0x14, 0x15, 0x16,
            0x17, 0x18, 0x19, 0x1A,
            0x1B, 0x1C, 0x1D, 0x1F,
        },
    }

    fakeKeys := EncryptionResult{
		key: []byte("1234567890123456"),
		mac: []byte("abcdefghijklmnop"),
		valid: true,
	}

    getFakeKeys := func(p EncryptionParameters, lastAttemptFailed bool) EncryptionResult {
        return fakeKeys
    }

    k := &keySupplierWrap{
        key:    []byte{0x01,0x02,0x03},
        mac:    []byte{0x04,0x05,0x06},
    }
    k.getKeys = getFakeKeys

    testContent := "this is an example for content"

    encryptedContent, err := encryptConfigContent(testContent, params, k)

    c.Assert(err, IsNil)
    c.Assert(encryptedContent, NotNil)

    var encryptedData *encryptedData
    err = json.Unmarshal(encryptedContent, &encryptedData)

    c.Assert(err, IsNil)

    c.Assert(encryptedData.Params.Nonce, DeepEquals, params.Nonce)
    c.Assert(encryptedData.Params.Salt, DeepEquals, params.Salt)

    c.Assert(encryptedData.Data, NotNil)
}

func (e *EncryptSuite) Test_encryptConfigContent_errorIfKeyIsNotValid(c *C) {

    params := &EncryptionParameters{
        nonceInternal: []byte{
            0x01, 0x02, 0x03, 0x04,
            0x05, 0x06, 0x07, 0x08,
            0x09, 0x0A, 0x0B, 0x0C,
        },
        saltInternal: []byte{
            0x0D, 0x0F, 0x11, 0x12,
            0x13, 0x14, 0x15, 0x16,
            0x17, 0x18, 0x19, 0x1A,
            0x1B, 0x1C, 0x1D, 0x1F,
        },
    }

    fakeKeys := EncryptionResult{
		key: []byte("1234567890123456"),
		mac: []byte("abcdefghijklmnop"),
		valid: false,
	}

    getFakeKeys := func(p EncryptionParameters, lastAttemptFailed bool) EncryptionResult {
        return fakeKeys
    }

    k := &keySupplierWrap{
        key:    []byte{0x01,0x02,0x03},
        mac:    []byte{0x04,0x05,0x06},
    }
    k.getKeys = getFakeKeys

    testContent := "this is an example for content"

    _, err := encryptConfigContent(testContent, params, k)

    c.Assert(err,NotNil)
}