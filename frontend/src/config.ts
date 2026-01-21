/**
 * Purpose: Centralized runtime configuration for API + WebSocket.
 * Values are taken from Vite env at build time.
 */

export const config = {
  apiBaseUrl: (import.meta.env.VITE_API_BASE_URL as string) || "http://localhost:8080",
  wsUrl: (import.meta.env.VITE_WS_URL as string) || "ws://localhost:8080/ws",
  streamUrl: (import.meta.env.VITE_STREAM_URL as string) || "http://localhost:8080/stream"
};
