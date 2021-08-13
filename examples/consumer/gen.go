package consumer

import (
	_ "git.fd.io/govpp.git/binapi"
)

//go:generate go run git.fd.io/govpp.git/cmd/binapi-generator --vpp $VPP_DIR -o ./impl
