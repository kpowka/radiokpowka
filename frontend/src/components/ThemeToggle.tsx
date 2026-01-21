/**
 * Purpose: Dark/Light theme toggle with small animation.
 */

import React from "react";
import { useAppStore } from "../store/useAppStore";
import { clsx } from "clsx";

export function ThemeToggle() {
  const theme = useAppStore((s) => s.ui.theme);
  const setTheme = useAppStore((s) => s.setTheme);

  const isDark = theme === "dark";

  return (
    <button
      onClick={() => setTheme(isDark ? "light" : "dark")}
      className={clsx(
        "group relative inline-flex items-center gap-2 rounded-2xl border px-3 py-2 text-sm shadow-sm transition",
        "border-slate-200 bg-white hover:shadow dark:border-white/10 dark:bg-white/5"
      )}
      aria-label="Toggle theme"
      title="Theme"
    >
      <span className="text-base transition group-hover:scale-105">{isDark ? "🌙" : "☀️"}</span>
      <span className="text-xs text-slate-600 dark:text-slate-300">{isDark ? "Dark" : "Light"}</span>
      <span
        className={clsx(
          "absolute -bottom-1 left-1/2 h-[2px] w-0 -translate-x-1/2 rounded-full bg-slate-900/20 transition-all dark:bg-white/20",
          "group-hover:w-3/4"
        )}
      />
    </button>
  );
}
