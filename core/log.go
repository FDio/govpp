package core

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	DebugEnvVar = "DEBUG_GOVPP"

	debugOptCore     = "core"
	debugOptConn     = "conn"
	debugOptMsgId    = "msgid"
	debugOptChannels = "channels"
)

var (
	debugMap map[string]struct{}
	debugOn  bool

	log *logrus.Logger
)

// init initializes global logger
func init() {
	debugMap = initDebugMap(os.Getenv(DebugEnvVar))
	debugOn = isDebugOn(debugOptCore)

	log = logrus.New()
	log.Formatter = &logrus.TextFormatter{
		EnvironmentOverrideColors: true,
	}
	if debugOn {
		log.Level = logrus.DebugLevel
		log.Debugf("govpp: debug enabled %v", debugMap)
	}
}

func isDebugOn(u string) bool {
	_, ok := debugMap[u]
	return ok
}

func initDebugMap(s string) map[string]struct{} {
	debugSet := make(map[string]struct{})
	for _, p := range splitString(s) {
		key := strings.SplitN(p, "=", 2)[0] // We only need the key
		debugSet[key] = struct{}{}
	}
	return debugSet
}

func splitString(s string) []string {
	return strings.FieldsFunc(s, func(c rune) bool {
		return c == ';' || c == ','
	})
}

// SetLogger sets global logger to l.
func SetLogger(l *logrus.Logger) {
	log = l
}

// SetLogLevel sets global logger level to lvl.
func SetLogLevel(lvl logrus.Level) {
	log.Level = lvl
}
