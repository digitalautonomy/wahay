package tor

type authenticationMethod func(tc torgoController) error

func authenticateNone(tc torgoController) error {
	return tc.AuthenticateNone()
}

func authenticateCookie(tc torgoController) error {
	return tc.AuthenticateCookie()
}

func authenticatePassword(password string) func(tc torgoController) error {
	return func(tc torgoController) error {
		return tc.AuthenticatePassword(password)
	}
}

func authenticateAny(am ...authenticationMethod) func(torgoController) error {
	return func(tc torgoController) error {
		var e error
		for _, a := range am {
			e = a(tc)
			if e == nil {
				return nil
			}
		}
		return e
	}
}
