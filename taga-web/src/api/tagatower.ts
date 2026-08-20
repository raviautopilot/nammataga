import API_BASE from "../config/api";
export interface CreateOrderResponse {
  key: string;
  order: {
    id: string;
    amount: number;
    currency: string;
    [key: string]: any;
  };
}

/**
 * 🔥 Smart fetch:
 * 1. Try normal URL
 * 2. If 404 → retry without /api
 */
const fetchWithFallback = async (url: string, options?: RequestInit) => {
  let res = await fetch(url, options);

  if (res.ok) return res;

  // Retry without /api if 404
  if (res.status === 404 && url.includes('/api/')) {
    const fallbackUrl = url.replace('/api', '');
    res = await fetch(fallbackUrl, options);
  }

  return res;
};

const TOWERS_API = `${API_BASE}/towers`;

export interface Room {
  id: string;
  name: string;
  type: 'apex-suite' | 'ac-room' | 'gents-dorm' | 'ladies-dorm';
  capacity: number;
  allowSingleBed: boolean;
}

export interface GuestDetail {
  name: string;
  age: number;
  contact: string;
  gender: 'male' | 'female' | '';
}

export interface CreateBookingRequest {
  roomId: string;
  checkInDate: string;
  checkOutDate: string;
  bookerPhone: string;
  bookingFor: 'self' | 'guest';
  bedCount: number;
  gender?: 'male' | 'female';
  guestDetails?: GuestDetail[];
  upiId?: string;
}

export interface BookingResponse {
  id: string;
  roomId: string;
  roomName: string;
  checkInDate: string;
  checkOutDate: string;
  bookerName: string;
  bookerId: string;
  bookingFor: 'self' | 'guest';
  bedCount: number;
  gender?: 'male' | 'female';
  guestDetails?: GuestDetail[];
  paymentStatus: 'pending' | 'confirmed' | 'cancelled' | 'refunded';
  advanceAmount: number;
  // ✅ NEW: computed booking status from backend
  bookingStatus: string;   // "upcoming" | "active" | "completed" | "cancelled"
}

export interface RoomAvailability {
  room: Room;
  available: boolean;
  availableBeds: number;
  genderRestriction?: 'male' | 'female';
}

const handleResponse = async (res: Response) => {
  if (!res.ok) {
    const data = await res.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(data.error || `Request failed: ${res.status}`);
  }
  return res.json();
};

const getAuthHeaders = (customHeaders?: Record<string, string>): Record<string, string> => {
  const token = localStorage.getItem('member_token') || localStorage.getItem('admin_token');
  const headers: Record<string, string> = { ...customHeaders };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
};

// ✅ Get all rooms
export const getAllRooms = async (): Promise<Room[]> => {
  try {
    const res = await fetchWithFallback(`${TOWERS_API}/rooms`, {
      headers: getAuthHeaders(),
    });
    return await handleResponse(res);
  } catch (error) {
    console.error('❌ Error fetching rooms:', error);
    throw error;
  }
};

// ✅ Check availability for a single room+date (kept for backward compat)
export const checkAvailability = async (
  roomId: string,
  date: string
): Promise<RoomAvailability> => {
  try {
    const res = await fetchWithFallback(
      `${TOWERS_API}/availability?roomId=${roomId}&date=${date}`,
      { headers: getAuthHeaders() }
    );
    return await handleResponse(res);
  } catch (error) {
    console.error('❌ Error checking availability:', error);
    throw error;
  }
};

// ✅ Bulk: Check availability of ALL rooms across a full date range in ONE API call.
export const checkAvailabilityRange = async (
  checkIn: string,  // YYYY-MM-DD
  checkOut: string  // YYYY-MM-DD
): Promise<Record<string, RoomAvailability>> => {
  const res = await fetchWithFallback(
    `${TOWERS_API}/availability-range?checkIn=${checkIn}&checkOut=${checkOut}`,
    { headers: getAuthHeaders() }
  );
  return await handleResponse(res);
};

// ✅ Create booking
export const createBooking = async (
  request: CreateBookingRequest,
  bookerId: string
): Promise<BookingResponse> => {
  try {
    const res = await fetchWithFallback(
      `${TOWERS_API}/bookings?bookerId=${encodeURIComponent(bookerId)}`,
      {
        method: 'POST',
        headers: getAuthHeaders({ 'Content-Type': 'application/json' }),
        body: JSON.stringify(request),
      }
    );
    return await handleResponse(res);
  } catch (error) {
    console.error('❌ Error creating booking:', error);
    throw error;
  }
};

// ✅ Get user bookings (Active)
export const getUserBookings = async (bookerId: string): Promise<BookingResponse[]> => {
  try {
    const res = await fetchWithFallback(
      `${TOWERS_API}/bookings?bookerId=${encodeURIComponent(bookerId)}`,
      { headers: getAuthHeaders() }
    );
    const bookings = await handleResponse(res);
    return bookings || [];
  } catch (error) {
    console.error('❌ Error fetching bookings:', error);
    throw error;
  }
};

// ✅ Get past user bookings (Archived)
export const getPastUserBookings = async (
  bookerId: string,
  year?: string,
  month?: string
): Promise<BookingResponse[]> => {
  try {
    let url = `${TOWERS_API}/bookings/past?bookerId=${encodeURIComponent(bookerId)}`;
    if (year) url += `&year=${year}`;
    if (month) url += `&month=${month}`;

    const res = await fetchWithFallback(url, { headers: getAuthHeaders() });
    const bookings = await handleResponse(res);
    return bookings || [];
  } catch (error) {
    console.error('❌ Error fetching past bookings:', error);
    throw error;
  }
};

// ✅ Confirm payment
export const confirmPayment = async (bookingId: string, upiId: string): Promise<void> => {
  try {
    const res = await fetchWithFallback(
      `${TOWERS_API}/bookings/${bookingId}/confirm-payment`,
      {
        method: 'POST',
        headers: getAuthHeaders({ 'Content-Type': 'application/json' }),
        body: JSON.stringify({ upiId }),
      }
    );

    if (!res.ok) {
      const text = await res.text();
      throw new Error(`Request failed: ${res.status} - ${text}`);
    }
  } catch (error) {
    console.error('❌ Error confirming payment:', error);
    throw error;
  }
};

// ✅ Cancel booking
export const cancelBooking = async (bookingId: string): Promise<void> => {
  try {
    const res = await fetchWithFallback(
      `${TOWERS_API}/bookings/${bookingId}`,
      {
        method: 'DELETE',
        headers: getAuthHeaders(),
      }
    );

    if (!res.ok) {
      const text = await res.text();
      throw new Error(`Request failed: ${res.status} - ${text}`);
    }
  } catch (error) {
    console.error('❌ Error cancelling booking:', error);
    throw error;
  }
};

// ✅ Create Razorpay Order (returns key + order)
export const createOrder = async (amount: number, notes?: Record<string, any>): Promise<CreateOrderResponse> => {
  try {
    const body: { amount: number; notes?: Record<string, any> } = { amount };
    if (notes) {
      body.notes = notes;
    }
    
    const res = await fetchWithFallback(
      `${TOWERS_API}/create-order`,
      {
        method: "POST",
        headers: getAuthHeaders({ "Content-Type": "application/json" }),
        body: JSON.stringify(body),
      }
    );
    return await handleResponse(res);
  } catch (error) {
    console.error("❌ Error creating Razorpay order:", error);
    throw error;
  }
};