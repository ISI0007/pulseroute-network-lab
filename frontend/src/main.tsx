import React, { useEffect, useRef, useState } from "react";
import { createRoot } from "react-dom/client";
import {
  Activity,
  ArrowDownLeft,
  ArrowUpRight,
  CheckCircle2,
  ChevronRight,
  Clock,
  Database,
  FileUp,
  FlaskConical,
  FolderOpen,
  History,
  Image as ImageIcon,
  MessageSquare,
  Network,
  Play,
  Plug,
  Radio,
  Send,
  Settings,
  ShieldCheck,
  Square,
  Sun,
  Trash2,
  X,
  Zap,
} from "lucide-react";
import "./style.css";
type Event = {
  kind: string;
  id: string;
  peer: string;
  text: string;
  name: string;
  path: string;
  sha256: string;
  bytes: number;
  total: number;
  status: string;
  at: string;
};
type Result = {
  protocol: string;
  address: string;
  sent: number;
  received: number;
  lossPercent: number;
  rttMs: number;
  jitterMs: number;
  mbps: number;
  durationMs: number;
  outOfOrder: number;
  duplicates: number;
  samples: { sequence: number; rttMs: number; received: boolean }[];
};
type NodeSettings = {
  name: string;
  tcp: string;
  udp: string;
  target: string;
  discovery: string;
};
type State = {
  settings?: NodeSettings;
  listening: string;
  udp: string;
  connected: string;
  error: string;
  history: Event[];
  progress: Event[];
  peers: { name: string; address: string }[];
  results: Result[];
  dropped: string[];
  theme: string;
  dataDir: string;
};
declare global {
  interface Window {
    go?: {
      main: { App: Record<string, (...args: unknown[]) => Promise<any>> };
    };
  }
}
const empty: State = {
  listening: "",
  udp: "",
  connected: "",
  error: "",
  history: [],
  progress: [],
  peers: [],
  results: [],
  dropped: [],
  theme: "dark",
  dataDir: "",
};
async function call<T = void>(method: string, ...args: unknown[]): Promise<T> {
  const bridge = window.go?.main.App;
  if (!bridge)
    throw Error("Open PulseRoute.exe to use the desktop networking bridge.");
  return bridge[method](...args);
}
const bytes = (n: number) =>
  n >= 1073741824
    ? (n / 1073741824).toFixed(2) + " GiB"
    : n >= 1048576
      ? (n / 1048576).toFixed(1) + " MiB"
      : n >= 1024
        ? (n / 1024).toFixed(1) + " KiB"
        : n + " B";
