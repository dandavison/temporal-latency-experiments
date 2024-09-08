package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	. "github.com/dandavison/temporal-latency-experiments/must"
	"github.com/dandavison/temporal-latency-experiments/tle"
	"github.com/dandavison/tle/experiments/signalquerypoll"
	"github.com/dandavison/tle/experiments/update"
	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
)

var experiments = map[string]func(client.Client, sdklog.Logger, int) tle.Results{
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
	run, workflow, iterations := parseArguments()
	c := Must(client.Dial(client.Options{Logger: l}))
	defer c.Close()
	wo := startWorker(c, workflow)
	defer wo.Stop()

	r := run(c, l, iterations)

	fmt.Println(string(Must(json.MarshalIndent(r, "", "  "))))
}

func startWorker(c client.Client, workflow interface{}) worker.Worker {
	wo := worker.New(c, tle.TaskQueue, worker.Options{})
	wo.RegisterWorkflow(workflow)
	Must1(wo.Start())
	return wo
}

func parseArguments() (func(client.Client, sdklog.Logger, int) tle.Results, interface{}, int) {
	iterations := flag.Int("iterations", 1, "Number of iterations")
	experimentName := flag.String("experiment", "", "Experiment to run")
	flag.Parse()
	if *experimentName == "" {
		panic("Experiment name is required")
	}
	run, ok := experiments[*experimentName]
	if !ok {
		panic("Experiment not found")
	}
	workflow, ok := workflows[*experimentName]
	if !ok {
		panic("Workflow not found")
	}
	fmt.Fprintf(os.Stderr, "Running experiment %s\n", *experimentName)
	return run, workflow, *iterations
}
