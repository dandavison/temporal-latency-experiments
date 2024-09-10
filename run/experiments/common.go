package experiments

import (
	"os"

	sdklog "go.temporal.io/sdk/log"
)

const TaskQueue string = "tle"

func Fatal(l sdklog.Logger, msg string, err error) {
	l.Error(msg, "error", err)
	os.Exit(1)
}
