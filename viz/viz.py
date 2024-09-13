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
    wfts: List[int]
    query_times: List[int]
    signal_times: List[int]

    @property
    def display_name(self) -> str:
        display_name = self.name
        if self.name == "signalquery":
            display_name = "signal+query"
        return display_name

    @property
    def env(self) -> str:
        return "cloud" if self.cloud else "localhost, in-memory"

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

    create_presentation_page(experiments).save(dst_root / "presentation-results.html")
    create_presentation_page(experiments, dark_mode=True).save(
        dst_root / "presentation-results-dark.html"
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
            title=f"{experiment.display_name} ({experiment.env}) p50: {p50:.1f}ms, p90: {p90:.1f}ms, p99: {p99:.1f}ms"
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
        .properties(title=f"{experiment.display_name} ({experiment.env}) Density Plot")
    )

    line_plot = (
        alt.Chart(df.reset_index())
        .mark_line()
        .encode(
            x=alt.X("index:Q", title="Sequence"),
            y=alt.Y("LatencyMs:Q", title="Latency (ms)"),
        )
        .properties(
            title=f"{experiment.display_name} ({experiment.env}) Latency Sequence"
        )
    )

    wft_df = pd.DataFrame(experiment.wfts, columns=["Wft"])
    wft_plot = (
        alt.Chart(wft_df.reset_index())
        .mark_line()
        .encode(
            x=alt.X("index:Q", title="Sequence"),
            y=alt.Y("Wft:Q", title="Wft"),
        )
        .properties(title=f"{experiment.display_name} ({experiment.env}) Wft Sequence")
    )

    query_df = pd.DataFrame(
        {
            "Sequence": range(len(experiment.query_times)),
            "QueryTimeNs": experiment.query_times,
        }
    )
    query_df["QueryTimeMs"] = query_df["QueryTimeNs"] / 1e6
    query_time_plot = (
        alt.Chart(query_df)
        .mark_line()
        .encode(
            x=alt.X("Sequence:Q", title="Sequence"),
            y=alt.Y(
                "QueryTimeMs:Q",
                title="Query Time (ms)",
                axis=alt.Axis(titleColor="blue"),
            ),
            color=alt.value("blue"),
        )
        .properties(
            title=f"{experiment.display_name} ({experiment.env}) Query Time Sequence"
        )
    )

    return alt.vconcat(histogram, density, line_plot, wft_plot, query_time_plot)


def create_combined_experiments_page(experiments: List[Experiment]) -> alt.VConcatChart:
    combined_df, xlim = create_combined_data(experiments)
    x_scale = alt.Scale(domain=[combined_df["LatencyMs"].min(), xlim])

    charts = []
    for key, display_name in [
        ("Cloud", "cloud"),
        ("Local", "localhost in-memory sqlite"),
    ]:
        df = combined_df[combined_df["Cloud"] == key]
        if len(df):
            charts.append(create_density_plot(df, display_name, x_scale))

    return alt.vconcat(*charts).resolve_scale(color="independent")


def create_presentation_page(
    experiments: List[Experiment], dark_mode: bool = False
) -> alt.VConcatChart:
    cloud_experiments = [exp for exp in experiments if exp.cloud]

    combined_df, p999 = create_combined_data(
        cloud_experiments, filter_names=["update", "signalquery"]
    )
    x_scale = alt.Scale(domain=[combined_df["LatencyMs"].min(), p999])

    charts = []
    for key, display_name in [
        ("Cloud", "cloud"),
    ]:
        df = combined_df[combined_df["Cloud"] == key]
        if len(df):
            charts.append(create_density_plot(df, "", x_scale))  # Remove title

    chart = alt.vconcat(*charts).resolve_scale(color="independent")

    if dark_mode:
        chart = chart.configure(
            background="#2E2E2E",
            axis=alt.AxisConfig(
                gridColor="#444444",
                domain=False,
                tickColor="#DDDDDD",
                labelColor="#DDDDDD",
                titleColor="#DDDDDD",
            ),
            legend=alt.LegendConfig(
                labelColor="#DDDDDD",
                titleColor="#DDDDDD",
            ),
            axisRight=alt.AxisConfig(
                gridColor="#444444",
                domainColor="#444444",
            ),
        ).configure_view(stroke=None)
    else:
        chart = chart.configure(axis=alt.AxisConfig(domain=False))

    return chart


def create_density_plot(df: pd.DataFrame, title: str, x_scale: alt.Scale) -> alt.Chart:
    return (
        alt.Chart(df)
        .transform_density(
            "LatencyMs",
            groupby=["Experiment"],
            as_=["LatencyMs", "density"],
        )
        .mark_line()
        .encode(
            x=alt.X("LatencyMs:Q", title="Latency (ms)", scale=x_scale),
            y=alt.Y("density:Q", title=None, axis=alt.Axis(ticks=False, labels=False)),
            color=alt.Color("Experiment:N", legend=alt.Legend(title="")),
        )
        .properties(title=title)
    )


def create_combined_data(
    experiments: List[Experiment], filter_names: List[str] = None
) -> pd.DataFrame:
    combined_data = []

    for experiment in experiments:
        if filter_names is None or experiment.name in filter_names:
            df = pd.DataFrame(experiment.latencies, columns=["LatencyNs"])
            df["LatencyMs"] = df["LatencyNs"] / 1e6
            p90 = df["LatencyMs"].quantile(0.9)
            display_name = experiment.display_name
            df["Experiment"] = f"{display_name} p90 = {p90:.0f}ms"
            df["Cloud"] = "Cloud" if experiment.cloud else "Local"
            combined_data.append(df)

    combined_df = pd.concat(combined_data)
    xlim = combined_df["LatencyMs"].quantile(0.997)
    combined_df = combined_df[combined_df["LatencyMs"] <= xlim]  # Filter data
    return combined_df, xlim


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
                    yield Experiment(
                        name=experiment_path.name,
                        cloud=cloud,
                        latencies=results.get("latenciesNs") or [],
                        wfts=results.get("wfts") or [],
                        query_times=results.get("queryTimes") or [],
                        signal_times=results.get("signalTimes") or [],
                    )


if __name__ == "__main__":
    main()
