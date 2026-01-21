/**
 * Purpose: Minimal API client with JWT support.
 * Security: We keep token in memory (Zustand store) by default.
 */

import { config } from "./config";
import { useAppStore } from "./store/useAppStore";

export type LoginRequest = { username: string; password: string };
export type LoginResponse = { token: string };

export type PlayerState = {
  isPlaying: boolean;
  isPaused: boolean; // server-authoritative pause
  volume: number; // 0..1
  current?: {
    id: string;
    title: string;
    url: string;
    addedByNick?: string;
  };
  positionSec: number;
  durationSec: number;
};

export type QueueEntry = {
  id: string;
  title: string;
  url: string;
  addedByNick?: string;
  addedAt: string;
  isDonation?: boolean;
  status: "prev" | "current" | "next";
};

type HttpMethod = "GET" | "POST";

async function request<T>(path: string, method: HttpMethod, body?: unknown): Promise<T> {
  const token = useAppStore.getState().auth.token;
  const res = await fetch(`${config.apiBaseUrl}${path}`, {
    method,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    },
    body: body ? JSON.stringify(body) : undefined,
    credentials: "omit"
  });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`HTTP ${res.status}: ${text || res.statusText}`);
  }

  if (res.status === 204 || res.headers.get("content-length") === "0") {
    return undefined as T;
  }

  const text = await res.text();
  if (!text) {
    return undefined as T;
  }

  return JSON.parse(text) as T;
}

export const api = {
  auth: {
    login: (payload: LoginRequest) => request<LoginResponse>("/api/auth/login", "POST", payload)
  },
  player: {
    state: () => request<PlayerState>("/api/player/state", "GET"),
    play: () => request<{ ok: true }>("/api/player/play", "POST"),
    pause: () => request<{ ok: true }>("/api/player/pause", "POST"),
    next: () => request<{ ok: true }>("/api/player/next", "POST"),
    prev: () => request<{ ok: true }>("/api/player/prev", "POST"),
    volume: (v: number) => request<{ ok: true }>("/api/player/volume", "POST", { volume: v })
  },
  playlist: {
    list: () => request<QueueEntry[]>("/api/playlist", "GET"),
    add: (url: string) => request<{ ok: true }>("/api/playlist/add", "POST", { url })
  },
  integrations: {
    donationalertsConnect: (payload: unknown) =>
      request<{ ok: true }>("/api/integrations/donationalerts/connect", "POST", payload),
    donxConnect: (payload: unknown) =>
      request<{ ok: true }>("/api/integrations/donx/connect", "POST", payload)
  }
};
