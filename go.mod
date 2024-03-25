module go.fd.io/govpp

go 1.18

require (
	github.com/bennyscetbun/jsongo v1.1.1
	github.com/docker/cli v26.0.0+incompatible
	github.com/fsnotify/fsnotify v1.7.0
	github.com/ftrvxmtrx/fd v0.0.0-20150925145434-c6d800382fff
	github.com/gookit/color v1.5.4
	github.com/lunixbochs/struc v0.0.0-20200521075829-a4cb8d33dbbe
	github.com/mitchellh/go-ps v1.0.0
	github.com/moby/term v0.5.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/gomega v1.32.0
	github.com/pkg/profile v1.7.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/text v0.14.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/felixge/fgprof v0.9.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20211214055906-6f57359322fd // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

// Versions v0.5.0 and older use old module path git.fd.io/govpp.git
retract [v0.1.0, v0.5.0]
