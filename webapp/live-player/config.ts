import { confLoader } from "@innoai-tech/config";

export const ENVS = {
  ONLINE: "ONLINE"
};

export const APP_CONFIG = {
  SRV_API: (env: string, feature: string) => {
    if (env === "local") {
      return `//0.0.0.0:777`;
    }
    return "";
  }
};

export const APP_MANIFEST = {
  name: "Live Player",
  crossorigin: "use-credentials"
};

export default confLoader<keyof typeof APP_CONFIG>();