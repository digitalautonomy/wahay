package main

import "errors"

// This file is only for demonstration, until we have actual production code
// Please delete when real code exists

func func1(one, two int) (int, error) {
	if one > 100 {
		return 0, errors.New("too large number")
	}

	if two > 100 {
		return 0, errors.New("another too large number")
	}

	return one + two, nil
}
