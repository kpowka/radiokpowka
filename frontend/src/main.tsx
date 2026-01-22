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
    if (theme === "dark") {
      el.classList.add("dark");
      el.style.colorScheme = "dark";
    } else {
      el.classList.remove("dark");
      el.style.colorScheme = "light";
    }
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
