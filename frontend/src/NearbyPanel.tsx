import React, { useEffect, useState } from "react";
import QRCode from "qrcode";
import {
  Bluetooth,
  Copy,
  FileUp,
  QrCode,
  Radio,
  Square,
  Trash2,
} from "lucide-react";
type SharedFile = { id: string; name: string; size: number };
type NearbyInfo = {
  active: boolean;
  listening: string;
  urls: string[];
  localUrl: string;
  expiresAt: string;
  files: SharedFile[];
  bindChoices: { label: string; address: string }[];
  maxUploadBytes: number;
};
type BluetoothInfo = { available: boolean; method: string; detail: string };
const empty: NearbyInfo = {
  active: false,
  listening: "",
  urls: [],
  localUrl: "",
  expiresAt: "",
  files: [],
  bindChoices: [],
  maxUploadBytes: 512 * 1024 * 1024,
};
async function invoke<T>(method: string, ...args: unknown[]): Promise<T> {
  const bridge = (
    window as unknown as {
      go?: {
        main?: { App?: Record<string, (...args: unknown[]) => Promise<T>> };
      };
    }
  ).go?.main?.App;
  if (!bridge?.[method])
    throw Error("Open the current PulseRoute desktop app to share nearby.");
  return bridge[method](...args);
}
const name = (path: string) => path.split(/[\\/]/).pop() || path;
export default function NearbyPanel({
  available,
  onError,
}: {
  available: boolean;
  onError: (message: string) => void;
}) {
  const [info, setInfo] = useState<NearbyInfo>(empty),
    [address, setAddress] = useState(""),
    [paths, setPaths] = useState<string[]>([]),
    [busy, setBusy] = useState(false),
    [notice, setNotice] = useState(""),
    [qr, setQR] = useState(""),
    [bluetooth, setBluetooth] = useState<BluetoothInfo | null>(null);
  const link = info.urls[0] || info.localUrl;
  useEffect(() => {
    if (!available) return;
    let active = true,
      reading = false;
    const refresh = async () => {
      if (reading || document.hidden) return;
      reading = true;
      try {
        const next = await invoke<NearbyInfo>("NearbyStatus");
        if (active) {
          setInfo(next);
          setAddress((old) => old || next.bindChoices[0]?.address || "");
        }
      } catch (error) {
        if (active) onError(String(error));
      } finally {
        reading = false;
      }
    };
    void refresh();
    void invoke<BluetoothInfo>("BluetoothStatus")
      .then((value) => {
        if (active) setBluetooth(value);
      })
      .catch(() => {});
    const timer = window.setInterval(refresh, 2000);
    document.addEventListener("visibilitychange", refresh);
    return () => {
      active = false;
      window.clearInterval(timer);
      document.removeEventListener("visibilitychange", refresh);
    };
  }, [available]);
  useEffect(() => {
    let active = true;
    setQR("");
    if (info.active && link)
      void QRCode.toDataURL(link, {
        width: 240,
        margin: 2,
        errorCorrectionLevel: "M",
      })
        .then((value) => {
          if (active) setQR(value);
        })
        .catch((error) => {
          if (active) onError(String(error));
        });
    return () => {
      active = false;
    };
  }, [info.active, link]);
  const run = async (action: () => Promise<void>) => {
    setBusy(true);
    setNotice("");
    try {
      await action();
    } catch (error) {
      onError(String(error));
    } finally {
      setBusy(false);
    }
  };
  return (
    <div className="nearby-layout">
      <section className="panel nearby-panel">
        <div className="panel-heading">
          <h2>
            <Radio size={18} /> Nearby Wi-Fi sharing
          </h2>
          <span className="small-label">QR / BROWSER COMPANION</span>
        </div>
        <p className="footnote">
          Share with a browser on Windows, Mac, Android, iPhone or iPad on the
          same trusted Wi-Fi. This is PulseRoute nearby sharing, not Apple's
          AirDrop protocol.
        </p>
        {!available && (
          <div className="notice" role="status">
            Open PulseRoute.exe to start a real nearby session. No session is
            active in this browser preview.
          </div>
        )}
        <label>
          Network interface
          <select
            aria-label="Nearby network interface"
            value={address}
            disabled={!available || info.active || busy}
            onChange={(e) => setAddress(e.target.value)}
          >
            <option value="" disabled>
              Select an interface
            </option>
            {info.bindChoices.map((choice) => (
              <option key={choice.address} value={choice.address}>
                {choice.label}
              </option>
            ))}
          </select>
        </label>
        <div className="nearby-file-list">
          {paths.length ? (
            paths.map((path) => (
              <div className="queue-row" key={path}>
                <span>{name(path)}</span>
                <button
                  disabled={busy || info.active}
                  aria-label={"Remove " + name(path) + " from nearby sharing"}
                  onClick={() =>
                    setPaths((old) => old.filter((p) => p !== path))
                  }
                >
                  <Trash2 size={15} />
                </button>
              </div>
            ))
          ) : (
            <p className="footnote">
              Choose files to offer for download, or start an empty session to
              receive files from another device.
            </p>
          )}
        </div>
        <div className="call-actions">
          <button
            className="secondary"
            disabled={!available || info.active || busy}
            onClick={() =>
              run(async () => {
                const files = await invoke<string[]>("ChooseFiles");
                setPaths((old) =>
                  [...new Set([...old, ...files])].slice(0, 128),
                );
              })
            }
          >
            <FileUp size={16} /> Choose files to share
          </button>
          {info.active ? (
            <button
              className="secondary"
              disabled={busy}
              onClick={() =>
                run(async () => {
                  await invoke<void>("StopNearby");
                  setInfo(await invoke<NearbyInfo>("NearbyStatus"));
                  setNotice(
                    "Nearby link revoked. Existing transfers were stopped.",
                  );
                })
              }
            >
              <Square size={15} /> Stop nearby sharing
            </button>
          ) : (
            <button
              className="primary"
              disabled={!available || !address || busy}
              onClick={() =>
                run(async () => {
                  setInfo(
                    await invoke<NearbyInfo>("StartNearby", address, paths),
                  );
                  setNotice(
                    "Temporary nearby session started. Share this link only with people you trust.",
                  );
                })
              }
            >
              <QrCode size={16} /> Start 15-minute sharing
            </button>
          )}
        </div>
        <div className="notice nearby-security">
          <strong>Trusted Wi-Fi only.</strong> Anyone with this temporary link
          can download the selected files and upload files to this computer.
          HTTP traffic is not encrypted. Stop the session to revoke access.
          Phone-browser uploads are limited to 512 MiB; native TCP transfers use
          their separate limit.
        </div>
        {notice && (
          <p role="status" className="footnote">
            {notice}
          </p>
        )}
        {info.active && (
          <div className="nearby-session">
            <div className="nearby-qr">
              {qr && (
                <img
                  src={qr}
                  alt="Nearby sharing QR code"
                  width={240}
                  height={240}
                />
              )}
              <span>Scan with your other device's camera</span>
            </div>
            <div className="nearby-link">
              <strong>
                {info.urls.length
                  ? "Open on your nearby device"
                  : "Local test only"}
              </strong>
              {!info.urls.length && (
                <p className="footnote">
                  This address works on this PC only. Choose a LAN interface for
                  another device.
                </p>
              )}
              <input
                aria-label="Nearby sharing link"
                value={link}
                readOnly
                onFocus={(e) => e.currentTarget.select()}
              />
              <button
                className="secondary"
                onClick={() =>
                  run(async () => {
                    if (!navigator.clipboard)
                      throw Error("Select and copy the link manually.");
                    await navigator.clipboard.writeText(link);
                    setNotice(
                      "Nearby link copied. It contains temporary access permission.",
                    );
                  })
                }
              >
                <Copy size={15} /> Copy nearby link
              </button>
              <p className="footnote">
                Expires {new Date(info.expiresAt).toLocaleTimeString()} ·{" "}
                {info.files.length} selected files · {info.listening}
              </p>
            </div>
          </div>
        )}
      </section>
      <section className="panel bluetooth-panel">
        <div className="panel-heading">
          <h2>
            <Bluetooth size={18} /> Bluetooth assistant
          </h2>
          <span className="small-label">WINDOWS INTEGRATION</span>
        </div>
        <p>
          Use the standard Windows Bluetooth File Transfer assistant with paired
          devices that support Bluetooth file exchange.
        </p>
        <p className="footnote">
          {bluetooth?.detail ||
            "Radio pairing and transfer support depend on the operating system and device. Use the Wi-Fi QR flow for Apple phones/tablets and other browser clients."}
        </p>
        <button
          className="secondary"
          disabled={!available || !bluetooth?.available || busy}
          onClick={() =>
            run(async () => {
              await invoke<void>("OpenBluetoothTransfer");
              setNotice(
                "Windows Bluetooth File Transfer opened. Choose Send or Receive and finish there; PulseRoute does not track radio-transfer completion.",
              );
            })
          }
        >
          <Bluetooth size={16} /> Open Windows Bluetooth transfer
        </button>
        {available && bluetooth && !bluetooth.available && (
          <p className="footnote">
            The Windows Bluetooth assistant is not installed on this machine.
          </p>
        )}
        <p className="footnote">
          Opening the assistant is not proof of a completed Bluetooth transfer.
          No Windows radio, pairing or security setting is changed
          automatically.
        </p>
      </section>
    </div>
  );
}
