module go.fd.io/govpp

go 1.18

require (
	github.com/bennyscetbun/jsongo v1.1.0
	github.com/fatih/color v1.13.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/ftrvxmtrx/fd v0.0.0-20150925145434-c6d800382fff
	github.com/lunixbochs/struc v0.0.0-20200521075829-a4cb8d33dbbe
	github.com/onsi/gomega v1.19.0
	github.com/pkg/profile v1.2.1
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/text v0.3.7
)

require (
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220503163025-988cb79eb6c6 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// Versions v0.5.0 and older use old module path git.fd.io/govpp.git
retract [v0.1.0, v0.5.0]
