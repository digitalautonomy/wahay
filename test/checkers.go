package test

import (
	"fmt"
	. "gopkg.in/check.v1"
	"strings"
)

type stringMethodChecker struct {
	*CheckerInfo
	check func(string, string) bool
}

var Contains Checker = &stringMethodChecker{
	&CheckerInfo{Name: "Contains", Params: []string{"value", "substring"}},
	strings.Contains,
}

func (checker *stringMethodChecker) Check(params []interface{}, _ []string) (result bool, error string) {
	exp, ok1 := params[1].(string)
	if !ok1 {
		return false, "Expected must be a string"
	}

	val, ok0 := params[0].(string)
	if !ok0 {
		if v, ok := params[0].(fmt.Stringer); ok {
			val = v.String()
		} else {
			return false, "Obtained value is not a string and has no .String()"
		}
	}

	return checker.check(val, exp), ""
}
