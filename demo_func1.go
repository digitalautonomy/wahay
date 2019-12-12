package main

import "errors"

func func1(one, two int) (int, error) {
	if one > 100 {
		return 0, errors.New("too large number")
	}
	return one + two, nil
}
