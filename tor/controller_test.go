package tor

import (
	"errors"
	"testing"

	"github.com/wybiral/torgo"
	. "gopkg.in/check.v1"
)

type TonioTorSuite struct{}

var _ = Suite(&TonioTorSuite{})

func Test(t *testing.T) { TestingT(t) }

type controllerMock struct {
	authenticatePasswordArg1   string
	authenticatePasswordCalled bool
	authenticatePasswordReturn error

	authenticateCookieCalled bool
	authenticateCookieReturn error

	authenticateNoneCalled bool
	authenticateNoneReturn error

	addOnionArg1           *torgo.Onion
	addOnionCalled         bool
	addOnionReturnError    error
	addOnionAddServiceInfo string

	deleteOnionArg         *string
	deleteOnionCalled      bool
	deleteOnionReturnError error

	getVersionReturn1 string
	getVersionReturn2 error
}

func (m *controllerMock) AuthenticateNone() error {
	m.authenticateNoneCalled = true
	return m.authenticateNoneReturn
}

func (m *controllerMock) AuthenticateCookie() error {
	m.authenticateCookieCalled = true
	return m.authenticateCookieReturn
}

func (m *controllerMock) AuthenticatePassword(v1 string) error {
	m.authenticatePasswordArg1 = v1
	m.authenticatePasswordCalled = true
	return m.authenticatePasswordReturn
}

func (m *controllerMock) AddOnion(v1 *torgo.Onion) error {
	m.addOnionCalled = true
	m.addOnionArg1 = v1
	if m.addOnionAddServiceInfo != "" {
		v1.ServiceID = m.addOnionAddServiceInfo
	}
	return m.addOnionReturnError
}

func (m *controllerMock) GetVersion() (string, error) {
	return m.getVersionReturn1, m.getVersionReturn2
}

func (m *controllerMock) DeleteOnion(serviceID string) error {
	m.deleteOnionCalled = true
	if serviceID != "" {
		m.deleteOnionArg = &serviceID
	}
	return m.deleteOnionReturnError
}

func (m *controllerMock) createTestGotor(addr string) (torgoController, error) {
	return m, nil
}

func (s *TonioTorSuite) Test_controller_CreateNewOnionService_returnsErrorIfAuthenticationFails(c *C) {
	mock := &controllerMock{}
	mock.authenticatePasswordReturn = errors.New("authentication failed bla bla")
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "doesntMatter",
		tc:       mock.createTestGotor,
	}

	_, e := cntrl.CreateNewOnionService("addr", "dport", "123")

	c.Assert(mock.authenticatePasswordCalled, Equals, true)
	c.Assert(e, ErrorMatches, "authentication failed.*")
}

func (s *TonioTorSuite) Test_controller_CreateNewOnionService_authenticatesWithGivenPassword(c *C) {
	mock := &controllerMock{}
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "11112222",
		tc:       mock.createTestGotor,
	}

	_, e := cntrl.CreateNewOnionService("addr", "dport", "123")

	c.Assert(e, IsNil)
	c.Assert(mock.authenticatePasswordArg1, Equals, "11112222")
}

func (s *TonioTorSuite) Test_controller_CreateNewOnionService_returnsErrorIfTorControllerCantBeCreated(c *C) {
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "11112222",
		tc: func(string) (torgoController, error) {
			return nil, errors.New("couldn't create torgocontroller")
		},
	}

	_, e := cntrl.CreateNewOnionService("addr", "dport", "123")

	c.Assert(e, ErrorMatches, "couldn't create torgocontroller")
}

func (s *TonioTorSuite) Test_controller_CreateNewOnionService_triesToCreateTorControllerWithGivenAddress(c *C) {
	var addrGiven *string
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "11112222",
		tc: func(v string) (torgoController, error) {
			addrGiven = &v
			return nil, errors.New("we want stop early. this error is not part of the test")
		},
	}

	_, _ = cntrl.CreateNewOnionService("128.0.42.1", "3535", "123")

	c.Assert(addrGiven, Not(IsNil))
	c.Assert(*addrGiven, Equals, "127.1.2.3:9052")
}

func (s *TonioTorSuite) Test_controller_CreateNewOnionService_returnsErrorIfAddOnionFails(c *C) {
	mock := &controllerMock{}
	mock.addOnionReturnError = errors.New("add onion failed")
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "doesntMatter",
		tc:       mock.createTestGotor,
	}

	_, e := cntrl.CreateNewOnionService("addr", "dport", "123")

	c.Assert(mock.addOnionCalled, Equals, true)
	c.Assert(e, ErrorMatches, "add onion failed.*")
}

func (s *TonioTorSuite) Test_controller_CreateNewOnionService_createsOnionWithGivenArguments(c *C) {
	mock := &controllerMock{}
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "doesntMatter",
		tc:       mock.createTestGotor,
	}

	_, e := cntrl.CreateNewOnionService("127.0.42.1", "42", "7877")

	c.Assert(mock.addOnionCalled, Equals, true)
	c.Assert(e, IsNil)
	o := mock.addOnionArg1
	c.Assert(o, Not(IsNil))
	c.Assert(o.PrivateKeyType, Equals, "NEW")
	c.Assert(o.PrivateKey, Equals, "ED25519-V3")
	c.Assert(o.Ports, HasLen, 1)
	c.Assert(o.Ports[7877], Equals, "127.0.42.1:42")
}

