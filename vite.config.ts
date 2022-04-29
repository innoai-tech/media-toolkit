import { defineConfig } from "vite";
import viteReact from "@vitejs/plugin-react";
import { injectWebAppConfig } from "@innoai-tech/config/vite-plugin-inject-config";
import nodeResolve from "@rollup/plugin-node-resolve";
import { join } from "path";

export default defineConfig({
  root: "./webapp/live-player",
  build: {
    outDir: "dist",
    emptyOutDir: true
  },
  plugins: [
    injectWebAppConfig() as any,
    nodeResolve({
      mainFields: ["browser", "jsnext:main", "module", "main"],
      moduleDirectories: [
        process.cwd(), // project root for mono repo
        join(process.cwd(), "node_modules"), // root node_modules first
        "node_modules" // then related node_modules
      ],
      extensions: [".ts", ".tsx", ".mjs", ".js", ".jsx"]
    }),
    viteReact({
      jsxRuntime: "automatic",
      jsxImportSource: "@emotion/react",
      babel: {
        plugins: ["@emotion/babel-plugin"]
      },
      jsxPure: true
    })
  ]
});
