/**
 * Purpose: Modal for adding a playlist link (owner flow).
 */

import React from "react";
import { useAppStore } from "../store/useAppStore";
import { api } from "../api";

export function AddPlaylistModal({
  onToast
}: {
  onToast: (t: { title: string; message?: string; kind?: "info" | "success" | "error" }) => void;
}) {
  const open = useAppStore((s) => s.ui.addPlaylistOpen);
  const close = useAppStore((s) => s.closeAddPlaylist);

  const [url, setUrl] = React.useState("");
  const [loading, setLoading] = React.useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      await api.playlist.add(url);
      onToast({ title: "Плейлист", message: "Добавлено в очередь", kind: "success" });
      close();
      setUrl("");
    } catch (e) {
      onToast({
        title: "Плейлист",
        message: e instanceof Error ? e.message : "Ошибка",
        kind: "error"
      });
    } finally {
      setLoading(false);
    }
  }

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-40 grid place-items-center bg-black/40 p-6 backdrop-blur-sm">
      <div className="w-full max-w-lg rounded-3xl border border-slate-200 bg-white p-6 shadow-soft dark:border-white/10 dark:bg-slate-950">
        <div className="flex items-start justify-between">
          <div>
            <div className="text-lg font-semibold tracking-tight">Добавить плейлист</div>
            <div className="mt-1 text-xs text-slate-500 dark:text-slate-400">
              Вставьте ссылку на плейлист YouTube (или трек). Бэкенд разберёт и добавит.
            </div>
          </div>
          <button
            onClick={close}
            className="rounded-xl px-2 py-1 text-xs text-slate-500 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-white/10"
          >
            ✕
          </button>
        </div>

        <form onSubmit={submit} className="mt-5 space-y-3">
          <input
            className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm outline-none transition focus:ring-2 focus:ring-slate-300 dark:border-white/10 dark:bg-white/5 dark:focus:ring-white/20"
            placeholder="https://www.youtube.com/playlist?list=..."
            value={url}
            onChange={(e) => setUrl(e.target.value)}
          />

          <button
            type="submit"
            disabled={loading || !url.trim()}
            className="w-full rounded-2xl bg-slate-900 px-4 py-3 text-sm font-semibold text-white shadow-soft transition hover:opacity-95 disabled:opacity-60 dark:bg-white dark:text-slate-900"
          >
            {loading ? "Отправка..." : "Добавить"}
          </button>
        </form>
      </div>
    </div>
  );
}
