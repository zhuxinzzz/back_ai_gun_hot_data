package lr

import (
	"github.com/sirupsen/logrus"
)

var (
	errorLogger *logrus.Logger
	infoLogger  *logrus.Logger
)

// lr -> logger
func Init() {
	errorLogger = createLogger(logrus.ErrorLevel)
	infoLogger = createLogger(logrus.InfoLevel)
}

func createLogger(level logrus.Level) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetReportCaller(true)
	return logger
}

type F = logrus.Fields

func E() *logrus.Logger {
	return errorLogger
}

func I() *logrus.Logger {
	return infoLogger
}
