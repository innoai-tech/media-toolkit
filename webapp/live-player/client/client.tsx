import useSWR from "swr";


export interface RequestConfig {
  method: string;
  url: string;
  params?: { [k: string]: any };
  headers?: { [k: string]: any };
  body?: { [k: string]: any };
}

export const createRequest = <TReq, TRespData>(createConfig: (input: TReq) => RequestConfig) => {
  return createConfig as RequestCreator<TReq, TRespData>;
};

export interface RequestCreator<TReq, TRespData> {
  (input: TReq): RequestConfig;

  TRespData: TRespData;
}

const baseURL = "http://0.0.0.0:777";

export const toHref = (requestConfig: RequestConfig): string => {
  return `${baseURL}${requestConfig.url}`;
};

export const createFetcher = <TRespData extends any>(requestConfig: RequestConfig) => {
  return (_: string) => {
    const req: RequestInit = {
      method: requestConfig.method,
      headers: requestConfig.headers
    };

    return fetch(toHref(requestConfig), req)
      .then(async (res) => {
        let data: any;
        if (res.headers.get("Content-Type")?.includes("application/json")) {
          data = await res.json();
        }

        return ({
          status: res.status,
          headers: res.headers,
          data: data as TRespData
        });
      });
  };
};

export const useSWRWith = <TReq, TRespData>(createConfig: RequestCreator<TReq, TRespData>, input: TReq) => {
  const c = createConfig(input);
  return useSWR(c.url, createFetcher<TRespData>(c));
};

