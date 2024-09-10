package signalquerypoll

import (
	"context"

	"time"

	. "github.com/dandavison/temporal-latency-experiments/must"
	"github.com/dandavison/temporal-latency-experiments/tle"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
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
	}, MyWorkflow))

	latencies := []int64{}
	polls := []int{}
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
	}
	Must1(c.SignalWorkflow(ctx, workflowID, "", DoneSignalName, nil))

	return tle.Results{
		LatenciesNs: latencies,
		Polls:       polls,
	}
}

func MyWorkflow(ctx workflow.Context) (int, error) {
	counter := 0

	workflow.SetQueryHandler(ctx, QueryName, func() (int, error) {
		return counter, nil
	})

	ch := workflow.GetSignalChannel(ctx, SignalName)

	sel := workflow.NewSelector(ctx)
	sel.AddReceive(ch, func(c workflow.ReceiveChannel, more bool) {
		var signal int
		c.Receive(ctx, &signal)
		counter++
	})

	doneCh := workflow.GetSignalChannel(ctx, DoneSignalName)

	for {
		sel.Select(ctx)
		if doneCh.ReceiveAsync(nil) {
			break
		}
	}

	return counter, nil
}
