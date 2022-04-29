import { createRoot } from "react-dom/client";
import { App } from "./App";
import { ReactNode, useMemo } from "react";
import { applyRequestInterceptors, createFetcher, FetcherProvider } from "../utils";
import conf from "./config";

const root = createRoot(document.getElementById("root") as any);

const Bootstrap = ({ children }: { children: ReactNode }) => {
  const fetcher = useMemo(() => applyRequestInterceptors(
    (requestConfig) => {
      if (!(requestConfig.url.startsWith("//") || requestConfig.url.startsWith("http:") || requestConfig.url.startsWith("https://"))) {
        requestConfig.url = `${conf().SRV_API}${requestConfig.url}`;
      }
      return requestConfig;
    }
  )(createFetcher()), []);

  return (
    <FetcherProvider fetcher={fetcher}>
      {children}
    </FetcherProvider>
  );
};

root.render(
  <Bootstrap>
    <App />
  </Bootstrap>
);
