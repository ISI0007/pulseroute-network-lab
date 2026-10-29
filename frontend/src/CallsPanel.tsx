import React, { useEffect, useRef, useState } from "react";
import {
  Check,
  Mic,
  MicOff,
  Phone,
  PhoneIncoming,
  PhoneOff,
  Video,
  VideoOff,
  Volume2,
  X,
} from "lucide-react";
import {
  CallSignalEnvelope,
  CallState,
  CallStats,
  CallSummary,
  LanCallController,
  emptyCallState,
  mediaFailure,
} from "./rtc";
export type CallPeer = { id: string; name: string };
export type CallsPanelProps = {
  connected: string;
  selfId: string;
  peers: CallPeer[];
  signals: CallSignalEnvelope[];
  onError?: (message: string) => void;
  onCallState?: (state: CallState) => void;
  onCallEnded?: (summary: CallSummary) => void;
};
type NativeWindow = {
  go?: {
    main?: {
      App?: {
        SendCallSignal?: (
          to: string,
          id: string,
          text: string,
        ) => Promise<void>;
      };
    };
  };
  runtime?: {
    EventsOn?: (
      name: string,
      callback: (...values: unknown[]) => void,
    ) => (() => void) | void;
  };
};
const activePhases = new Set([
  "incoming",
  "acquiring",
  "ringing",
  "connecting",
  "connected",
  "reconnecting",
]);
const phaseLabel: Record<CallState["phase"], string> = {
  idle: "Ready for a call",
  incoming: "Incoming call",
  acquiring: "Preparing your devices",
  ringing: "Calling",
  connecting: "Connecting media",
  connected: "Live call",
  reconnecting: "Reconnecting media",
  ended: "Call ended",
  error: "Call unavailable",
};
const byteLabel = (n: number) =>
  n >= 1048576
    ? (n / 1048576).toFixed(1) + " MiB"
    : n >= 1024
      ? (n / 1024).toFixed(1) + " KiB"
      : n + " B";
const durationLabel = (n: number) =>
  Math.floor(n / 60)
    .toString()
    .padStart(2, "0") +
  ":" +
  (n % 60).toString().padStart(2, "0");
