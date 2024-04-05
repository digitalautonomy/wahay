package hosting

import (
	"errors"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type hostingSuite struct{}

var _ = Suite(&hostingSuite{})

func noErrHelper() error {
	return nil
}
func errHelper() error {
	return errors.New("error2")
}
