package binapi

// Generate Go code from the VPP APIs located in the /usr/share/vpp/api directory.

//go:generate binapi-generator --import-types=false mactime af_packet interface interface_types ip vpe sr acl memif ip_types fib_types
