package core

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	debug       = os.Getenv("DEBUG_GOVPP") != ""
	debugMsgIDs = strings.Contains(os.Getenv("DEBUG_GOVPP"), "msgid")

	log = logrus.New()
)

// init initializes global logger
func init() {
	log.Formatter = &logrus.TextFormatter{
		EnvironmentOverrideColors: true,
	}
	if debug {
		log.Level = logrus.DebugLevel
		log.Debugf("govpp: debug level enabled")
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
