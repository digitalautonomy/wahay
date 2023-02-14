package config

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type boolEqualsChecker struct {
	*CheckerInfo
	value bool
}

var IsTrue Checker = &boolEqualsChecker{
	&CheckerInfo{Name: "IsTrue", Params: []string{"value"}},
	true,
}

var IsFalse Checker = &boolEqualsChecker{
	&CheckerInfo{Name: "IsFalse", Params: []string{"value"}},
	false,
}

func (checker *boolEqualsChecker) Check(params []interface{}, names []string) (result bool, error string) {
	ob, ok := params[0].(bool)
	if !ok {
		return false, "Obtained value is not a boolean"
	}
	return ob == checker.value, ""
}

type checkerSuite struct{}

var _ = Suite(&checkerSuite{})

func (s *checkerSuite) Test_boolEqualsChecker_Check_comparesABooleanValue(c *C) {
	ch := &boolEqualsChecker{value: true}

	res, e := ch.Check([]interface{}{false}, nil)
	c.Assert(res, Equals, false)
	c.Assert(e, Equals, "")

	res, e = ch.Check([]interface{}{true}, nil)
	c.Assert(res, Equals, true)
	c.Assert(e, Equals, "")

	ch = &boolEqualsChecker{value: false}

	res, e = ch.Check([]interface{}{false}, nil)
	c.Assert(res, Equals, true)
	c.Assert(e, Equals, "")

	res, e = ch.Check([]interface{}{true}, nil)
	c.Assert(res, Equals, false)
	c.Assert(e, Equals, "")
}

func (s *checkerSuite) Test_boolEqualsChecker_Check_failsIfObtainedValueIsNotBool(c *C) {
	ch := &boolEqualsChecker{value: true}
	res, e := ch.Check([]interface{}{42}, nil)
	c.Assert(res, Equals, false)
	c.Assert(e, Equals, "Obtained value is not a boolean")
}
