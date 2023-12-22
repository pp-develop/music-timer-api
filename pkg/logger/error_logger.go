package logger

import (
    "log"
    "runtime/debug"
)

func LogError(err error) {
    log.Printf("ERROR: %v\nStack Trace:\n%s", err, debug.Stack())
}
