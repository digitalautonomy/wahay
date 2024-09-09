//go:build !windows

package client

func (c *client) binaryEnv() []string {
	// This is a temporary fix for making sure that
	// Mumble doesn't run under Wayland.
	env := []string{"QT_QPA_PLATFORM=xcb"}
	if c.isValid && c.binary != nil {
		return append(env, c.binary.envIfBundle()...)
	}
	return env
}
