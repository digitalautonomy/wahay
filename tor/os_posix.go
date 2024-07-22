//go:build !windows

package tor

func searchProxyTool() error {
	return findTorsocksBinary()
}
