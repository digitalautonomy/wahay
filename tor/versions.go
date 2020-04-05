package tor

import (
	"errors"
	"strconv"
	"strings"
)

// minSupportedVersion is the Tor minimun supported version
const minSupportedVersion = "0.3.2"

// Simple utilities to manage version comparisons

func errorsAny(es ...error) error {
	for _, e := range es {
		if e != nil {
			return e
		}
	}
	return nil
}

func parseVersion(v string) (major, minor, patch int, err error) {
	res := strings.Split(v, ".")

	if len(res) < 3 {
		err = errors.New("invalid version string")
		return
	}

	mj, e1 := strconv.ParseInt(res[0], 10, 16)
	mn, e2 := strconv.ParseInt(res[1], 10, 16)
	pa, e3 := strconv.ParseInt(res[2], 10, 16)

	if errorsAny(e1, e2, e3) != nil {
		err = errors.New("invalid version number")
		return
	}

	return int(mj), int(mn), int(pa), nil
}

func cmpInt(l, r int) int {
	if l == r {
		return 0
	}
	if l < r {
		return -1
	}
	return 1
}

func compareVersions(v1 string, v2 string) (diff int, err error) {
	v1major, v1minor, v1patch, e1 := parseVersion(v1)
	v2major, v2minor, v2patch, e2 := parseVersion(v2)

	if errorsAny(e1, e2) != nil {
		err = errors.New("invalid version string")
		return
	}

	majorDiff := cmpInt(v1major, v2major)
	minorDiff := cmpInt(v1minor, v2minor)
	patchDiff := cmpInt(v1patch, v2patch)

	if majorDiff != 0 {
		return majorDiff, nil
	}

	if minorDiff != 0 {
		return minorDiff, nil
	}

	return patchDiff, nil
}
