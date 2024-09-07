import json
import os

import altair as alt
import pandas as pd


def create_charts(data, title):
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


def main():
    src_root = "../run/experiments"
    dst_root = "./experiments"

    for root, dirs, files in os.walk(src_root):
        for dir_name in dirs:
            experiment_path = os.path.join(root, dir_name)
            results_path = os.path.join(experiment_path, "results.json")

            if os.path.exists(results_path):
                with open(results_path, "r") as f:
                    results = json.load(f)
                    latencies = results.get("latenciesNs", [])

                    if latencies:
                        chart = create_charts(latencies, dir_name)
                        dst_dir = os.path.join(dst_root, dir_name)
                        os.makedirs(dst_dir, exist_ok=True)
                        chart.save(os.path.join(dst_dir, "index.html"))


if __name__ == "__main__":
    main()
