package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

func Logger() *logrus.Logger {
	temp := logrus.New()
	temp.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
		ForceQuote:  true,
	})
	temp.SetOutput(os.Stdout)
	logLevel := os.Getenv("BIMA_LOG")
	temp.SetLevel(logrus.FatalLevel)
	if logLevel != "" {
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			temp.Fatalf("Failed to parse log level: %v", err)
		}
		temp.SetLevel(level)
	}
	return temp
}
