// Package log logging utils which wrap the logrus lib
package log

import (
	"os"
	"strings"
	"time"

	"github.com/Akagi201/udplb/config"
	"github.com/sirupsen/logrus"
)

func init() {
	// set log level
	if level, err := logrus.ParseLevel(strings.ToLower(config.Opts.LogLevel)); err != nil {
		logrus.SetLevel(level)
	}

	// set log format
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	// set log output
	switch config.Opts.LogFile {
	case "stdout":
		logrus.SetOutput(os.Stdout)
	case "stderr":
		logrus.SetOutput(os.Stderr)
	default:
		f, err := os.OpenFile(config.Opts.LogFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
		if err != nil {
			logrus.Fatalln(err)
		}
		logrus.SetOutput(f)
	}
}

// For add log node_type field
func For(name string) *logrus.Entry {
	return logrus.WithField("node_type", name)
}

/* ----- Wrap logrus ------ */

func Debug(args ...interface{}) {
	logrus.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

func Debugln(args ...interface{}) {
	logrus.Debugln(args...)
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func Infoln(args ...interface{}) {
	logrus.Infoln(args...)
}

func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Warnln(args ...interface{}) {
	logrus.Warnln(args...)
}

func Error(args ...interface{}) {
	logrus.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Errorln(args ...interface{}) {
	logrus.Errorln(args...)
}

func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

func Fatalln(args ...interface{}) {
	logrus.Fatalln(args...)
}
