package tor

import "github.com/wybiral/torgo"

type torgoController interface {
	AuthenticatePassword(string) error
	AddOnion(*torgo.Onion) error
	GetVersion() (string, error)
	DeleteOnion(string) error
}