func (s *TonioTorSuite) Test_controller_CreateNewOnionService_signalsErrorForInvalidPorts(c *C) {
	mock := &controllerMock{}
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "doesntMatter",
		tc:       mock.createTestGotor,
	}

	_, e := cntrl.CreateNewOnionService("127.0.42.1", "42", "not port")
	c.Assert(e, ErrorMatches, "invalid source port")

	_, e = cntrl.CreateNewOnionService("127.0.42.1", "42", "-1")
	c.Assert(e, ErrorMatches, "invalid source port")

	_, e = cntrl.CreateNewOnionService("127.0.42.1", "42", "65536")
	c.Assert(e, ErrorMatches, "invalid source port")
}

func (s *TonioTorSuite) Test_controller_CreateNewOnionService_returnsTheServiceID(c *C) {
	mock := &controllerMock{}
	mock.addOnionAddServiceInfo = "123abcfff"
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "doesntMatter",
		tc:       mock.createTestGotor,
	}

	serviceID, e := cntrl.CreateNewOnionService("127.0.42.1", "42", "7877")

	c.Assert(e, IsNil)
	c.Assert(serviceID, Equals, "123abcfff.onion")
}

func (s *TonioTorSuite) Test_controller_EnsureTorCompatibility_authenticatesCorrectly(c *C) {
	mock := &controllerMock{}
	mock.getVersionReturn1 = "1.2.3.4"
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "11112223",
		tc:       mock.createTestGotor,
	}

	e := cntrl.EnsureTorCompatibility()

	c.Assert(e, IsNil)
	c.Assert(mock.authenticatePasswordArg1, Equals, "11112223")
}

func (s *TonioTorSuite) Test_controller_EnsureTorCompatibility_failsOnAuthenticationError(c *C) {
	mock := &controllerMock{}
	mock.authenticatePasswordReturn = errors.New("oh no, godzilla")
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "11112223",
		tc:       mock.createTestGotor,
	}

	e := cntrl.EnsureTorCompatibility()

	c.Assert(e, ErrorMatches, "oh no, godzilla")
}

func (s *TonioTorSuite) Test_controller_EnsureTorCompatibility_reportsErrorWhenFailingToCreateController(c *C) {
	cntrl := &controller{
		torHost:  "127.1.2.4",
		torPort:  "9053",
		password: "11112223",
		tc: func(string) (torgoController, error) {
			return nil, errors.New("couldn't create torgocontroller")
		},
	}

	e := cntrl.EnsureTorCompatibility()

	c.Assert(e, ErrorMatches, "couldn't create torgocontroller")
}

func (s *TonioTorSuite) Test_controller_EnsureTorCompatibility_triesToCreateTorControllerWithGivenAddress(c *C) {
	var addrGiven *string
	cntrl := &controller{
		torHost:  "127.1.2.4",
		torPort:  "9053",
		password: "11112222",
		tc: func(v string) (torgoController, error) {
			addrGiven = &v
			return nil, errors.New("we want stop early. this error is not part of the test")
		},
	}

	_ = cntrl.EnsureTorCompatibility()

	c.Assert(addrGiven, Not(IsNil))
	c.Assert(*addrGiven, Equals, "127.1.2.4:9053")
}

func (s *TonioTorSuite) Test_controller_EnsureTorCompatibility_returnsErrorIfCallForCurrentVersionFails(c *C) {
	mock := &controllerMock{}
	mock.getVersionReturn2 = errors.New("connection failed")
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "11112223",
		tc:       mock.createTestGotor,
	}

	e := cntrl.EnsureTorCompatibility()

	c.Assert(e, ErrorMatches, "connection failed")
}

func (s *TonioTorSuite) Test_controller_EnsureTorCompatibility_returnsErrorIfVersionGivenFromTorIsBadlyFormatted(c *C) {
	mock := &controllerMock{}
	mock.getVersionReturn1 = "1.x.12"
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "11112223",
		tc:       mock.createTestGotor,
	}

	e := cntrl.EnsureTorCompatibility()

	c.Assert(e, ErrorMatches, "invalid version string")
}

func (s *TonioTorSuite) Test_controller_EnsureTorCompatibility_returnsErrorIfTorVersionIsTooLow(c *C) {
	mock := &controllerMock{}
	mock.getVersionReturn1 = "0.1.42.x"
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "11112223",
		tc:       mock.createTestGotor,
	}

	e := cntrl.EnsureTorCompatibility()

	c.Assert(e, ErrorMatches, "version of Tor is not compatible")
}

func (s *TonioTorSuite) Test_controller_EnsureTorCompatibility_returnsNilIfTorIsCompatible(c *C) {
	mock := &controllerMock{}
	mock.getVersionReturn1 = "0.4.1"
	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "11112223",
		tc:       mock.createTestGotor,
	}

	e := cntrl.EnsureTorCompatibility()

	c.Assert(e, IsNil)
}

func (s *TonioTorSuite) Test_controller_DeleteOnionService_returnsErrorIfServiceIDIsEmpty(c *C) {
	mock := &controllerMock{}
	mock.deleteOnionReturnError = errors.New("the service ID cannot be empty")

	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "11112223",
		tc:       mock.createTestGotor,
	}

	_, _ = cntrl.CreateNewOnionService("127.1.2.3", "9052", "7877")

	e := cntrl.c.DeleteOnion("")

	//error if delete fail
	c.Assert(e, ErrorMatches, "the service ID cannot be empty")
}

func (s *TonioTorSuite) Test_controller_DeleteOnionService_returnsErrorIfFails(c *C) {
	mock := &controllerMock{}
	mock.deleteOnionReturnError = errors.New("service deletion error")

	cntrl := &controller{
		torHost:  "127.1.2.3",
		torPort:  "9052",
		password: "11112223",
		tc:       mock.createTestGotor,
	}

	_, _ = cntrl.CreateNewOnionService("127.1.2.3", "9052", "7877")

	e := cntrl.c.DeleteOnion("123456")

	//error if delete fail
	c.Assert(e, ErrorMatches, "service deletion error")
}
