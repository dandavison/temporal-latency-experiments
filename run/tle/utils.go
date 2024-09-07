package tle

import (
	"os"

	"go.temporal.io/sdk/log"
)

func Fatal(l log.Logger, msg string, err error) {
	l.Error(msg, "error", err)
	os.Exit(1)
}
