// LoginConfig is injected by the hub as window.__LOGIN_CONFIG__ before serving the
// page. It carries only per-request dynamic values; branding (titles, logos,
// background images) is baked into the bundle at build time.
export interface LoginConfig {
  // Selects which page component to render.
  page: "login" | "device-verify" | "device-success" | "session-expired";

  // Login page: POST target URL including authRequestId query param.
  formAction?: string;

  // Login page: error message to display on failed authentication.
  errorMessage?: string;

  // TODO Add support for >1 federated IdP button
  // Login page: federated IdP button shown when an external IdP is enabled
  // but not configured as primary (so the local form is also shown).
  federatedIdp?: {
    name: string;
    loginUrl: string;
  };

  // Device verify page: POST target URL for user-code form submission.
  deviceFormAction?: string;
}

declare global {
  interface Window {
    __LOGIN_CONFIG__?: LoginConfig;
  }
}
