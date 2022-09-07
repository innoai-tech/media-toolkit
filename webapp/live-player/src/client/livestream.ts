import { createRequest } from "./client";

export interface Stream {
	id: string;
	name: string;
}

export const listStreams = createRequest<void, Stream[]>(
	() => ({
		method: "GET",
		url: "/api/live-streams",
	}),
);

export const liveStreamStatus = createRequest<
	{ id: string },
	{ active: boolean; observers: { [k: string]: number } }
>(
	({ id }) => ({
		method: "GET",
		url: `/api/live-streams/${id}/status`,
	}),
);

export const wsMp4fStreamFor = createRequest<{ id: string }, void>(
	({ id }) => ({
		method: "GET",
		url: `/api/live-streams/${id}/wsmp4f`,
	}),
);

export const takePic = createRequest<{ id: string }, void>(
	({ id }) => ({
		method: "PUT",
		url: `/api/live-streams/${id}/takepic`,
	}),
);

export const takeVideo = createRequest<{ id: string; stop?: boolean }, void>(
	({ id, stop }) => ({
		method: "PUT",
		url: `/api/live-streams/${id}/takevideo`,
		params: {
			stop,
		},
	}),
);

export interface BlobInfo {
	// BlobRef
	ref: string;
	userID: string;
	from: string;
	through: string;
	labels: { [k: string]: string[] };
}

export const exportDataset = createRequest<
	{ time: string; filter?: string },
	Blob
>(
	({ time, filter }) => ({
		method: "GET",
		url: "/api/datasets",
		params: {
			time,
			filter,
		},
	}),
);

export const listBlob = createRequest<
	{ time: string; filter?: string },
	BlobInfo[]
>(
	({ time, filter }) => ({
		method: "GET",
		url: "/api/blobs",
		params: {
			time,
			filter,
		},
	}),
);

export const getBlob = createRequest<{ ref: string }, any>(
	({ ref }) => ({
		method: "GET",
		url: `/api/blobs/${ref}`,
	}),
);

export const deleteBlob = createRequest<{ ref: string }, any>(
	({ ref }) => ({
		method: "DELETE",
		url: `/api/blobs/${ref}`,
	}),
);

export const labelBlob = createRequest<
	{ ref: string; label: string; value: string },
	void
>(
	({ ref, label, value }) => ({
		method: "PUT",
		url: `/api/blobs/${ref}/labels/${label}/${value}`,
	}),
);

export const unLabelBlob = createRequest<
	{ ref: string; label: string; value: string },
	void
>(
	({ ref, label, value }) => ({
		method: "DELETE",
		url: `/api/blobs/${ref}/labels/${label}/${value}`,
	}),
);
