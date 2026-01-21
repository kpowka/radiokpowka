import { describe, it, expect } from "vitest";
import { useAppStore } from "../store/useAppStore";

// This test is a “light” integration: we simulate WS events via store setters.
// Real WS (WebSocket) E2E is usually done with playwright/cypress, but we keep it minimal.

describe("ws integration (store)", () => {
  it("applies player_state and queue_update into store", () => {
    const now = new Date().toISOString();

    useAppStore.getState().setPlayerState({
      isPlaying: true,
      isPaused: false,
      volume: 0.7,
      positionSec: 10,
      durationSec: 100,
      current: { id: "t1", title: "Song", url: "https://y", addedByNick: "nick" }
    });

    useAppStore.getState().setQueue([
      {
        id: "q1",
        title: "Song",
        url: "https://y",
        addedByNick: "nick",
        addedAt: now,
        status: "current"
      }
    ]);

    const s = useAppStore.getState();
    expect(s.player?.isPlaying).toBe(true);
    expect(s.queue.length).toBe(1);
    expect(s.queue[0]?.title).toBe("Song");
  });
});
