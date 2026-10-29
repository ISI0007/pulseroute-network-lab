/** Real LAN WebRTC call lifecycle. Signaling is supplied by the native TCP bridge. */
export type CallMode = "audio" | "video";
export type CallPhase =
  | "idle"
  | "incoming"
  | "acquiring"
  | "ringing"
  | "connecting"
  | "connected"
  | "reconnecting"
  | "ended"
  | "error";
export type CallSignal = {
  type: "offer" | "answer" | "ice" | "reject" | "hangup" | "busy";
  mode?: CallMode;
  sdp?: string;
  candidate?: RTCIceCandidateInit;
  reason?: string;
};
export type CallSignalEnvelope = {
  seq: number;
  from: string;
  peer: string;
  id: string;
  text: string;
  error?: string;
};
export type CallState = {
  phase: CallPhase;
  id: string;
  peerId: string;
  peerName: string;
  mode: CallMode;
  reason: string;
  muted: boolean;
  cameraEnabled: boolean;
  cameraBusy: boolean;
  startedAt: number | null;
};
export type CallSummary = {
  id: string;
  peer: string;
  mode: CallMode;
  status: string;
  durationSeconds: number;
};
export type CallStats = {
  bytesSent: number;
  bytesReceived: number;
  packetsLost: number;
  rttMs: number | null;
};
export const emptyCallState: CallState = {
  phase: "idle",
  id: "",
  peerId: "",
  peerName: "",
  mode: "audio",
  reason: "",
  muted: false,
  cameraEnabled: false,
  cameraBusy: false,
  startedAt: null,
};
const validId = (id: string) => /^[0-9a-f]{32}$/i.test(id);
export function newCallId(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return [...bytes].map((n) => n.toString(16).padStart(2, "0")).join("");
}
export function mediaFailure(error: unknown): string {
  const name = error instanceof DOMException ? error.name : "";
  if (name === "NotAllowedError")
    return "Camera or microphone permission was denied. Review the app and Windows privacy permissions, then try again.";
  if (name === "NotFoundError")
    return "No requested camera or microphone was found. Connect a device or choose a voice-only call.";
  if (name === "NotReadableError")
    return "The camera or microphone is unavailable or in use by another application.";
  if (name === "SecurityError")
    return "Media capture requires a supported secure application context.";
  return error instanceof Error
    ? error.message
    : "The call could not be started.";
}
export function parseCallSignal(text: string): CallSignal {
  if (new TextEncoder().encode(text).length > 60 * 1024)
    throw Error("Call signal exceeds the size limit.");
  const value: unknown = JSON.parse(text);
  if (!value || typeof value !== "object") throw Error("Invalid call signal.");
  const s = value as CallSignal;
  if (!["offer", "answer", "ice", "reject", "hangup", "busy"].includes(s.type))
    throw Error("Unknown call signal.");
  if (s.type === "offer" && s.mode !== "audio" && s.mode !== "video")
    throw Error("Invalid call mode.");
  if (
    (s.type === "offer" || s.type === "answer") &&
    (typeof s.sdp !== "string" || !s.sdp.trim())
  )
    throw Error("Missing session description.");
  if (
    s.type === "ice" &&
    (!s.candidate || typeof s.candidate.candidate !== "string")
  )
    throw Error("Invalid network candidate.");
  return s;
}
type CallOptions = {
  send: (to: string, id: string, signal: CallSignal) => Promise<void>;
  onState: (state: CallState) => void;
  onLocalStream: (stream: MediaStream | null) => void;
  onRemoteStream: (stream: MediaStream | null) => void;
  onError?: (message: string) => void;
  onEnded?: (summary: CallSummary) => void;
  /** Tests may inject explicitly synthetic tracks; the application uses real devices. */
  acquireMedia?: (constraints: MediaStreamConstraints) => Promise<MediaStream>;
  createPeerConnection?: () => RTCPeerConnection;
};
type Session = {
  id: string;
  peerId: string;
  peerName: string;
  mode: CallMode;
  pc: RTCPeerConnection | null;
  local: MediaStream | null;
  remote: MediaStream | null;
  offer?: string;
  candidates: RTCIceCandidateInit[];
  outbound: RTCIceCandidateInit[];
  outboundReady: boolean;
  sentCandidates: number;
  receivedCandidates: number;
  videoSender?: RTCRtpSender;
  timer?: ReturnType<typeof setTimeout>;
  disconnectedTimer?: ReturnType<typeof setTimeout>;
};
export class LanCallController {
  private session: Session | null = null;
  private state: CallState = { ...emptyCallState };
  private work: Promise<void> = Promise.resolve();
  private earlyIce = new Map<
    string,
    { at: number; candidates: RTCIceCandidateInit[] }
  >();
  private finished = new Map<string, number>();
  private disposed = false;
  private queuedSignals = 0;
  constructor(private options: CallOptions) {}
  private isCurrent(session: Session): boolean {
    return this.session === session;
  }
  private publish(patch: Partial<CallState>) {
    if (this.disposed) return;
    this.state = { ...this.state, ...patch };
    this.options.onState({ ...this.state });
  }
  private createSession(
    id: string,
    peerId: string,
    peerName: string,
    mode: CallMode,
  ): Session {
    const key = peerId + ":" + id;
    const session: Session = {
      id,
      peerId,
      peerName,
      mode,
      pc: null,
      local: null,
      remote: null,
      candidates: this.earlyIce.get(key)?.candidates || [],
      outbound: [],
      outboundReady: false,
      sentCandidates: 0,
      receivedCandidates: 0,
    };
    this.earlyIce.delete(key);
    this.session = session;
    this.publish({ ...emptyCallState, id, peerId, peerName, mode });
    session.timer = setTimeout(() => {
      if (this.isCurrent(session))
        this.end(
          "No answer or media connection within 45 seconds.",
          "missed",
          true,
        );
    }, 45000);
    return session;
  }
  private async acquire(session: Session) {
    if (
      !globalThis.isSecureContext ||
      typeof RTCPeerConnection === "undefined" ||
      !navigator.mediaDevices?.getUserMedia
    )
      throw Error(
        "Live calls require a secure desktop WebView with camera/microphone support. LAN HTTP phone browsers cannot capture media.",
      );
    const acquire =
      this.options.acquireMedia ||
      ((constraints) => navigator.mediaDevices.getUserMedia(constraints));
    const stream = await acquire({
      audio: { echoCancellation: true, noiseSuppression: true },
      video:
        session.mode === "video"
          ? {
              width: { ideal: 1280 },
              height: { ideal: 720 },
              frameRate: { ideal: 24, max: 30 },
            }
          : false,
    });
    if (!this.isCurrent(session) || this.disposed) {
      stream.getTracks().forEach((t) => t.stop());
      return false;
    }
    session.local = stream;
    this.options.onLocalStream(stream);
    this.publish({
      muted: false,
      cameraEnabled: stream
        .getVideoTracks()
        .some((t) => t.readyState === "live"),
    });
    return true;
  }
  private makePeer(session: Session) {
    const pc =
      this.options.createPeerConnection?.() ||
      new RTCPeerConnection({ iceServers: [], bundlePolicy: "max-bundle" });
    session.pc = pc;
    session.remote = new MediaStream();
    this.options.onRemoteStream(session.remote);
    for (const track of session.local!.getTracks()) {
      const sender = pc.addTrack(track, session.local!);
      if (track.kind === "video") session.videoSender = sender;
    }
    pc.onicecandidate = (event) => {
      if (!event.candidate || !this.isCurrent(session)) return;
      if (++session.sentCandidates > 128) {
        this.fail(session, Error("Too many network candidates."));
        return;
      }
      const candidate = event.candidate.toJSON();
      if (!session.outboundReady) session.outbound.push(candidate);
      else
        void this.options
          .send(session.peerId, session.id, { type: "ice", candidate })
          .catch((error) => this.fail(session, error));
    };
    pc.ontrack = (event) => {
      if (!this.isCurrent(session)) return;
      if (!session.remote!.getTracks().some((t) => t.id === event.track.id))
        session.remote!.addTrack(event.track);
      this.options.onRemoteStream(new MediaStream(session.remote!.getTracks()));
    };
    pc.onconnectionstatechange = () => {
      if (!this.isCurrent(session)) return;
      if (pc.connectionState === "connected") {
        clearTimeout(session.timer);
        clearTimeout(session.disconnectedTimer);
        this.publish({
          phase: "connected",
          reason: "",
          startedAt: this.state.startedAt ?? Date.now(),
        });
      } else if (pc.connectionState === "failed")
        this.fail(
          session,
          Error(
            "The LAN media connection failed. Check both devices are reachable and permissions are allowed.",
          ),
        );
      else if (pc.connectionState === "disconnected") {
        this.publish({
          phase: "reconnecting",
          reason: "Media connection interrupted.",
        });
        clearTimeout(session.disconnectedTimer);
        session.disconnectedTimer = setTimeout(() => {
          if (this.isCurrent(session) && pc.connectionState !== "connected")
            this.end("Media connection lost.", "failed", true);
        }, 8000);
      }
    };
    return pc;
  }
  private async flushCandidates(session: Session) {
    if (!session.pc?.remoteDescription) return;
    for (const candidate of session.candidates.splice(0)) {
      if (!this.isCurrent(session)) return;
      await session.pc.addIceCandidate(candidate);
    }
  }
  private async descriptionSent(session: Session) {
    if (!this.isCurrent(session)) return;
    session.outboundReady = true;
    for (const candidate of session.outbound.splice(0)) {
      if (!this.isCurrent(session)) return;
      await this.options.send(session.peerId, session.id, {
        type: "ice",
        candidate,
      });
    }
  }
  private fail(session: Session, error: unknown) {
    if (!this.isCurrent(session)) return;
    const message = mediaFailure(error);
    this.options.onError?.(message);
    this.end(message, "failed", true);
  }
  async start(peerId: string, peerName: string, mode: CallMode): Promise<void> {
    if (this.disposed || this.session)
      throw Error("Finish the current call first.");
    if (!validId(peerId)) throw Error("Select an online PulseRoute peer.");
    const session = this.createSession(newCallId(), peerId, peerName, mode);
    this.publish({
      phase: "acquiring",
      reason:
        "Requesting microphone" +
        (mode === "video" ? " and camera" : "") +
        " permission.",
    });
    try {
      if (!(await this.acquire(session))) return;
      const pc = this.makePeer(session);
      const offer = await pc.createOffer();
      if (!this.isCurrent(session)) return;
      await pc.setLocalDescription(offer);
      if (!this.isCurrent(session)) return;
      await this.options.send(peerId, session.id, {
        type: "offer",
        mode,
        sdp: pc.localDescription!.sdp,
      });
      if (!this.isCurrent(session)) return;
      this.publish({
        phase: "ringing",
        reason: "Waiting for the other person to accept.",
      });
      await this.descriptionSent(session);
    } catch (error) {
      this.fail(session, error);
    }
  }
  async accept(): Promise<void> {
    const session = this.session;
    if (!session?.offer || this.state.phase !== "incoming") return;
    this.publish({
      phase: "acquiring",
      reason:
        "Requesting microphone" +
        (session.mode === "video" ? " and camera" : "") +
        " permission.",
    });
    try {
      if (!(await this.acquire(session))) return;
      const pc = this.makePeer(session);
      await pc.setRemoteDescription({ type: "offer", sdp: session.offer });
      await this.flushCandidates(session);
      const answer = await pc.createAnswer();
      if (!this.isCurrent(session)) return;
      await pc.setLocalDescription(answer);
      if (!this.isCurrent(session)) return;
      await this.options.send(session.peerId, session.id, {
        type: "answer",
        sdp: pc.localDescription!.sdp,
      });
      if (!this.isCurrent(session)) return;
      this.publish({
        phase: "connecting",
        reason: "Establishing the LAN media path.",
      });
      await this.descriptionSent(session);
    } catch (error) {
      this.fail(session, error);
    }
  }
  receive(envelope: CallSignalEnvelope): void {
    if (this.disposed) return;
    if (this.queuedSignals >= 256) {
      if (this.session && this.session.id === envelope.id)
        this.fail(
          this.session,
          Error("Call signaling queue exceeded its limit."),
        );
      return;
    }
    this.queuedSignals++;
    this.work = this.work
      .then(() => this.process(envelope))
      .catch((error) => {
        if (
          this.session &&
          this.session.id === envelope.id &&
          this.session.peerId === envelope.from
        )
          this.fail(this.session, error);
      })
      .finally(() => {
        this.queuedSignals--;
      });
  }
  private async process(envelope: CallSignalEnvelope) {
    if (this.disposed) return;
    if (envelope.error) {
      if (this.session && (!envelope.id || envelope.id === this.session.id))
        this.end(envelope.error, "failed", false);
      return;
    }
    if (!validId(envelope.id) || !validId(envelope.from)) return;
    const signal = parseCallSignal(envelope.text);
    const key = envelope.from + ":" + envelope.id;
    const now = Date.now();
    for (const [k, item] of this.earlyIce)
      if (now - item.at > 45000) this.earlyIce.delete(k);
    for (const [k, at] of this.finished)
      if (now - at > 60000) this.finished.delete(k);
    if (this.finished.has(key)) return;
    let session = this.session;
    if (signal.type === "offer") {
      if (session) {
        if (session.id === envelope.id && session.peerId === envelope.from)
          return;
        await this.options.send(envelope.from, envelope.id, {
          type: "busy",
          reason: "Already handling a call.",
        });
        return;
      }
      session = this.createSession(
        envelope.id,
        envelope.from,
        envelope.peer || "PulseRoute peer",
        signal.mode!,
      );
      session.offer = signal.sdp;
      this.publish({
        phase: "incoming",
        reason:
          "Incoming " +
          signal.mode +
          " call. Your camera and microphone are off until you accept.",
      });
      return;
    }
    if (
      !session ||
      session.id !== envelope.id ||
      session.peerId !== envelope.from
    ) {
      if (signal.type === "ice") {
        if (!this.earlyIce.has(key) && this.earlyIce.size < 8)
          this.earlyIce.set(key, { at: now, candidates: [] });
        const early = this.earlyIce.get(key);
        if (early && early.candidates.length < 16)
          early.candidates.push(signal.candidate!);
      } else if (signal.type === "hangup" || signal.type === "reject") {
        if (this.finished.size >= 128)
          this.finished.delete(this.finished.keys().next().value!);
        this.finished.set(key, now);
      }
      return;
    }
    if (signal.type === "answer") {
      if (!session.pc || session.pc.signalingState !== "have-local-offer")
        return;
      await session.pc.setRemoteDescription({
        type: "answer",
        sdp: signal.sdp!,
      });
      await this.flushCandidates(session);
      if (this.isCurrent(session) && session.pc.connectionState !== "connected")
        this.publish({
          phase: "connecting",
          reason: "Establishing the LAN media path.",
        });
    } else if (signal.type === "ice") {
      if (++session.receivedCandidates > 128)
        throw Error("Too many remote network candidates.");
      if (!session.pc?.remoteDescription)
        session.candidates.push(signal.candidate!);
      else await session.pc.addIceCandidate(signal.candidate!);
    } else
      this.end(
        signal.reason ||
          (signal.type === "busy"
            ? "The other person is busy."
            : signal.type === "reject"
              ? "Call declined."
              : "The other person ended the call."),
        signal.type === "hangup" ? "ended" : "rejected",
        false,
      );
  }
  reject() {
    const session = this.session;
    if (!session || this.state.phase !== "incoming") return;
    void this.options
      .send(session.peerId, session.id, {
        type: "reject",
        reason: "Call declined.",
      })
      .catch(() => {});
    this.end("Call declined.", "rejected", false);
  }
  mute() {
    if (!this.session?.local) return;
    const muted = !this.state.muted;
    this.session.local.getAudioTracks().forEach((track) => {
      track.enabled = !muted;
    });
    this.publish({ muted });
  }
  async toggleCamera() {
    const session = this.session;
    if (
      !session?.local ||
      !session.videoSender ||
      this.state.cameraBusy ||
      session.mode !== "video"
    )
      return;
    this.publish({ cameraBusy: true });
    try {
      if (this.state.cameraEnabled) {
        await session.videoSender.replaceTrack(null);
        session.local.getVideoTracks().forEach((track) => {
          track.stop();
          session.local!.removeTrack(track);
        });
        if (this.isCurrent(session)) this.publish({ cameraEnabled: false });
      } else {
        const stream = await (
          this.options.acquireMedia ||
          ((c) => navigator.mediaDevices.getUserMedia(c))
        )({ audio: false, video: true });
        const track = stream.getVideoTracks()[0];
        if (!this.isCurrent(session) || !track) {
          stream.getTracks().forEach((t) => t.stop());
          return;
        }
        try {
          await session.videoSender.replaceTrack(track);
        } catch (error) {
          stream.getTracks().forEach((t) => t.stop());
          throw error;
        }
        if (!this.isCurrent(session)) {
          stream.getTracks().forEach((t) => t.stop());
          return;
        }
        session.local.addTrack(track);
        this.options.onLocalStream(new MediaStream(session.local.getTracks()));
        this.publish({ cameraEnabled: true });
      }
    } catch (error) {
      this.options.onError?.(mediaFailure(error));
    } finally {
      if (this.isCurrent(session)) this.publish({ cameraBusy: false });
    }
  }
  async stats(): Promise<CallStats | null> {
    const pc = this.session?.pc;
    if (!pc || pc.connectionState === "closed") return null;
    const report = await pc.getStats();
    const value: CallStats = {
      bytesSent: 0,
      bytesReceived: 0,
      packetsLost: 0,
      rttMs: null,
    };
    report.forEach((item) => {
      if (item.type === "outbound-rtp" && !item.isRemote)
        value.bytesSent += item.bytesSent || 0;
      if (item.type === "inbound-rtp" && !item.isRemote) {
        value.bytesReceived += item.bytesReceived || 0;
        value.packetsLost += Math.max(0, item.packetsLost || 0);
      }
      if (
        item.type === "candidate-pair" &&
        item.state === "succeeded" &&
        item.nominated &&
        typeof item.currentRoundTripTime === "number"
      )
        value.rttMs = item.currentRoundTripTime * 1000;
    });
    return value;
  }
  hangup() {
    this.end(
      "Call ended on this device.",
      this.state.phase === "incoming" ? "rejected" : "ended",
      true,
    );
  }
  private end(reason: string, result: string, notify: boolean) {
    const session = this.session;
    if (!session) return;
    this.session = null;
    if (notify)
      void this.options
        .send(session.peerId, session.id, { type: "hangup", reason })
        .catch(() => {});
    clearTimeout(session.timer);
    clearTimeout(session.disconnectedTimer);
    session.pc?.close();
    session.local?.getTracks().forEach((track) => track.stop());
    session.remote?.getTracks().forEach((track) => track.stop());
    const key = session.peerId + ":" + session.id;
    if (this.finished.size >= 128)
      this.finished.delete(this.finished.keys().next().value!);
    this.finished.set(key, Date.now());
    this.earlyIce.delete(key);
    this.options.onLocalStream(null);
    this.options.onRemoteStream(null);
    this.options.onEnded?.({
      id: session.id,
      peer: session.peerName,
      mode: session.mode,
      status: result,
      durationSeconds: this.state.startedAt
        ? Math.max(0, Math.floor((Date.now() - this.state.startedAt) / 1000))
        : 0,
    });
    this.publish({
      phase: result === "failed" ? "error" : "ended",
      reason,
      muted: false,
      cameraEnabled: false,
      cameraBusy: false,
      startedAt: null,
    });
  }
  destroy() {
    this.end("Signaling connection closed.", "ended", true);
    this.disposed = true;
    this.earlyIce.clear();
    this.finished.clear();
  }
}
