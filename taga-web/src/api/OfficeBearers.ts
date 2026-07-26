/**
 * OfficeBearers Utility Module
 * Handles formatting and utility functions for office bearers data
 */

import { API_BASE_URL } from '../config/api';

interface OfficeBearerData {
  name: string;
  designation?: string;
  position?: string;
  department?: string;
  location?: string;
  tenure?: string;
  email?: string;
  phone?: string;
  experience?: string;
  qualification?: string;
  education?: string;
  image?: string;
}

interface FormattedBearer {
  name: string;
  position: string;
  department: string;
  location: string;
  tenure: string;
  email: string;
  phone: string;
  experience: string;
  education: string;
  image: string | null;
}

interface BearerStatistics {
  totalBearers: number;
  positionCount: number;
  locationsCount: number;
  withEmail: number;
  withPhone: number;
}

/**
 * Format office bearer data for display
 */
export const formatOfficeBearerData = (
  bearer: OfficeBearerData
): FormattedBearer => {
  return {
    name: bearer.name || "",
    position: bearer.designation || bearer.position || "",
    department: bearer.department || "",
    location: bearer.location || "",
    tenure: bearer.tenure || "2023-2025",
    email: bearer.email || "",
    phone: bearer.phone || "",
    experience: bearer.experience || "",
    education: bearer.qualification || bearer.education || "",
    image:
      bearer.image && bearer.image.startsWith("http")
        ? bearer.image
        : bearer.image
          ? `${API_BASE_URL}${bearer.image}`
          : null,
  };
};

/**
 * Group office bearers by position
 */
export const groupByPosition = (
  bearers: FormattedBearer[]
): Record<string, FormattedBearer[]> => {
  return bearers.reduce(
    (groups, bearer) => {
      const position = bearer.position || "Other";
      if (!groups[position]) {
        groups[position] = [];
      }
      groups[position].push(bearer);
      return groups;
    },
    {} as Record<string, FormattedBearer[]>
  );
};

/**
 * Group office bearers by location/district
 */
export const groupByLocation = (
  bearers: FormattedBearer[]
): Record<string, FormattedBearer[]> => {
  return bearers.reduce(
    (groups, bearer) => {
      const location = bearer.location || "Unknown";
      if (!groups[location]) {
        groups[location] = [];
      }
      groups[location].push(bearer);
      return groups;
    },
    {} as Record<string, FormattedBearer[]>
  );
};

/**
 * Sort office bearers by tenure or name
 */
export const sortBearers = (
  bearers: FormattedBearer[],
  sortBy: keyof FormattedBearer = "name",
  order: "asc" | "desc" = "asc"
): FormattedBearer[] => {
  const sorted = [...bearers].sort((a, b) => {
    let valA: any = a[sortBy] || "";
    let valB: any = b[sortBy] || "";

    if (typeof valA === "string") {
      valA = valA.toLowerCase();
      valB = valB.toLowerCase();
    }

    if (valA < valB) return order === "asc" ? -1 : 1;
    if (valA > valB) return order === "asc" ? 1 : -1;
    return 0;
  });

  return sorted;
};

/**
 * Search office bearers by name or position
 */
export const searchBearers = (
  bearers: FormattedBearer[],
  query: string
): FormattedBearer[] => {
  const lowerQuery = query.toLowerCase();

  return bearers.filter(
    (bearer) =>
      bearer.name.toLowerCase().includes(lowerQuery) ||
      bearer.position.toLowerCase().includes(lowerQuery) ||
      (bearer.location && bearer.location.toLowerCase().includes(lowerQuery))
  );
};

/**
 * Get office bearer statistics
 */
export const getStatistics = (
  bearers: FormattedBearer[]
): BearerStatistics => {
  return {
    totalBearers: bearers.length,
    positionCount: [...new Set(bearers.map((b) => b.position))].length,
    locationsCount: [
      ...new Set(bearers.map((b) => b.location || "Unknown")),
    ].length,
    withEmail: bearers.filter((b) => b.email).length,
    withPhone: bearers.filter((b) => b.phone).length,
  };
};

/**
 * Export office bearers to CSV format
 */
export const exportToCSV = (bearers: FormattedBearer[]): string => {
  if (bearers.length === 0) return "";

  const headers = [
    "Name",
    "Position",
    "Location",
    "Department",
    "Email",
    "Phone",
    "Experience",
    "Education",
  ];

  const rows = bearers.map((bearer) => [
    bearer.name,
    bearer.position,
    bearer.location || "",
    bearer.department || "",
    bearer.email || "",
    bearer.phone || "",
    bearer.experience || "",
    bearer.education || "",
  ]);

  const csvContent = [
    headers.join(","),
    ...rows.map((row) =>
      row.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(",")
    ),
  ].join("\n");

  return csvContent;
};

/**
 * Download CSV file
 */
export const downloadCSV = (
  csvContent: string,
  filename: string = "office-bearers.csv"
): void => {
  const element = document.createElement("a");
  element.setAttribute(
    "href",
    "data:text/csv;charset=utf-8," + encodeURIComponent(csvContent)
  );
  element.setAttribute("download", filename);
  element.style.display = "none";
  document.body.appendChild(element);
  element.click();
  document.body.removeChild(element);
};
