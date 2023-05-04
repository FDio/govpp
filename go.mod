module go.fd.io/govpp

go 1.18

require (
	github.com/bennyscetbun/jsongo v1.1.1
	github.com/fsnotify/fsnotify v1.4.9
	github.com/ftrvxmtrx/fd v0.0.0-20150925145434-c6d800382fff
	github.com/ghodss/yaml v1.0.0
	github.com/gookit/color v1.5.2
	github.com/lunixbochs/struc v0.0.0-20200521075829-a4cb8d33dbbe
	github.com/mitchellh/go-ps v1.0.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/gomega v1.19.0
	github.com/pkg/profile v1.5.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/text v0.7.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// Versions v0.5.0 and older use old module path git.fd.io/govpp.git
retract [v0.1.0, v0.5.0]
