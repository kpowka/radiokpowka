/**
 * Purpose: Optional overlay to preview incoming donation events (from WS).
 */

import React from "react";
import { useAppStore } from "../store/useAppStore";
import { clsx } from "clsx";

export function DonationOverlay() {
  const preview = useAppStore((s) => s.ui.donationPreview);
  const clear = useAppStore((s) => s.clearDonationPreview);

  React.useEffect(() => {
    if (!preview) return;
    const t = setTimeout(() => clear(), 6000);
    return () => clearTimeout(t);
  }, [preview, clear]);

  if (!preview) return null;

  return (
    <div className="pointer-events-none fixed bottom-4 left-1/2 z-40 w-[520px] max-w-[calc(100vw-2rem)] -translate-x-1/2">
      <div
        className={clsx(
          "pointer-events-auto rounded-3xl border p-4 shadow-soft backdrop-blur transition",
          "border-amber-500/30 bg-amber-500/10",
          "dark:border-amber-400/25 dark:bg-amber-400/10"
        )}
      >
        <div className="flex items-start justify-between gap-3">
          <div>
            <div className="text-sm font-semibold">Донат</div>
            <div className="mt-1 text-xs text-slate-700 dark:text-slate-200">
              <b>{preview.donorNick}</b>: {preview.message}
            </div>
            {preview.trackUrl ? (
              <div className="mt-1 text-[11px] text-slate-600 dark:text-slate-300">
                Трек из сообщения: <span className="underline">{preview.trackUrl}</span>
              </div>
            ) : null}
          </div>
          <button
            className="rounded-xl px-2 py-1 text-xs text-slate-700 hover:bg-white/40 dark:text-slate-200 dark:hover:bg-white/10"
            onClick={clear}
          >
            ✕
          </button>
        </div>
      </div>
    </div>
  );
}
