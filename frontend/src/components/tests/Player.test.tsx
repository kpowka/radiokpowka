import React from "react";
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { Player } from "../components/Player";
import { useAppStore } from "../store/useAppStore";

describe("Player", () => {
  it("shows placeholder title when no track", () => {
    // prepare store
    useAppStore.setState({
      ...useAppStore.getState(),
      role: "listener",
      player: {
        isPlaying: false,
        isPaused: true,
        volume: 0.8,
        positionSec: 0,
        durationSec: 0
      }
    });

    render(<Player onToast={() => {}} />);

    expect(screen.getByText("Сейчас играет")).toBeTruthy();
    expect(screen.getByText("—")).toBeTruthy();
  });
});
