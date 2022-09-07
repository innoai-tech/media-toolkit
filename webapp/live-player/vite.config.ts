import { defineConfig } from "vite";
import { app, presetReact } from "@innoai-tech/vite-presets";
import { injectWebAppConfig } from "@innoai-tech/config/vite-plugin-inject-config";

export default defineConfig({
  plugins: [
    app("live-player"),
    injectWebAppConfig(),
    presetReact({
      chunkGroups: {
        core: /rollup|core-js|tslib|babel|scheduler|history|object-assign|hey-listen|react|react-router/,
        utils: /innoai-tech|date-fns|lodash|rxjs|filesize|buffer/,
        styling: /emotion|react-spring|mui/
      }
    })
  ]
});

