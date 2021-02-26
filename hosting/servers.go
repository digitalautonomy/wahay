package hosting

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/grumble/pkg/logtarget"
	grumbleServer "github.com/digitalautonomy/grumble/server"
	"github.com/digitalautonomy/wahay/tor"
)

// Servers serves
type Servers interface {
	CreateServer(...serverModifier) (Server, error)
	DestroyServer(Server) error
	DataDir() string
	Cleanup()
	NewService(port string, t tor.Instance) (Service, error)
}

// MeetingData is a representation of the data used to create a Mumble url
// More information at https://wiki.mumble.info/wiki/Mumble_URL
type MeetingData struct {
	MeetingID string
	Port      int
	Password  string
	Username  string
}

func create() (Servers, error) {
	s := &servers{}
	e := s.create()

	return s, e
}

type servers struct {
	dataDir string
	started bool
	nextID  int
	servers map[int64]*grumbleServer.Server
	log     *log.Logger
}

// GenerateURL is a helper function for creating Mumble valid URLs
func (d *MeetingData) GenerateURL() string {
	u := url.URL{
		Scheme: "mumble",
		User:   url.UserPassword(d.Username, d.Password),
		Host:   fmt.Sprintf("%s:%d", d.MeetingID, d.Port),
	}

	return u.String()
}

func (s *servers) initializeSharedObjects() {
	s.servers = make(map[int64]*grumbleServer.Server)
	grumbleServer.SetServers(s.servers)
}

func (s *servers) initializeDataDirectory() error {
	var e error
	s.dataDir, e = ioutil.TempDir("", "wahay")
	if e != nil {
		s.log.Debug(e.Error())
		return e
	}

	grumbleServer.Args.DataDir = s.dataDir

	e = os.MkdirAll(filepath.Join(s.dataDir, "servers"), 0700)
	if e != nil {
		s.log.Debug(e.Error())
		return e
	}

	return nil
}

func (s *servers) initializeLogging() error {
	logDir := path.Join(s.dataDir, "grumble.log")
	grumbleServer.Args.LogPath = logDir

	err := logtarget.Target.OpenFile(logDir)
	if err != nil {
		return err
	}

	l := log.New()
	l.SetOutput(&logtarget.Target)
	s.log = l
	s.log.Info("Grumble")
	s.log.Infof("Using data directory: %s", s.dataDir)

	return nil
}

func (s *servers) initializeCertificates() error {
	s.log.Debug("Generating 4096-bit RSA keypair for self-signed certificate...")

	certFn := filepath.Join(s.dataDir, "cert.pem")
	keyFn := filepath.Join(s.dataDir, "key.pem")
	err := grumbleServer.GenerateSelfSignedCert(certFn, keyFn)
	if err != nil {
		return err
	}

	s.log.Debugf("Certificate output to %v", certFn)
	s.log.Debugf("Private key output to %v", keyFn)
	return nil
}

func callAll(fs ...func() error) error {
	for _, f := range fs {
		if e := f(); e != nil {
			return e
		}
	}
	return nil
}

// create will initialize all grumble things
// because the grumble server package uses global
// state it is NOT advisable to call this function
// more than once in a program
func (s *servers) create() error {
	s.initializeSharedObjects()

	return callAll(
		s.initializeDataDirectory,
		s.initializeLogging,
		s.initializeCertificates,
	)
}

func (s *servers) startListener() {
	if !s.started {
		go grumbleServer.SignalHandler()
		s.started = true
	}
}

type serverModifier func(*grumbleServer.Server)

func setDefaultOptions(serv *grumbleServer.Server) {
	serv.Set("NoWebServer", "true")
	serv.Set("Address", defaultHost())
}

func setWelcomeText(t string) serverModifier {
	return func(serv *grumbleServer.Server) {
		if len(t) != 0 {
			serv.Set("WelcomeText", t)
		}
	}
}

func setPort(port string) serverModifier {
	return func(serv *grumbleServer.Server) {
		serv.Set("Port", port)
	}
}

func setPassword(password string) serverModifier {
	return func(serv *grumbleServer.Server) {
		if len(password) != 0 {
			serv.SetServerPassword(password)
		}
	}
}

func setSuperUser(username, password string) serverModifier {
	return func(serv *grumbleServer.Server) {
		if len(username) != 0 && len(password) != 0 {
			serv.SetSuperUserName(username)
			serv.SetSuperUserPassword(password)
		}
	}
}

func (s *servers) CreateServer(modifiers ...serverModifier) (Server, error) {
	s.nextID++

	serv, err := grumbleServer.NewServer(int64(s.nextID))
	if err != nil {
		return nil, err
	}

	s.servers[serv.Id] = serv

	err = os.Mkdir(filepath.Join(s.dataDir, "servers", fmt.Sprintf("%v", serv.Id)), 0750)
	if err != nil {
		return nil, err
	}

	for _, m := range modifiers {
		m(serv)
	}

	return &server{s, serv}, nil
}

func (s *servers) DestroyServer(Server) error {
	// For now, this function will do nothing. We will still call it,
	// in case we need it in the server
	return nil
}

func (s *servers) DataDir() string {
	return s.dataDir
}

func (s *servers) Cleanup() {
	err := os.RemoveAll(s.dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Error cleaning up temporaries: "+err.Error())
	}
}
