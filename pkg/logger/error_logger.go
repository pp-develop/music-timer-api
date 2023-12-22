package logger

import (
	"log"
	"runtime/debug"
)

func LogError(err error) {
	log.Printf("Error: %v, Type: %T, Stacktrace: %s", err, err, debug.Stack())
}
