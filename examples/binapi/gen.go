package binapi

// Generate Go code from the VPP APIs located in the /usr/share/vpp/api directory.

//go:generate binapi-generator --import-prefix=git.fd.io/govpp.git/examples/binapi af_packet interface ip memclnt vpe sr acl memif ip_types fib_types
