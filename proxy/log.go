package proxy

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	debug = os.Getenv("DEBUG_GOVPP_PROXY") != ""

	log = logrus.New()
)

func init() {
	if debug {
		log.Level = logrus.DebugLevel
		log.Debugf("govpp/proxy: debug mode enabled")
	}
}

// SetLogger sets logger.
func SetLogger(l *logrus.Logger) {
	log = l
}

// SetLogLevel sets log level for logger.
func SetLogLevel(lvl logrus.Level) {
	log.Level = lvl
}

// SetOutput sets log output for logger.
func SetLogOutput(out io.Writer) {
	log.Out = out
}
