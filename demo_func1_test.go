package main

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type TonioFunc1Suite struct{}

var _ = Suite(&TonioFunc1Suite{})

func (s *TonioFunc1Suite) Test_func1_adds(c *C) {
	result, e := func1(1, 42)
	c.Assert(e, IsNil)
	c.Assert(result, Equals, 43)

	result, e = func1(101, 42)
	c.Assert(e, ErrorMatches, "too large number")
	c.Assert(result, Equals, 0)
}

type TonioFunc2Suite struct{}

var _ = Suite(&TonioFunc2Suite{})

func (s *TonioFunc2Suite) Test_func2_returnsCorrectValue(c *C) {
	c.Assert(func2(), Equals, 42)
}
