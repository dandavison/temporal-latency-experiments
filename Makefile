run:
	cd run && \
	for experiment in update signalquerypoll; do \
		go run main.go --iterations 2000 --experiment $$experiment > experiments/$$experiment/results.json; \
	done

viz:
	cd viz && uv run viz.py

.PHONY: run viz