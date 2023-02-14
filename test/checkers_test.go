package test

import (
	. "gopkg.in/check.v1"
	"testing"
)

type testSuite struct{}

var _ = Suite(&testSuite{})

func Test(t *testing.T) { TestingT(t) }

func (s *testSuite) Test_stringMethodChecker_Check_returnsTheBooleanResultFromTheCheckFunction(c *C) {
	toReturn := false
	check := func(string, string) bool {
		return toReturn
	}

	sMC := &stringMethodChecker{}
	sMC.check = check

	result, _ := sMC.Check([]interface{}{"", ""}, nil)
	c.Assert(result, Equals, false)

	toReturn = true
	result, _ = sMC.Check([]interface{}{"", ""}, nil)
	c.Assert(result, Equals, true)
}

func (s *testSuite) Test_stringMethodChecker_Check_callsTheCheckFunctionWithTheTwoFirstItemsInTheSliceOfArguments(c *C) {
	sMC := &stringMethodChecker{}

	called := false
	sMC.check = func(p1, p2 string) bool {
		called = true
		c.Assert(p1, Equals, "a first value")
		c.Assert(p2, Equals, "a second value")
		return false
	}

	_, _ = sMC.Check([]interface{}{"a first value", "a second value"}, nil)
	c.Assert(called, Equals, true)

	called = false
	sMC.check = func(p1, p2 string) bool {
		called = true
		c.Assert(p1, Equals, "some more value")
		c.Assert(p2, Equals, "even more value")
		return false
	}

	_, _ = sMC.Check([]interface{}{"some more value", "even more value"}, nil)
	c.Assert(called, Equals, true)
}

func (s *testSuite) Test_stringMethodChecker_Check_returnsAnErrorIfTheExpectedValueIsNotAString(c *C) {
	sMC := &stringMethodChecker{}

	_, er := sMC.Check([]interface{}{"a first value", 42}, nil)
	c.Assert(er, Equals, "Expected must be a string")
}

func (s *testSuite) Test_stringMethodChecker_Check_returnsAnErrorIfTheValueIsNotAString(c *C) {
	sMC := &stringMethodChecker{}

	_, er := sMC.Check([]interface{}{23, "expected value"}, nil)
	c.Assert(er, Equals, "Obtained value is not a string and has no .String()")
}

type somethingStringable struct{}

func (*somethingStringable) String() string { return "hello world" }

func (s *testSuite) Test_stringMethodChecker_Check_callToTheCheckFunctionIfTheValueCanBeConvertedToAnInterfaceThatImplementsTheStringMethod(c *C) {
	sMC := &stringMethodChecker{}

	called := false
	sMC.check = func(s string, s2 string) bool {
		called = true
		return true
	}

	_, e := sMC.Check([]interface{}{&somethingStringable{}, "expected value"}, nil)
	c.Assert(e, Equals, "")
	c.Assert(called, Equals, true)
}

func (s *testSuite) Test_stringMethodChecker_Check_(c *C) {
	sMC := &stringMethodChecker{}

	called := false
	sMC.check = func(p string, p1 string) bool {
		called = true
		c.Assert(p, Equals, "hello world")
		c.Assert(p1, Equals, "expected value")
		return true
	}

	_, e := sMC.Check([]interface{}{&somethingStringable{}, "expected value"}, nil)

	c.Assert(e, Equals, "")
	c.Assert(called, Equals, true)
}
