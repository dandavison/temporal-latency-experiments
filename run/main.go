package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	. "github.com/dandavison/temporal-latency-experiments/must"
	"github.com/dandavison/temporal-latency-experiments/tle"
	"github.com/dandavison/tle/experiments/query"
	"github.com/dandavison/tle/experiments/signal"
	"github.com/dandavison/tle/experiments/signalquery"
	"github.com/dandavison/tle/experiments/update"
	"go.temporal.io/sdk/client"
	sdklog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
)

var experiments = map[string]func(client.Client, sdklog.Logger, int) tle.Results{
	"query":       query.Run,
	"signal":      signal.Run,
	"signalquery": signalquery.Run,
	"update":      update.Run,
}

var workflows = map[string]interface{}{
	"query":       signalquery.MyWorkflow,
	"signal":      signalquery.MyWorkflow,
	"signalquery": signalquery.MyWorkflow,
	"update":      update.MyWorkflow,
}

func main() {
	l := sdklog.NewStructuredLogger(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})))
	run, workflow, iterations, cc := parseArguments()
	c := makeClient(cc, l)
	defer c.Close()
	wo := startWorker(c, workflow)
	defer wo.Stop()

	r := run(c, l, iterations)

	fmt.Fprintf(os.Stderr, "p90: %.1f\n", float64(tle.Quantile(r.LatenciesNs, 0.9))/1e6)

	fmt.Println(string(Must(json.MarshalIndent(r, "", "  "))))
}

func makeClient(cc *ClientConfig, l sdklog.Logger) client.Client {
	if cc == nil {
		return Must(client.Dial(client.Options{Logger: l}))
	}
	cert := Must(tls.LoadX509KeyPair(cc.ClientCertPath, cc.ClientKeyPath))
	return Must(client.Dial(client.Options{
		HostPort:  cc.HostPort,
		Namespace: cc.Namespace,
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{Certificates: []tls.Certificate{cert}},
		},
		Logger: l,
	}))
}

func startWorker(c client.Client, workflow interface{}) worker.Worker {
	wo := worker.New(c, tle.TaskQueue, worker.Options{})
	wo.RegisterWorkflow(workflow)
	Must1(wo.Start())
	return wo
}

type ClientConfig struct {
	ClientKeyPath  string
	ClientCertPath string
	HostPort       string
	Namespace      string
}

func parseArguments() (func(client.Client, sdklog.Logger, int) tle.Results, interface{}, int, *ClientConfig) {
	iterations := flag.Int("iterations", 1, "Number of iterations")
	experimentName := flag.String("experiment", "", "Experiment to run")

	var cc = new(ClientConfig)
	flag.StringVar(&cc.ClientKeyPath, "client-key", "", "Path to client key")
	flag.StringVar(&cc.ClientCertPath, "client-cert", "", "Path to client cert")
	flag.StringVar(&cc.HostPort, "address", "", "Address of the Temporal server")
	flag.StringVar(&cc.Namespace, "namespace", "", "Namespace of the Temporal server")

	flag.Parse()

	if cc.IsZero() {
		cc = nil
	} else if cc.ClientKeyPath == "" || cc.ClientCertPath == "" || cc.HostPort == "" || cc.Namespace == "" {
		panic(fmt.Sprintf("If any client config flag is set, all must be set: %+v", cc))
	}

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
	return run, workflow, *iterations, cc
}

func (c *ClientConfig) IsZero() bool {
	return c.ClientKeyPath == "" && c.ClientCertPath == "" && c.HostPort == "" && c.Namespace == ""
}
