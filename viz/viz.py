import json
from dataclasses import dataclass
from pathlib import Path
from typing import List

import altair as alt
import pandas as pd


@dataclass
class Experiment:
    name: str
    latencies: List[int]


def main() -> None:
    src_root = Path("../run/experiments")
    dst_root = Path("./experiments")

    experiments = collect_experiments(src_root)
    for experiment in experiments:
        chart = create_charts(experiment.latencies, experiment.name)
        dst_dir = dst_root / experiment.name
        dst_dir.mkdir(parents=True, exist_ok=True)
        chart.save(dst_dir / "index.html")


def create_charts(data: List[int], title: str) -> alt.VConcatChart:
    df = pd.DataFrame(data, columns=["LatencyNs"])
    df["LatencyMs"] = df["LatencyNs"] / 1e6

    p50 = df["LatencyMs"].quantile(0.5)
    p90 = df["LatencyMs"].quantile(0.9)

    histogram = (
        alt.Chart(df)
        .mark_bar()
        .encode(
            alt.X("LatencyMs:Q", bin=alt.Bin(maxbins=100), title="Latency (ms)"),
            y=alt.Y("count()", title=None),
        )
        .properties(title=f"{title} Histogram (p50: {p50:.2f} ms, p90: {p90:.2f} ms)")
    )

    density = (
        alt.Chart(df)
        .transform_density(
            "LatencyMs",
            as_=["LatencyMs", "density"],
        )
        .mark_line()
        .encode(
            x=alt.X("LatencyMs:Q", title="Latency (ms)"),
            y=alt.Y("density:Q", title=None),
        )
        .properties(title=f"{title} Density Plot")
    )

    line_plot = (
        alt.Chart(df.reset_index())
        .mark_line()
        .encode(
            x=alt.X("index:Q", title="Sequence"),
            y=alt.Y("LatencyMs:Q", title="Latency (ms)"),
        )
        .properties(title=f"{title} Latency Sequence")
    )

    combined_chart = alt.vconcat(histogram, density, line_plot)

    return combined_chart


def collect_experiments(src_root: Path) -> List[Experiment]:
    experiments = []
    for experiment_path in src_root.iterdir():
        if experiment_path.is_dir():
            results_path = experiment_path / "results.json"
            if results_path.exists():
                with open(results_path, "r") as f:
                    results = json.load(f)
                    latencies = results.get("latenciesNs", [])
                    if latencies:
                        experiments.append(
                            Experiment(name=experiment_path.name, latencies=latencies)
                        )
    return experiments


if __name__ == "__main__":
    main()
