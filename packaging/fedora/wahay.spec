Name:    wahay
Version: 20200304
Release: 1%{?dist}
Summary: Easy-to-use, secure and decentralized conference call app

License: GPLv3
Source0: wahay-20200304.tar.bz2
Requires: mumble, tor >= 0.3.5, xclip
BuildRequires:  compiler(go-compiler)
BuildRequires:  desktop-file-utils
BuildRequires:  gtk3-devel

%description
 Wahay -  easy-to-use, secure and decentralized conference calls Wahay
 (https://wahay.org) is an application that allows you to easily host and
 participate in conference calls, without the need for any centralized
 servers or services. We are building a voice call application that is
 meant to be as easy-to-use as possible, while still providing extremely
 high security and privacy out of the box.
 .
 In order to do this, we use Tor (https://torproject.org) Onion Services
 in order to communicate between the end-points, and we use the Mumble
 (https://www.mumble.info) protocol for the actual voice communication. We
 are doing extensive user testing in order to ensure that the usability of
 the application is as good as possible.  Installing For end-users, please
 refer to installation instructions on the website (https://wahay.org). We
 provide several different options for installation there. If you are a
 developer, installing the application should be as easy as cloning the
 repository and running make build.  Security warning Wahay is currently
 under active development. There have been no security audits of the
 code, and you should currently not use this for anything sensitive.
 Compatibility The current version of Wahay is compatible with all major
 Linux distributions. It is possible that the application can run on OS X
 or Windows, but at this moment we have not tested this. We are planning
 on adding official OS X and Windows compatibility in the near future.
 About the developers Wahay is developed by the NGO Centro de Autonomía
 Digital (https://autonomia.digital), based in Quito, Ecuador.  License
 Wahay is licensed under the GPL version 3.

%prep
%setup -q


%build
mkdir -p src/github.com/digitalautonomy
ln -s ../../../  src/github.com/digitalautonomy/wahay
export GOPATH=$(pwd):%{gopath}

cd src/github.com/digitalautonomy/wahay
make deps
make build


%install
install -d %{buildroot}/%{_bindir}
install -p -m 755 bin/%{name} %{buildroot}/%{_bindir}
install -d %{buildroot}/%{_mandir}/man1/
install -p packaging/ubuntu/ubuntu/usr/share/man/man1/%{name}.1.gz %{buildroot}/%{_mandir}/man1/

sed "s/__NAME__/Wahay/g" gui/config_files/wahay.desktop | sed "s/__EXEC__/\/usr\/bin\/wahay/g" | sed "s/__ICON__/wahay/" | sed "s/Internet/Network/"  > %{name}.desktop

desktop-file-install --dir=${RPM_BUILD_ROOT}%{_datadir}/applications %{name}.desktop


for size in 192x192 256x256 512x512; do
  install -d %{buildroot}%{_datadir}/icons/hicolor/${size}/apps
  install -p gui/images/wahay-${size}.png %{buildroot}%{_datadir}/icons/hicolor/${size}/apps/%{name}.png
done

%files
%{_bindir}/%{name}
%license LICENSE
%doc README.md
%{_mandir}/man1/%{name}.*
%{_datadir}/icons/hicolor/*
%{_datadir}/applications/*.desktop


%changelog
