package binapi

// Generate Go code from the VPP APIs located in the /usr/share/vpp/api directory.

//go:generate binapi-generator -include-services=false memclnt
