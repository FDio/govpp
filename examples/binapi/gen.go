package binapi

// Generate Go code from the VPP APIs located in the /usr/share/vpp/api directory.
//go:generate -command binapigen binapi-generator --output-dir=. --include-services

// core
//go:generate binapigen --input-file=/usr/share/vpp/api/core/af_packet.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/interface.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/ip.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/memclnt.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/vpe.api.json
// plugins
//go:generate binapigen --input-file=/usr/share/vpp/api/plugins/acl.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/plugins/memif.api.json

// VPP version
//go:generate sh -ec "dpkg-query -f '$DOLLAR{Version}' -W vpp > VPP_VERSION"
