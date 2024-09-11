import json
from dataclasses import dataclass
from pathlib import Path
from typing import Iterator, List

import altair as alt
import pandas as pd


@dataclass
class Experiment:
    name: str
    cloud: bool
    latencies: List[int]

    @property
    def display_name(self) -> str:
        return f"{self.name} ({'cloud' if self.cloud else 'localhost, in-memory'})"

    @property
    def html_filename(self) -> str:
        return f"results-{'cloud' if self.cloud else 'local'}.html"


def main() -> None:
    src_root = Path("../run/experiments")
    dst_root = Path("./experiments")

    experiments = list(collect_experiments(src_root))
    for experiment in experiments:
        dst_dir = dst_root / experiment.name
        dst_dir.mkdir(parents=True, exist_ok=True)
        create_per_experiment_page(experiment).save(dst_dir / experiment.html_filename)

    create_combined_experiments_page(experiments).save(
        dst_root / "combined-results.html"
    )


def create_per_experiment_page(experiment: Experiment) -> alt.VConcatChart:
    df = pd.DataFrame(experiment.latencies, columns=["LatencyNs"])
    df["LatencyMs"] = df["LatencyNs"] / 1e6

    p50 = df["LatencyMs"].quantile(0.5)
    p90 = df["LatencyMs"].quantile(0.9)
    p99 = df["LatencyMs"].quantile(0.99)

    histogram = (
        alt.Chart(df)
        .mark_bar()
        .encode(
            alt.X("LatencyMs:Q", bin=alt.Bin(maxbins=100), title="Latency (ms)"),
            y=alt.Y("count()", title=None),
        )
        .properties(
            title=f"{experiment.display_name} p50: {p50:.1f}ms, p90: {p90:.1f}ms, p99: {p99:.1f}ms"
        )
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
        .properties(title=f"{experiment.display_name} Density Plot")
    )

    line_plot = (
        alt.Chart(df.reset_index())
        .mark_line()
        .encode(
            x=alt.X("index:Q", title="Sequence"),
            y=alt.Y("LatencyMs:Q", title="Latency (ms)"),
        )
        .properties(title=f"{experiment.display_name} Latency Sequence")
    )

    return alt.vconcat(histogram, density, line_plot)


def create_combined_experiments_page(experiments: List[Experiment]) -> alt.Chart:
    combined_data = []
    for experiment in sorted(experiments, key=lambda x: (x.cloud, x.name)):
        df = pd.DataFrame(experiment.latencies, columns=["LatencyNs"])
        df["LatencyMs"] = df["LatencyNs"] / 1e6
        df["Experiment"] = f"{experiment.display_name}"
        combined_data.append(df)

    combined_df = pd.concat(combined_data)

    combined_density = (
        alt.Chart(combined_df)
        .transform_density(
            "LatencyMs",
            groupby=["Experiment"],
            as_=["LatencyMs", "density"],
        )
        .mark_line()
        .encode(
            x=alt.X("LatencyMs:Q", title="Latency (ms)"),
            y=alt.Y("density:Q", title=None, axis=alt.Axis(ticks=False, labels=False)),
            color=alt.Color("Experiment:N", legend=alt.Legend(title="Experiment")),
        )
    )

    return combined_density


def collect_experiments(src_root: Path) -> Iterator[Experiment]:
    for experiment_path in src_root.iterdir():
        if experiment_path.is_dir():
            for file_name, cloud in [
                ("results-cloud.json", True),
                ("results-local.json", False),
            ]:
                results_path = experiment_path / file_name
                if results_path.exists():
                    with open(results_path, "r") as f:
                        results = json.load(f)
                        latencies = results.get("latenciesNs", [])
                        if latencies:
                            yield Experiment(
                                name=experiment_path.name,
                                cloud=cloud,
                                latencies=latencies,
                            )


if __name__ == "__main__":
    main()
