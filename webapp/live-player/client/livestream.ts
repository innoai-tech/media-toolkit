import { RequestConfig, createRequest } from "./client";

export interface Stream {
  id: string;
  name: string;
}


export const listStreams = createRequest<void, Stream[]>(() => ({
  method: "GET",
  url: "/api/live-streams"
}));


export const wsMp4fStreamFor = createRequest<{ id: string }, Blob>(({ id }) => ({
  method: "GET",
  url: `/api/live-streams/${id}/wsmp4f`
}));