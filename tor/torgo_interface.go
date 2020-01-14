package tor

import "github.com/wybiral/torgo"

type torgoController interface {
	AuthenticatePassword(string) error
	AuthenticateCookie() error
	AuthenticateNone() error
	AddOnion(*torgo.Onion) error
	GetVersion() (string, error)
	DeleteOnion(string) error
}
