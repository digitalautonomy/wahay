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
	return stringMethodCheck(params[0], params[1], checker.check)
}

func stringMethodCheck(value, expected interface{}, f func(string, string) bool) (result bool, error string) {
	exStr, ok := expected.(string)
	if !ok {
		return false, "Expected must be a string"
	}
	valueStr, valueIsStr := value.(string)
	if !valueIsStr {
		if valueWithStr, valueHasStr := value.(fmt.Stringer); valueHasStr {
			valueStr, valueIsStr = valueWithStr.String(), true
		}
	}
	if valueIsStr {
		checkOk := f(valueStr, exStr)
		return checkOk, ""
	}
	return false, "Obtained value is not a string and has no .String()"
}
