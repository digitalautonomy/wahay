package tor

import "github.com/wybiral/torgo"

type torgoController interface {
	AuthenticateCookie() error
	AddOnion(*torgo.Onion) error
	GetVersion() (string, error)
	DeleteOnion(string) error
}
