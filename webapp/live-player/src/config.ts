import { AppConfig, AppContext, confLoader } from "@innoai-tech/config";

const APP_CONFIG = {
	SRV_API: ({ env }: AppContext) => {
		if (env === "local") {
			return "//localhost:777";
		}
		return "";
	},
};

export const CONFIG: AppConfig = {
	name: "live-stream",
	group: "",
	manifest: {
		crossorigin: "use-credentials",
	},
	config: APP_CONFIG,
};

export default confLoader<keyof typeof APP_CONFIG>();
