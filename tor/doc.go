/*

Package tor implements functionality necessary to manage the lifecycle of Tor instances.

There are three main types that are exposed from this package and which can be used to control Tor. These are the
Instance, Control and Service types. The Instance type represents a running instance of Tor. It can either be a system
Tor instance, or an instance that the Tor package controls. The Tor package has an internal singleton Instance that
represents the one Tor Instance currently being used. The Control interface is tightly connected with an Instance -
specifically, the Control interface represents the actions that can be used to control the Tor instance over the Tor
control port. Finally, the Service interface represents a running executable which was started using a Tor instance to
allow all of its traffic to be proxied over the specific Tor instance.

In order to start using this package, you must first initialize a Tor instance. This is done by calling
InitializeInstance. This function will try to figure out different ways of running Tor - first trying to use the system
Tor instance, and if that's not possible, start its own Tor instance that we have the possibility of controlling. The
InitializeInstance function should only be called once, at startup.

All the top level functions that use Tor in this package, such as NewOnionServiceWithMultiplePorts and NewService, will
use the instance and controller inside of the global instance in the package.

*/
package tor
