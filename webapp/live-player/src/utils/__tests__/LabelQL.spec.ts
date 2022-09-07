import { describe, it, expect } from "vitest";
import { formatLabelQL, parseLabelQL, quote } from "../LabelQL";

describe("LabelQL", () => {
	it("LabelQL", () => {
		const labels = {
			a: ["1", "2"],
			b: [`"x,={||}x"`],
		};

		const ret = formatLabelQL(labels);

		expect(ret).toBe(`{a="1",a="2",b=${quote(`"x,={||}x"`)}}`);

		const parsed = parseLabelQL(ret);

		expect(parsed).toEqual(labels);
	});
});
