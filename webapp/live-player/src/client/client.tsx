import type { RequestConfig, RequestConfigCreator } from "../utils";

export const createRequest = <TInputs, TRespData>(
	createConfig: (input: TInputs) => Omit<RequestConfig<TInputs>, "inputs">,
) => {
	const create = (inputs: TInputs) => ({ ...createConfig(inputs), inputs });
	return create as RequestConfigCreator<TInputs, TRespData>;
};
