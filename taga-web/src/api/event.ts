/* --------------------------------
   Base URL (ENV SAFE)
-------------------------------- */

import { API_BASE_URL } from '../config/api';

const BASE_URL = API_BASE_URL;

// API endpoint for events
const API_BASE = `${BASE_URL}/events`;

/* --------------------------------
   Types
-------------------------------- */

export interface Event {
  id: string;
  title: string;
  date: Date | string;
  location: string;
  description: string;
  attendees?: number;
  imageUrl?: string;
  status: "upcoming" | "completed";
  category?: string;
  time?: string;
}

export interface EventsByYear {
  [year: number]: Event[];
}

export interface Gallery {
  id: string;
  year: number;
  images: string[];
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
   Get All Years
-------------------------------- */

export const getEventYears = async (): Promise<number[]> => {
  try {
    const res = await fetch(`${API_BASE}/years`);
    return await handleResponse(res);
  } catch (error) {
    console.error("❌ Error fetching event years:", error);
    throw error;
  }
};

/* --------------------------------
   Get Events by Year
-------------------------------- */

export const getEventsByYear = async (year: number): Promise<Event[]> => {
  try {
    const res = await fetch(`${API_BASE}/year/${year}`);
    return await handleResponse(res);
  } catch (error) {
    console.error(`❌ Error fetching events for year ${year}:`, error);
    throw error;
  }
};

/* --------------------------------
   Get Upcoming Events
-------------------------------- */

export const getUpcomingEvents = async (): Promise<Event[]> => {
  try {
    const res = await fetch(`${API_BASE}/upcoming`);
    return await handleResponse(res);
  } catch (error) {
    console.error("❌ Error fetching upcoming events:", error);
    throw error;
  }
};

/* --------------------------------
   Get Past Events
-------------------------------- */

export const getPastEvents = async (): Promise<Event[]> => {
  try {
    const res = await fetch(`${API_BASE}/past`);
    return await handleResponse(res);
  } catch (error) {
    console.error("❌ Error fetching past events:", error);
    throw error;
  }
};

/* --------------------------------
   Get Event by ID
-------------------------------- */

export const getEventById = async (eventId: string): Promise<Event> => {
  try {
    const res = await fetch(`${API_BASE}/${eventId}`);
    return await handleResponse(res);
  } catch (error) {
    console.error(`❌ Error fetching event ${eventId}:`, error);
    throw error;
  }
};

/* --------------------------------
   Get Gallery by Year
-------------------------------- */

export const getGalleryByYear = async (year: number): Promise<Gallery> => {
  try {
    const res = await fetch(`${API_BASE}/gallery/${year}`);
    return await handleResponse(res);
  } catch (error) {
    console.error(`❌ Error fetching gallery for year ${year}:`, error);
    throw error;
  }
};

/* --------------------------------
   Create Event (Admin)
-------------------------------- */

export const createEvent = async (
  event: Omit<Event, "id">,
  token: string
): Promise<Event> => {
  try {
    const res = await fetch(`${API_BASE}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(event),
    });
    return await handleResponse(res);
  } catch (error) {
    console.error("❌ Error creating event:", error);
    throw error;
  }
};

/* --------------------------------
   Update Event (Admin)
-------------------------------- */

export const updateEvent = async (
  eventId: string,
  event: Partial<Event>,
  token: string
): Promise<Event> => {
  try {
    const res = await fetch(`${API_BASE}/${eventId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(event),
    });
    return await handleResponse(res);
  } catch (error) {
    console.error(`❌ Error updating event ${eventId}:`, error);
    throw error;
  }
};

/* --------------------------------
   Delete Event (Admin)
-------------------------------- */

export const deleteEvent = async (
  eventId: string,
  token: string
): Promise<void> => {
  try {
    const res = await fetch(`${API_BASE}/${eventId}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(`Delete failed: ${res.status} - ${text}`);
    }
  } catch (error) {
    console.error(`❌ Error deleting event ${eventId}:`, error);
    throw error;
  }
};
