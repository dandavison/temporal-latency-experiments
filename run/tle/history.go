package tle

import (
	"context"

	. "github.com/dandavison/temporal-latency-experiments/must"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

func GetHistory(c client.Client, wid, rid string) []enumspb.EventType {
	iter := c.GetWorkflowHistory(context.Background(), wid, rid, false, 0)
	var events []enumspb.EventType
	for iter.HasNext() {
		event := Must(iter.Next())
		events = append(events, event.GetEventType())
	}
	return events
}

func CountWorkflowTasks(c client.Client, wid, rid string) int {
	history := GetHistory(c, wid, rid)
	var count int
	for _, event := range history {
		if event == enumspb.EVENT_TYPE_WORKFLOW_TASK_COMPLETED {
			count++
		}
	}
	return count
}
