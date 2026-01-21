/**
 * Purpose: Login modal shown on first load.
 * Flow: calls POST /api/auth/login, stores JWT in memory (Zustand).
 */

import React from "react";
import { api } from "../api";
import { useAppStore } from "../store/useAppStore";

export function LoginModal({
  onLoggedIn
}: {
  onLoggedIn?: () => void;
}) {
  const open = useAppStore((s) => s.ui.loginOpen);
  const closeLogin = useAppStore((s) => s.closeLogin);
  const setToken = useAppStore((s) => s.setToken);

  const [username, setUsername] = React.useState("admin");
  const [password, setPassword] = React.useState("admin123");
  const [loading, setLoading] = React.useState(false);
  const [err, setErr] = React.useState<string | null>(null);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr(null);
    setLoading(true);
    try {
      const res = await api.auth.login({ username, password });
      setToken(res.token);
      closeLogin();
      onLoggedIn?.();
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-40 grid place-items-center bg-black/40 p-6 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-3xl border border-slate-200 bg-white p-6 shadow-soft dark:border-white/10 dark:bg-slate-950">
        <div className="flex items-start justify-between gap-3">
          <div>
            <div className="text-lg font-semibold tracking-tight">Вход</div>
            <div className="mt-1 text-xs text-slate-500 dark:text-slate-400">
              Введите логин/пароль. Токен хранится в памяти (без localStorage).
            </div>
          </div>
        </div>

        <form onSubmit={submit} className="mt-5 space-y-3">
          <label className="block">
            <div className="text-xs font-medium text-slate-600 dark:text-slate-300">Username</div>
            <input
              className="mt-1 w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm outline-none transition focus:ring-2 focus:ring-slate-300 dark:border-white/10 dark:bg-white/5 dark:focus:ring-white/20"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              autoComplete="username"
            />
          </label>

          <label className="block">
            <div className="text-xs font-medium text-slate-600 dark:text-slate-300">Password</div>
            <input
              type="password"
              className="mt-1 w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm outline-none transition focus:ring-2 focus:ring-slate-300 dark:border-white/10 dark:bg-white/5 dark:focus:ring-white/20"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="current-password"
            />
          </label>

          {err ? (
            <div className="rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-xs text-rose-700 dark:text-rose-200">
              {err}
            </div>
          ) : null}

          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-2xl bg-slate-900 px-4 py-3 text-sm font-semibold text-white shadow-soft transition hover:opacity-95 disabled:opacity-60 dark:bg-white dark:text-slate-900"
          >
            {loading ? "Входим..." : "Войти"}
          </button>
        </form>

        <div className="mt-4 text-center text-[11px] text-slate-500 dark:text-slate-400">
          По умолчанию: <b>admin</b> / <b>admin123</b> (будет сид в бэке).
        </div>
      </div>
    </div>
  );
}
