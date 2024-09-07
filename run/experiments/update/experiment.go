package update

import (
	"context"
	"strconv"
	"time"

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
func Run(c client.Client, l sdklog.Logger, iterations int) (tle.Results, error) {
	ctx := context.Background()
	_, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             tle.TaskQueue,
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}, MyWorkflow)
	if err != nil {
		tle.Fatal(l, "Failed to start workflow", err)
	}

	latencies := []int64{}
	for i := 0; i < iterations; i++ {
		start := time.Now()
		u, err := c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
			WorkflowID:   workflowID,
			UpdateName:   UpdateName,
			UpdateID:     strconv.Itoa(i),
			WaitForStage: client.WorkflowUpdateStageCompleted,
		})
		if err != nil {
			tle.Fatal(l, "Failed to start update", err)
		}

		var updateResult int
		if err = u.Get(ctx, &updateResult); err != nil {
			tle.Fatal(l, "Failed to fetch update result", err)
		}

		latency := time.Since(start).Nanoseconds()
		latencies = append(latencies, latency)

	}
	if err := c.SignalWorkflow(ctx, workflowID, "", DoneSignalName, nil); err != nil {
		tle.Fatal(l, "Failed to signal workflow", err)
	}

	return tle.Results{
		LatenciesNs: latencies,
	}, nil
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
