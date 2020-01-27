package gui

import (
	"errors"

	"autonomia.digital/tonio/app/config"
	"github.com/coyim/gotk3adapter/gtki"
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
		r := config.GenerateKeysBasedOnPassw(password, p)
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

// This function should be only called on startup, never call this function
// or should not be called during the execution of this app
func (u *gtkUI) getMasterPassword(p config.EncryptionParameters) config.EncryptionResult {
	u.hideLoadingWindow()

	passwordResultCh := make(chan string)

	builder := u.g.uiBuilderFor("MasterPasswordWindow")
	win := builder.get("masterPasswordWindow").(gtki.Window)
	txtPassword := builder.get("entryPassword").(gtki.Entry)
	btnTogglePassword := builder.get("btnTogglePassword").(gtki.CheckButton)

	win.SetApplication(u.app)

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
				u.closeApplication()
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

	// Show the loading window while checking entered the password
	u.displayLoadingWindow()

	if len(password) == 0 {
		return config.EncryptionResult{}
	}

	return config.GenerateKeysBasedOnPassw(password, p)
}

func (u *gtkUI) captureMasterPassword(onSuccess func(), onCancel func()) {
	builder := u.g.uiBuilderFor("MasterPasswordWindow")
	passwordWindow := builder.get("captureMasterPassword").(gtki.Window)

	txtPassword := builder.get("txtPassword").(gtki.Entry)
	txtPasswordRepeat := builder.get("txtPasswordRepeat").(gtki.Entry)

	lblValidation := builder.get("lblValidation").(gtki.Label)
	lblValidation.SetVisible(false)

	isValidPassword := false

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
		"on_cancel": func() {
			passwordWindow.Destroy()
			u.enableWindow(u.currentWindow)
			if !isValidPassword {
				onCancel()
			}
		},
		"on_close": func() {
			passwordWindow.Destroy()
			u.enableWindow(u.currentWindow)
			if !isValidPassword {
				onCancel()
			}
		},
	})

	// u.currentWindow should be the settings window
	u.disableWindow(u.currentWindow)
	passwordWindow.SetTransientFor(u.currentWindow)
	passwordWindow.Show()
}

func validatePasswords(pass1, pass2 string) error {
	if len(pass1) == 0 {
		return errors.New("please enter a valid password")
	}

	if len(pass2) == 0 {
		return errors.New("enter the password confirmation")
	}

	if pass1 != pass2 {
		return errors.New("passwords do not match")
	}

	if len(pass1) < passwordMinSize {
		return errors.New("enter a password of 6 chars of minimun length")
	}

	return nil
}
