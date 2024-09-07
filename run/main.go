package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/dandavison/temporal-latency-experiments/tle"
	"github.com/dandavison/tle/experiments/signalquerypoll"
	"github.com/dandavison/tle/experiments/update"
	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
)

var experiments = map[string]func(client.Client, sdklog.Logger, int) (tle.Results, error){
	"update":          update.Run,
	"signalquerypoll": signalquerypoll.Run,
}

var workflows = map[string]interface{}{
	"update":          update.MyWorkflow,
	"signalquerypoll": signalquerypoll.MyWorkflow,
}

func main() {
	l := sdklog.NewStructuredLogger(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})))

	// prepare
	run, workflow, iterations := parseArguments(l)
	c := getClient(l)
	defer c.Close()
	wo := startWorker(l, c, workflow)
	defer wo.Stop()

	// run
	r, err := run(c, l, iterations)
	if err != nil {
		tle.Fatal(l, "Failed to run experiment", err)
	}

	jsonOutput, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		tle.Fatal(l, "Failed to marshal results to JSON", err)
	}
	fmt.Println(string(jsonOutput))
}

func getClient(l sdklog.Logger) client.Client {
	c, err := client.Dial(client.Options{Logger: l})
	if err != nil {
		tle.Fatal(l, "Failed to create client", err)
	}
	return c
}

func startWorker(l sdklog.Logger, c client.Client, workflow interface{}) worker.Worker {
	wo := worker.New(c, tle.TaskQueue, worker.Options{})
	wo.RegisterWorkflow(workflow)

	err := wo.Start()
	if err != nil {
		tle.Fatal(l, "Failed to start worker", err)
	}
	return wo
}

func parseArguments(l sdklog.Logger) (func(client.Client, sdklog.Logger, int) (tle.Results, error), interface{}, int) {
	iterations := flag.Int("iterations", 1, "Number of iterations")
	experimentName := flag.String("experiment", "", "Experiment to run")
	flag.Parse()
	if *experimentName == "" {
		tle.Fatal(l, "Experiment name is required", nil)
	}
	run, ok := experiments[*experimentName]
	if !ok {
		tle.Fatal(l, "Experiment not found", nil)
	}
	workflow, ok := workflows[*experimentName]
	if !ok {
		tle.Fatal(l, "Workflow not found", nil)
	}
	fmt.Fprintf(os.Stderr, "Running experiment %s\n", *experimentName)
	return run, workflow, *iterations
}
