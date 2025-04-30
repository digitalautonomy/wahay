<p>
  <a href="https://wahay.org/" target="_blank" rel="noopener noreferrer">
    <img src="https://wahay.org/assets/img/logo.svg" alt="Wahay" />
  </a><br><br>
  An easy-to-use, secure and decentralized conference call application.
</p>

<p>
  <a href="https://github.com/digitalautonomy/wahay/issues"><img alt="GitHub issues" src="https://img.shields.io/github/issues/digitalautonomy/wahay"></a>
  <a href="https://github.com/digitalautonomy/wahay/network"><img alt="GitHub forks" src="https://img.shields.io/github/forks/digitalautonomy/wahay"></a>
  <a href="https://github.com/digitalautonomy/wahay/stargazers"><img alt="GitHub stars" src="https://img.shields.io/github/stars/digitalautonomy/wahay"></a>
  <a href="https://www.gnu.org/licenses/gpl-3.0"><img alt="License: GPL v3" src="https://img.shields.io/badge/License-GPLv3-blue.svg"></a>
</p>

------

[Wahay](https://wahay.org) is an application that allows you to easily host and participate in conference calls, without the need for any
centralized servers or services. We are building a voice call application that is meant to be as easy-to-use as possible, while still
providing extremely high security and privacy out of the box.

In order to do this, we use [Tor](https://torproject.org) Onion Services in order to communicate between the end-points, and we use the [Mumble](https://www.mumble.info) protocol for the actual voice communication. We are doing extensive user testing in order to ensure that the usability of the application is as good as possible.

## Documentation

For full documentation, visit [wahay.org](https://wahay.org/documentation/index.html).

## Installing

For end-users, please refer to installation instructions on the
[website](https://wahay.org/documentation/getting-started/installation/). We provide several different options for
installation there.

## Building from Source

To build Wahay from source, you'll need to install the following dependencies:

1. Install build dependencies:
```bash
sudo apt install make pkg-config golang libgtk-3-dev ruby-sass
```

2. Install runtime dependencies:
```bash
sudo apt install tor mumble
```

3. Clone the repository

4. Build the project:
```bash
make build
```

The build process will create the binary in the `bin` directory. You can run it with:
```bash
./bin/wahay
```

## Security warning

Wahay is currently under active development. There have been no security audits
of the code, and you should currently not use this for anything sensitive.

## Language

The language to be used is the same configured under `LANG` environment variable.

Example:

```bash
$ export LANG="en_US.utf8"
```

## Compatibility

The current version of Wahay is compatible with all major Linux distributions. It is possible that the application can
run on macOS or Windows, but at the moment we have not tested it. We are planning on adding official macOS and Windows
compatibility in the near future.

## About the developers

Wahay is developed by the NGO [Centro de Autonomía Digital](https://autonomia.digital), based in Quito, Ecuador.

## License

Wahay is licensed under the [GPL version 3](https://www.gnu.org/licenses/gpl-3.0.html).
