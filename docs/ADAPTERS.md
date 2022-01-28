# Adapters

## Pure-Go implementation (recommended)

GoVPP now supports pure Go implementation for VPP binary API. This does
not depend on CGo or any VPP library and can be easily compiled.

There are two packages providing pure Go implementations in GoVPP:
- [`socketclient`](adapter/socketclient) - for VPP binary API (via unix socket)
- [`statsclient`](adapter/statsclient) - for VPP stats API (via shared memory)

## CGo wrapper for vppapiclient library (deprecated)

GoVPP also provides vppapiclient package which actually uses
`vppapiclient.so` library from VPP codebase to communicate with VPP API.
To build GoVPP, `vpp-dev` package must be installed,
either [from packages][from-packages] or [from sources][from-sources].

To build & install `vpp-dev` from sources:

```sh
git clone https://gerrit.fd.io/r/vpp
cd vpp
make install-dep
make pkg-deb
cd build-root
sudo dpkg -i vpp*.deb
```

To build & install GoVPP:

```sh
go get -u git.fd.io/govpp.git
cd $GOPATH/src/git.fd.io/govpp.git
make test
make install
```