const fileName = (p: string) => p.split(/[\\/]/).pop() || p;
const nav = [
  ["Overview", Activity],
  ["Messages", MessageSquare],
  ["Transfers", ArrowUpRight],
  ["Network Lab", FlaskConical],
  ["History", History],
  ["Settings", Settings],
] as const;
function Chart({
  values,
  color = "var(--cyan)",
}: {
  values: number[];
  color?: string;
}) {
  if (!values.length)
    return (
      <div className="chart-empty">
        <Activity size={28} />
        <span>Run an experiment to plot measured RTT</span>
      </div>
    );
  const max = Math.max(...values, 0.01);
  return (
    <svg
      viewBox="0 0 700 150"
      role="img"
      aria-label="Measured round-trip times"
      preserveAspectRatio="none"
    >
      <defs>
        <linearGradient id="fade" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity=".22" />
          <stop offset="100%" stopColor={color} stopOpacity="0" />
        </linearGradient>
      </defs>
      {[30, 65, 100, 135].map((y) => (
        <line
          key={y}
          x1="0"
          x2="700"
          y1={y}
          y2={y}
          stroke="var(--border)"
          strokeDasharray="4 6"
        />
      ))}
      <polyline
        fill="none"
        stroke={color}
        strokeWidth="2.5"
        points={values
          .map(
            (v, i) =>
              `${(i * 700) / Math.max(values.length - 1, 1)},${135 - (v / max) * 115}`,
          )
          .join(" ")}
      />
      <text x="4" y="15" fill="var(--muted)" fontSize="11">
        {max.toFixed(2)} ms
      </text>
    </svg>
  );
}
function App() {
  const [s, setS] = useState<State>(empty),
    [tab, setTab] = useState("Overview"),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(""),
    [name, setName] = useState("Operator"),
    [tcp, setTcp] = useState("127.0.0.1:42191"),
    [udp, setUdp] = useState("127.0.0.1:42192"),
    [target, setTarget] = useState("127.0.0.1:42191"),
    [discovery, setDiscovery] = useState("255.255.255.255:42190"),
    [text, setText] = useState(""),
    [queue, setQueue] = useState<string[]>([]),
    [running, setRunning] = useState(false),
    [protocol, setProtocol] = useState("tcp"),
    [count, setCount] = useState(100),
    [size, setSize] = useState(1024),
    [preview, setPreview] = useState(""),
    [previewName, setPreviewName] = useState(""),
    [filter, setFilter] = useState("all"),
    [speeds, setSpeeds] = useState<Record<string, number>>({});
  const hydrated = useRef(false);
  const samples = useRef<Record<string, { at: number; bytes: number }[]>>({}),
    stop = useRef(false);
  useEffect(() => {
    let active = true,
      loading = false;
    const poll = async () => {
      if (loading) return;
      loading = true;
      try {
        const next = await call<State>("Snapshot");
        if (!active) return;
        setS(next);
        if (!hydrated.current && next.settings) {
          setName(next.settings.name);
          setTcp(next.settings.tcp);
          setUdp(next.settings.udp);
          setTarget(next.settings.target);
          setDiscovery(next.settings.discovery);
          hydrated.current = true;
        }
        if (next.dropped?.length)
          setQueue((q) => [...new Set([...q, ...next.dropped])].slice(0, 100));
        document.documentElement.dataset.theme = next.theme || "dark";
        const rates: Record<string, number> = {};
        for (const p of next.progress) {
          const list = samples.current[p.id] || [];
          const now = Date.now();
          list.push({ at: now, bytes: p.bytes });
          while (list.length > 2 && now - list[0].at > 2000) list.shift();
          samples.current[p.id] = list;
          const dt = (now - list[0].at) / 1000;
          rates[p.id] =
            dt > 0 ? Math.max(0, (p.bytes - list[0].bytes) / dt) : 0;
        }
        for (const id of Object.keys(samples.current)) {
          if (!next.progress.some((p) => p.id === id))
            delete samples.current[id];
        }
        setSpeeds(rates);
      } catch (e) {
        if (active) setError(String(e));
      } finally {
        loading = false;
      }
    };
    poll();
    const id = setInterval(poll, 500);
    return () => {
      active = false;
      clearInterval(id);
    };
  }, []);
  async function act(label: string, fn: () => Promise<unknown>) {
    setBusy(label);
    setError("");
    try {
      await fn();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy("");
    }
  }
  const transfers = s.history.filter((e) => e.kind === "transfer"),
    messages = s.history
      .filter((e) => e.kind === "message")
      .filter((e, i, a) => a.findIndex((x) => x.id === e.id) === i)
      .reverse(),
    verified = transfers.filter(
      (e) => e.status === "verified" || e.status === "sent-verified",
    ),
    latest = s.results.at(-1),
    current = s.progress.filter((e) =>
      ["hashing", "sending", "receiving"].includes(e.status),
    ),
    verifiedBytes = verified.reduce((n, e) => n + e.bytes, 0);
  async function startQueue() {
    setRunning(true);
    stop.current = false;
    try {
      for (const p of queue) {
        if (stop.current) break;
        await call("SendFile", target, p);
        setQueue((q) => q.filter((x) => x !== p));
      }
    } catch (e) {
      setError(String(e));
    } finally {
      setRunning(false);
    }
  }
  async function showPreview(path: string) {
    await act("Preview", async () => {
      const url = await call<string>("Preview", path);
      setPreview(url);
      setPreviewName(fileName(path));
    });
  }
  const connectPanel = (
    <section className="panel connection">
      <div className="panel-heading">
        <h2>
          <Plug size={17} /> Connection desk
        </h2>
        <span className="small-label">TCP / LAN</span>
      </div>
      <div className="connection-grid">
        <label>
          Peer name
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={64}
          />
        </label>
        <label>
          Remote TCP endpoint
          <input
            value={target}
            onChange={(e) => setTarget(e.target.value)}
            placeholder="192.168.1.20:42191"
          />
        </label>
        <button
          className="primary"
          disabled={!!busy}
          onClick={() =>
            act("Connecting", () =>
              s.connected ? call("Disconnect") : call("Connect", target, name),
            )
          }
        >
          <Plug size={16} />
          {s.connected ? "Disconnect" : "Connect"}
        </button>
      </div>
      <div className="listener-row">
        <span className={s.listening ? "dot live" : "dot"} />
        <span>
          {s.listening
            ? "Receiving on " + s.listening
            : "Local receiver is stopped"}
        </span>
        <button
          className="text-button"
          disabled={!!busy}
          onClick={() =>
            act("Listener", () =>
              s.listening
                ? call("StopHost")
                : call("StartHost", tcp, udp, name),
            )
          }
        >
          {s.listening ? "Stop receiver" : "Start receiver"}
          <ChevronRight size={14} />
        </button>
      </div>
    </section>
  );
  return (
    <div className="app">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-icon">
            <Activity />
          </div>
          <div>
            PulseRoute<small>NETWORK LABORATORY</small>
          </div>
        </div>
        <div className="workspace-label">
          WORKSPACE <span>01</span>
        </div>
        <nav>
          {nav.map(([label, Icon]) => (
            <button
              key={label}
              className={tab === label ? "active" : ""}
              onClick={() => setTab(label)}
            >
              <Icon size={18} />
              {label}
              {label === "Transfers" && queue.length > 0 && (
                <b>{queue.length}</b>
              )}
            </button>
          ))}
        </nav>
        <div className="sidebar-note">
          <ShieldCheck size={21} />
          <strong>Integrity, built in.</strong>
          <p>Every completed file is verified with SHA-256.</p>
        </div>
        <div className="sidebar-bottom">
          <span className={s.connected ? "dot live" : "dot"} />
          <div>
            {s.connected ? "Connection active" : "Ready to connect"}
            <small>{s.connected || "Trusted local networks"}</small>
          </div>
          <button
            aria-label="Toggle theme"
            onClick={() =>
              act("Theme", () =>
                call("SetTheme", s.theme === "light" ? "dark" : "light"),
              )
            }
          >
            <Sun size={17} />
          </button>
        </div>
      </aside>
      <main>
        <header>
          <div className="breadcrumb">
            WORKSPACE <ChevronRight size={12} /> {tab.toUpperCase()}
          </div>
          <div className="header-right">
            <span className="tag">v0.1 / DEVELOPMENT</span>
            <span className="avatar">{name.slice(0, 2).toUpperCase()}</span>
          </div>
        </header>
        <div className="page">
          <div className="page-heading">
            <div>
              <div className="eyebrow">
                PULSEROUTE /{" "}
                {tab === "Overview" ? "CONTROL ROOM" : tab.toUpperCase()}
              </div>
              <h1>{tab === "Overview" ? "Your network. In motion." : tab}</h1>
              <p>
                {
                  (
                    {
                      Overview:
                        "Connect peers, move files, and see what happens on the wire.",
                      Messages: "Real TCP sessions. Clear delivery semantics.",
                      Transfers:
                        "Stream binary files with end-to-end integrity verification.",
                      "Network Lab":
                        "Measure the transport. Keep the raw evidence.",
                      History:
                        "A local record of messages and transfer outcomes.",
                      Settings: "Configure your node and local workspace.",
                    } as Record<string, string>
                  )[tab]
                }
              </p>
            </div>
            <button className="secondary" onClick={() => setTab("Network Lab")}>
              <FlaskConical size={16} /> Open experiment
            </button>
          </div>
          {(error || s.error) && (
            <div className="error" role="alert">
              <span>{error || s.error}</span>
              <button
                aria-label="Dismiss error"
                onClick={() => {
                  setError("");
                  call("ClearError").catch(() => {});
                }}
              >
                <X size={16} />
              </button>
            </div>
          )}
          {tab === "Overview" && (
            <>
              <div className="stats">
                <Stat
                  icon={<Network />}
                  label="Session"
                  value={s.connected ? "Connected" : "Offline"}
                  detail={s.connected || "Connect through the desk below"}
                />
                <Stat
                  icon={<ArrowUpRight />}
                  label="Verified traffic"
                  value={bytes(verifiedBytes)}
                  detail={verified.length + " verified history entries"}
                />
                <Stat
                  icon={<Activity />}
                  label="Latest RTT"
                  value={latest ? latest.rttMs.toFixed(2) + " ms" : "N/A"}
                  detail={
                    latest
                      ? latest.protocol.toUpperCase() + " application echo"
                      : "Awaiting an experiment"
                  }
                />
                <Stat
                  icon={<ShieldCheck />}
                  label="Active transfers"
                  value={String(current.length)}
                  detail="Streaming in bounded chunks"
                />
              </div>
              <div className="overview-grid">
                <section className="panel topology">
                  <div className="panel-heading">
                    <h2>Connection topology</h2>
                    <span className="small-label">
                      {s.connected ? "LIVE SESSION" : "IDLE"}
                    </span>
                  </div>
                  <div className="network-stage">
                    <div className="node local">
                      <Network size={25} />
                      <strong>{name || "This node"}</strong>
                      <small>{s.listening || "Local node"}</small>
                    </div>
                    <div className={"route " + (s.connected ? "online" : "")}>
                      <span>TCP / LENGTH-PREFIXED</span>
                      <div />
                      <small>
                        {s.connected
                          ? "Ordered byte stream"
                          : "No active route"}
                      </small>
                    </div>
                    <div className="node remote">
                      <Radio size={25} />
                      <strong>
                        {s.connected ? "Remote endpoint" : "Peer / Hub"}
                      </strong>
                      <small>{s.connected || "Awaiting connection"}</small>
                    </div>
                  </div>
                  <div className="topology-footer">
                    <span>
                      <span className="dot live" /> Direct or client-server
                    </span>
                    <span>LAN scope / no NAT traversal</span>
                  </div>
                </section>
                <section className="panel quick">
                  <div className="eyebrow">MOVE SOMETHING</div>
                  <h2>
                    From your disk.
                    <br />
                    To their desktop.
                  </h2>
                  <p>
                    Images, documents, or video files.
                    <br />
                    One reliable transfer engine.
                  </p>
                  <button
                    className="primary"
                    onClick={() => setTab("Transfers")}
                  >
                    <FileUp size={17} /> Prepare a transfer
                    <ChevronRight size={15} />
                  </button>
                </section>
              </div>
              {connectPanel}
              <div className="bottom-grid">
                <section className="panel">
                  <div className="panel-heading">
                    <h2>Round-trip signal</h2>
                    <button
                      className="text-button"
                      onClick={() => setTab("Network Lab")}
                    >
                      Explore <ArrowUpRight size={15} />
                    </button>
                  </div>
                  <Chart
                    values={
                      latest?.samples
                        .filter((x) => x.received)
                        .map((x) => x.rttMs) || []
                    }
                  />
                  <div className="chart-caption">
                    Application echo RTT / actual recorded samples
                  </div>
                </section>
                <section className="panel">
                  <div className="panel-heading">
                    <h2>Recent activity</h2>
                    <button
                      className="text-button"
                      onClick={() => setTab("History")}
                    >
                      View all
                    </button>
                  </div>
                  {s.history.length ? (
                    s.history.slice(0, 4).map((e, i) => (
                      <div className="activity" key={i}>
                        <span className="activity-icon">
                          {e.kind === "message" ? (
                            <MessageSquare size={17} />
                          ) : (
                            <ArrowDownLeft size={17} />
                          )}
                        </span>
                        <div>
                          <strong>{e.name || e.peer || e.kind}</strong>
                          <small>
                            {e.status} / {new Date(e.at).toLocaleTimeString()}
                          </small>
                        </div>
                      </div>
                    ))
                  ) : (
                    <Empty
                      icon={<Clock />}
                      title="Your activity starts here"
                      text="Connect a peer or complete a transfer."
                    />
                  )}
                </section>
              </div>
            </>
          )}
          {tab === "Messages" && (
            <>
              {connectPanel}
              <section className="panel chat">
                <div className="panel-heading">
                  <h2>
                    <MessageSquare size={17} /> Session conversation
                  </h2>
                  <span className="small-label">
                    {messages.length} MESSAGES
                  </span>
                </div>
                <div className="messages">
                  {messages.length ? (
                    messages.map((m, i) => (
                      <div className="message" key={m.id + i}>
                        <span className="avatar">
                          {m.peer.slice(0, 2).toUpperCase()}
                        </span>
                        <div>
                          <strong>
                            {m.peer}
                            <small>{new Date(m.at).toLocaleTimeString()}</small>
                          </strong>
                          <p>{m.text}</p>
                        </div>
                      </div>
                    ))
                  ) : (
                    <Empty
                      icon={<MessageSquare />}
                      title="Start a conversation"
                      text="Messages are broadcast to clients connected to the same listener."
                    />
                  )}
                </div>
                <form
                  onSubmit={(e) => {
                    e.preventDefault();
                    act("Sending", async () => {
                      await call("SendMessage", text);
                      setText("");
                    });
                  }}
                >
                  <input
                    aria-label="Message"
                    value={text}
                    onChange={(e) => setText(e.target.value)}
                    maxLength={8192}
                    placeholder="Write a message..."
                    disabled={!s.connected}
                  />
                  <button
                    className="primary"
                    disabled={!text.trim() || !s.connected || !!busy}
                  >
                    <Send size={17} /> Send
                  </button>
                </form>
                <p className="footnote">
                  Acknowledgment means server acceptance and socket writes. It
                  does not mean the recipient has read the message.
                </p>
              </section>
            </>
          )}
          {tab === "Transfers" && (
            <>
              {connectPanel}
              <section className="panel">
                <div className="panel-heading">
                  <h2>Transfer queue</h2>
                  <span className="small-label">{queue.length} FILES</span>
                </div>
                <div
                  className="dropzone"
                  style={
                    { "--wails-drop-target": "drop" } as React.CSSProperties
                  }
                >
                  <div className="drop-icon">
                    <FileUp size={29} />
                  </div>
                  <h2>Drop files into the flow</h2>
                  <p>
                    Binary-safe transfer / SHA-256 verification / Up to 8 GiB
                    per file
                  </p>
                  <button
                    className="secondary"
                    disabled={running || !!busy}
                    onClick={() =>
                      act("Choose files", async () => {
                        const paths = await call<string[]>("ChooseFiles");
                        setQueue((q) =>
                          [...new Set([...q, ...paths])].slice(0, 100),
                        );
                      })
                    }
                  >
                    <FolderOpen size={16} /> Browse files
                  </button>
                </div>
                {queue.map((p) => (
                  <div className="queue-row" key={p}>
                    <FileUp size={18} />
                    <div>
                      <strong>{fileName(p)}</strong>
                      <small>{p}</small>
                    </div>
                    <button title="Preview" onClick={() => showPreview(p)}>
                      <ImageIcon size={17} />
                    </button>
                    <button
                      title="Remove from queue"
                      disabled={running}
                      onClick={() => setQueue((q) => q.filter((x) => x !== p))}
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                ))}
                <div className="queue-actions">
                  <span>
                    Destination: <code>{target}</code>
                  </span>
                  {running ? (
                    <button
                      className="secondary"
                      onClick={() => {
                        stop.current = true;
                        call("CancelTransfer").catch((e) =>
                          setError(String(e)),
                        );
                      }}
                    >
                      <Square size={15} /> Cancel transfer
                    </button>
                  ) : (
                    <button
                      className="primary"
                      disabled={!queue.length}
                      onClick={startQueue}
                    >
                      <Play size={16} /> Start queue
                    </button>
                  )}
                </div>
              </section>
              {s.progress.map((p) => (
                <section className="panel progress-card" key={p.id}>
                  <div>
                    <strong>{p.name}</strong>
                    <span className="tag">{p.status}</span>
                  </div>
                  <progress value={p.bytes} max={p.total || 1} />
                  <div>
                    <small>
                      {bytes(p.bytes)} / {bytes(p.total)}
                    </small>
                    <small>
                      {bytes(Math.round(speeds[p.id] || 0))}/s /{" "}
                      {(speeds[p.id] || 0) > 0
                        ? Math.ceil((p.total - p.bytes) / speeds[p.id]) +
                          "s remaining"
                        : "ETA pending"}
                    </small>
                  </div>
                </section>
              ))}
              <p className="footnote">
                Video files are transferred, not streamed live. Interrupted
                files restart from the beginning. File names must use letters,
                numbers, spaces, dots, hyphens, or underscores.
              </p>
            </>
          )}
          {tab === "Network Lab" && (
            <>
              <div className="lab-grid">
                <section className="panel">
                  <div className="panel-heading">
                    <h2>Configure experiment</h2>
                    <FlaskConical size={18} />
                  </div>
                  <label>
                    Transport
                    <select
                      value={protocol}
                      onChange={(e) => setProtocol(e.target.value)}
                    >
                      <option value="tcp">TCP / reliable stream</option>
                      <option value="udp">UDP / best-effort datagrams</option>
                    </select>
                  </label>
                  <label>
                    Endpoint
                    <input
                      value={protocol === "tcp" ? target : udp}
                      onChange={(e) =>
                        protocol === "tcp"
                          ? setTarget(e.target.value)
                          : setUdp(e.target.value)
                      }
                    />
                  </label>
                  <div className="form-pair">
                    <label>
                      Probes
                      <input
                        type="number"
                        min="1"
                        max="1000"
                        value={count}
                        onChange={(e) => setCount(+e.target.value)}
                      />
                    </label>
                    <label>
                      Payload bytes
                      <input
                        type="number"
                        min="16"
                        max="1200"
                        value={size}
                        onChange={(e) => setSize(+e.target.value)}
                      />
                    </label>
                  </div>
                  <button
                    className="primary full"
                    disabled={!!busy}
                    onClick={() =>
                      act("Measuring", () =>
                        call(
                          "Benchmark",
                          protocol,
                          protocol === "tcp" ? target : udp,
                          count,
                          size,
                        ),
                      )
                    }
                  >
                    <Play size={16} />
                    {busy === "Measuring" ? "Measuring..." : "Run experiment"}
                  </button>
                  <p className="footnote">
                    Start the remote receiver first. Echo goodput is an
                    application metric, not a link-capacity estimate.
                  </p>
                </section>
                <section className="panel">
                  <div className="panel-heading">
                    <h2>Measured round-trip times</h2>
                    <span className="tag">
                      {latest?.protocol.toUpperCase() || "NO RUN"}
                    </span>
                  </div>
                  <Chart
                    values={
                      latest?.samples
                        .filter((x) => x.received)
                        .map((x) => x.rttMs) || []
                    }
                  />
                  <div className="mini-stats">
                    <div>
                      <small>MEAN RTT</small>
                      <strong>
                        {latest ? latest.rttMs.toFixed(3) : "N/A"} <em>ms</em>
                      </strong>
                    </div>
                    <div>
                      <small>LOSS</small>
                      <strong>
                        {latest ? latest.lossPercent.toFixed(1) : "N/A"}{" "}
                        <em>%</em>
                      </strong>
                    </div>
                    <div>
                      <small>RTT JITTER</small>
                      <strong>
                        {latest ? latest.jitterMs.toFixed(3) : "N/A"}{" "}
                        <em>ms</em>
                      </strong>
                    </div>
                  </div>
                </section>
              </div>
              <section className="panel">
                <div className="panel-heading">
                  <h2>Experiment ledger</h2>
                  <button
                    className="text-button"
                    disabled={!s.results.length}
                    onClick={() => act("Export", () => call("ExportResults"))}
                  >
                    Export raw JSON <ArrowUpRight size={15} />
                  </button>
                </div>
                <table>
                  <thead>
                    <tr>
                      <th>Transport</th>
                      <th>Endpoint</th>
                      <th>Received</th>
                      <th>Loss</th>
                      <th>Mean RTT</th>
                      <th>Echo goodput</th>
                    </tr>
                  </thead>
                  <tbody>
                    {s.results.map((r, i) => (
                      <tr key={i}>
                        <td>
                          <span className="tag">
                            {r.protocol.toUpperCase()}
                          </span>
                        </td>
                        <td>
                          <code>{r.address}</code>
                        </td>
                        <td>
                          {r.received}/{r.sent}
                        </td>
                        <td>{r.lossPercent.toFixed(1)}%</td>
                        <td>{r.rttMs.toFixed(3)} ms</td>
                        <td>{r.mbps.toFixed(3)} Mbps</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                {!s.results.length && (
                  <Empty
                    icon={<FlaskConical />}
                    title="No measurements yet"
                    text="Run both transports against the same host and payload size."
                  />
                )}
              </section>
              <section className="panel">
                <div className="panel-heading">
                  <h2>Known peers and LAN discovery</h2>
                  <Radio size={18} />
                </div>
                <div className="inline-form">
                  <input
                    aria-label="Discovery destination"
                    value={discovery}
                    onChange={(e) => setDiscovery(e.target.value)}
                  />
                  <button
                    className="secondary"
                    disabled={!!busy}
                    onClick={() =>
                      act("Discovering", () => call("Discover", discovery))
                    }
                  >
                    Discover peers
                  </button>
                </div>
                {s.peers.map((p) => (
                  <div className="queue-row" key={p.address}>
                    <Radio size={18} />
                    <div>
                      <strong>{p.name}</strong>
                      <small>{p.address}</small>
                    </div>
                    <button
                      className="text-button"
                      onClick={() => {
                        setTarget(p.address);
                        setTab("Transfers");
                      }}
                    >
                      Use endpoint <ChevronRight size={15} />
                    </button>
                  </div>
                ))}
                <p className="footnote">
                  Discovery uses UDP broadcast on port 42190. Firewalls and
                  subnet boundaries may prevent replies; manual addressing
                  remains available.
                </p>
              </section>
            </>
          )}
          {tab === "History" && (
            <section className="panel">
              <div className="panel-heading">
                <h2>
                  <Database size={17} /> Local history
                </h2>
                <select
                  aria-label="History filter"
                  value={filter}
                  onChange={(e) => setFilter(e.target.value)}
                >
                  <option value="all">All events</option>
                  <option value="message">Messages</option>
                  <option value="transfer">Transfers</option>
                </select>
              </div>
              {s.history
                .filter((e) => filter === "all" || e.kind === filter)
                .map((e, i) => (
                  <div className="history-row" key={i}>
                    <div className="activity-icon">
                      {e.kind === "transfer" ? (
                        <ShieldCheck size={18} />
                      ) : (
                        <MessageSquare size={18} />
                      )}
                    </div>
                    <div>
                      <strong>{e.name || e.peer || e.kind}</strong>
                      <p>{e.text || e.sha256 || e.status}</p>
                      <small>
                        {new Date(e.at).toLocaleString()} / {e.status}
                        {e.bytes ? " / " + bytes(e.bytes) : ""}
                      </small>
                    </div>
                    {e.path && (
                      <button
                        className="text-button"
                        onClick={() => showPreview(e.path)}
                      >
                        Preview
                      </button>
                    )}
                  </div>
                ))}
              {!s.history.length && (
                <Empty
                  icon={<History />}
                  title="A clean slate"
                  text="Completed activity will be recorded in SQLite."
                />
              )}
            </section>
          )}
          {tab === "Settings" && (
            <section className="panel settings">
              <h2>Node configuration</h2>
              <p>
                Use 127.0.0.1 for same-machine experiments. Use 0.0.0.0 to
                listen on your LAN interfaces.
              </p>
              <label>
                TCP listen address
                <input
                  value={tcp}
                  onChange={(e) => setTcp(e.target.value)}
                  disabled={!!s.listening}
                />
              </label>
              <label>
                UDP echo address
                <input
                  value={udp}
                  onChange={(e) => setUdp(e.target.value)}
                  disabled={!!s.listening}
                />
              </label>
              <label>
                Display name
                <input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  maxLength={64}
                />
              </label>
              <button
                className="primary"
                onClick={() =>
                  act("Listener", () =>
                    s.listening
                      ? call("StopHost")
                      : call("StartHost", tcp, udp, name),
                  )
                }
              >
                {s.listening ? "Stop receiver" : "Start receiver"}
              </button>
              <button
                className="secondary"
                disabled={!!busy}
                onClick={() =>
                  act("Saving settings", () =>
                    call("SaveSettings", { name, tcp, udp, target, discovery }),
                  )
                }
              >
                <Database size={16} /> Save node settings
              </button>
              <p className="footnote">
                Saved settings apply to the next connection or receiver start.
                Running sockets keep their existing addresses.
              </p>
              <hr />
              <h2>Local storage</h2>
              <p className="mono">
                {s.dataDir || "Available inside the desktop application"}
              </p>
              <p className="footnote">
                History and files stay on this machine. Theme, saved addresses,
                peer name, known peers and recent experiments are stored in
                SQLite.
              </p>
              <button
                className="secondary"
                onClick={() =>
                  act("Theme", () =>
                    call("SetTheme", s.theme === "light" ? "dark" : "light"),
                  )
                }
              >
                <Sun size={16} /> Switch to{" "}
                {s.theme === "light" ? "dark" : "light"} theme
              </button>
              <hr />
              <h2>Laboratory boundaries</h2>
              <p>
                Trusted LAN only. No encryption or authenticated peer identity.
                No internet NAT traversal, live video calling, or custom TCP
                congestion-control implementation.
              </p>
            </section>
          )}
          <footer>
            <span>
              <Activity size={13} /> PULSEROUTE NETWORK LABORATORY
            </span>
            <span>Real sockets. Recorded evidence.</span>
          </footer>
        </div>
      </main>
      {preview && (
        <div
          className="modal"
          role="dialog"
          aria-modal="true"
          aria-label="File preview"
        >
          <section>
            <div className="panel-heading">
              <h2>{previewName}</h2>
              <button aria-label="Close preview" onClick={() => setPreview("")}>
                <X />
              </button>
            </div>
            {preview.startsWith("data:video") ? (
              <video src={preview} controls />
            ) : (
              <img src={preview} alt={previewName} />
            )}
          </section>
        </div>
      )}
    </div>
  );
}
function Stat({
  icon,
  label,
  value,
  detail,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  detail: string;
}) {
  return (
    <section className="stat">
      <div>
        <span>{label}</span>
        {icon}
      </div>
      <strong>{value}</strong>
      <small>{detail}</small>
    </section>
  );
}
function Empty({
  icon,
  title,
  text,
}: {
  icon: React.ReactNode;
  title: string;
  text: string;
}) {
  return (
    <div className="empty">
      {icon}
      <strong>{title}</strong>
      <p>{text}</p>
    </div>
  );
}
createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
