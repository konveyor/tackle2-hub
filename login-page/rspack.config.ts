import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { rspack } from "@rspack/core";
import type { Configuration } from "@rspack/core";
import { TsCheckerRspackPlugin } from "ts-checker-rspack-plugin";

const __dirname = fileURLToPath(new URL(".", import.meta.url));
const pathTo = (...parts: string[]) => path.resolve(__dirname, ...parts);

// Branding: default to ./branding, overridden by BRANDING env var at build time.
// This allows container image builds to swap in custom branding without changing source.
const baseBrandingPath = process.env.BRANDING ?? "./branding";
const brandingPath = pathTo(baseBrandingPath);
const brandingStrings = JSON.parse(
  readFileSync(path.resolve(brandingPath, "strings.json"), "utf8")
);
console.log("Using branding from:", brandingPath);

const isDev = process.env.NODE_ENV === "development";

// Build output goes into the Go embed package so that `go:embed dist` in
// internal/loginpage/embed.go picks it up directly. The frontend source
// intentionally lives in login-page/ (project root) while the compiled
// artifacts are placed where the Go build needs them.
const distPath = pathTo("..", "internal", "loginpage", "dist");

const config: Configuration = {
  mode: isDev ? "development" : "production",
  entry: {
    main: "./src/index.tsx",
  },
  output: {
    path: distPath,
    // /oidc/assets/ matches the static file handler registered in internal/api/auth.go.
    // In dev mode, assets are served by the rspack dev server at its own port.
    publicPath: isDev ? "/" : "/oidc/assets/",
    filename: isDev ? "[name].js" : "[name].[contenthash:8].js",
    cssFilename: isDev ? "[name].css" : "[name].[contenthash:8].css",
    clean: true,
  },
  resolve: {
    extensions: [".ts", ".tsx", ".js", ".jsx"],
  },
  module: {
    rules: [
      {
        test: /\.(?:js|mjs|jsx|ts|tsx)$/,
        exclude: /node_modules/,
        use: {
          loader: "builtin:swc-loader",
          options: {
            detectSyntax: "auto",
            jsc: {
              parser: {
                syntax: "typescript",
              },
            }
          },
        },
      },
      {
        test: /\.css$/,
        type: "css/auto",
      },
      {
        test: /\.(png|jpg|jpeg|gif|svg|woff|woff2|eot|ttf|otf)$/,
        type: "asset/resource",
      },
    ],
  },
  plugins: [
    new TsCheckerRspackPlugin(),
    new rspack.HtmlRspackPlugin({
      template: "./src/index.html",
      filename: "index.html",
      // TODO Add support for branding.styles.favicon
      // TODO Add support for branding.styles.themeCss
    }),
    new rspack.CssExtractRspackPlugin({
      filename: isDev ? "[name].css" : "[name].[contenthash:8].css",
    }),
    // Bake branding strings into the bundle at build time as a global constant.
    // Components import brandingStrings from branding.ts which reads __BRANDING_STRINGS__.
    new rspack.DefinePlugin({
      __BRANDING_STRINGS__: JSON.stringify(brandingStrings),
    }),
    // Copy image/asset files from the branding directory into dist/branding/.
    new rspack.CopyRspackPlugin({
      patterns: [
        {
          from: brandingPath,
          to: path.join(distPath, "branding"),
          noErrorOnMissing: true,
          globOptions: {
            ignore: ["**/strings.json"],
          },
        },
      ],
    }),
  ],
  optimization: {
    minimize: !isDev,
  },
  devServer: isDev
    ? {
        port: 3001,
        open: true,
      }
    : undefined,
};

export default config;
