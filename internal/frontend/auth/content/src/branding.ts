// Branding strings are baked into the bundle at build time from branding/strings.json
// via rspack's DefinePlugin in rspack.config.ts. The BRANDING env var selects which
// branding directory to read at build time, allowing container image builds to swap
// branding without code changes.
//
// __BRANDING_STRINGS__ is replaced by a literal JSON object during compilation.

export interface LoginBrandingStrings {
  application: {
    title: string;
    name?: string;
    description?: string;
  };
  loginPage: {
    title: string;
    subtitle?: string;
    textContent?: string;
  };
  devicePage: {
    title: string;
    subtitle?: string;
    textContent?: string;
    successTitle: string;
    successMessage: string;
  };
  styles: {
    brandImage?: string;
    backgroundImage?: string;
    favicon?: string; // TODO: Add favicon support (in rspack.config.ts)
    themeCss?: string; // TODO: Add theme CSS support (in rspack.config.ts or index.tsx)
  };
}

declare const __BRANDING_STRINGS__: LoginBrandingStrings;

export const brandingStrings: LoginBrandingStrings = __BRANDING_STRINGS__;
