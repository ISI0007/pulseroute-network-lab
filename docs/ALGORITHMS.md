# Algorithms and interpretation
## Framing
N = one type byte + payload length. A four-byte big-endian prefix carries N. Validate 1 <= N <= 131072 before allocation; io.ReadFull handles fragmented TCP reads. Write loops handle short writes. Time O(N), memory O(N), with an explicit cap.
## SHA-256 and file transfer
The sender hashes the source in a preliminary pass, rewinds, and sends 65536-byte chunks. The receiver hashes bytes while writing a temporary file. Success requires receivedSize = declaredSize and receivedHash = declaredHash, followed by sync, rename and recorded completion. Time O(fileBytes); bounded chunk memory; sender performs two disk reads. SHA-256 checks integrity but does not authenticate a malicious peer who can supply its own hash.
## RTT
RTT_i = local receive time - local send time. Go elapsed durations use the local monotonic clock where supported. TCP establishment is excluded from the echo loop. This is application request/response latency, including scheduling and echo handling, not a packet-capture estimate.
## Echo loss
LossPercent = 100 * (sent - uniqueReceived) / sent. UDP requests have a 250 ms response deadline. Late replies observed during later requests count; replies after the experiment ends do not. An absent echo may reflect outbound loss, inbound loss, server behavior or delay. It does not isolate one-way network loss.
## Duplicates and ordering
Keep a set of valid sequence numbers. Repeat arrivals increment duplicates but do not increase received bytes or mean RTT. A new sequence below the highest previously received sequence increments outOfOrder. Sequential probes rarely exercise reordering; synthetic test vectors exercise the calculation. Time O(samples), space O(sent).
## RTT variation
J_i = J_(i-1) + (abs(RTT_i - RTT_(i-1)) - J_(i-1)) / 16.
This is an exponentially smoothed RTT-variation statistic, not RTP interarrival jitter or a one-way delay measure. Initialize J to zero; ignore duplicate samples.
## Echo goodput
Mbps = unique echoed payload bytes * 8 / elapsedSeconds / 1,000,000.
Payload includes the experiment's sequence header; wire-level TCP/IP/Ethernet overhead is excluded. Stop-and-wait probes and timeout waits limit the reported rate. Do not interpret this as link capacity or a fair saturation-throughput comparison.
## Sliding-window transfer rate and ETA
The UI polls every 500 ms and retains recent byte counters spanning approximately two seconds.
bytesPerSecond = max(0, newestBytes - oldestBytes) / elapsedSeconds.
ETA = remainingBytes / bytesPerSecond, only when the measured rate is positive.
Hash preparation is a separate state. Buffered writes are not recipient verification. A final SHA-256 acknowledgment determines verified completion. Terminal samples and progress entries are bounded.
## Operating-system TCP
The OS implements TCP retransmission, ordering, congestion control and flow control. PulseRoute neither implements nor selects Tahoe, Reno or CUBIC. Its application frame/chunk size is not the TCP congestion window.

## Reproducible empirical checks
Twenty hosted loopback experiments are retained under evidence/ci/4ef6e0c. scripts/analyze-experiments.py independently recalculates delivered sequences, loss, duplicate/out-of-order counts, RTT, smoothed RTT variation and payload goodput, then emits CSV and measured charts. Medians and nearest-rank 95th-percentile RTTs are calculated from actual observed samples. These measurements do not estimate physical link capacity.
