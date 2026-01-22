/**
 * Purpose: Main player UI with <audio> stream consumer + controls.
 * Rule: YouTube video is never embedded. We only consume backend audio stream URL.
 *
 * Pause semantics: server-authoritative. When server is paused, we pause audio locally.
 */

import React from "react";
import { api } from "../api";
import type { PlayerState } from "../api";
import { config } from "../config";
import { useAppStore } from "../store/useAppStore";
import { clsx } from "clsx";

export function Player({
  onToast
}: {
  onToast: (t: { title: string; message?: string; kind?: "info" | "success" | "error" }) => void;
}) {
  const role = useAppStore((s) => s.role);
  const player = useAppStore((s) => s.player);
  const setPlayerState = useAppStore((s) => s.setPlayerState);

  const [localVolume, setLocalVolume] = React.useState<number>(player?.volume ?? 0.8);
  const audioRef = React.useRef<HTMLAudioElement | null>(null);
  const [streamKey, setStreamKey] = React.useState(0); // for reconnect
  const [progress, setProgress] = React.useState(0);

  const isOwner = role === "owner";
  const controlsDisabled = role === "unknown";

  // Pull initial state on mount (WS will keep it fresh later)
  React.useEffect(() => {
    (async () => {
      try {
        const st = await api.player.state();
        setPlayerState(st);
        setLocalVolume(st.volume);
      } catch (e) {
        onToast({ title: "Player", message: "Не удалось получить state", kind: "error" });
      }
    })();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Apply volume to audio element
  React.useEffect(() => {
    const a = audioRef.current;
    if (!a) return;
    a.volume = clamp(localVolume, 0, 1);
  }, [localVolume]);

  // If server is paused -> pause locally; if server is playing -> try play (autoplay may be blocked)
  React.useEffect(() => {
    const a = audioRef.current;
    if (!a || !player) return;

    if (player.isPaused || !player.isPlaying) {
      a.pause();
    } else {
      const p = a.play();
      if (p) p.catch(() => void 0);
    }
  }, [player?.isPaused, player?.isPlaying]);

  // Track progress locally (UI only)
  React.useEffect(() => {
    const id = window.setInterval(() => {
      const a = audioRef.current;
      if (!a) return;
      setProgress(a.currentTime || 0);
    }, 250);
    return () => window.clearInterval(id);
  }, []);

  const currentTitle = player?.current?.title ?? "—";
  const duration = player?.durationSec ?? 0;
  const serverPos = player?.positionSec ?? 0;

  // Prefer local currentTime if playing, otherwise fallback to server-reported
  const shownPos = player?.isPlaying && !player?.isPaused ? progress : serverPos;

  const pct = duration > 0 ? Math.min(100, Math.max(0, (shownPos / duration) * 100)) : 0;

  async function play() {
    try {
      await api.player.play();
      onToast({ title: "Player", message: "Play", kind: "success" });
    } catch (e) {
      onToast({ title: "Player", message: errMsg(e), kind: "error" });
    }
  }

  async function pause() {
    try {
      await api.player.pause();
      onToast({ title: "Player", message: "Pause", kind: "success" });
    } catch (e) {
      onToast({ title: "Player", message: errMsg(e), kind: "error" });
    }
  }

  async function next() {
    try {
      await api.player.next();
      onToast({ title: "Player", message: "Next", kind: "success" });
    } catch (e) {
      onToast({ title: "Player", message: errMsg(e), kind: "error" });
    }
  }

  async function prev() {
    try {
      await api.player.prev();
      onToast({ title: "Player", message: "Prev", kind: "success" });
    } catch (e) {
      onToast({ title: "Player", message: errMsg(e), kind: "error" });
    }
  }

  async function commitVolume(v: number) {
    try {
      await api.player.volume(v);
    } catch (e) {
      onToast({ title: "Volume", message: errMsg(e), kind: "error" });
    }
  }

  // Reconnect stream if it drops
  function handleAudioError() {
    onToast({ title: "Stream", message: "Поток оборвался, переподключаем...", kind: "info" });
    setStreamKey((k) => k + 1);
  }

  return (
    <section className="rounded-3xl border border-slate-200 bg-white p-6 shadow-soft dark:border-white/10 dark:bg-slate-900/40">
      <div className="flex flex-col gap-5">
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <div className="text-xs text-slate-500 dark:text-slate-400">Сейчас играет</div>
            <div className="mt-1 truncate text-xl font-semibold tracking-tight">{currentTitle}</div>
            <div className="mt-2 text-[11px] text-slate-500 dark:text-slate-400">
              {player?.current?.addedByNick ? (
                <>
                  Добавил: <b className="text-slate-700 dark:text-slate-200">{player.current.addedByNick}</b>
                </>
              ) : (
                "—"
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Badge label={player?.isPaused ? "PAUSED (server)" : player?.isPlaying ? "PLAYING" : "STOPPED"} />
            <Badge label={isOwner ? "OWNER" : "LISTENER"} subtle />
          </div>
        </div>

        <div className="h-2 w-full overflow-hidden rounded-full bg-slate-100 dark:bg-white/10">
          <div
            className="h-full rounded-full bg-slate-900/60 transition-[width] duration-300 ease-smooth dark:bg-white/60"
            style={{ width: `${pct}%` }}
          />
        </div>

        <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div className="flex flex-wrap items-center gap-2">
            <ControlButton onClick={prev} disabled={controlsDisabled} label="Prev" icon="⏮" />
            {player?.isPaused || !player?.isPlaying ? (
              <ControlButton onClick={play} disabled={controlsDisabled} label="Play" icon="▶" primary />
            ) : (
              <ControlButton onClick={pause} disabled={controlsDisabled} label="Pause" icon="⏸" primary />
            )}
            <ControlButton onClick={next} disabled={controlsDisabled} label="Next" icon="⏭" />
          </div>

          <div className="flex items-center gap-3">
            <div className="text-xs text-slate-500 dark:text-slate-400">Volume</div>
            <input
              type="range"
              min={0}
              max={100}
              value={Math.round(localVolume * 100)}
              onChange={(e) => setLocalVolume(Number(e.target.value) / 100)}
              onMouseUp={() => commitVolume(localVolume)}
              onTouchEnd={() => commitVolume(localVolume)}
              className="w-44 accent-slate-900 dark:accent-white"
            />
          </div>
        </div>

        {/* HTML5 audio stream (no video) */}
        <audio
          key={streamKey}
          ref={audioRef}
          src={`${config.streamUrl}?v=${streamKey}`}
          preload="none"
          onError={handleAudioError}
          onCanPlay={() => {
            // try play if server says playing
            if (player?.isPlaying && !player?.isPaused) {
              audioRef.current?.play().catch(() => void 0);
            }
          }}
        />
      </div>
    </section>
  );
}

function ControlButton({
  onClick,
  disabled,
  label,
  icon,
  primary
}: {
  onClick: () => void;
  disabled?: boolean;
  label: string;
  icon: string;
  primary?: boolean;
}) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={clsx(
        "inline-flex items-center gap-2 rounded-2xl px-4 py-2 text-sm font-semibold shadow-sm transition",
        primary
          ? "bg-slate-900 text-white hover:opacity-95 dark:bg-white dark:text-slate-900"
          : "border border-slate-200 bg-white hover:bg-slate-50 dark:border-white/10 dark:bg-white/5 dark:hover:bg-white/10",
        disabled && "opacity-50"
      )}
      aria-label={label}
      title={label}
    >
      <span className="text-base">{icon}</span>
      <span className="hidden sm:inline">{label}</span>
    </button>
  );
}

function Badge({ label, subtle }: { label: string; subtle?: boolean }) {
  return (
    <span
      className={clsx(
        "rounded-full px-3 py-1 text-[11px] font-semibold",
        subtle
          ? "bg-slate-100 text-slate-700 dark:bg-white/10 dark:text-slate-200"
          : "bg-emerald-500/15 text-emerald-800 dark:text-emerald-200"
      )}
    >
      {label}
    </span>
  );
}

function clamp(n: number, a: number, b: number) {
  return Math.max(a, Math.min(b, n));
}

function errMsg(e: unknown) {
  return e instanceof Error ? e.message : "Ошибка";
}
