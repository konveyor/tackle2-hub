import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import Handlebars from "handlebars";
import { rspack } from "@rspack/core";
import type { Configuration } from "@rspack/core";
import { TsCheckerRspackPlugin } from "ts-checker-rspack-plugin";
import { LoginBrandingStrings } from "./src/branding";

const __dirname = fileURLToPath(new URL(".", import.meta.url));
const pathTo = (...parts: string[]) => path.resolve(__dirname, ...parts);

const brandingPath = pathTo("./branding");
console.log("Using branding from:", brandingPath);

const isDev = process.env.NODE_ENV === "development";

const distPath = pathTo("dist");

const publicPath = isDev ? "/" : "/hub/frontend/auth/";

const brandingRaw = readFileSync(path.resolve(brandingPath, "strings.json"), "utf8");
const brandingStrings = JSON.parse(
  Handlebars.compile(brandingRaw)({ publicPath })
) as LoginBrandingStrings;

const faviconPath = (JSON.parse(
  Handlebars.compile(brandingRaw)({ publicPath: `${brandingPath}/../` })
) as LoginBrandingStrings)?.styles?.favicon || undefined;

const config: Configuration = {
  mode: isDev ? "development" : "production",
  entry: {
    main: "./src/index.tsx",
  },
  output: {
    path: distPath,
    publicPath,
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
      filename: "index.html.tmpl",
      template: "./src/index.html",
      templateParameters: {
        title: brandingStrings.application.title,
        name: brandingStrings.application.name ?? brandingStrings.application.title,
        description: brandingStrings.application.description ?? "",
      },
      favicon: faviconPath,

      // Minification is disabled so that Go template actions ({{ . }})
      // in the source HTML pass through to the output unchanged.
      minify: false,
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
