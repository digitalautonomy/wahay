package test

import (
	. "gopkg.in/check.v1"
)

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
