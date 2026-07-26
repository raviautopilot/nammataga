/// <reference path="../../vite-env.d.ts" />
// src/config/api.ts
declare global {
    interface Window {
        _env_?: {
            VITE_API_BASE_URL?: string;
            VITE_RAZORPAY_KEY?: string; // ✅ ADD THIS LINE
            API_URL?: string;
        };
    }
}
type Source =
    | "domain match: nammataga.com"
    | "domain match: stg.nammataga.com"
    | "domain match: tst.nammataga.com"
    | "domain match: dev.nammataga.com"
    | "window._env_.VITE_API_BASE_URL"
    | "window._env_.API_URL"
    | "import.meta.env.VITE_API_BASE_URL"
    | "empty default";

let rawUrl: string = "";
let source: Source = "empty default";

// 0️⃣ Domain-based auto-detection (highest priority to prevent misconfiguration across environments)
const hostname = typeof window !== "undefined" ? window.location.hostname : "";

if (hostname === "nammataga.com" || hostname === "www.nammataga.com") {
    rawUrl = "https://api.nammataga.com/api";
    source = "domain match: nammataga.com";
} else if (hostname === "stg.nammataga.com") {
    rawUrl = "https://stgapi.nammataga.com/api";
    source = "domain match: stg.nammataga.com";
} else if (hostname === "tst.nammataga.com") {
    rawUrl = "https://tstapi.nammataga.com/api";
    source = "domain match: tst.nammataga.com";
} else if (hostname === "dev.nammataga.com") {
    rawUrl = "https://devapi.nammataga.com/api";
    source = "domain match: dev.nammataga.com";
}

// 1️⃣ Runtime injected env (highest priority fallback for local dev / alternative IPs)
else if (window._env_?.VITE_API_BASE_URL !== undefined) {
    rawUrl = window._env_.VITE_API_BASE_URL;
    source = "window._env_.VITE_API_BASE_URL";
} else if (window._env_?.API_URL !== undefined) {
    rawUrl = window._env_.API_URL;
    source = "window._env_.API_URL";
}

// 2️⃣ Build-time env (Vite)
else if (import.meta.env.VITE_API_BASE_URL !== undefined) {
    rawUrl = import.meta.env.VITE_API_BASE_URL;
    source = "import.meta.env.VITE_API_BASE_URL";
}

// 3️⃣ Final normalization (NO external fallback)
rawUrl = rawUrl ?? "";

// remove trailing slashes
export const API_BASE_URL = (rawUrl || "").replace(/\/+$/, "");

// helper to safely build paths
export function apiPath(path?: string): string {
    if (!path) return "";

    const cleanPath = path.replace(/^\/+/, "");
    return `${API_BASE_URL}/${cleanPath}`;
}

// // debug output (keep during setup)
// console.group("API URL Sources");
// console.log("window._env_.API_URL:", window._env_?.API_URL);
// console.log("window._env_.VITE_API_BASE_URL:", window._env_?.VITE_API_BASE_URL);
// console.log("import.meta.env.VITE_API_BASE_URL:", import.meta.env.VITE_API_BASE_URL);
// console.log(`✅ Using API_BASE_URL: "${API_BASE_URL}" (from: ${source})`);
// console.groupEnd();

export default API_BASE_URL;

export const RAZORPAY_KEY =
    window._env_?.VITE_RAZORPAY_KEY ||
    import.meta.env.VITE_RAZORPAY_KEY ||
    "";
console.log("💳 Razorpay Key:", RAZORPAY_KEY);