/**
 * Purpose: App bootstrap + theme class + WS lifecycle.
 */

import React from "react";
import ReactDOM from "react-dom/client";
import "./styles.css";
import App from "./App";
import { useAppStore } from "./store/useAppStore";
import { WsClient } from "./ws";

const wsClient = new WsClient();

function Root() {
  const theme = useAppStore((s) => s.ui.theme);

  React.useEffect(() => {
    const el = document.documentElement;
    if (theme === "dark") el.classList.add("dark");
    else el.classList.remove("dark");
  }, [theme]);

  React.useEffect(() => {
    wsClient.start();
    return () => wsClient.stop();
  }, []);

  return <App />;
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <Root />
  </React.StrictMode>
);
