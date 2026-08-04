/* --------------------------------
   Utility: Sort Documents
   - Newest year first (descending)
   - Same year: alphabetical by title
   - Missing year: uses current year
-------------------------------- */
export const sortDocuments = (docs: Document[]): Document[] => {
  const currentYear = new Date().getFullYear().toString();

  return [...docs].sort((a, b) => {
    // Get years, fallback to current year if missing
    const yearA = (a.year?.trim() || currentYear).substring(0, 4); // Handle "2026-2027" -> "2026"
    const yearB = (b.year?.trim() || currentYear).substring(0, 4);

    // Convert to numbers for comparison
    const yearNumA = parseInt(yearA) || parseInt(currentYear);
    const yearNumB = parseInt(yearB) || parseInt(currentYear);

    // Sort by year descending (newest first)
    if (yearNumB !== yearNumA) {
      return yearNumB - yearNumA;
    }

    // Same year: sort alphabetically, case-insensitive
    const titleA = (a.title || '').toLowerCase().trim();
    const titleB = (b.title || '').toLowerCase().trim();

    return titleA.localeCompare(titleB);
  });
};

/* --------------------------------
   Utility: Sort External Links
   - Alphabetical by title only
-------------------------------- */
export const sortExternalLinks = (links: any[]): any[] => {
  return [...links].sort((a, b) => {
    const titleA = (a.title || '').toLowerCase().trim();
    const titleB = (b.title || '').toLowerCase().trim();
    return titleA.localeCompare(titleB);
  });
};
import { API_BASE_URL } from "../config/api";

/* --------------------------------
   Base URL (ENV SAFE)
-------------------------------- */

const BASE_URL = API_BASE_URL;


// API endpoint (for categories & documents)
const API_BASE = `${BASE_URL}/resources`;

/* --------------------------------
   Types
-------------------------------- */

export interface Category {
  id: string;
  name: string;
}

export interface Document {
  title: string;
  year: string;
  subcategory?: string;
  url?: string;
}

/* --------------------------------
   Helper: Handle API Response
-------------------------------- */

const handleResponse = async (res: Response) => {
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Request failed: ${res.status} - ${text}`);
  }
  return res.json();
};

/* --------------------------------
   Helper: Get Full File URL (PDF FIX)
-------------------------------- */

export const getFileUrl = (url: string) => {
  const base = API_BASE_URL;
  const cleanBase = base.replace(/\/api\/?$/, "");

  let cleanUrl = url.startsWith('/') ? url : `/${url}`;

  // Normalize /data/docs/ -> /docs/
  if (cleanUrl.startsWith('/data/docs/')) {
    cleanUrl = cleanUrl.replace('/data/docs/', '/docs/');
  }

  // Route through /api/docs/ for guaranteed Nginx proxying to Go backend
  let finalPath = cleanUrl;
  if (cleanUrl.startsWith('/docs/')) {
    finalPath = cleanUrl.replace('/docs/', '/api/docs/');
  }

  // Safely encode URI components (spaces -> %20, & -> %26)
  const encodedUrl = finalPath.split('/').map(segment => encodeURIComponent(segment)).join('/');

  return `${cleanBase}${encodedUrl}`;
};

/* --------------------------------
   Helper: Auth Header Extraction
-------------------------------- */

const getAuthHeaders = (): Record<string, string> => {
  const token = localStorage.getItem("member_token") || localStorage.getItem("admin_token");
  if (token) {
    return { Authorization: `Bearer ${token}` };
  }
  return {};
};

/* --------------------------------
   Get Categories
-------------------------------- */

export const getCategories = async (): Promise<Category[]> => {
  try {
    const res = await fetch(`${API_BASE}`, {
      headers: getAuthHeaders(),
    });
    return await handleResponse(res);
  } catch (error) {
    console.error("❌ Error fetching categories:", error);
    throw error;
  }
};

/* --------------------------------
   Get Documents by Category
-------------------------------- */

export const getDocumentsByCategory = async (
  categoryId: string,
  subcategory?: string
): Promise<Document[]> => {
  try {
    let url = `${API_BASE}/${categoryId}`;

    if (subcategory) {
      url += `?subcategory=${encodeURIComponent(subcategory)}`;
    }

    const res = await fetch(url, {
      headers: getAuthHeaders(),
    });
    return await handleResponse(res);
  } catch (error) {
    console.error("❌ Error fetching documents:", error);
    throw error;
  }
};
export const getResourcesBanner = async (): Promise<string> => {
  const res = await fetch(`${BASE_URL}/resources-banner`);
  const data = await res.json();
  return data.image;
};
export const getExternalLinks = async () => {
  const res = await fetch(`${API_BASE}/external-links`, {
    headers: getAuthHeaders(),
  });
  return await handleResponse(res);
};