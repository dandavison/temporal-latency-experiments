package signal

import (
	"github.com/dandavison/temporal-latency-experiments/experiments/signalquerypoll"
	"github.com/dandavison/temporal-latency-experiments/tle"
	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
)

// Send a signal and wait for the response.
func Run(c client.Client, l sdklog.Logger, iterations int) tle.Results {
	return signalquerypoll.SignalQueryPollRunHelper(c, l, iterations, true, false)
}
