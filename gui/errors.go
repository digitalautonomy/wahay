package gui

import (
	"fmt"
)

func fatal(v interface{}) {
	panic(fmt.Sprintf("failing on error: %v", v))
}

func fatalf(format string, v ...interface{}) {
	//	log.Printf(format, v...)
	panic(fmt.Sprintf(format, v...))
}
