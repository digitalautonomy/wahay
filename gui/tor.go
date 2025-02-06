package gui

import (
	"errors"
	"sync"

	"github.com/digitalautonomy/wahay/tor"
)

var errTorNoBinary = errors.New("tor can't be used")

func (u *gtkUI) ensureTor(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer u.torInitialized.Done()

		instance, e := tor.NewInstance(u.config, u.onTorInstanceCreated)
		if e != nil {
			u.errorHandler.addNewStartupError(e, errGroupTor)
			return
		}

		if instance == nil {
			u.errorHandler.addNewStartupError(errTorNoBinary, errGroupTor)
			return
		}

		u.tor = instance
	}()
}

func (u *gtkUI) onTorInstanceCreated(i tor.Instance) {
	// Tor instance has been successfully created, so we
	// add a new cleanup callback to destroy the given Tor
	// instance so when Wahay closes Tor can cleanup things
	u.onExit(i.Destroy)
}

func (u *gtkUI) waitForTorInstance(f func(tor.Instance)) {
	go func() {
		u.torInitialized.Wait()
		f(u.tor)
	}()
}

const errGroupTor errGroupType = "tor"

func init() {
	initStartupErrorGroup(errGroupTor, torErrorTranslator)
}

func torErrorTranslator(err error) string {
	switch err {
	case tor.ErrTorBinaryNotFound:
		return i18n().Sprintf("In order to run Wahay, you must have Tor installed in your system.\n\n" +
			"You can also download the Wahay bundle with Tor from our website:\n\n" +
			"https://wahay.org/download.html\n")

	case tor.ErrTorInstanceCantStart:
		return i18n().Sprintf("The Tor instance can't be started.\n")

	case tor.ErrTorConnectionTimeout:
		return i18n().Sprintf("The Tor instance can't connect to the Tor network.\n\n" +
			"Please check the information available at " +
			"https://tb-manual.torproject.org/troubleshooting/ to know what you can do.\n")

	case tor.ErrPartialTorNoControlPort:
		return i18n().Sprintf("No valid Tor Control Port found on the system in order to run Wahay.\n")

	case tor.ErrPartialTorNoValidAuth:
		return i18n().Sprintf("No valid Tor Control Port authentication method found on the system.\n")

	case tor.ErrFatalTorNoConnectionAllowed:
		return i18n().Sprintf("We found a valid Tor in the system but the connection over Tor network " +
			"is not available.\n\nPlease check the information available at " +
			"https://tb-manual.torproject.org/troubleshooting/ to know what you can do.\n")

	case tor.ErrTorVersionNotCompatible:
		return i18n().Sprintf("The current version of Tor is incompatible with Wahay.\n")

	case tor.ErrInvalidConfiguredTorBinary:
		return i18n().Sprintf("The configured path to the Tor binary is not valid or can't be used.\n\n" +
			"Please configure another path or download a bundled Wahay with Tor in the following url:" +
			"\n\nhttps://wahay.org/download.html\n")

	case tor.ErrInvalidTorPath:
	default:
		return i18n().Sprintf("No valid Tor binary found on the system in order to run Wahay.\n")
	}

	return err.Error()
}
