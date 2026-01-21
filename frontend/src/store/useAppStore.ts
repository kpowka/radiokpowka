/**
 * Purpose: Global state for auth/player/queue + UI flags.
 * Security: token is kept in memory by default (no localStorage).
 */

import { create } from "zustand";
import type { PlayerState, QueueEntry } from "../api";

type Role = "owner" | "listener" | "unknown";
type WsStatus = "connected" | "disconnected";

type DonationPreview = {
  donorNick: string;
  trackUrl?: string;
  message: string;
} | null;

type AppState = {
  auth: {
    token: string | null;
  };
  role: Role;

  wsStatus: WsStatus;

  player: PlayerState | null;
  queue: QueueEntry[];

  ui: {
    loginOpen: boolean;
    burgerOpen: boolean;
    addPlaylistOpen: boolean;
    donationPreview: DonationPreview;
    theme: "dark" | "light";
  };

  // actions
  setToken: (token: string | null) => void;
  setRole: (role: Role) => void;

  setWsStatus: (s: WsStatus) => void;

  setPlayerState: (s: PlayerState) => void;
  setQueue: (q: QueueEntry[]) => void;
  pushQueue: (e: QueueEntry) => void;

  openLogin: () => void;
  closeLogin: () => void;

  toggleBurger: () => void;
  closeBurger: () => void;

  openAddPlaylist: () => void;
  closeAddPlaylist: () => void;

  showDonationPreview: (p: NonNullable<DonationPreview>) => void;
  clearDonationPreview: () => void;

  setTheme: (t: "dark" | "light") => void;
};

function getInitialTheme(): "dark" | "light" {
  // по умолчанию — dark, но можно переключать в UI
  return "dark";
}

export const useAppStore = create<AppState>((set, get) => ({
  auth: { token: null },
  role: "unknown",
  wsStatus: "disconnected",

  player: null,
  queue: [],

  ui: {
    loginOpen: true,
    burgerOpen: false,
    addPlaylistOpen: false,
    donationPreview: null,
    theme: getInitialTheme()
  },

  setToken: (token) => set({ auth: { token } }),
  setRole: (role) => set({ role }),

  setWsStatus: (s) => set({ wsStatus: s }),

  setPlayerState: (s) => set({ player: s }),
  setQueue: (q) => set({ queue: q }),
  pushQueue: (e) => set({ queue: [...get().queue, e] }),

  openLogin: () => set({ ui: { ...get().ui, loginOpen: true } }),
  closeLogin: () => set({ ui: { ...get().ui, loginOpen: false } }),

  toggleBurger: () => set({ ui: { ...get().ui, burgerOpen: !get().ui.burgerOpen } }),
  closeBurger: () => set({ ui: { ...get().ui, burgerOpen: false } }),

  openAddPlaylist: () => set({ ui: { ...get().ui, addPlaylistOpen: true } }),
  closeAddPlaylist: () => set({ ui: { ...get().ui, addPlaylistOpen: false } }),

  showDonationPreview: (p) => set({ ui: { ...get().ui, donationPreview: p } }),
  clearDonationPreview: () => set({ ui: { ...get().ui, donationPreview: null } }),

  setTheme: (t) => set({ ui: { ...get().ui, theme: t } })
}));
