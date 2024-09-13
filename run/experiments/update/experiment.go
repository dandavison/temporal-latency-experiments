package update

import (
	"context"
	"strconv"
	"time"

	. "github.com/dandavison/temporal-latency-experiments/must"
	"github.com/dandavison/temporal-latency-experiments/tle"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

const (
	UpdateName     = "my-update"
	DoneSignalName = "Done"
	workflowID     = "my-workflow-id"
)

// Execute an update (i.e., send it and wait for the result).
func Run(c client.Client, l sdklog.Logger, iterations int) tle.Results {
	ctx := context.Background()

	latencies := []int64{}
	wfts := []int{}
	for i := 0; i < iterations; i++ {
		if i%2000 == 0 {
			Must(c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
				ID:                    workflowID,
				TaskQueue:             tle.TaskQueue,
				WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
			}, MyWorkflow))
		}

		start := time.Now()
		u := Must(c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
			WorkflowID:   workflowID,
			UpdateName:   UpdateName,
			UpdateID:     strconv.Itoa(i),
			WaitForStage: client.WorkflowUpdateStageCompleted,
		}))

		var updateResult int
		Must1(u.Get(ctx, &updateResult))

		latency := time.Since(start).Nanoseconds()
		latencies = append(latencies, latency)
		time.Sleep(100 * time.Millisecond)

	}
	Must1(c.SignalWorkflow(ctx, workflowID, "", DoneSignalName, nil))

	return tle.Results{
		LatenciesNs: latencies,
		Wfts:        wfts,
	}
}

func MyWorkflow(ctx workflow.Context) (int, error) {
	counter := 0
	err := workflow.SetUpdateHandler(
		ctx,
		UpdateName,
		func(ctx workflow.Context, val int) (int, error) {
			counter += val
			return counter, nil
		})
	if err != nil {
		return 0, err
	}
	workflow.GetSignalChannel(ctx, DoneSignalName).Receive(ctx, nil)
	return counter, nil
}
