import React from "react";
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { TrackList } from "../components/TrackList";
import type { QueueEntry } from "../api";

describe("TrackList", () => {
  it("renders current track and donation badge", () => {
    const q: QueueEntry[] = [
      {
        id: "1",
        title: "Prev Song",
        url: "https://x",
        addedByNick: "nick",
        addedAt: new Date().toISOString(),
        status: "prev"
      },
      {
        id: "2",
        title: "Current Song",
        url: "https://y",
        addedByNick: "donor",
        addedAt: new Date().toISOString(),
        status: "current",
        isDonation: true
      },
      {
        id: "3",
        title: "Next Song",
        url: "https://z",
        addedByNick: "nick2",
        addedAt: new Date().toISOString(),
        status: "next"
      }
    ];

    render(<TrackList queue={q} />);

    expect(screen.getByText("Очередь")).toBeTruthy();
    expect(screen.getByText("Current Song")).toBeTruthy();
    expect(screen.getByText("DONATION")).toBeTruthy();
  });
});
