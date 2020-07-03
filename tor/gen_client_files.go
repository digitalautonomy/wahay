//go:generate ../.build-tools/esc -o files.go -modtime 1489449600 -pkg tor files/

package tor

import (
	"github.com/digitalautonomy/wahay/codegen"
)

func getTorrc() string {
	return codegen.GetFileWithFallback("torrc", "tor/files", FSString)
}

func getTorrcLogConfig() string {
	return codegen.GetFileWithFallback("torrc-logs", "tor/files", FSString)
}
