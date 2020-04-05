package tor

import (
	. "gopkg.in/check.v1"
)

type TorAcceptanceSuite struct{}

var _ = Suite(&TorAcceptanceSuite{})

// These tests will try to document the expected behavior
// with regard to what Tor instance will be used or not used
// depending on what the system figures out about what
// is going on

// Specifically, a method should be called ONCE
// at system start to initialize the Tor subsystem. This method
// will first try to detect circumstances of the system it's
// running on, and then based on that create an instance that
// will be used for all subsequent Tor connections.

// The rules look more or less like this:
//   If Tor is available and running already:
//     - Test authentication method NONE
//     - Test authentication method COOKIE
//     - Test authentication method PASSWORD
//   If any of the authentication methods succeed:
//     - Check that the version of Tor is acceptable
//     - Check that Tor is actually connected to the
//       internet and can be used to do connections
//
//   If the System Tor is is acceptable, use it, with the
//   detected authentication method. In this case, do
//   not try to create a configuration file or data dir
//   for Tor. Also, do not try to stop it at the end of
//   Wahay.
//
//   If the System Tor is not possible to use, check whether
//   the Tor binary is available and acceptable (version is correct)
//
//   If the System Tor is not acceptable or not available, try to
//   find another Tor executable that can be used.
//
//   If no acceptable Tor executable is found, we have to give up
//
//   If an acceptable Tor executable is found, create a new data dir
//   and configuration file. Start Tor with this.
//   Run checks to make sure it's acceptable. If not, we have to give up.
//
//   At the end of Wahay, when running a custom Tor instance, stop the Tor
//   instance. Also clean up and destroy the created data directory
