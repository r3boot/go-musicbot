package log

import (
	"github.com/sirupsen/logrus"
	"strings"
)

type Fields map[string]interface{}

var (
	logger *logrus.Logger
)

func NewLogger(level string, useJson bool) {
	logger = logrus.New()

	if useJson {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	switch strings.ToUpper(level) {
	case "INFO":
		{
			Infof(Fields{
				"package":   "log",
				"function":  "NewLogger",
				"log_level": "info",
			}, "loglevel set")
			logger.SetLevel(logrus.InfoLevel)
		}
	case "DEBUG":
		{
			Infof(Fields{
				"package":   "log",
				"function":  "NewLogger",
				"log_level": "debug",
			}, "loglevel set")
			logger.SetLevel(logrus.DebugLevel)
		}
	}
}

func GetLogger() logrus.Logger {
	return *logger
}

func toLogrusFields(fields Fields) logrus.Fields {
	logFields := logrus.Fields{}
	for k, v := range fields {
		logFields[k] = v
	}
	return logFields
}

func Infof(fields Fields, msg string, args ...interface{}) {
	logger.WithFields(toLogrusFields(fields)).Infof(msg, args...)
}

func Debugf(fields Fields, msg string, args ...interface{}) {
	logger.WithFields(toLogrusFields(fields)).Debugf(msg, args...)
}

func Warningf(fields Fields, msg string, args ...interface{}) {
	logger.WithFields(toLogrusFields(fields)).Warningf(msg, args...)
}

func Fatalf(fields Fields, msg string, args ...interface{}) {
	logrus.WithFields(toLogrusFields(fields)).Fatalf(msg, args...)
}
