/**
 * Purpose: Lightweight toast notifications (no external libs).
 */

import React from "react";
import { clsx } from "clsx";

export type Toast = {
  id: string;
  title: string;
  message?: string;
  kind?: "info" | "success" | "error";
};

export function ToastHost({
  toasts,
  onDismiss
}: {
  toasts: Toast[];
  onDismiss: (id: string) => void;
}) {
  return (
    <div className="pointer-events-none fixed right-4 top-4 z-50 flex w-[360px] max-w-[calc(100vw-2rem)] flex-col gap-2">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={clsx(
            "pointer-events-auto rounded-2xl border p-3 shadow-soft backdrop-blur transition",
            "bg-white/90 dark:bg-slate-900/70",
            "border-slate-200 dark:border-white/10",
            t.kind === "success" && "ring-1 ring-emerald-500/30",
            t.kind === "error" && "ring-1 ring-rose-500/30"
          )}
        >
          <div className="flex items-start justify-between gap-3">
            <div>
              <div className="text-sm font-semibold">{t.title}</div>
              {t.message ? (
                <div className="mt-1 text-xs text-slate-600 dark:text-slate-300">{t.message}</div>
              ) : null}
            </div>
            <button
              className="rounded-xl px-2 py-1 text-xs text-slate-500 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-white/10"
              onClick={() => onDismiss(t.id)}
              aria-label="Dismiss toast"
            >
              ✕
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}
