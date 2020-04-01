# Wahay -  easy-to-use, secure and decentralized conference calls

[Wahay](https://wahay.org) is an application that allows you to easily host and participate in conference calls, without the need for any
centralized servers or services. We are building a voice call application that is meant to be as easy-to-use as possible, while still
providing extremely high security and privacy out of the box.

In order to do this, we use [Tor](https://torproject.org) Onion Services in order to communicate between the end-points, and we use the
[Mumble](https://www.mumble.info) protocol for the actual voice communication. We are doing extensive user testing in order to ensure that
the usability of the application is as good as possible.

## Bundle Distribution

In order to make Wahay simple to use, we have created bundles that allows the use the application without installing it on the system or requiring root privileges. For this we are including Tor and Mumble with their respective libraries.

In the case of Tor we have downloaded the Tor Browser Bundle for Linux and copy the folder Tor to Wahay’s bundles. This folder includes the Tor binary and the libraries it needs to work. The respective libraries are located at the “Licenses/Tor/” directory inside this bundle.

For Mumble, we have copied the binary from a Linux distribution and follow the instructions to create a [portable mumble](https://wiki.mumble.info/wiki/Mumble_Portable_Linux). The Mumble bundle is dependent of the distribution and right now Debian Buster, Fedora 30, Fedora 31 and Ubuntu 18.04 are supported. The respective libraries are located at the “Licenses/Mumble/” directory inside this bundle.
