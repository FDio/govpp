package consumer

import (
	_ "git.fd.io/govpp.git/binapi"
)

//go:generate go run ../vpplink/cmd/ --binapi-package "git.fd.io/govpp.git/binapi" --output-dir ./impl