export function CallsPanel(props: CallsPanelProps) {
  const [state, setState] = useState<CallState>({ ...emptyCallState });
  const [selected, setSelected] = useState("");
  const [error, setError] = useState("");
  const [local, setLocal] = useState<MediaStream | null>(null);
  const [remote, setRemote] = useState<MediaStream | null>(null);
  const [remoteVideoReady, setRemoteVideoReady] = useState(false);
  const [audioBlocked, setAudioBlocked] = useState(false);
  const [stats, setStats] = useState<CallStats | null>(null);
  const [elapsed, setElapsed] = useState(0);
  const callbacks = useRef(props);
  callbacks.current = props;
  const controller = useRef<LanCallController | null>(null);
  const processed = useRef(new Set<number>());
  const localVideo = useRef<HTMLVideoElement>(null);
  const remoteVideo = useRef<HTMLVideoElement>(null);
  const receive = useRef<(value: unknown) => void>(() => {});
  receive.current = (value) => {
    if (!value || typeof value !== "object" || !controller.current) return;
    const envelope = value as CallSignalEnvelope;
    if (
      !Number.isSafeInteger(envelope.seq) ||
      envelope.seq < 1 ||
      processed.current.has(envelope.seq)
    )
      return;
    processed.current.add(envelope.seq);
    if (processed.current.size > 512)
      processed.current.delete(processed.current.values().next().value!);
    controller.current.receive(envelope);
  };
  useEffect(() => {
    setState({ ...emptyCallState });
    setLocal(null);
    setRemote(null);
    setStats(null);
    setError("");
    setAudioBlocked(false);
    if (!props.connected || !props.selfId) {
      callbacks.current.onCallState?.({ ...emptyCallState });
      return;
    }
    const call = new LanCallController({
      send: async (to, id, signal) => {
        const send = (window as unknown as NativeWindow).go?.main?.App
          ?.SendCallSignal;
        if (!send)
          throw Error(
            "Open the current PulseRoute desktop app to place calls.",
          );
        await send(to, id, JSON.stringify(signal));
      },
      onState: (next) => {
        setState(next);
        callbacks.current.onCallState?.(next);
      },
      onLocalStream: setLocal,
      onRemoteStream: setRemote,
      onError: (message) => {
        setError(message);
        callbacks.current.onError?.(message);
      },
      onEnded: (summary) => callbacks.current.onCallEnded?.(summary),
    });
    controller.current = call;
    return () => {
      call.destroy();
      if (controller.current === call) controller.current = null;
    };
  }, [props.connected, props.selfId]);
  useEffect(() => {
    for (const signal of [...props.signals].sort((a, b) => a.seq - b.seq))
      receive.current(signal);
  }, [props.signals, props.connected, props.selfId]);
  useEffect(() => {
    if (!props.connected || !props.selfId) return;
    let unsubscribe: (() => void) | void;
    const subscribe = () => {
      const events = (window as unknown as NativeWindow).runtime?.EventsOn;
      if (!events) return false;
      unsubscribe = events("pulseroute:call", (...values) =>
        receive.current(values[0]),
      );
      return true;
    };
    const timer = subscribe()
      ? undefined
      : window.setInterval(() => {
          if (subscribe()) window.clearInterval(timer);
        }, 1000);
    return () => {
      window.clearInterval(timer);
      if (typeof unsubscribe === "function") unsubscribe();
    };
  }, [props.connected, props.selfId]);
  useEffect(() => {
    const video = localVideo.current;
    if (!video) return;
    video.srcObject = local;
    if (local) void video.play().catch(() => {});
    return () => {
      video.srcObject = null;
    };
  }, [local, state.mode]);
  useEffect(() => {
    const video = remoteVideo.current;
    if (!video) return;
    video.srcObject = remote;
    setAudioBlocked(false);
    if (remote) void video.play().catch(() => setAudioBlocked(true));
    return () => {
      video.srcObject = null;
    };
  }, [remote, state.mode]);
  useEffect(() => {
    const tracks = remote?.getVideoTracks() || [];
    const refresh = () =>
      setRemoteVideoReady(
        tracks.some((t) => t.readyState === "live" && !t.muted),
      );
    refresh();
    tracks.forEach((t) => {
      t.addEventListener("mute", refresh);
      t.addEventListener("unmute", refresh);
      t.addEventListener("ended", refresh);
    });
    return () =>
      tracks.forEach((t) => {
        t.removeEventListener("mute", refresh);
        t.removeEventListener("unmute", refresh);
        t.removeEventListener("ended", refresh);
      });
  }, [remote]);
  useEffect(() => {
    if (state.phase !== "connected" && state.phase !== "reconnecting") {
      setStats(null);
      setElapsed(0);
      return;
    }
    let alive = true,
      reading = false;
    const refresh = async () => {
      if (!alive || reading || document.hidden) return;
      reading = true;
      try {
        const value = await controller.current?.stats();
        if (alive) {
          setStats(value || null);
          setElapsed(
            state.startedAt
              ? Math.max(0, Math.floor((Date.now() - state.startedAt) / 1000))
              : 0,
          );
        }
      } catch {
        if (alive) setStats(null);
      } finally {
        reading = false;
      }
    };
    void refresh();
    const timer = window.setInterval(refresh, 1000);
    return () => {
      alive = false;
      window.clearInterval(timer);
    };
  }, [state.phase, state.startedAt]);
  const peers = props.peers.filter((peer) => peer.id !== props.selfId);
  useEffect(() => {
    if (!peers.some((peer) => peer.id === selected))
      setSelected(peers[0]?.id || "");
  }, [props.peers, props.selfId, selected]);
  const active = activePhases.has(state.phase);
  const canStart = Boolean(
    props.connected && props.selfId && selected && !active,
  );
  const mediaReady = Boolean(local && state.phase !== "incoming");
  const start = async (mode: "audio" | "video") => {
    const peer = peers.find((p) => p.id === selected);
    if (!peer || !controller.current) return;
    setError("");
    try {
      await controller.current.start(peer.id, peer.name, mode);
    } catch (failure) {
      const message = mediaFailure(failure);
      setError(message);
      callbacks.current.onError?.(message);
    }
  };
  const resumeAudio = async () => {
    try {
      await remoteVideo.current?.play();
      setAudioBlocked(false);
    } catch {
      setAudioBlocked(true);
    }
  };
  return (
    <section className="panel calls-panel" aria-label="Live calls">
      <div className="panel-heading">
        <div>
          <h2>
            <Phone size={18} /> Live voice and video
          </h2>
          <p className="footnote">
            One-to-one calls between PulseRoute users on the same LAN.
          </p>
        </div>
        <span className="small-label">WEBRTC / LOCAL NETWORK</span>
      </div>
      {!props.connected && (
        <div className="notice" role="status">
          Connect both computers to the same PulseRoute hub before calling.
        </div>
      )}
      {props.connected && !props.selfId && (
        <div className="notice" role="status">
          This connection does not support calling. Update the hub and reconnect
          both apps.
        </div>
      )}
      {error && (
        <div className="error" role="alert">
          {error}
        </div>
      )}
      <div className="call-peer-list">
        <label>
          Call a connected peer
          <select
            aria-label="Call peer"
            value={selected}
            onChange={(event) => setSelected(event.target.value)}
            disabled={!props.connected || !props.selfId || active}
          >
            {!peers.length && (
              <option value="">No other connected peers</option>
            )}
            {peers.map((peer) => (
              <option key={peer.id} value={peer.id}>
                {peer.name} / {peer.id.slice(0, 8)}
              </option>
            ))}
          </select>
        </label>
        <div className="call-actions">
          <button
            className="primary"
            disabled={!canStart}
            onClick={() => void start("audio")}
          >
            <Phone size={16} /> Voice call
          </button>
          <button
            className="secondary"
            disabled={!canStart}
            onClick={() => void start("video")}
          >
            <Video size={16} /> Video call
          </button>
        </div>
      </div>
      <div
        className={
          "call-status " + (state.phase === "incoming" ? "incoming" : "")
        }
        role={state.phase === "incoming" ? "alert" : "status"}
        aria-live="polite"
      >
        {state.phase === "incoming" ? (
          <PhoneIncoming size={22} />
        ) : state.phase === "connected" ? (
          <span className="dot live" />
        ) : (
          <Phone size={20} />
        )}
        <div>
          <strong>
            {phaseLabel[state.phase]}
            {active && state.peerName ? " / " + state.peerName : ""}
          </strong>
          <p>{state.reason || "Choose a peer and call when you are ready."}</p>
        </div>
        {(state.phase === "connected" || state.phase === "reconnecting") && (
          <span className="mono">{durationLabel(elapsed)}</span>
        )}
      </div>
      {state.phase === "incoming" && (
        <div className="call-actions">
          <button
            className="primary"
            onClick={() => {
              setError("");
              void controller.current?.accept();
            }}
          >
            <Check size={16} /> Accept{" "}
            {state.mode === "video" ? "video" : "voice"} call
          </button>
          <button
            className="secondary"
            onClick={() => controller.current?.reject()}
          >
            <X size={16} /> Decline
          </button>
        </div>
      )}
      {active && state.phase !== "incoming" && (
        <div
          className="call-stage"
          aria-label={
            state.mode === "video" ? "Live video call" : "Live voice call"
          }
        >
          <div className="call-remote">
            <video
              className={
                state.mode === "audio" ? "call-video audio-only" : "call-video"
              }
              ref={remoteVideo}
              autoPlay
              playsInline
              aria-label="Remote call media"
            />
            {(!remote || state.mode === "audio" || !remoteVideoReady) && (
              <div className="call-placeholder">
                {state.mode === "video" ? (
                  <VideoOff size={42} />
                ) : (
                  <Phone size={42} />
                )}
                <strong>{state.peerName}</strong>
                <span>
                  {state.mode === "audio"
                    ? "Voice call"
                    : "Waiting for remote camera video"}
                </span>
              </div>
            )}
            <span className="call-media-label">
              {state.peerName || "Remote peer"}
            </span>
          </div>
          {state.mode === "video" && (
            <div className="call-local">
              <video
                className="call-video"
                ref={localVideo}
                autoPlay
                muted
                playsInline
                aria-label="Your camera preview"
              />
              {!state.cameraEnabled && (
                <div className="call-placeholder">
                  <VideoOff size={24} />
                  <span>Camera off</span>
                </div>
              )}
              <span className="call-media-label">You / muted preview</span>
            </div>
          )}
        </div>
      )}
      {audioBlocked && (
        <button className="primary" onClick={() => void resumeAudio()}>
          <Volume2 size={16} /> Play call audio
        </button>
      )}
      {active && state.phase !== "incoming" && (
        <div className="call-actions">
          <button
            className="secondary"
            disabled={!mediaReady}
            aria-pressed={state.muted}
            onClick={() => controller.current?.mute()}
          >
            {state.muted ? <MicOff size={16} /> : <Mic size={16} />}
            {state.muted ? "Unmute microphone" : "Mute microphone"}
          </button>
          {state.mode === "video" && (
            <button
              className="secondary"
              disabled={!mediaReady || state.cameraBusy}
              aria-pressed={state.cameraEnabled}
              onClick={() => void controller.current?.toggleCamera()}
            >
              {state.cameraEnabled ? (
                <VideoOff size={16} />
              ) : (
                <Video size={16} />
              )}
              {state.cameraBusy
                ? "Changing camera..."
                : state.cameraEnabled
                  ? "Turn camera off"
                  : "Turn camera on"}
            </button>
          )}
          <button
            className="secondary call-end"
            onClick={() => controller.current?.hangup()}
          >
            <PhoneOff size={16} /> End call
          </button>
        </div>
      )}
      {stats && (
        <div className="call-stats" aria-label="Measured call statistics">
          <span>
            Media sent <strong>{byteLabel(stats.bytesSent)}</strong>
          </span>
          <span>
            Media received <strong>{byteLabel(stats.bytesReceived)}</strong>
          </span>
          <span>
            Reported packet loss <strong>{stats.packetsLost}</strong>
          </span>
          <span>
            ICE RTT{" "}
            <strong>
              {stats.rttMs === null ? "N/A" : stats.rttMs.toFixed(1) + " ms"}
            </strong>
          </span>
        </div>
      )}
      <p className="footnote">
        Camera and microphone start only when you call or accept. Ending a call
        releases them. No recording is performed. Media uses WebRTC; signaling
        uses your selected hub. Names are not verified identities. No
        phone-number calling, public STUN/TURN service or Internet connectivity
        guarantee.
      </p>
    </section>
  );
}
export default CallsPanel;
