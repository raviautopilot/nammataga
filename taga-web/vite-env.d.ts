/// <reference types="vite/client" />
// TODO: need to remove duplicate value, use the same variable accross
interface ImportMetaEnv {
    readonly VITE_API_URL?: string;
    // add other custom vars here...
}

interface ImportMeta {
    readonly env: ImportMetaEnv;
}