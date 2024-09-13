ITERATIONS=10000
EXPERIMENTS=query signal signalquery update

run:
	cd run && \
	set -e && \
	for experiment in $(EXPERIMENTS); do \
		go run main.go \
			--iterations $(ITERATIONS) \
			--experiment $$experiment \
			> experiments/$$experiment/results-local.json; \
	done

run-cloud:
	cd run && \
	set -e && \
	for experiment in $(EXPERIMENTS); do \
		go run main.go \
			--iterations $(ITERATIONS) \
			--experiment $$experiment \
			--address sdk-ci.a2dd6.tmprl.cloud:7233 \
			--namespace sdk-ci.a2dd6 \
			--client-cert /tmp/client.crt \
			--client-key /tmp/client.key \
			> experiments/$$experiment/results-cloud.json; \
	done

viz:
	cd viz && uv run viz.py

setup:
	cd viz && uv sync

.PHONY: run run-cloud viz setup
