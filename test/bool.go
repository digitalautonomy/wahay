package test

import (
	. "gopkg.in/check.v1"
)

type boolEqualsChecker struct {
	*CheckerInfo
	value bool
}

func (checker *boolEqualsChecker) Check(params []interface{}, names []string) (result bool, error string) {
	ob, ok := params[0].(bool)
	if !ok {
		return false, "Obtained value is not a boolean"
	}
	return ob == checker.value, ""
}

var IsTrue Checker = &boolEqualsChecker{
	&CheckerInfo{Name: "IsTrue", Params: []string{"value"}},
	true,
}

var IsFalse Checker = &boolEqualsChecker{
	&CheckerInfo{Name: "IsFalse", Params: []string{"value"}},
	false,
}
