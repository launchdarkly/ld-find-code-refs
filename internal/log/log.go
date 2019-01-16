package log

import (
	"os"

	logger "github.com/sirupsen/logrus"
)

func init() {
	logger.SetFormatter(&logger.TextFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logger.DebugLevel)
}

func Fatal(msg string, err error) {
	logger.WithError(err).Fatal(msg)
}

func Error(msg string, err error, fields map[string]interface{}) {
	logger.WithFields(logger.Fields(fields)).WithError(err).Error(msg)
}

func Debug(msg string, fields map[string]interface{}) {
	logger.WithFields(logger.Fields(fields)).Debug(msg)
}

func Info(msg string, fields map[string]interface{}) {
	logger.WithFields(logger.Fields(fields)).Info(msg)
}

func Field(key string, val interface{}) map[string]interface{} {
	return map[string]interface{}{key: val}
}
