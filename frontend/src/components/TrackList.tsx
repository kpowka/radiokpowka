/**
 * Purpose: Right-side queue with previous/current/next sections.
 * Visual: current track highlighted; donation tracks marked with badge.
 */

import React from "react";
import type { QueueEntry } from "../api";
import { clsx } from "clsx";

function fmtTime(iso: string) {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString();
}

export function TrackList({
  queue,
  className
}: {
  queue: QueueEntry[];
  className?: string;
}) {
  const prev = queue.filter((q) => q.status === "prev");
  const current = queue.find((q) => q.status === "current") ?? null;
  const next = queue.filter((q) => q.status === "next");

  return (
    <section
      className={clsx(
        "rounded-3xl border border-slate-200 bg-white shadow-soft dark:border-white/10 dark:bg-slate-900/40",
        className
      )}
    >
      <div className="flex items-center justify-between px-5 py-4">
        <div className="text-sm font-semibold">Очередь</div>
        <div className="text-xs text-slate-500 dark:text-slate-400">
          {queue.length} трек(ов)
        </div>
      </div>

      <div className="max-h-[70vh] overflow-auto px-3 pb-4">
        <div className="px-2 py-2 text-[11px] font-semibold text-slate-500 dark:text-slate-400">
          Сейчас
        </div>

        {current ? (
          <Item entry={current} highlight />
        ) : (
          <div className="px-4 py-3 text-xs text-slate-500 dark:text-slate-400">
            Нет активного трека
          </div>
        )}

        <div className="px-2 py-2 text-[11px] font-semibold text-slate-500 dark:text-slate-400">
          Далее
        </div>

        {next.length ? (
          next.map((e) => <Item key={e.id} entry={e} />)
        ) : (
          <div className="px-4 py-3 text-xs text-slate-500 dark:text-slate-400">
            Очередь пуста
          </div>
        )}

        <div className="px-2 py-2 text-[11px] font-semibold text-slate-500 dark:text-slate-400">
          Было
        </div>

        {prev.length ? (
          prev.slice(-10).reverse().map((e) => <Item key={e.id} entry={e} muted />)
        ) : (
          <div className="px-4 py-3 text-xs text-slate-500 dark:text-slate-400">
            История пуста
          </div>
        )}

        <div className="px-4 pt-3 text-[11px] text-slate-500 dark:text-slate-400">
          Показывается до 10 последних в истории.
        </div>
      </div>
    </section>
  );

  function Item({
    entry,
    highlight,
    muted
  }: {
    entry: QueueEntry;
    highlight?: boolean;
    muted?: boolean;
  }) {
    return (
      <a
        href={entry.url}
        target="_blank"
        rel="noreferrer"
        className={clsx(
          "group mx-2 block rounded-2xl border p-3 transition",
          "border-slate-200 hover:bg-slate-50 dark:border-white/10 dark:hover:bg-white/5",
          highlight && "ring-1 ring-slate-900/15 dark:ring-white/15",
          muted && "opacity-70"
        )}
      >
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="truncate text-sm font-semibold">{entry.title}</div>
            <div className="mt-1 flex flex-wrap items-center gap-2 text-[11px] text-slate-500 dark:text-slate-400">
              <span className="truncate">
                Добавил: <b className="text-slate-700 dark:text-slate-200">{entry.addedByNick || "—"}</b>
              </span>
              <span>•</span>
              <span>{fmtTime(entry.addedAt)}</span>
              {entry.isDonation ? (
                <>
                  <span>•</span>
                  <span className="rounded-full bg-amber-500/15 px-2 py-0.5 text-[10px] font-semibold text-amber-700 dark:text-amber-200">
                    DONATION
                  </span>
                </>
              ) : null}
            </div>
          </div>
          <span className="text-xs text-slate-400 transition group-hover:text-slate-600 dark:group-hover:text-slate-200">
            ↗
          </span>
        </div>
      </a>
    );
  }
}
