export const quote = (s: string = "") => {
	return `"${s.replace(/"/g, "\\\"")}"`;
};

export const unquote = (s: string = "") => {
	return s.slice(1, s.length - 1).replace(/\\"/g, '"');
};

export const formatLabelQL = (labels: { [k: string]: string[] }) => {
	return `{${Object.keys(labels)
		.map((k) => labels[k].map((v) => `${k}=${quote(v)}`).join(","))
		.join(",")}}`;
};

export const parseLabelQL = (
	labelQL: string = "",
): { [k: string]: string[] } => {
	let ql = labelQL.trim();

	const labels: { [k: string]: string[] } = {};

	const commit = (key: string, value: string) => {
		if (key && value) {
			labels[key] = [...(labels[key] || []), unquote(value)];
		}
	};

	let stringOpen = false;
	let key = "";
	let buf = "";

	for (const c of ql) {
		if (!stringOpen) {
			if (c === "{") {
				continue;
			}

			if (c === "=") {
				key = buf;
				buf = "";
				continue;
			}

			if (c === "}" || c === ",") {
				commit(key, buf.trim());
				buf = "";
				continue;
			}
		}

		if (c === '"') {
			if (buf[buf.length - 1] !== "\\") {
				stringOpen = !stringOpen;
			}
		}

		buf += c;
	}

	return labels;
};
