package gui

import (
	"errors"

	"github.com/coyim/gotk3adapter/gtki"
	"github.com/digitalautonomy/wahay/config"
)

const passwordMinSize = 1

type onetimeSavedPassword struct {
	savedPassword  string
	realKeySuplier config.KeySupplier
}

func (o *onetimeSavedPassword) GenerateKey(p config.EncryptionParameters) config.EncryptionResult {
	// We should check the password is checked only ONE time
	if len(o.savedPassword) > 0 {
		password := o.savedPassword
		o.savedPassword = ""

		// Before returning the result, we update the data for the
		// real key supplier so we remove the plain password from memory
		// but we keep the password key during this session
		r := config.GenerateKeysBasedOnPassword(password, p)
		_ = o.realKeySuplier.CacheFromResult(r)

		return r
	}

	return o.realKeySuplier.GenerateKey(p)
}

func (o *onetimeSavedPassword) Invalidate() {
	o.realKeySuplier.Invalidate()
}

func (o *onetimeSavedPassword) LastAttemptFailed() {
	o.realKeySuplier.LastAttemptFailed()
}

func (o *onetimeSavedPassword) CacheFromResult(r config.EncryptionResult) error {
	return o.realKeySuplier.CacheFromResult(r)
}

func (u *gtkUI) getMasterPasswordBuilder() *uiBuilder {
	builder := u.g.uiBuilderFor("MasterPasswordWindow")

	builder.i18nProperties(
		"title", "captureMasterPassword",
		"label", "lblPasswordIntro",
		"placeholder", "txtPassword",
		"placeholder", "txtPasswordRepeat",
		"button", "btnCancel",
		"button", "btnContinue",

		"title", "masterPasswordWindow",
		"label", "lblMasterPasswordIntro",
		"label", "lblError",
		"placeholder", "entryPassword",
		"tooltip", "btnTogglePassword",
		"label", "lblMasterPasswordText",
		"button", "btnMasterPasswordCancel",
		"button", "btnMasterPasswordContinue")

	return builder
}

// This function should be only called on startup, never call this function
// or should not be called during the execution of this app
func (u *gtkUI) getMasterPassword(p config.EncryptionParameters, lastAttemptFailed bool) config.EncryptionResult {
	u.hideLoadingWindow()

	passwordResultCh := make(chan string)

	builder := u.getMasterPasswordBuilder()

	win := builder.get("masterPasswordWindow").(gtki.Window)
	txtPassword := builder.get("entryPassword").(gtki.Entry)
	btnTogglePassword := builder.get("btnTogglePassword").(gtki.CheckButton)

	win.SetApplication(u.app)

	lblError := builder.get("lblError").(gtki.Label)
	lblError.SetVisible(lastAttemptFailed)

	hadSubmission := false
	togglePassword := btnTogglePassword.GetActive()

	builder.ConnectSignals(map[string]interface{}{
		"on_toggle_password": func() {
			togglePassword = !togglePassword
			txtPassword.SetVisibility(togglePassword)
		},
		"on_cancel": func() {
			if !hadSubmission {
				hadSubmission = true
				close(passwordResultCh)
				u.quit()
			}
		},
		"on_save": func() {
			if !hadSubmission {
				hadSubmission = true
				text, err := txtPassword.GetText()
				if err != nil || len(text) == 0 {
					// TODO: show alert warning the user
					hadSubmission = false
				} else {
					txtPassword.SetSensitive(false)
					passwordResultCh <- text
					close(passwordResultCh)
				}
			}
		},
	})

	u.doInUIThread(win.Show)

	password := <-passwordResultCh

	u.doInUIThread(win.Destroy)

	u.displayLoadingWindow()

	if len(password) == 0 {
		return config.EncryptionResult{}
	}

	return config.GenerateKeysBasedOnPassword(password, p)
}

func (u *gtkUI) captureMasterPassword(onSuccess func(), onCancel func()) {
	builder := u.getMasterPasswordBuilder()

	passwordWindow := builder.get("captureMasterPassword").(gtki.Window)

	txtPassword := builder.get("txtPassword").(gtki.Entry)
	txtPasswordRepeat := builder.get("txtPasswordRepeat").(gtki.Entry)

	lblValidation := builder.get("lblValidation").(gtki.Label)
	lblValidation.SetVisible(false)

	isValidPassword := false
	cleanup := func() {
		u.doInUIThread(func() {
			passwordWindow.Destroy()
			u.enableWindow(u.currentWindow)
		})
		if !isValidPassword {
			onCancel()
		}
	}

	builder.ConnectSignals(map[string]interface{}{
		"on_save": func() {
			lblValidation.SetVisible(true)

			password, err1 := txtPassword.GetText()
			repeat, err2 := txtPasswordRepeat.GetText()

			if err1 != nil || err2 != nil {
				passwordWindow.Destroy()
				u.enableWindow(u.currentWindow)
				onCancel()
			}

			err := validatePasswords(password, repeat)
			if err != nil {
				txtPassword.GrabFocus()
				lblValidation.SetText(err.Error())
				lblValidation.SetVisible(true)
			} else {
				u.keySupplier = &onetimeSavedPassword{
					savedPassword:  password,
					realKeySuplier: u.keySupplier,
				}
				isValidPassword = true
				passwordWindow.Destroy()
				u.enableWindow(u.currentWindow)
				onSuccess()
			}
		},
		"on_cancel": cleanup,
		"on_close":  cleanup,
	})

	if u.currentWindow != nil {
		u.disableWindow(u.currentWindow)
		passwordWindow.SetTransientFor(u.currentWindow)
	}

	u.doInUIThread(passwordWindow.Show)
}

func validatePasswords(pass1, pass2 string) error {
	if len(pass1) == 0 {
		return errors.New(i18n.Sprintf("please enter a valid password"))
	}

	if len(pass2) == 0 {
		return errors.New(i18n.Sprintf("enter the password confirmation"))
	}

	if pass1 != pass2 {
		return errors.New(i18n.Sprintf("passwords do not match"))
	}

	if len(pass1) < passwordMinSize {
		return errors.New(i18n.Sprintf("enter a password at least 6 characters long"))
	}

	return nil
}
