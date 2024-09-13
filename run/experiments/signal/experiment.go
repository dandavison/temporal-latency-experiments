package signal

import (
	"context"
	"time"

	"github.com/dandavison/temporal-latency-experiments/experiments/signalquery"
	. "github.com/dandavison/temporal-latency-experiments/must"
	"github.com/dandavison/temporal-latency-experiments/tle"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
)

// Send a signal and wait for the response.
func Run(c client.Client, l sdklog.Logger, iterations int) tle.Results {
	ctx := context.Background()

	latencies := []int64{}
	wfts := []int{}
	for i := 0; i < iterations; i++ {
		if i%2000 == 0 {
			Must(c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
				ID:                    signalquery.WorkflowID,
				TaskQueue:             tle.TaskQueue,
				WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
			}, signalquery.MyWorkflow))
		}
		start := time.Now()

		Must1(c.SignalWorkflow(ctx, signalquery.WorkflowID, "", signalquery.SignalName, i))

		latency := time.Since(start).Nanoseconds()
		latencies = append(latencies, latency)
		time.Sleep(100 * time.Millisecond)
	}
	Must1(c.SignalWorkflow(ctx, signalquery.WorkflowID, "", signalquery.DoneSignalName, nil))

	return tle.Results{
		LatenciesNs: latencies,
		Wfts:        wfts,
	}
}
