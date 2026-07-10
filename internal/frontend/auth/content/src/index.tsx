import React from "react";
import { createRoot } from "react-dom/client";

import { UserLoginPage } from "./UserLoginPage";
import { DeviceVerifyPage } from "./DeviceVerifyPage";
import { DeviceSuccessPage } from "./DeviceSuccessPage";
import { SessionExpiredPage } from "./SessionExpiredPage";
import type { LoginConfig } from "./types";

import "@patternfly/react-core/dist/styles/base.css";
// TODO Add support for branding.styles.themeCss (or in rspack.config.ts)

// The hub injects window.__LOGIN_CONFIG__ into the served HTML before the
// bundle script runs. It carries all per-request dynamic values (page type,
// form actions, error messages, federated IdP config).
function getConfig(): LoginConfig {
  if (window.__LOGIN_CONFIG__) {
    return window.__LOGIN_CONFIG__;
  }
  // Fallback for local development without a running hub.
  console.warn("window.__LOGIN_CONFIG__ not set; using development defaults.");
  return { page: "login", formAction: "/oidc/login" };
}

const App = () => {
  const config = getConfig();

  switch (config.page) {
    case "login":
      return <UserLoginPage config={config} />;
    case "device-verify":
      return <DeviceVerifyPage config={config} />;
    case "device-success":
      return <DeviceSuccessPage config={config} />;
    case "session-expired":
      return <SessionExpiredPage config={config} />;
    default:
      return <UserLoginPage config={config} />;
  }
}

const container = document.getElementById("root");
if (!container) {
  throw new Error("Root element #root not found in HTML template.");
}

createRoot(container).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
