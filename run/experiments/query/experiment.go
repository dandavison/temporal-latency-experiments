package query

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

// Send a query and wait for the response.
func Run(c client.Client, l sdklog.Logger, iterations int) tle.Results {
	ctx := context.Background()
	Must(c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                    signalquery.WorkflowID,
		TaskQueue:             tle.TaskQueue,
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}, signalquery.MyWorkflow))

	latencies := []int64{}
	wfts := []int{}
	for i := 0; i < iterations; i++ {
		start := time.Now()

		queryResult := Must(c.QueryWorkflow(ctx, signalquery.WorkflowID, "", signalquery.QueryName))
		var result int
		Must1(queryResult.Get(&result))

		latency := time.Since(start).Nanoseconds()
		latencies = append(latencies, latency)
		wfts = append(wfts, tle.CountWorkflowTasks(c, signalquery.WorkflowID, ""))
	}
	Must1(c.SignalWorkflow(ctx, signalquery.WorkflowID, "", signalquery.DoneSignalName, nil))

	return tle.Results{
		LatenciesNs: latencies,
		Wfts:        wfts,
	}
}
