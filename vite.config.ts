import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import nodeResolve from "@rollup/plugin-node-resolve";
import { join } from "path";

export default defineConfig({
  root: "./webapp/live-player",
  build: {
    outDir: "../dist",
    emptyOutDir: true
  },
  plugins: [
    nodeResolve({
      mainFields: ["browser", "jsnext:main", "module", "main"],
      moduleDirectories: [
        process.cwd(), // project root for mono repo
        join(process.cwd(), "node_modules"), // root node_modules first
        "node_modules" // then related node_modules
      ],
      extensions: [".ts", ".tsx", ".mjs", ".js", ".jsx"]
    }),
    react({
      jsxRuntime: "automatic",
      jsxImportSource: "@emotion/react",
      babel: {
        plugins: ["@emotion/babel-plugin"]
      },
      jsxPure: true
    })
  ]
});
