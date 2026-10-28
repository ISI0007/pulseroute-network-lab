"""Recalculate saved echo results and export measured summaries (not link capacity)."""
import argparse
import csv
import json
import math
import pathlib
import statistics
import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--input", type=pathlib.Path, required=True)
parser.add_argument("--output", type=pathlib.Path, required=True)
args = parser.parse_args()
args.output.mkdir(parents=True, exist_ok=True)
records = []
for source in sorted(args.input.rglob("*-loopback.json")):
    data = json.loads(source.read_text(encoding="utf-8-sig"))
    assert data["environment"]["network"] == "loopback", source
    for i, run in enumerate(data["runs"], 1):
        seen = set()
        rtts = []
        jitter = 0.0
        last_sequence = -1
        duplicates = reordered = delivered_bytes = 0
        for sample in run["samples"]:
            seq = sample["sequence"]
            if not sample["received"] or not 0 <= seq < run["sent"]:
                continue
            if seq in seen:
                duplicates += 1
                continue
            seen.add(seq)
            assert math.isfinite(sample["rttMs"]) and sample["rttMs"] >= 0
            if rtts:
                jitter += (abs(sample["rttMs"] - rtts[-1]) - jitter) / 16
            rtts.append(sample["rttMs"])
            delivered_bytes += sample["bytes"]
            reordered += int(seq < last_sequence)
            last_sequence = max(seq, last_sequence)
        calculated = {
            "received": len(seen), "duplicates": duplicates,
            "outOfOrder": reordered,
            "lossPercent": 100 * (run["sent"] - len(seen)) / run["sent"],
            "rttMs": statistics.mean(rtts) if rtts else 0,
            "jitterMs": jitter,
            "mbps": delivered_bytes * 8 / (run["durationMs"] / 1000) / 1e6,
        }
        for field, value in calculated.items():
            assert math.isclose(run[field], value, rel_tol=1e-8, abs_tol=1e-9), (source, i, field, run[field], value)
        sorted_rtts = sorted(rtts)
        records.append({
            "source": source.name, "os": data["environment"]["os"],
            "arch": data["environment"]["arch"], "payloadBytes": run["samples"][0]["bytes"], "go": data["environment"]["go"],
            "run": i, "protocol": run["protocol"], "sent": run["sent"],
            **calculated, "medianRttMs": statistics.median(rtts) if rtts else 0,
            "p95RttMs": sorted_rtts[max(0, math.ceil(.95 * len(rtts)) - 1)] if rtts else 0,
            "durationMs": run["durationMs"], "recordedAt": data["recordedAt"],
        })
assert records, "No complete loopback measurement files found"
with (args.output / "run-summary.csv").open("w", newline="", encoding="utf-8") as f:
    writer = csv.DictWriter(f, fieldnames=list(records[0]))
    writer.writeheader()
    writer.writerows(records)

plt.rcParams.update({"font.family": "DejaVu Sans", "font.size": 10})
systems = sorted({r["os"] for r in records})
fig, axes = plt.subplots(1, 2, figsize=(11, 4.5), layout="constrained")
for ax, key, title, units in zip(axes, ["rttMs", "mbps"],
        ["Measured sequential echo RTT", "Measured echo payload goodput"],
        ["Mean RTT per run (ms)", "Echo payload (Mbit/s)"]):
    labels, values = [], []
    for system in systems:
        for protocol in ("tcp", "udp"):
            values.append([r[key] for r in records if r["os"] == system and r["protocol"] == protocol])
            labels.append(system + "\n" + protocol.upper())
    positions = list(range(1, len(values) + 1))
    for pos, group in zip(positions, values):
        ax.scatter([pos] * len(group), group, s=35, color="#008f9a", zorder=3)
        ax.plot([pos - .22, pos + .22], [statistics.median(group)] * 2, color="#17324a", linewidth=2)
    ax.set_xticks(positions, labels)
    ax.set_ylabel(units)
    ax.set_title(title, fontweight="bold")
    ax.set_ylim(bottom=0)
    ax.grid(axis="y", alpha=.2)
counts=", ".join(str(n) for n in sorted({r["sent"] for r in records}))
sizes=", ".join(str(n) for n in sorted({r["payloadBytes"] for r in records}))
fig.suptitle(f"PulseRoute / hosted CI loopback / {counts} probes per run / {sizes}-byte payload", fontsize=12)
fig.savefig(args.output / "loopback-measurements.png", dpi=180)
fig.savefig(args.output / "loopback-measurements.svg")
plt.close(fig)
lines = ["# Recorded loopback experiments", "",
    "Results are actual hosted CI measurements, independently recalculated from raw probe samples.",
    "Each point is a run; the horizontal mark is the median of run means. No improvement or link-capacity claim is made.", "",
    "| OS | Protocol | Runs | Delivered / sent | Median run RTT (ms) | Median echo goodput (Mbit/s) |",
    "|---|---|---:|---:|---:|---:|"]
for system in systems:
    for protocol in ("tcp", "udp"):
        group = [r for r in records if r["os"] == system and r["protocol"] == protocol]
        lines.append(f"| {system} | {protocol.upper()} | {len(group)} | {sum(r['received'] for r in group)} / {sum(r['sent'] for r in group)} | {statistics.median(r['rttMs'] for r in group):.6f} | {statistics.median(r['mbps'] for r in group):.6f} |")
lines += ["", "## Interpretation limits", "",
    "The experiment uses sequential request/reply over loopback. Its goodput includes application pacing and response waiting; it is not physical link throughput. Hosted runners have uncontrolled scheduling and shared resources. Windows and Linux runs are separate environments, not a controlled OS comparison. No two-device LAN result, one-way latency, large-file speed, or before/after optimization is inferred.", "",
    "Loss is the fraction of unique probe sequences not observed during the experiment. A missing response can reflect either direction or a deadline. RTT jitter is the smoothed absolute difference between successive observed RTTs, not RFC 3550 one-way interarrival jitter.", "",
    "Reproduce with the CLI selftest and run this analyzer against the downloaded verification artifacts. Consult the adjacent provenance JSON for the exact commit, run and artifact hashes."]
(args.output / "RESULTS.md").write_text("\n".join(lines) + "\n", encoding="utf-8")
print(json.dumps({"status": "PASS", "runsRecalculated": len(records), "sourceFiles": len(list(args.input.rglob('*-loopback.json')))}, indent=2))

svg=args.output / "loopback-measurements.svg"
svg.write_text("\n".join(line.rstrip() for line in svg.read_text().splitlines())+"\n",encoding="utf-8")
