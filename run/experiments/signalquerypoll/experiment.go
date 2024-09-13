package signalquerypoll

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

const (
	SignalName     = "my-signal"
	QueryName      = "my-query"
	DoneSignalName = "Done"
	workflowID     = "my-workflow-id"
)

// Send a signal and immediately start executing queries until a query result is
// received indicating that it read the signal's writes to local workflow state.
func Run(c client.Client, l sdklog.Logger, iterations int) tle.Results {
	ctx := context.Background()
	Must(c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             tle.TaskQueue,
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}, signalquery.MyWorkflow))

	latencies := []int64{}
	polls := []int{}
	wfts := []int{}
	for i := 0; i < iterations; i++ {
		start := time.Now()

		go Must1(c.SignalWorkflow(ctx, workflowID, "", SignalName, i))

		for j := 1; ; j++ {
			queryResult := Must(c.QueryWorkflow(ctx, workflowID, "", QueryName))
			var result int
			Must1(queryResult.Get(&result))
			if result == i+1 {
				polls = append(polls, j)
				break
			}
		}
		latency := time.Since(start).Nanoseconds()
		latencies = append(latencies, latency)
		time.Sleep(100 * time.Millisecond)
	}
	Must1(c.SignalWorkflow(ctx, workflowID, "", DoneSignalName, nil))

	return tle.Results{
		LatenciesNs: latencies,
		Polls:       polls,
		Wfts:        wfts,
	}
}
