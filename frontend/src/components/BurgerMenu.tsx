/**
 * Purpose: Left side burger settings menu with slide animation.
 * Includes placeholders for integration connect actions.
 */

import React from "react";
import { useAppStore } from "../store/useAppStore";
import { clsx } from "clsx";
import { api } from "../api";

export function BurgerMenu({
  onToast
}: {
  onToast: (t: { title: string; message?: string; kind?: "info" | "success" | "error" }) => void;
}) {
  const burgerOpen = useAppStore((s) => s.ui.burgerOpen);
  const toggleBurger = useAppStore((s) => s.toggleBurger);
  const closeBurger = useAppStore((s) => s.closeBurger);
  const openAddPlaylist = useAppStore((s) => s.openAddPlaylist);

  async function connectDonAlerts() {
    try {
      // placeholder payload (backend will validate)
      await api.integrations.donationalertsConnect({ mode: "placeholder" });
      onToast({ title: "DonAlerts", message: "Подключение отправлено", kind: "success" });
    } catch (e) {
      onToast({
        title: "DonAlerts",
        message: e instanceof Error ? e.message : "Ошибка",
        kind: "error"
      });
    }
  }

  async function connectDonX() {
    try {
      await api.integrations.donxConnect({ mode: "placeholder" });
      onToast({ title: "DonX", message: "Подключение отправлено", kind: "success" });
    } catch (e) {
      onToast({
        title: "DonX",
        message: e instanceof Error ? e.message : "Ошибка",
        kind: "error"
      });
    }
  }

  function connectTwitchBot() {
    // UI placeholder: real connect flow depends on backend OAuth / token
    onToast({
      title: "Twitch Bot",
      message: "Здесь будет подключение бота (OAuth/token).",
      kind: "info"
    });
  }

  return (
    <>
      <button
        onClick={toggleBurger}
        className={clsx(
          "inline-flex items-center gap-2 rounded-2xl border px-3 py-2 text-sm shadow-sm transition hover:shadow",
          "border-slate-200 bg-white dark:border-white/10 dark:bg-white/5"
        )}
        aria-label="Open menu"
      >
        <span className="text-base">☰</span>
        <span className="hidden sm:inline">Settings</span>
      </button>

      {/* Backdrop */}
      <div
        className={clsx(
          "fixed inset-0 z-30 bg-black/30 backdrop-blur-sm transition-opacity",
          burgerOpen ? "opacity-100" : "pointer-events-none opacity-0"
        )}
        onClick={closeBurger}
      />

      {/* Drawer */}
      <aside
        className={clsx(
          "fixed left-4 top-4 z-40 h-[calc(100vh-2rem)] w-[320px] max-w-[calc(100vw-2rem)]",
          "rounded-3xl border border-slate-200 bg-white shadow-soft dark:border-white/10 dark:bg-slate-950",
          "transition-transform duration-300 ease-smooth",
          burgerOpen ? "translate-x-0" : "-translate-x-[120%]"
        )}
        role="dialog"
        aria-label="Settings drawer"
      >
        <div className="flex h-full flex-col">
          <div className="flex items-center justify-between px-5 py-4">
            <div className="text-sm font-semibold">Настройки</div>
            <button
              onClick={closeBurger}
              className="rounded-xl px-2 py-1 text-xs text-slate-500 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-white/10"
            >
              ✕
            </button>
          </div>

          <div className="px-5">
            <div className="text-xs text-slate-500 dark:text-slate-400">
              Подключения и управление плейлистом
            </div>
          </div>

          <div className="mt-4 flex flex-col gap-2 px-5">
            <button
              onClick={connectTwitchBot}
              className="rounded-2xl border border-slate-200 px-4 py-3 text-left text-sm transition hover:bg-slate-50 dark:border-white/10 dark:hover:bg-white/5"
            >
              <div className="font-semibold">Connect Twitch Bot</div>
              <div className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                Команды !track и !track spam N
              </div>
            </button>

            <button
              onClick={connectDonAlerts}
              className="rounded-2xl border border-slate-200 px-4 py-3 text-left text-sm transition hover:bg-slate-50 dark:border-white/10 dark:hover:bg-white/5"
            >
              <div className="font-semibold">Connect DonAlerts</div>
              <div className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                Webhook / API коннектор (placeholder)
              </div>
            </button>

            <button
              onClick={connectDonX}
              className="rounded-2xl border border-slate-200 px-4 py-3 text-left text-sm transition hover:bg-slate-50 dark:border-white/10 dark:hover:bg-white/5"
            >
              <div className="font-semibold">Connect DonX</div>
              <div className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                Webhook / API коннектор (placeholder)
              </div>
            </button>

            <button
              onClick={() => {
                openAddPlaylist();
                closeBurger();
              }}
              className="rounded-2xl bg-slate-900 px-4 py-3 text-left text-sm font-semibold text-white transition hover:opacity-95 dark:bg-white dark:text-slate-900"
            >
              Add Playlist Link
              <div className="mt-1 text-xs font-normal opacity-80">
                Добавить плейлист (owner)
              </div>
            </button>
          </div>

          <div className="mt-auto px-5 py-4 text-[11px] text-slate-500 dark:text-slate-400">
            UI готов. Реальные интеграции будут в бэкенде (части 3–5).
          </div>
        </div>
      </aside>
    </>
  );
}
