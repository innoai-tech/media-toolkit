import { createRoot } from "react-dom/client";
import { App } from "./App";
import { ReactNode, useMemo } from "react";
import {
	applyRequestInterceptors,
	createFetcher,
	FetcherProvider,
} from "./utils";
import conf from "./config";
import { BrowserRouter } from "react-router-dom";

const root = createRoot(document.getElementById("root") as any);

const fixBaseURL = (baseURL: string) => {
	if (baseURL) {
		if (baseURL.startsWith("//")) {
			return `${location.protocol}${baseURL}`;
		}
		return baseURL;
	}
	return location.origin;
};

const Bootstrap = ({ children }: { children: ReactNode }) => {
	const fetcher = useMemo(
		() =>
			applyRequestInterceptors((requestConfig) => {
				const c = conf();
				if (
					!(
						requestConfig.url.startsWith("//") ||
						requestConfig.url.startsWith("http:") ||
						requestConfig.url.startsWith("https://")
					)
				) {
					requestConfig.url = `${fixBaseURL(c.SRV_API)}${requestConfig.url}`;
				}

				if (location.protocol.endsWith("s:")) {
					if (!requestConfig.params) {
						requestConfig.params = {};
					}
					requestConfig.params["secure"] = true;
				}

				return requestConfig;
			})(createFetcher()),
		[],
	);

	return (
		<BrowserRouter>
			<FetcherProvider fetcher={fetcher}>{children}</FetcherProvider>
		</BrowserRouter>
	);
};

root.render(
	<Bootstrap>
		<App />
	</Bootstrap>,
);
