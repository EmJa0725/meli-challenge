package logger

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var level = INFO
var std = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

func init() {
	// Attempt to load .env so LOG_LEVEL set in the project .env is available
	_ = godotenv.Load()

	l := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	// print log level
	log.Printf("Log level set to: %s", l)
	switch l {
	case "DEBUG":
		level = DEBUG
	case "INFO":
		level = INFO
	case "WARN":
		level = WARN
	case "ERROR":
		level = ERROR
	default:
		level = INFO
	}
}

func Debugf(format string, v ...interface{}) {
	if level <= DEBUG {
		std.Printf("[DEBUG] "+format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if level <= INFO {
		std.Printf("[INFO] "+format, v...)
	}
}

func Warnf(format string, v ...interface{}) {
	if level <= WARN {
		std.Printf("[WARN] "+format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if level <= ERROR {
		std.Printf("[ERROR] "+format, v...)
	}
}
