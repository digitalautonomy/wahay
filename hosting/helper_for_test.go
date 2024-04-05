package hosting

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type hostingSuite struct{}

var _ = Suite(&hostingSuite{})
