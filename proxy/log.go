package proxy

import (
	"github.com/sirupsen/logrus"
	"os"
)

var (
	debug = os.Getenv("DEBUG_GOVPP_PROXY") != ""

	log = logrus.New()
)

func init() {
	log.Out = os.Stdout
	if debug {
		log.Level = logrus.DebugLevel
		log.Debugf("govpp/proxy: debug mode enabled")
	}
}

// SetLogger sets global logger to l.
func SetLogger(l *logrus.Logger) {
	log = l
}

// SetLogLevel sets global logger level to lvl.
func SetLogLevel(lvl logrus.Level) {
	log.Level = lvl
}