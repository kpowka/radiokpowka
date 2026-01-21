/**
 * Purpose: Main UI layout wiring: LoginModal -> Player + TrackList + BurgerMenu + modals.
 */

import React from "react";
import { useAppStore } from "./store/useAppStore";
import { api, type QueueEntry } from "./api";
import { ThemeToggle } from "./components/ThemeToggle";
import { ToastHost, type Toast } from "./components/ToastHost";
import { LoginModal } from "./components/LoginModal";
import { BurgerMenu } from "./components/BurgerMenu";
import { Player } from "./components/Player";
import { TrackList } from "./components/TrackList";
import { AddPlaylistModal } from "./components/AddPlaylistModal";
import { DonationOverlay } from "./components/DonationOverlay";

function uid() {
  return Math.random().toString(16).slice(2);
}

export default function App() {
  const token = useAppStore((s) => s.auth.token);
  const openLogin = useAppStore((s) => s.openLogin);
  const closeLogin = useAppStore((s) => s.closeLogin);
  const queue = useAppStore((s) => s.queue);
  const setQueue = useAppStore((s) => s.setQueue);

  const [toasts, setToasts] = React.useState<Toast[]>([]);

  function toast(t: { title: string; message?: string; kind?: "info" | "success" | "error" }) {
    const id = uid();
    setToasts((x) => [...x, { id, ...t }]);
    window.setTimeout(() => {
      setToasts((x) => x.filter((q) => q.id !== id));
    }, 4500);
  }

  React.useEffect(() => {
    // если нет токена — показываем логин
    if (!token) openLogin();
  }, [token, openLogin]);

  async function refreshQueue() {
    try {
      const q = await api.playlist.list();
      setQueue(q);
    } catch {
      // тихо: WS обычно обновит
    }
  }

  React.useEffect(() => {
    // при логине подтянуть очередь (WS тоже обновит)
    if (token) void refreshQueue();
  }, [token]);

  return (
    <div className="min-h-screen bg-white text-slate-900 dark:bg-slate-950 dark:text-slate-50">
      <ToastHost
        toasts={toasts}
        onDismiss={(id) => setToasts((x) => x.filter((t) => t.id !== id))}
      />

      <header className="mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-6">
        <div className="flex items-center gap-3">
          <div className="grid h-10 w-10 place-items-center rounded-2xl bg-slate-900/10 text-sm font-bold dark:bg-white/10">
            RK
          </div>
          <div>
            <div className="text-lg font-semibold tracking-tight">RadioKpowka</div>
            <div className="text-xs text-slate-500 dark:text-slate-400">
              Серверный стрим аудио • Реалтайм WebSocket
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <BurgerMenu onToast={toast} />
          <ThemeToggle />
          <button
            className="rounded-2xl border border-slate-200 px-3 py-2 text-sm shadow-sm transition hover:shadow dark:border-white/10"
            onClick={() => {
              closeLogin();
              useAppStore.getState().setToken(null);
              toast({ title: "Сессия", message: "Вы вышли", kind: "info" });
            }}
            title="Logout"
          >
            Logout
          </button>
        </div>
      </header>

      <main className="mx-auto w-full max-w-6xl px-6 pb-10">
        <div className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
          <Player onToast={toast} />
          <TrackList queue={queue as QueueEntry[]} />
        </div>

        <div className="mt-6 rounded-3xl border border-slate-200 bg-white p-4 text-xs text-slate-600 shadow-soft dark:border-white/10 dark:bg-slate-900/40 dark:text-slate-300">
          Подсказка: owner-кнопки управления (play/pause/next/prev) будут доступны после того, как бэкенд
          начнёт отправлять роль через WS/event <b>auth_update</b>.
        </div>
      </main>

      <LoginModal
        onLoggedIn={() => {
          toast({ title: "Вход", message: "Успешно", kind: "success" });
        }}
      />
      <AddPlaylistModal onToast={toast} />
      <DonationOverlay />
    </div>
  );
}
