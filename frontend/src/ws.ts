/**
 * Purpose: WebSocket client with auto-reconnect and event dispatching into Zustand store.
 * Realtime events: player_state, queue_update, track_added, donation_received, auth_update
 */

import { config } from "./config";
import { useAppStore } from "./store/useAppStore";
import type { PlayerState, QueueEntry } from "./api";

export type WsEvent =
  | { type: "player_state"; data: PlayerState }
  | { type: "queue_update"; data: QueueEntry[] }
  | { type: "track_added"; data: { entry: QueueEntry } }
  | { type: "donation_received"; data: { donorNick: string; trackUrl?: string; message: string } }
  | { type: "auth_update"; data: { role: "owner" | "listener" } };

type WsClientOptions = {
  reconnectMinMs?: number;
  reconnectMaxMs?: number;
};

export class WsClient {
  private ws: WebSocket | null = null;
  private stopped = false;
  private attempt = 0;
  private readonly reconnectMinMs: number;
  private readonly reconnectMaxMs: number;

  constructor(opts: WsClientOptions = {}) {
    this.reconnectMinMs = opts.reconnectMinMs ?? 600;
    this.reconnectMaxMs = opts.reconnectMaxMs ?? 6000;
  }

  start() {
    this.stopped = false;
    this.connect();
  }

  stop() {
    this.stopped = true;
    this.ws?.close();
    this.ws = null;
  }

  private connect() {
    if (this.stopped) return;

    const token = useAppStore.getState().auth.token;
    const url = new URL(config.wsUrl);
    if (token) url.searchParams.set("token", token);

    this.ws = new WebSocket(url.toString());

    this.ws.onopen = () => {
      this.attempt = 0;
      useAppStore.getState().setWsStatus("connected");
    };

    this.ws.onmessage = (evt) => {
      try {
        const msg = JSON.parse(String(evt.data)) as WsEvent;
        this.onEvent(msg);
      } catch {
        // ignore malformed messages
      }
    };

    this.ws.onclose = () => {
      useAppStore.getState().setWsStatus("disconnected");
      this.scheduleReconnect();
    };

    this.ws.onerror = () => {
      // onclose will handle reconnect
    };
  }

  private scheduleReconnect() {
    if (this.stopped) return;
    this.attempt += 1;
    const backoff = Math.min(
      this.reconnectMaxMs,
      this.reconnectMinMs * Math.pow(1.6, this.attempt)
    );
    const jitter = Math.floor(Math.random() * 250);
    setTimeout(() => this.connect(), backoff + jitter);
  }

  private onEvent(evt: WsEvent) {
    const store = useAppStore.getState();
    switch (evt.type) {
      case "player_state":
        store.setPlayerState(evt.data);
        break;
      case "queue_update":
        store.setQueue(evt.data);
        break;
      case "track_added":
        store.pushQueue(evt.data.entry);
        break;
      case "donation_received":
        store.showDonationPreview(evt.data);
        break;
      case "auth_update":
        store.setRole(evt.data.role);
        break;
      default:
        break;
    }
  }
}
