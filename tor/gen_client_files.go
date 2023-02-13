package tor

import (
	_ "embed"
)

//go:embed files/torrc
var torrcContent string

func getTorrc() string {
	return torrcContent
}

//go:embed files/torrc-logs
var torrclogsContent string

func getTorrcLogConfig() string {
	return torrclogsContent
}
