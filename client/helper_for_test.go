package client

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type clientSuite struct{}

var _ = Suite(&clientSuite{})

func init() {
	logrus.SetOutput(io.Discard)
}

var fakeCert = `-----BEGIN CERTIFICATE-----
ZzBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQCp5hnG7ogBhtlynpOS21cBewKE/B7j
V14qeyslnr26xZUsSVko36ZnhiaO/zbMOoRcKK9vEcgMtcLFuQTWDl3RAgMBAAGj
gbEwga4wHQYDVR0OBBYEFFXI70krXeQDxZgbaCQoR4jUDncEMH8GA1UdIwR4MHaA
BgNVBAMTC0hlcm9uZyBZYW5nggEAMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQEE
FFXI70krXeQDxZgbaCQoR4jUDncEoVukWTBXMQswCQYDVQQGEwJDTjELMAkGA1UE
Wm7DCfrPNGVwFWUQOmsPue9rZBgO
-----END CERTIFICATE-----`

type mockTempDir struct {
	mock.Mock
}

func (m *mockTempDir) tempDir(dir, prefix string) (string, error) {
	args := m.Called(dir, prefix)
	return args.String(0), args.Error(1)
}

type mockCmd struct {
	mock.Mock
}

func (m *mockCmd) Output() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}
