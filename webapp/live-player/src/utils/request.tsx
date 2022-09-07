import { forEach, isArray, isObject, isUndefined } from "@innoai-tech/lodash";
import { createContext, ReactNode, useContext, useMemo } from "react";
import {
	BehaviorSubject,
	catchError,
	from,
	mergeMap,
	Observable,
	Subject,
	tap,
} from "rxjs";

export interface RequestConfig<TInputs> {
	method: string;
	url: string;
	params?: { [k: string]: any };
	headers?: { [k: string]: any };
	body?: { [k: string]: any };
	inputs: TInputs;
}

export interface RequestConfigCreator<TInputs, TRespData> {
	(input: TInputs): RequestConfig<TInputs>;
	TRespData: TRespData;
}

export interface FetcherResponse<TInputs, TData> {
	config: RequestConfig<TInputs>;
	status: number;
	headers: { [k: string]: string };
	body: TData;
}

export interface FetcherErrorResponse<TInputs extends any, TError extends any>
	extends FetcherResponse<TInputs, any> {
	error: TError;
}

export interface Fetcher {
	toHref: (requestConfig: RequestConfig<any>) => string;
	request: <TInputs extends any, TData extends any>(
		requestConfig: RequestConfig<TInputs>,
	) => Promise<FetcherResponse<TInputs, TData>>;
}

const getContentType = (headers: any = {}) =>
	headers["Content-Type"] || headers["content-type"] || "";

export const isContentTypeMultipartFormData = (headers: any) =>
	getContentType(headers).includes("multipart/form-data");
export const isContentTypeFormURLEncoded = (headers: any) =>
	getContentType(headers).includes("application/x-www-form-urlencoded");

export const paramsSerializer = (params: any) => {
	const searchParams = new URLSearchParams();

	const append = (k: string, v: any) => {
		if (isArray(v)) {
			forEach(v, (vv) => {
				append(k, vv);
			});
			return;
		}
		if (isObject(v)) {
			append(k, JSON.stringify(v));
			return;
		}
		if (isUndefined(v) || `${v}`.length === 0) {
			return;
		}
		searchParams.append(k, `${v}`);
	};

	forEach(params, (v, k) => {
		append(k, v);
	});

	return searchParams.toString();
};

export const transformRequestBody = (data: any, headers: any) => {
	if (isContentTypeMultipartFormData(headers)) {
		const formData = new FormData();

		const appendValue = (k: string, v: any) => {
			if (v instanceof File || v instanceof Blob) {
				formData.append(k, v);
			} else if (isArray(v)) {
				forEach(v, (item) => appendValue(k, item));
			} else if (isObject(v)) {
				formData.append(k, JSON.stringify(v));
			} else {
				formData.append(k, v as string);
			}
		};

		forEach(data, (v, k) => appendValue(k, v));

		return formData;
	}

	if (isContentTypeFormURLEncoded(headers)) {
		return paramsSerializer(data);
	}

	if (isArray(data) || isObject(data)) {
		return JSON.stringify(data);
	}

	return data;
};

export type RequestInterceptor = (
	requestConfig: RequestConfig<any>,
) => RequestConfig<any>;

export const applyRequestInterceptors =
	(...requestInterceptors: RequestInterceptor[]) =>
	(fetcher: Fetcher) => {
		return {
			request: <TInputs extends any, TRespData extends any>(
				requestConfig: RequestConfig<TInputs>,
			) => {
				for (const requestInterceptor of requestInterceptors) {
					requestConfig = requestInterceptor(requestConfig);
				}
				return fetcher.request<TInputs, TRespData>(requestConfig);
			},
			toHref: (requestConfig: RequestConfig<any>): string => {
				for (const requestInterceptor of requestInterceptors) {
					requestConfig = requestInterceptor(requestConfig);
				}
				return fetcher.toHref(requestConfig);
			},
		};
	};

