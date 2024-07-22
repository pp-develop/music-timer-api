package logger

import (
	"log"

	"github.com/pkg/errors"
)

func LogError(err error) {
	wrappedErr := errors.WithStack(err)
	log.Printf("Error: %v, Type: %T", err, err)
	log.Printf("Detailed Error: %+v", wrappedErr)
}
