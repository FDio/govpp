package core

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	DebugEnvVar = "DEBUG_GOVPP"

	debugOptConn     = "conn"
	debugOptMsgId    = "msgid"
	debugOptChannels = "channels"
)

var (
	debugOn  = os.Getenv(DebugEnvVar) != ""
	debugMap = initDebugMap(os.Getenv(DebugEnvVar))

	log = logrus.New()
)

// init initializes global logger
func init() {
	if os.Getenv("DEBUG_GOVPP_CONN") != "" {
		debugMap[debugOptConn] = "true"
	}
	log.Formatter = &logrus.TextFormatter{
		EnvironmentOverrideColors: true,
	}
	if debugOn {
		log.Level = logrus.DebugLevel
		log.Debugf("govpp: debug enabled %+v", debugMap)
	}
}

func isDebugOn(u string) bool {
	_, ok := debugMap[u]
	return ok
}

func initDebugMap(s string) map[string]string {
	debugMap := make(map[string]string)
	for _, p := range splitString(s) {
		var key, val string
		kv := strings.SplitN(p, "=", 2)
		key = kv[0]
		if len(kv) > 1 {
			val = kv[1]
		} else {
			val = "true"
		}
		debugMap[key] = val
	}
	return debugMap
}

func splitString(s string) []string {
	return strings.FieldsFunc(s, func(c rune) bool {
		switch c {
		case ';', ',', ' ':
			return true
		}
		return false
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
