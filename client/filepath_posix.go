//go:build !windows

package client

func generateDestination(path string) string {
	return filepathJoin(path, "mumble")
}
