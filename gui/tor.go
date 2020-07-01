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

		instance, e := tor.InitializeInstance(u.config)
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
		return "ErrTorBinaryNotFound description"

	case tor.ErrTorInstanceCantStart:
		return "ErrTorInstanceCantStart description"

	case tor.ErrTorConnectionTimeout:
		return "ErrTorConnectionTimeout description"

	case tor.ErrPartialTorNoControlPort:
		return "ErrPartialTorNoControlPort description"

	case tor.ErrPartialTorNoValidAuth:
		return "ErrPartialTorNoValidAuth description"

	case tor.ErrFatalTorNoConnectionAllowed:
		return "ErrFatalTorNoConnectionAllowed description"

	case tor.ErrInvalidTorPath:
		return "ErrInvalidTorPath description"

	case tor.ErrTorVersionNotCompatible:
		return "ErrTorVersionNotCompatible description"

	case tor.ErrInvalidConfiguredTorBinary:
		return "ErrInvalidConfiguredTorBinary description"

	case tor.ErrTorsocksNotInstalled:
		return i18n.Sprintf("Ensure you have installed Torsocks in your system.\n\n" +
			"For more information please visit:\n\nhttps://trac.torproject.org/projects/tor/wiki/doc/torsocks")

	case errTorNoBinary:
		return "errTorNoBinary description"
	}

	return err.Error()
}
