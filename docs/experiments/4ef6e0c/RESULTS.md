# Recorded loopback experiments

Results are actual hosted CI measurements, independently recalculated from raw probe samples.
Each point is a run; the horizontal mark is the median of run means. No improvement or link-capacity claim is made.

| OS | Protocol | Runs | Delivered / sent | Median run RTT (ms) | Median echo goodput (Mbit/s) |
|---|---|---:|---:|---:|---:|
| linux | TCP | 5 | 500 / 500 | 0.043488 | 187.972081 |
| linux | UDP | 5 | 500 / 500 | 0.017393 | 468.722822 |
| windows | TCP | 5 | 500 / 500 | 0.033374 | 245.460538 |
| windows | UDP | 5 | 500 / 500 | 0.016960 | 483.018868 |

## Interpretation limits

The experiment uses sequential request/reply over loopback. Its goodput includes application pacing and response waiting; it is not physical link throughput. Hosted runners have uncontrolled scheduling and shared resources. Windows and Linux runs are separate environments, not a controlled OS comparison. No two-device LAN result, one-way latency, large-file speed, or before/after optimization is inferred.

Loss is the fraction of unique probe sequences not observed during the experiment. A missing response can reflect either direction or a deadline. RTT jitter is the smoothed absolute difference between successive observed RTTs, not RFC 3550 one-way interarrival jitter.

Reproduce with the CLI selftest and run this analyzer against the downloaded verification artifacts. Consult the adjacent provenance JSON for the exact commit, run and artifact hashes.
