package consumer

import (
	_ "go.fd.io/govpp/binapi"
)

//go:generate go run go.fd.io/govpp/cmd/binapi-generator --vpp $VPP_DIR -o ./impl
