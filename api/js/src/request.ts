// Contains functions for getting data from the server and from localStorage.

import { csrf } from "./csrf";

const src = findServer();

export async function fetchJson<T>(url: string | URL, options: RequestInit): Promise<T> {
    if (url instanceof URL) {
        url = url.href;
    }
    const request = new Request(url, options);
    const response = await fetch(request);
    return await response.json();
}

export function submitJson<T>(url: string | URL, data: unknown): Promise<T> {
    const options = {
        body: JSON.stringify(data),
        headers: {
            Accept: "application/json",
            "Content-Type": "application/json",
            "X-CSRF-Token": csrf(),
        },
        method: "POST",
        mode: "cors" as RequestMode,
    };
    return fetchJson<T>(url, options);
}

function findServer(): string {
    const url = new URL(location.href);
    if (document.currentScript == null) {
        return url.origin;
    }

    const { origin, port } = document.currentScript.dataset;
    if (origin != null) {
        return origin;
    }
    if (port != null) {
        url.port = port;
    }
    return url.origin;
}

export function resolve(path: string): URL {
    return new URL(path, src);
}
