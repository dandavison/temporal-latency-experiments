run:
	cd run && \
	for experiment in update; do \
		go run main.go --iterations 1000 --experiment $$experiment > experiments/$$experiment/results.json; \
	done

viz:
	cd viz && uv run viz.py

.PHONY: run viz