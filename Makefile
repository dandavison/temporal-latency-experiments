run:
	cd run && \
	for experiment in update signalquerypoll; do \
		go run main.go --iterations 2000 --experiment $$experiment > experiments/$$experiment/results.json; \
	done

run-cloud:
	cd run && \
	for experiment in update signalquerypoll; do \
		go run main.go \
			--iterations 2000 \
			--experiment $$experiment \
			--address sdk-ci.a2dd6.tmprl.cloud:7233 \
			--namespace sdk-ci.a2dd6 \
			--client-cert /tmp/client.crt \
			--client-key /tmp/client.key \
			> experiments/$$experiment/results.json; \
	done

viz:
	cd viz && uv run viz.py

.PHONY: run viz