export const createFetcher = (): Fetcher => {
	return {
		toHref: (requestConfig: RequestConfig<any>) => {
			return `${requestConfig.url}?${paramsSerializer(requestConfig.params)}`;
		},
		request: <TInputs extends any, TRespData extends any>(
			requestConfig: RequestConfig<TInputs>,
		) => {
			const reqInit: RequestInit = {
				method: requestConfig.method,
				headers: requestConfig.headers,
				body: transformRequestBody(requestConfig.body, requestConfig.headers),
			};

			return fetch(
				`${requestConfig.url}?${paramsSerializer(requestConfig.params)}`,
				reqInit,
			)
				.then(async (res) => {
					let body: any;

					if (res.headers.get("Content-Type")?.includes("application/json")) {
						body = await res.json();
					} else if (
						res.headers.get("Content-Type")?.includes(
							"application/octet-stream",
						)
					) {
						body = await res.blob();
					} else {
						body = await res.text();
					}

					const resp: any = {
						config: requestConfig,
						status: res.status,
						headers: {},
					};

					res.headers.forEach((value, key) => {
						resp.headers[key] = value;
					});

					resp.body = body as TRespData;

					return resp as FetcherResponse<TInputs, TRespData>;
				})
				.then((resp) => {
					if (resp.status >= 400) {
						(resp as FetcherErrorResponse<TInputs, any>).error = resp.body;
						throw resp;
					}
					return resp;
				});
		},
	};
};

const FetcherContext = createContext<{ fetcher?: Fetcher }>({});

export const FetcherProvider = ({ fetcher, children }: {
	fetcher: Fetcher;
	children: ReactNode;
}) => {
	return (
		<FetcherContext.Provider value={{ fetcher: fetcher }}>
			{children}
		</FetcherContext.Provider>
	);
};

const useFetcher = () => {
	return useContext(FetcherContext).fetcher || createFetcher();
};

class RequestSubject<TInputs, TBody, TError> extends Observable<
	FetcherResponse<TInputs, TBody>
> {
	requesting$ = new BehaviorSubject<boolean>(false);
	error$ = new Subject<FetcherErrorResponse<TInputs, TError>>();

	_success$ = new Subject<FetcherResponse<TInputs, TBody>>();
	_input$ = new Subject<TInputs>();

	constructor(
		private fetcher: Fetcher,
		private createConfig: RequestConfigCreator<TInputs, TBody>,
	) {
		super((subscriber) => {
			return this._success$.subscribe(subscriber);
		});

		this._input$.pipe(
			mergeMap((input) => {
				this.requesting$.next(true);

				return from(fetcher.request<TInputs, TBody>(createConfig(input)));
			}),
			tap((resp) => {
				return this._success$.next(resp);
			}),
			catchError((errorResp) => {
				this.error$.next(errorResp);
				return errorResp;
			}),
			tap(() => {
				this.requesting$.next(false);
			}),
		).subscribe();
	}

	next(value: TInputs) {
		this._input$.next(value);
	}

	toHref(value: TInputs) {
		return this.fetcher.toHref(this.createConfig(value));
	}
}

interface RespError {
	code: number;
	msg: string;
	desc: string;
}

export const useRequest = <TReq, TRespData>(
	createConfig: RequestConfigCreator<TReq, TRespData>,
) => {
	const fetcher = useFetcher();

	return useMemo(() => {
		return new RequestSubject<TReq, TRespData, RespError>(
			fetcher,
			createConfig,
		);
	}, [fetcher]);
};

export const parseSearch = (s: string): { [k: string]: string[] } => {
	if (s[0] === "?") {
		s = s.slice(1);
	}

	const p = new URLSearchParams(s);

	const labels: { [k: string]: string[] } = {};

	for (const k in p) {
		labels[k] = p.getAll(k);
	}

	return labels;
};

export const stringifySearch = (query: { [k: string]: string[] }): string => {
	const p = new URLSearchParams();
	for (const k in query) {
		for (const v of query[k]) {
			p.append(k, v);
		}
	}
	return `?${p.toString()}`;
};
