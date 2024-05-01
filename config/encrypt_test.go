package config

import (
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