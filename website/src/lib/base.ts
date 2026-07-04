// BASE_URL may or may not carry a trailing slash depending on config;
// normalize so consumers can always write `${base}docs/...`.
export const base = import.meta.env.BASE_URL.replace(/\/+$/, "") + "/";
