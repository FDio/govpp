module go.fd.io/govpp

go 1.18

require (
	github.com/bennyscetbun/jsongo v1.1.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/ftrvxmtrx/fd v0.0.0-20150925145434-c6d800382fff
	github.com/lunixbochs/struc v0.0.0-20200521075829-a4cb8d33dbbe
	github.com/mitchellh/go-ps v1.0.0
	github.com/onsi/gomega v1.19.0
	github.com/pkg/profile v1.2.1
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/text v0.7.0
)

require (
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/stretchr/testify v1.3.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// Versions v0.5.0 and older use old module path git.fd.io/govpp.git
retract [v0.1.0, v0.5.0]
