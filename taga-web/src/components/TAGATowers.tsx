import React, { useState, useEffect, useCallback, useMemo } from 'react';
import API_BASE from "../config/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { RadioGroup, RadioGroupItem } from './ui/radio-group';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Calendar } from './ui/calendar';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog';
import { Alert, AlertDescription } from './ui/alert';
import { Separator } from './ui/separator';
import { ScrollArea } from './ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs';
import {
  Building2,
  Calendar as CalendarIcon,
  CheckCircle,
  XCircle,
  Phone,
  Mail,
  MapPin,
  User,
  Clock,
  BedDouble,
  AlertCircle,
  Users,
  Smartphone,
  IndianRupee,
  ShieldCheck,
  ChevronDown,
  Loader2,
} from 'lucide-react';
import { format, addDays, isBefore, startOfDay, isAfter } from 'date-fns';
import { toast } from 'sonner';
import {
  getAllRooms,
  checkAvailability,
  checkAvailabilityRange,
  createBooking,
  getUserBookings,
  getPastUserBookings,
  confirmPayment,
  cancelBooking,
  createOrder,
} from '../api/tagatower';
import type { Room, BookingResponse, RoomAvailability } from '../api/tagatower';
import { loadRazorpayScript } from '../utils/razorpay';

// ═══════════════════════════════════════════════════════════════════════════
// CONSTANTS & TYPES
// ═══════════════════════════════════════════════════════════════════════════

const ADVANCE_AMOUNTS = {
  self: 1,
  guest: 1,
} as const;

const BOOKING_DAYS_LIMIT = 10;
const MIN_STAY_DAYS = 1;
const PHONE_REGEX = /^(\+91|0)?[6-9]\d{9}$/;
const PHONE_PLACEHOLDER = '+91 9876543210';
const INITIAL_PHONE_PREFIX = '+91 ';

const ROOM_TYPE_LABELS: Record<string, string> = {
  'apex-suite': 'Apex Suite A/C',
  'ac-room': 'A/C Room',
  'gents-dorm': 'Gents Dormitory',
  'ladies-dorm': 'Ladies Dormitory',
};

const ROOM_NUMBER_MAP: Record<string, number> = {
  'apex-1': 1,
  'kurinchi': 3,
  'pavalam': 7,
  'malligai': 9,
  'kavery': 11,
  'vasantham': 12,
  'pasumai': 18,
  'gents-dorm': 20,
  'ladies-dorm': 22,
};

const TOTAL_BED_CAPACITY = 35;

const CARETAKER_INFO = {
  caretakers: [
    { name: 'Mr. Muthu', phone: '96007 63744' },
    { name: 'Mr. Mariyappan', phone: '96770 10300' }
  ],
  email: 'tagatower@nammataga.com',
  address: 'TAGA Towers, 123 Agriculture Complex Road, T. Nagar, Chennai - 600017, Tamil Nadu',
};

interface TAGATowersProps {
  isLoggedIn: boolean;
  isPaidMember: boolean;
  isAdmin?: boolean;
}

interface GuestDetail {
  name: string;
  age: number;
  contact: string;
}

interface LoggedInUser {
  name: string;
  tagaId: string;
  email: string;
}

interface BookingStatusConfig {
  label: string;
  icon: string;
  badgeClass: string;
}

// ═══════════════════════════════════════════════════════════════════════════
// UTILITY FUNCTIONS
// ═══════════════════════════════════════════════════════════════════════════

/**
 * Safely retrieves logged-in user from localStorage
 */
function getLoggedInUser(): LoggedInUser {
  try {
    const raw = localStorage.getItem('user');
    if (!raw) return { name: 'Member', tagaId: '', email: '' };
    const u = JSON.parse(raw);
    return {
      name: u.name || 'Member',
      tagaId: u.tagaId || '',
      email: u.emailId || u.username || '',
    };
  } catch (error) {
    console.warn('Failed to parse user from localStorage:', error);
    return { name: 'Member', tagaId: '', email: '' };
  }
}

/**
 * Validates Indian mobile number format
 */
function validateMobileNumber(phone: string): { isValid: boolean; error?: string } {
  if (!phone || phone.trim() === '' || phone === INITIAL_PHONE_PREFIX) {
    return { isValid: false, error: 'Mobile number is required' };
  }

  const cleaned = phone.replace(/[\s\-\(\)]/g, '');
  if (!PHONE_REGEX.test(cleaned)) {
    return {
      isValid: false,
      error: 'Enter a valid 10-digit mobile number (e.g., 9876543210)',
    };
  }
  return { isValid: true };
}

/**
 * Computes booking status from booking data
 */
function computeBookingStatus(booking: BookingResponse): string {
  if (booking.paymentStatus === 'cancelled' || booking.paymentStatus === 'refunded') {
    return 'cancelled';
  }
  const now = new Date();
  const checkIn = new Date(booking.checkInDate);
  const checkOut = new Date(booking.checkOutDate);
  if (now < checkIn) return 'upcoming';
  if (now >= checkOut) return 'completed';
  return 'active';
}

/**
 * Gets effective booking status (backend or computed)
 */
function getEffectiveBookingStatus(booking: BookingResponse): string {
  return booking.bookingStatus || computeBookingStatus(booking);
}

/**
 * Returns booking status styling configuration
 */
function getBookingStatusConfig(status: string): BookingStatusConfig {
  switch (status) {
    case 'upcoming':
      return {
        label: 'Upcoming',
        icon: '🟡',
        badgeClass: 'bg-yellow-100 text-yellow-800 border-yellow-300',
      };
    case 'active':
      return {
        label: 'Active Stay',
        icon: '🔵',
        badgeClass: 'bg-blue-100 text-blue-800 border-blue-300',
      };
    case 'completed':
      return {
        label: 'Completed',
        icon: '✅',
        badgeClass: 'bg-green-100 text-green-800 border-green-300',
      };
    case 'cancelled':
      return {
        label: 'Cancelled',
        icon: '❌',
        badgeClass: 'bg-red-100 text-red-800 border-red-300',
      };
    default:
      return {
        label: 'Confirmed',
        icon: '✅',
        badgeClass: 'bg-green-100 text-green-800 border-green-300',
      };
  }
}

/**
 * Gets room type display label
 */
function getRoomTypeLabel(type: string): string {
  return ROOM_TYPE_LABELS[type] || type;
}

// ═══════════════════════════════════════════════════════════════════════════
// MAIN COMPONENT
// ═══════════════════════════════════════════════════════════════════════════

export function TAGATowers({ isLoggedIn, isPaidMember, isAdmin = false }: TAGATowersProps) {
  const loggedInUser = getLoggedInUser();
  const BOOKER_NAME = loggedInUser.name;
  const BOOKER_ID = loggedInUser.tagaId;

  // ─────────────────────────────────────────────────────────────────────────
  // STATE: Dates
  // ─────────────────────────────────────────────────────────────────────────
  const [checkInDate, setCheckInDate] = useState<Date | undefined>(new Date());
  const [checkOutDate, setCheckOutDate] = useState<Date | undefined>(
    addDays(new Date(), MIN_STAY_DAYS)
  );
  const [calendarMonth, setCalendarMonth] = useState<Date>(new Date());

  // ─────────────────────────────────────────────────────────────────────────
  // STATE: Rooms & Availability
  // ─────────────────────────────────────────────────────────────────────────
  const [rooms, setRooms] = useState<Room[]>([]);
  const [roomsLoading, setRoomsLoading] = useState(false);
  const [roomsError, setRoomsError] = useState<string | null>(null);

  const [availabilityMap, setAvailabilityMap] = useState<Record<string, RoomAvailability>>({});
  const [availabilityLoading, setAvailabilityLoading] = useState(false);
  const [availabilityError, setAvailabilityError] = useState<string | null>(null);
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  // ─────────────────────────────────────────────────────────────────────────
  // STATE: Bookings
  // ─────────────────────────────────────────────────────────────────────────
  const [myBookings, setMyBookings] = useState<BookingResponse[]>([]);
  const [bookingsLoading, setBookingsLoading] = useState(false);
  const [bookingsError, setBookingsError] = useState<string | null>(null);

  const [pastBookings, setPastBookings] = useState<BookingResponse[]>([]);
  const [pastBookingsLoading, setPastBookingsLoading] = useState(false);
  const [pastYear, setPastYear] = useState<string>(new Date().getFullYear().toString());

  const [allBookings, setAllBookings] = useState<BookingResponse[]>([]);
  const [allBookingsLoading, setAllBookingsLoading] = useState(false);
  const [allBookingsError, setAllBookingsError] = useState<string | null>(null);

  // ─────────────────────────────────────────────────────────────────────────
  // STATE: Dialogs & Selection
  // ─────────────────────────────────────────────────────────────────────────
  const [bookingDialogOpen, setBookingDialogOpen] = useState(false);
  const [cancelDialogOpen, setCancelDialogOpen] = useState(false);
  const [selectedRoom, setSelectedRoom] = useState<Room | null>(null);
  const [selectedBooking, setSelectedBooking] = useState<BookingResponse | null>(null);

  // ─────────────────────────────────────────────────────────────────────────
  // STATE: Payment Processing
  // ─────────────────────────────────────────────────────────────────────────
  const [isProcessingPayment, setIsProcessingPayment] = useState(false);
  const [pendingBookingId, setPendingBookingId] = useState<string | null>(null);

  // ─────────────────────────────────────────────────────────────────────────
  // STATE: Booking Form
  // ─────────────────────────────────────────────────────────────────────────
  const [bookingFor, setBookingFor] = useState<'self' | 'guest'>('self');
  const [bedCount, setBedCount] = useState<number>(1);
  const [bookerPhone, setBookerPhone] = useState(INITIAL_PHONE_PREFIX);
  const [guestDetails, setGuestDetails] = useState<GuestDetail[]>([
    { name: '', age: 0, contact: '' },
  ]);
  const [bookingGender, setBookingGender] = useState<'male' | 'female'>('male');

  // ─────────────────────────────────────────────────────────────────────────
  // COMPUTED VALUES
  // ─────────────────────────────────────────────────────────────────────────
  const today = startOfDay(new Date());
  // 🔥 FIX: Max checkout is 10 days from the chosen check-in (not from today),
  // so a user who picks a check-in late in the month can still select a full 10-night stay.
  const maxCheckoutDate = addDays(today, BOOKING_DAYS_LIMIT);
  const bannerImage = `${API_BASE}/images/taga-towers.jpg`;
  const advanceAmount = ADVANCE_AMOUNTS[bookingFor];

  // ─────────────────────────────────────────────────────────────────────────
  // EVENT HANDLERS: Phone Input
  // ─────────────────────────────────────────────────────────────────────────

  const handlePhoneChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    let value = e.target.value;
    // Allow only digits and '+'
    value = value.replace(/[^\d\+]/g, '');
    // Prevent multiple '+'
    const plusCount = (value.match(/\+/g) || []).length;
    if (plusCount > 1) {
      value = value.replace(/\+/g, '');
    }
    // Limit length to 15 characters
    if (value.length <= 15) {
      setBookerPhone(value);
    }
  }, []);

  // ─────────────────────────────────────────────────────────────────────────
  // EVENT HANDLERS: Date Validation
  // ─────────────────────────────────────────────────────────────────────────

  const disabledDates = useCallback(
    (date: Date) => {
      const day = startOfDay(date);
      if (isBefore(day, today)) return true;

      // The absolute furthest date anyone can interact with (either check-in or check-out) is strictly today + 10 days
      const absoluteMaxDate = addDays(today, BOOKING_DAYS_LIMIT);

      if (checkInDate) {
        const checkInStart = startOfDay(checkInDate);
        // Check-out date must be on or after check-in, and cannot exceed today + 10 days
        return isBefore(day, checkInStart) || isAfter(day, absoluteMaxDate);
      }

      // If no check-in is selected yet, check-in date must be within 10 days from today
      return isAfter(day, absoluteMaxDate);
    },
    [today, checkInDate]
  );

  // ─────────────────────────────────────────────────────────────────────────
  // FETCH: Rooms
  // ─────────────────────────────────────────────────────────────────────────

  useEffect(() => {
    setRoomsLoading(true);
    setRoomsError(null);
    getAllRooms()
      .then(setRooms)
      .catch((error) => {
        console.error('Failed to load rooms:', error);
        setRoomsError('Failed to load rooms. Please refresh the page.');
        toast.error('Failed to load rooms');
      })
      .finally(() => setRoomsLoading(false));
  }, []);

  // ─────────────────────────────────────────────────────────────────────────
  // FETCH: Availability
  // ─────────────────────────────────────────────────────────────────────────

  useEffect(() => {
    if (!checkInDate || !checkOutDate || rooms.length === 0) return;

    // Stale-result guard — if deps change mid-flight, discard results from old run
    let cancelled = false;

    const checkInStr = format(checkInDate, 'yyyy-MM-dd');
    const checkOutStr = format(checkOutDate, 'yyyy-MM-dd');

    setAvailabilityLoading(true);
    setAvailabilityError(null);
    setAvailabilityMap({});

    // ── PRIMARY: Bulk endpoint (1 round trip for all rooms × all days) ──────
    checkAvailabilityRange(checkInStr, checkOutStr)
      .then((bulkResult) => {
        if (cancelled) return;
        // Bulk result is roomId → RoomAvailability — use directly
        setAvailabilityMap(bulkResult);
      })
      .catch(async (bulkErr) => {
        // ── FALLBACK: Old endpoint returned error (old backend / network issue) ──
        // Per-room parallel fetch — all dates in parallel per room
        console.warn('Bulk availability endpoint failed, falling back to per-room fetch:', bulkErr);

        const dates: string[] = [];
        for (let i = 0; i < BOOKING_DAYS_LIMIT; i++) {
          const d = new Date(checkInDate);
          d.setDate(d.getDate() + i);
          if (d >= checkOutDate) break;
          dates.push(format(d, 'yyyy-MM-dd'));
        }

        try {
          const results = await Promise.all(
            rooms.map(async (room) => {
              try {
                const dayResults = await Promise.all(
                  dates.map((date) => checkAvailability(room.id, date))
                );
                let finalAvailable = true;
                let minBeds = room.capacity;
                let genderRestriction: string | undefined;
                for (const avail of dayResults) {
                  if (!avail.available) finalAvailable = false;
                  minBeds = Math.min(minBeds, avail.availableBeds);
                  if (avail.genderRestriction) genderRestriction = avail.genderRestriction;
                }
                return {
                  [room.id]: { room, available: finalAvailable, availableBeds: minBeds, genderRestriction },
                };
              } catch {
                // Individual room fetch failed — show as available (don't block booking)
                return {
                  [room.id]: { room, available: true, availableBeds: room.capacity },
                };
              }
            })
          );
          if (!cancelled) {
            const map: Record<string, RoomAvailability> = Object.assign({}, ...results);
            setAvailabilityMap(map);
          }
        } catch (fallbackErr) {
          if (cancelled) return;
          console.error('Fallback availability fetch also failed:', fallbackErr);
          setAvailabilityError('Failed to load room availability. Please refresh.');
          toast.error('Failed to load room availability');
        }
      })
      .finally(() => {
        if (!cancelled) setAvailabilityLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [checkInDate, checkOutDate, rooms, refreshTrigger]);

  // ─────────────────────────────────────────────────────────────────────────
  // FETCH: My Bookings
  // ─────────────────────────────────────────────────────────────────────────

  const fetchMyBookings = useCallback(() => {
    if (!BOOKER_ID) return;
    setBookingsLoading(true);
    setBookingsError(null);
    getUserBookings(BOOKER_ID)
      .then((data: BookingResponse[]) => setMyBookings(data))
      .catch((error) => {
        console.error('Failed to load bookings:', error);
        setBookingsError('Failed to load your bookings');
        toast.error('Failed to load bookings');
      })
      .finally(() => setBookingsLoading(false));
  }, [BOOKER_ID]);

  useEffect(() => {
    if (isPaidMember) fetchMyBookings();
  }, [isPaidMember, BOOKER_ID, fetchMyBookings]);

  const fetchPastBookings = useCallback(() => {
    if (!BOOKER_ID || !pastYear) return;
    setPastBookingsLoading(true);
    getPastUserBookings(BOOKER_ID, pastYear)
      .then((data: BookingResponse[]) => setPastBookings(data))
      .catch((error) => {
        console.error('Failed to load past bookings:', error);
        toast.error('Failed to load past bookings');
      })
      .finally(() => setPastBookingsLoading(false));
  }, [BOOKER_ID, pastYear]);

  useEffect(() => {
    if (isPaidMember) fetchPastBookings();
  }, [isPaidMember, fetchPastBookings]);

  // ─────────────────────────────────────────────────────────────────────────
  // FETCH: All Bookings (Admin)
  // ─────────────────────────────────────────────────────────────────────────

  const fetchAllBookings = useCallback(() => {
    if (!isAdmin) return;
    setAllBookingsLoading(true);
    setAllBookingsError(null);
    fetch(`${API_BASE}/towers/admin/bookings`)
      .then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.json();
      })
      .then((data: BookingResponse[]) => setAllBookings(data || []))
      .catch((error) => {
        console.error('Failed to load occupancy data:', error);
        setAllBookingsError('Failed to load occupancy data');
        toast.error('Failed to load occupancy data');
      })
      .finally(() => setAllBookingsLoading(false));
  }, [isAdmin]);

  useEffect(() => {
    if (isAdmin) fetchAllBookings();
  }, [isAdmin, fetchAllBookings]);

  // ─────────────────────────────────────────────────────────────────────────
  // HANDLERS: Room Booking
  // ─────────────────────────────────────────────────────────────────────────

  const handleBookRoom = useCallback(
    (room: Room) => {
      if (!isPaidMember) {
        toast.error('Only paid subscribers can book rooms');
        return;
      }
      setSelectedRoom(room);
      setBedCount(1);
      setGuestDetails([{ name: '', age: 0, contact: '' }]);
      setBookingFor('self');
      setBookerPhone(INITIAL_PHONE_PREFIX);
      setBookingDialogOpen(true);
    },
    [isPaidMember]
  );

  const addGuestDetail = useCallback(() => {
    setGuestDetails((prev) => [...prev, { name: '', age: 0, contact: '' }]);
  }, []);

  const updateGuestDetail = useCallback(
    (index: number, field: keyof GuestDetail, value: string | number) => {
      setGuestDetails((prev) => {
        const updated = [...prev];
        updated[index] = { ...updated[index], [field]: value };
        return updated;
      });
    },
    []
  );

  const refreshAvailability = useCallback(() => {
    setRefreshTrigger((prev) => prev + 1);
  }, []);

  // ─────────────────────────────────────────────────────────────────────────
  // HANDLERS: Booking Confirmation & Payment
  // ─────────────────────────────────────────────────────────────────────────

  const confirmBooking = async () => {
    // Validation: Dates
    if (!selectedRoom || !checkInDate || !checkOutDate) {
      toast.error('Please select valid dates');
      return;
    }

    if (checkOutDate <= checkInDate) {
      toast.error('Check-out must be after check-in (minimum 1 day stay)');
      return;
    }

    // Validation: Phone
    const phoneValidation = validateMobileNumber(bookerPhone);
    if (!phoneValidation.isValid) {
      toast.error(phoneValidation.error || 'Invalid phone number');
      return;
    }

    // Validation: Guest Details
    if (bookingFor === 'guest') {
      if (guestDetails.some((g) => !g.name || !g.age || !g.contact)) {
        toast.error('Please fill in all guest details');
        return;
      }
      // 🔥 NEW: Validate guest contact numbers
      for (const guest of guestDetails) {
        const guestPhoneValidation = validateMobileNumber(guest.contact);
        if (!guestPhoneValidation.isValid) {
          toast.error(`Invalid phone number for guest: ${guest.name}`);
          return;
        }
      }
    }

    try {
      const finalGender = 
        selectedRoom.type === 'gents-dorm' ? 'male' :
        selectedRoom.type === 'ladies-dorm' ? 'female' :
        bookingGender;

      const response = await createBooking(
        {
          roomId: selectedRoom.id,
          checkInDate: format(checkInDate, 'yyyy-MM-dd'),
          checkOutDate: format(checkOutDate, 'yyyy-MM-dd'),
          bookerPhone,
          bookingFor,
          bedCount: selectedRoom.allowSingleBed ? bedCount : selectedRoom.capacity,
          ...(selectedRoom.allowSingleBed && bedCount < selectedRoom.capacity && { gender: finalGender }),
          ...(bookingFor === 'guest' && { guestDetails }),
        },
        BOOKER_ID
      );

      setPendingBookingId(response.id);
      setBookingDialogOpen(false);
      await openRazorpay(response.id);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Booking failed';

      // 🔥 Enhanced error messages
      if (message.includes('only') && message.includes('guests')) {
        toast.error(message);
      } else if (message.includes('not enough beds')) {
        toast.error('Room is full for selected dates');
      } else if (message.includes('already booked')) {
        toast.error('Room already booked for these dates');
      } else if (message.includes('past')) {
        toast.error('Cannot book dates in the past');
      } else {
        toast.error('Booking failed. Please try again.');
      }

      console.error('Booking error:', err);
    }
  };

  const openRazorpay = async (bookingId: string) => {
    try {
      setIsProcessingPayment(true);
      const loaded = await loadRazorpayScript();
      if (!loaded || !(window as any).Razorpay) {
        toast.error('Payment system failed to load. Please try again.');
        setIsProcessingPayment(false);
        return;
      }
      const amount = advanceAmount * 100; // ₹ → paise

      // Prepare notes for Razorpay (for admin email notification)
      const guestDetailsJSON = guestDetails.length > 0 ? JSON.stringify(guestDetails) : "";

      const roomNumber = selectedRoom ? ROOM_NUMBER_MAP[selectedRoom.id] || 0 : 0;

      const notes = {
        room_name: selectedRoom?.name || "",
        room_number: roomNumber.toString(),
        bed_count: selectedRoom?.allowSingleBed ? bedCount.toString() : (selectedRoom?.capacity || 0).toString(),
        check_in: checkInDate ? format(checkInDate, 'yyyy-MM-dd') : "",
        check_out: checkOutDate ? format(checkOutDate, 'yyyy-MM-dd') : "",
        booker_name: BOOKER_NAME,
        booker_taga_id: BOOKER_ID,
        booker_phone: bookerPhone.replace(/\s+/g, ''),
        booker_email: loggedInUser.email || "",
        booking_type: bookingFor,
        guest_details: guestDetailsJSON,
      };

      const { key, order } = await createOrder(amount, notes);

      const options = {
        key,
        amount: order.amount,
        currency: order.currency,
        name: 'TAGA Towers',
        description: 'Room Booking Advance',
        order_id: order.id,
        // image: `${import.meta.env.VITE_API_BASE_URL}/images/logo/logo1.jpg`,
        prefill: {
          name: BOOKER_NAME,
          email: loggedInUser.email || '',
          contact: bookerPhone.replace(/\s+/g, '') || '',
        },
        method: {
          upi: true,
          card: true,
          netbanking: true,
          wallet: true
        },
        theme: {
          color: '#297248',
        },
        handler: async function (response: any) {
          try {
            const verifyRes = await fetch(`${API_BASE}/towers/verify-payment`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                bookingId,
                razorpay_order_id: response.razorpay_order_id,
                razorpay_payment_id: response.razorpay_payment_id,
                razorpay_signature: response.razorpay_signature,
              }),
            });

            if (!verifyRes.ok) {
              throw new Error('Payment verification failed');
            }

            const roomName = selectedRoom?.name || 'Room';
            toast.success(`🎉 Booking Successful! Your reservation for ${roomName} has been confirmed.`);
            setPendingBookingId(null);
            setSelectedRoom(null);

            // 🔥 FIX: Reset booking form
            setGuestDetails([{ name: '', age: 0, contact: '' }]);
            setBookingFor('self');
            setBookerPhone(INITIAL_PHONE_PREFIX);

            fetchMyBookings();
            if (isAdmin) fetchAllBookings();
            refreshAvailability();
          } catch (error) {
            console.error('Payment verification error:', error);
            toast.error('Payment verification failed. Please contact support.');
          }
        },
        modal: {
          ondismiss: async function () {
            try {
              if (bookingId) {
                await cancelBooking(bookingId);
                fetchMyBookings();
                refreshAvailability();
              }
            } catch (error) {
              console.error('Booking cancellation error:', error);
            }
            setPendingBookingId(null);
            toast.error('Payment cancelled. Booking removed.');
          },
        },
      };

      const rzp = new (window as any).Razorpay(options);
      rzp.open();
    } catch (err: any) {
      console.error('Razorpay error:', err);
      toast.error(err.message || 'Payment system error');
    } finally {
      setIsProcessingPayment(false);
    }
  };

  // ─────────────────────────────────────────────────────────────────────────
  // HANDLERS: Booking Cancellation
  // ─────────────────────────────────────────────────────────────────────────

  const canCancelBooking = useCallback((booking: BookingResponse): boolean => {
    const todayDate = startOfDay(new Date());
    const checkInDateObj = new Date(booking.checkInDate);
    return todayDate < checkInDateObj;
  }, []);

  const handleCancelBooking = useCallback(
    (booking: BookingResponse) => {
      if (!canCancelBooking(booking)) {
        toast.error('Cannot cancel booking on or after check-in date');
        return;
      }
      setSelectedBooking(booking);
      setCancelDialogOpen(true);
    },
    [canCancelBooking]
  );

  const confirmCancellation = async () => {
    if (!selectedBooking) return;

    // Double-check eligibility
    if (!canCancelBooking(selectedBooking)) {
      toast.error('Cannot cancel booking on or after check-in date');
      setCancelDialogOpen(false);
      return;
    }

    try {
      // Instant UI update
      setMyBookings((prev) =>
        prev.map((b) =>
          b.id === selectedBooking.id
            ? { ...b, paymentStatus: 'cancelled', bookingStatus: 'cancelled' }
            : b
        )
      );

      await cancelBooking(selectedBooking.id);

      setCancelDialogOpen(false);
      setSelectedBooking(null);
      toast.success('Booking cancelled successfully.');

      fetchMyBookings();
      if (isAdmin) fetchAllBookings();
      refreshAvailability();
    } catch (err) {
      console.error('Cancellation error:', err);
      toast.error('Failed to cancel booking. Please try again.');
      // Revert UI on error
      fetchMyBookings();
    }
  };

  // ─────────────────────────────────────────────────────────────────────────
  // HELPERS: Occupancy Schedule (Admin)
  // ─────────────────────────────────────────────────────────────────────────

  const getOccupancyForDate = useCallback(
    (date: Date) => {
      const dateStr = format(date, 'yyyy-MM-dd');
      return allBookings.filter((b) => {
        if (b.paymentStatus === 'cancelled' || b.paymentStatus === 'refunded') return false;
        return b.checkInDate <= dateStr && b.checkOutDate > dateStr;
      });
    },
    [allBookings]
  );

  // ─────────────────────────────────────────────────────────────────────────
  // RENDER: Locked Page (Non-Paid Member)
  // ─────────────────────────────────────────────────────────────────────────

  if (!isPaidMember) {
    return (
      <div className="space-y-6" data-testid="testid-taga-towers-locked-page">
        <div className="relative overflow-hidden rounded-2xl shadow-2xl">
          <div className="absolute inset-0">
            <img
              src={bannerImage}
              alt="TAGA Towers"
              className="w-full h-full object-cover"
            />
            <div className="absolute inset-0 bg-gradient-to-r from-green-900/95 via-green-800/90 to-green-900/80" />
          </div>
          <div className="relative p-8 md:p-12">
            <div className="max-w-3xl">
              <Badge className="mb-4 bg-green-600 text-white">
                <Building2 className="w-3 h-3 mr-1" />
                For Paid Members Only
              </Badge>
              <h1 className="text-4xl md:text-5xl font-bold text-white mb-4">TAGA Towers</h1>
              <p className="text-lg text-green-50">
                Comfortable accommodation facility for TAGA members in Chennai
              </p>
            </div>
          </div>
        </div>

        <Card className="max-w-2xl mx-auto bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800">
          <CardContent className="pt-6 text-center py-12">
            <AlertCircle className="w-16 h-16 mx-auto text-yellow-600 mb-4" />
            <h3 className="text-xl font-semibold text-yellow-900 dark:text-yellow-100 mb-2">
              Subscriber Access Only
            </h3>
            <p className="text-yellow-800 dark:text-yellow-200 mb-4">
              TAGA Towers booking is available only for paid subscribers. Please ensure your
              subscription is active to book rooms.
            </p>
            <Button data-testid="testid-manage-subscription-button">Manage Subscription</Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // ─────────────────────────────────────────────────────────────────────────
  // RENDER: Admin Dashboard
  // ─────────────────────────────────────────────────────────────────────────

  if (isAdmin) {
    return (
      <div className="space-y-6" data-testid="testid-taga-towers-page">
        {/* Hero Banner */}
        <div className="relative overflow-hidden rounded-2xl shadow-2xl">
          <div className="absolute inset-0">
            <img
              src={bannerImage}
              alt="TAGA Towers Building"
              className="w-full h-full object-cover"
            />
            <div className="absolute inset-0 bg-gradient-to-r from-green-900/95 via-green-800/85 to-green-900/90" />
          </div>
          <div className="relative p-12 text-white">
            <div className="text-center space-y-4">
              <div className="flex items-center justify-center space-x-3">
                <div className="p-3 bg-white/10 rounded-full backdrop-blur-sm">
                  <ShieldCheck className="w-12 h-12" />
                </div>
              </div>
              <h1 className="text-5xl font-bold">Admin Dashboard</h1>
              <p className="text-xl text-green-50 max-w-3xl mx-auto">
                TAGA Towers Occupancy & Management
              </p>
            </div>
          </div>
        </div>

        {/* Occupancy Schedule */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <ShieldCheck className="w-5 h-5 text-primary" />
              Occupancy Schedule — Next {BOOKING_DAYS_LIMIT} Days
            </CardTitle>
            <CardDescription>
              Live room occupancy for the upcoming {BOOKING_DAYS_LIMIT} days — confirmed bookings
              only
            </CardDescription>
          </CardHeader>
          <CardContent>
            {allBookingsError && (
              <Alert className="mb-4 border-red-200 bg-red-50">
                <AlertCircle className="w-4 h-4 text-red-600" />
                <AlertDescription className="text-red-800">{allBookingsError}</AlertDescription>
              </Alert>
            )}
            {allBookingsLoading ? (
              <div className="text-center py-12">
                <Loader2 className="w-8 h-8 animate-spin mx-auto text-muted-foreground mb-2" />
                <p className="text-muted-foreground">Loading occupancy data...</p>
              </div>
            ) : (
              <div className="space-y-6">
                {Array.from({ length: BOOKING_DAYS_LIMIT }).map((_, i) => {
                  const date = addDays(today, i);
                  const dateStr = format(date, 'yyyy-MM-dd');
                  const dayBookings = getOccupancyForDate(date);
                  const occupiedBeds = dayBookings.reduce((sum, b) => sum + (b.bedCount || 0), 0);
                  const occupancyPercentage = Math.min(
                    100,
                    Math.round((occupiedBeds / TOTAL_BED_CAPACITY) * 100)
                  );
                  const byRoom: Record<string, BookingResponse[]> = {};
                  dayBookings.forEach((b) => {
                    if (!byRoom[b.roomId]) byRoom[b.roomId] = [];
                    byRoom[b.roomId].push(b);
                  });

                  return (
                    <div key={i} className="space-y-3">
                      <div className="flex items-center justify-between">
                        <div>
                          <h3 className="text-lg font-semibold">
                            {i === 0 ? 'Today — ' : ''}
                            {format(date, 'EEEE, MMMM dd, yyyy')}
                          </h3>
                          <p className="text-sm text-muted-foreground">
                            {occupiedBeds} beds occupied • {TOTAL_BED_CAPACITY - occupiedBeds} beds
                            free
                          </p>
                        </div>
                        <Badge
                          className={
                            occupancyPercentage >= 80
                              ? 'bg-red-600 text-white'
                              : occupancyPercentage >= 50
                                ? 'bg-yellow-500 text-white'
                                : 'bg-green-600 text-white'
                          }
                        >
                          {occupancyPercentage}% occupied
                        </Badge>
                      </div>
                      <div className="w-full bg-muted rounded-full h-2.5 overflow-hidden">
                        <div
                          className={`h-full transition-all rounded-full ${occupancyPercentage >= 80
                              ? 'bg-red-500'
                              : occupancyPercentage >= 50
                                ? 'bg-yellow-400'
                                : 'bg-green-600'
                            }`}
                          style={{ width: `${occupancyPercentage}%` }}
                        />
                      </div>
                      {Object.keys(byRoom).length > 0 ? (
                        <div className="ml-1 grid grid-cols-1 md:grid-cols-2 gap-2">
                          {Object.entries(byRoom).map(([roomId, bookings]) => {
                            const roomName =
                              rooms.find((r) => r.id === roomId)?.name || roomId;
                            const roomNo = ROOM_NUMBER_MAP[roomId];
                            return (
                              <div
                                key={roomId}
                                className="text-sm bg-muted/50 rounded-lg px-3 py-2 space-y-1"
                              >
                                <p className="font-medium text-foreground">
                                  {roomNo ? `Room ${roomNo} — ` : ''}
                                  {roomName}
                                </p>
                                {bookings.map((b, idx) => (
                                  <p key={idx} className="text-muted-foreground text-xs">
                                    {b.bookerName}
                                    {b.bookerId ? ` (${b.bookerId})` : ''}
                                    {' · '}
                                    {b.bedCount} bed{b.bedCount > 1 ? 's' : ''}
                                    {b.bookingFor === 'guest' ? ' · Guest booking' : ''}
                                  </p>
                                ))}
                              </div>
                            );
                          })}
                        </div>
                      ) : (
                        <p className="text-sm text-muted-foreground ml-1">
                          No bookings for this day
                        </p>
                      )}
                      {i < BOOKING_DAYS_LIMIT - 1 && <Separator className="mt-2" />}
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    );
  }

  // ─────────────────────────────────────────────────────────────────────────
  // RENDER: Member View
  // ─────────────────────────────────────────────────────────────────────────

  return (
    <div className="space-y-6" data-testid="testid-taga-towers-page">
      {/* Hero Banner */}
      <div className="relative overflow-hidden rounded-2xl shadow-2xl">
        <div className="absolute inset-0">
          <img
            src={bannerImage}
            alt="TAGA Towers Building"
            className="w-full h-full object-cover object-center"
            style={{ objectPosition: "center 20%" }}
          />
          <div className="absolute inset-0 bg-gradient-to-r from-green-900/95 via-green-800/85 to-green-900/90" />
        </div>
        <div className="relative p-12 text-white">
          <div className="text-center space-y-4">
            <div className="flex items-center justify-center space-x-3">
              <div className="p-3 bg-white/10 rounded-full backdrop-blur-sm">
                <Building2 className="w-12 h-12" />
              </div>
            </div>
            <h1 className="text-5xl font-bold">TAGA Towers</h1>
            <p className="text-xl text-green-50 max-w-3xl mx-auto">
              Comfortable accommodation facility for TAGA subscribers in Chennai
            </p>
            <div className="flex flex-wrap items-center justify-center gap-6 text-sm text-green-100">
              <div className="flex items-center space-x-2 bg-white/10 backdrop-blur-sm px-4 py-2 rounded-lg">
                <MapPin className="w-4 h-4" />
                <span>Chennai Location</span>
              </div>
              <div className="flex items-center space-x-2 bg-white/10 backdrop-blur-sm px-4 py-2 rounded-lg">
                <Clock className="w-4 h-4" />
                <span>Book up to {BOOKING_DAYS_LIMIT} days ahead</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <Tabs defaultValue="room-availability" className="w-full">
        <TabsList className="grid w-full max-w-md mx-auto mb-6 grid-cols-1">
          <TabsTrigger value="room-availability">Room Availability</TabsTrigger>
        </TabsList>

        <TabsContent value="room-availability" className="space-y-6">
          {/* Date Selection + Room Availability */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Date Selection Card */}
            <Card data-testid="testid-taga-towers-date-selection-card">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <CalendarIcon className="w-5 h-5" />
                  <span>Select Date</span>
                </CardTitle>
                <CardDescription>Choose your check-in and check-out dates</CardDescription>
              </CardHeader>
              <CardContent>
                <Calendar
                  mode="range"
                  data-testid="testid-room-date-range-calendar"
                  selected={{ from: checkInDate, to: checkOutDate }}
                  // 🔥 FIX: Auto-navigate to the check-in month so cross-month
                  // 🔥 FIX: Controlled month navigation so developers and users
                  // can select dates across multiple months (like Aug 31 to Sep 8)
                  // while still centering on checkInDate when it's selected.
                  month={calendarMonth}
                  onMonthChange={setCalendarMonth}
                  onSelect={(range: any) => {
                    const from = range?.from;
                    let to = range?.to;
                    const maxAllowed = addDays(today, BOOKING_DAYS_LIMIT);

                    if (from && to && to <= from) {
                      to = addDays(from, MIN_STAY_DAYS);
                    }

                    // 🔥 FIX: Programmatic validation guard to ensure check-out date
                    // never exceeds today + 10 days limit under any circumstances.
                    if (to && isAfter(startOfDay(to), maxAllowed)) {
                      to = maxAllowed;
                    }

                    // 🔥 FIX: If check-in is the absolute max date itself, block it
                    // because a minimum stay of 1 night would require checking out on day 11.
                    if (from && (isAfter(startOfDay(from), maxAllowed) || startOfDay(from).getTime() === maxAllowed.getTime())) {
                      toast.error("Cannot check in on this date. Bookings are capped at 10 days from today.");
                      return;
                    }

                    setCheckInDate(from);
                    setCheckOutDate(to);
                    if (from) {
                      setCalendarMonth(from);
                    }
                  }}
                  disabled={disabledDates}
                  className="rounded-md border"
                />
                {checkInDate && checkOutDate && (
                  <div className="mt-4 p-3 bg-primary/10 rounded-lg">
                    <p className="text-sm font-medium">Selected Dates</p>
                    <p className="text-lg font-bold text-primary">
                      {format(checkInDate, 'MMM dd, yyyy')} → {format(checkOutDate, 'MMM dd, yyyy')}
                    </p>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Room Availability Card */}
            <Card className="lg:col-span-2" data-testid="testid-taga-towers-room-availability-card">
              <CardHeader>
                <CardTitle>Available Rooms</CardTitle>
                <CardDescription>
                  {checkInDate
                    ? `Showing availability from ${format(checkInDate, 'MMM dd')} to ${format(
                      checkOutDate || checkInDate,
                      'MMM dd, yyyy'
                    )}`
                    : 'Select a date to view available rooms'}
                </CardDescription>
              </CardHeader>
              <CardContent>
                {availabilityError && (
                  <Alert className="mb-4 border-red-200 bg-red-50">
                    <AlertCircle className="w-4 h-4 text-red-600" />
                    <AlertDescription className="text-red-800">{availabilityError}</AlertDescription>
                  </Alert>
                )}
                {roomsError && (
                  <Alert className="mb-4 border-red-200 bg-red-50">
                    <AlertCircle className="w-4 h-4 text-red-600" />
                    <AlertDescription className="text-red-800">{roomsError}</AlertDescription>
                  </Alert>
                )}
                {checkInDate ? (
                  roomsLoading || availabilityLoading ? (
                    <div className="text-center py-12">
                      <Loader2 className="w-8 h-8 animate-spin mx-auto text-muted-foreground mb-2" />
                      <p className="text-muted-foreground">Loading rooms...</p>
                    </div>
                  ) : (
                    <ScrollArea className="h-[400px] pr-4">
                      <div className="space-y-4">
                        {rooms.map((room) => {
                          const avail = availabilityMap[room.id];
                          const isAvailable = avail?.available ?? false;
                          const availableBeds = avail?.availableBeds ?? 0;
                          const genderRestriction = avail?.genderRestriction;
                          return (
                            <Card
                              key={room.id}
                              data-testid={`testid-room-${room.id}-card`}
                              className={`transition-all ${isAvailable ? 'hover:shadow-lg hover:border-primary/50' : 'opacity-60'
                                }`}
                            >
                              <CardContent className="pt-6">
                                <div className="flex items-start justify-between mb-3">
                                  <div>
                                    <h3 className="font-bold text-lg">{room.name}</h3>
                                    <p className="text-sm text-muted-foreground">
                                      {getRoomTypeLabel(room.type)}
                                    </p>
                                  </div>
                                  <Badge
                                    className={
                                      !avail
                                        ? 'bg-gray-400'
                                        : !avail.available
                                          ? 'bg-red-600'
                                          : avail.availableBeds < room.capacity
                                            ? 'bg-yellow-500'
                                            : 'bg-green-600'
                                    }
                                  >
                                    {!avail
                                      ? 'Loading'
                                      : !avail.available
                                        ? 'Full'
                                        : avail.availableBeds < room.capacity
                                          ? `${avail.availableBeds} beds left`
                                          : 'Available'}
                                  </Badge>
                                </div>
                                <div className="space-y-2 text-sm text-muted-foreground mb-3">
                                  <div className="flex items-center">
                                    <Users className="w-4 h-4 mr-2" />
                                    <span>Capacity: {room.capacity} persons</span>
                                  </div>
                                  {isAvailable && (
                                    <div className="flex items-center text-green-600">
                                      <BedDouble className="w-4 h-4 mr-2" />
                                      <span>{availableBeds} bed(s) available</span>
                                    </div>
                                  )}
                                  {genderRestriction && (
                                    <Alert className="mt-2">
                                      <AlertCircle className="w-4 h-4" />
                                      <AlertDescription className="text-xs">
                                        Partially booked — {genderRestriction} guests only
                                      </AlertDescription>
                                    </Alert>
                                  )}
                                </div>
                                <Button
                                  className="w-full"
                                  size="sm"
                                  data-testid={`testid-room-${room.id}-book-button`}
                                  disabled={!avail?.available || isProcessingPayment}
                                  onClick={() => handleBookRoom(room)}
                                >
                                  {isAvailable ? 'Book Now' : 'Not Available'}
                                </Button>
                              </CardContent>
                            </Card>
                          );
                        })}
                      </div>
                    </ScrollArea>
                  )
                ) : (
                  <div className="text-center py-12">
                    <CalendarIcon className="w-16 h-16 mx-auto text-muted-foreground mb-4" />
                    <p className="text-muted-foreground">
                      Please select a date to view room availability
                    </p>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* My Bookings */}
          <Card data-testid="testid-my-bookings-card">
            <CardHeader>
              <CardTitle>My Bookings</CardTitle>
              <CardDescription>View and manage your room bookings</CardDescription>
            </CardHeader>
            <CardContent>
              {bookingsError && (
                <Alert className="mb-4 border-red-200 bg-red-50">
                  <AlertCircle className="w-4 h-4 text-red-600" />
                  <AlertDescription className="text-red-800">{bookingsError}</AlertDescription>
                </Alert>
              )}
              {bookingsLoading ? (
                <div className="text-center py-8">
                  <Loader2 className="w-8 h-8 animate-spin mx-auto text-muted-foreground mb-2" />
                  <p className="text-muted-foreground">Loading bookings...</p>
                </div>
              ) : myBookings.length > 0 ? (
                <Tabs defaultValue="active-upcoming" className="w-full">
                  <TabsList className="grid w-full max-w-md grid-cols-2">
                    <TabsTrigger value="active-upcoming">Active & Upcoming</TabsTrigger>
                    <TabsTrigger value="past">Past Bookings</TabsTrigger>
                  </TabsList>

                  {/* Active & Upcoming Tab */}
                  <TabsContent value="active-upcoming" className="space-y-4 mt-4">
                    {(() => {
                      const filtered = myBookings.filter((b) => {
                        const status = getEffectiveBookingStatus(b);
                        return status === 'upcoming' || status === 'active';
                      });
                      if (filtered.length === 0) {
                        return (
                          <div className="text-center py-8 text-muted-foreground">
                            No active or upcoming bookings
                          </div>
                        );
                      }
                      return (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                          {filtered.map((booking) => {
                            const status = getEffectiveBookingStatus(booking);
                            const isCancellable = canCancelBooking(booking);
                            const statusConfig = getBookingStatusConfig(status);

                            return (
                              <Card key={booking.id} className="border-2">
                                <CardContent className="pt-4">
                                  <div className="space-y-3">
                                    <div>
                                      <div className="flex items-center justify-between">
                                        <h4 className="font-semibold">{booking.roomName}</h4>
                                        <Badge className={statusConfig.badgeClass}>
                                          {statusConfig.icon} {statusConfig.label}
                                        </Badge>
                                      </div>
                                      <p className="text-sm text-muted-foreground mt-1">
                                        {format(new Date(booking.checkInDate), 'EEE, MMM dd, yyyy')}
                                        {' → '}
                                        {format(new Date(booking.checkOutDate), 'MMM dd, yyyy')}
                                      </p>
                                    </div>
                                    <div className="flex items-center justify-between text-sm">
                                      <span className="text-muted-foreground">
                                        {booking.bookingFor === 'self' ? 'Self' : 'Guest'} •{' '}
                                        {booking.bedCount} bed(s)
                                      </span>
                                    </div>
                                    {booking.paymentStatus !== 'cancelled' &&
                                      booking.paymentStatus !== 'refunded' &&
                                      isCancellable && (
                                        <Button
                                          size="sm"
                                          variant="destructive"
                                          className="w-full"
                                          data-testid={`testid-booking-${booking.id}-cancel-button`}
                                          onClick={() => handleCancelBooking(booking)}
                                        >
                                          Cancel Booking
                                        </Button>
                                      )}
                                    {!isCancellable && status === 'active' && (
                                      <p className="text-xs text-center text-muted-foreground">
                                        Cannot cancel after check-in
                                      </p>
                                    )}
                                  </div>
                                </CardContent>
                              </Card>
                            );
                          })}
                        </div>
                      );
                    })()}
                  </TabsContent>

                  {/* Past Bookings Tab */}
                  <TabsContent value="past" className="space-y-4 mt-4">
                    <div className="flex items-center gap-2 mb-4">
                      <Select value={pastYear} onValueChange={setPastYear}>
                        <SelectTrigger className="w-[120px]">
                          <SelectValue placeholder="Year" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value={(new Date().getFullYear()).toString()}>{new Date().getFullYear()}</SelectItem>
                          <SelectItem value={(new Date().getFullYear() - 1).toString()}>{new Date().getFullYear() - 1}</SelectItem>
                        </SelectContent>
                      </Select>
                      {pastBookingsLoading && <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />}
                    </div>

                    {(() => {
                      const combined = [...myBookings, ...pastBookings];
                      // unique by id
                      const uniqueMap = new Map();
                      combined.forEach(b => uniqueMap.set(b.id, b));
                      const unique = Array.from(uniqueMap.values());

                      const filtered = unique.filter((b) => {
                        const status = getEffectiveBookingStatus(b);
                        const matchesYear = b.checkOutDate.startsWith(pastYear);
                        return (status === 'completed' || status === 'cancelled') && matchesYear;
                      });
                      if (filtered.length === 0) {
                        return (
                          <div className="text-center py-8 text-muted-foreground">
                            No past bookings
                          </div>
                        );
                      }
                      return (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                          {filtered.map((booking) => {
                            const status = getEffectiveBookingStatus(booking);
                            const statusConfig = getBookingStatusConfig(status);

                            return (
                              <Card key={booking.id} className="border-2 opacity-75">
                                <CardContent className="pt-4">
                                  <div className="space-y-3">
                                    <div>
                                      <div className="flex items-center justify-between">
                                        <h4 className="font-semibold">{booking.roomName}</h4>
                                        <Badge className={statusConfig.badgeClass}>
                                          {statusConfig.icon} {statusConfig.label}
                                        </Badge>
                                      </div>
                                      <p className="text-sm text-muted-foreground mt-1">
                                        {format(new Date(booking.checkInDate), 'EEE, MMM dd, yyyy')}
                                        {' → '}
                                        {format(new Date(booking.checkOutDate), 'MMM dd, yyyy')}
                                      </p>
                                    </div>
                                    <div className="flex items-center justify-between text-sm">
                                      <span className="text-muted-foreground">
                                        {booking.bookingFor === 'self' ? 'Self' : 'Guest'} •{' '}
                                        {booking.bedCount} bed(s)
                                      </span>
                                    </div>
                                  </div>
                                </CardContent>
                              </Card>
                            );
                          })}
                        </div>
                      );
                    })()}
                  </TabsContent>
                </Tabs>
              ) : (
                <p className="text-center text-muted-foreground py-6">No bookings found</p>
              )}
            </CardContent>
          </Card>

          {/* Caretaker Contact */}
          <Card
            className="bg-gradient-to-r from-primary/5 to-primary/10"
            data-testid="testid-caretaker-contact-card"
          >
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Phone className="w-5 h-5" />
                <span>Caretaker Contact</span>
              </CardTitle>
              <CardDescription>For inquiries and assistance during your stay</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="flex items-start space-x-3">
                  <User className="w-5 h-5 text-primary mt-1 flex-shrink-0" />
                  <div className="space-y-1.5 w-full">
                    <p className="text-sm text-muted-foreground">Name</p>
                    {CARETAKER_INFO.caretakers.map((caretaker, idx) => (
                      <p key={idx} className="font-semibold h-6 flex items-center">{caretaker.name}</p>
                    ))}
                  </div>
                </div>
                <div className="flex items-start space-x-3">
                  <Phone className="w-5 h-5 text-primary mt-1 flex-shrink-0" />
                  <div className="space-y-1.5 w-full">
                    <p className="text-sm text-muted-foreground">Phone</p>
                    {CARETAKER_INFO.caretakers.map((caretaker, idx) => (
                      <a
                        key={idx}
                        href={`tel:${caretaker.phone.replace(/\s+/g, '')}`}
                        className="font-semibold text-sm hover:text-primary h-6 flex items-center"
                      >
                        {caretaker.phone}
                      </a>
                    ))}
                  </div>
                </div>
                <div className="flex items-start space-x-3">
                  <Mail className="w-5 h-5 text-primary mt-1 flex-shrink-0" />
                  <div className="space-y-1.5 w-full">
                    <p className="text-sm text-muted-foreground">Email</p>
                    <a
                      href={`mailto:${CARETAKER_INFO.email}`}
                      className="font-semibold hover:text-primary break-all h-6 flex items-center"
                    >
                      {CARETAKER_INFO.email}
                    </a>
                  </div>
                </div>
              </div>
              <Separator className="my-4" />
              <div className="flex items-start space-x-3">
                <MapPin className="w-5 h-5 text-primary mt-1 flex-shrink-0" />
                <div>
                  <p className="text-sm text-muted-foreground">Address</p>
                  <p className="font-semibold">{CARETAKER_INFO.address}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Booking Dialog */}
      <Dialog open={bookingDialogOpen} onOpenChange={(open: boolean) => setBookingDialogOpen(open)}>
        <DialogContent
          className="max-w-2xl max-h-[90vh] overflow-y-auto"
          data-testid="testid-room-booking-modal"
        >
          <DialogHeader>
            <DialogTitle>Book Room - {selectedRoom?.name}</DialogTitle>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {/* Booker Details */}
            <div className="space-y-3">
              <h4 className="font-semibold">Booker Details</h4>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label>Name</Label>
                  <Input value={BOOKER_NAME} disabled />
                </div>
                <div>
                  <Label>TAGA ID</Label>
                  <Input value={BOOKER_ID || '—'} disabled />
                </div>
                <div className="col-span-2">
                  <Label>Contact Number</Label>
                  <Input
                    value={bookerPhone}
                    data-testid="testid-booker-phone-input"
                    onChange={handlePhoneChange}
                    placeholder={PHONE_PLACEHOLDER}
                    aria-label="Booker phone number"
                  />
                  {bookerPhone && bookerPhone !== INITIAL_PHONE_PREFIX && (
                    <p className="text-xs text-muted-foreground mt-1">
                      {validateMobileNumber(bookerPhone).isValid ? (
                        <span className="text-green-600">✓ Valid number</span>
                      ) : (
                        <span className="text-red-600">✗ Invalid format</span>
                      )}
                    </p>
                  )}
                </div>
              </div>
            </div>

            <Separator />

            {/* Booking For */}
            <div className="space-y-3">
              <Label>Booking For</Label>
              <RadioGroup
                value={bookingFor}
                data-testid="testid-booking-for-radio-group"
                onValueChange={(value: string) => setBookingFor(value as 'self' | 'guest')}
              >
                <div className="flex items-center space-x-2">
                  <RadioGroupItem value="self" id="self" />
                  <Label htmlFor="self">Self (Advance: ₹{ADVANCE_AMOUNTS.self})</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <RadioGroupItem value="guest" id="guest" />
                  <Label htmlFor="guest">Guest (Advance: ₹{ADVANCE_AMOUNTS.guest})</Label>
                </div>
              </RadioGroup>
            </div>

            {/* Bed Count Selection */}
            {selectedRoom?.allowSingleBed && (
              <div className="space-y-3">
                <Label>Number of Beds</Label>
                <Select
                  value={bedCount.toString()}
                  onValueChange={(v: string) => setBedCount(Number(v))}
                >
                  <SelectTrigger data-testid="testid-bed-count-select">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {Array.from({ length: selectedRoom.capacity }, (_, i) => (
                      <SelectItem key={i + 1} value={(i + 1).toString()}>
                        {i + 1} bed{i + 1 > 1 ? 's' : ''}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            {/* Gender Selection */}
            {selectedRoom?.allowSingleBed &&
              bedCount < selectedRoom.capacity && (
                <div className="space-y-3">
                  <Label>Gender</Label>
                  
                  {selectedRoom.type === 'gents-dorm' ? (
                    <div className="inline-flex items-center px-3 py-1 rounded-md bg-blue-50 text-blue-700 font-medium text-sm border border-blue-200">
                      Male Only
                    </div>
                  ) : selectedRoom.type === 'ladies-dorm' ? (
                    <div className="inline-flex items-center px-3 py-1 rounded-md bg-pink-50 text-pink-700 font-medium text-sm border border-pink-200">
                      Female Only
                    </div>
                  ) : (
                    <RadioGroup
                      value={bookingGender}
                      data-testid="testid-booking-gender-radio-group"
                      onValueChange={(value: string) =>
                        setBookingGender(value as 'male' | 'female')
                      }
                    >
                      <div className="flex items-center space-x-2">
                        <RadioGroupItem value="male" id="male" />
                        <Label htmlFor="male">Male</Label>
                      </div>
                      <div className="flex items-center space-x-2">
                        <RadioGroupItem value="female" id="female" />
                        <Label htmlFor="female">Female</Label>
                      </div>
                    </RadioGroup>
                  )}
                </div>
              )}

            {/* Guest Details */}
            {bookingFor === 'guest' && (
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="font-semibold">Guest Details</h4>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={addGuestDetail}
                    data-testid="testid-add-guest-button"
                  >
                    Add Guest
                  </Button>
                </div>
                {guestDetails.map((guest, index) => (
                  <Card key={index} className="p-4" data-testid={`testid-guest-${index + 1}-details-card`}>
                    <div className="grid grid-cols-2 gap-3">
                      <div className="col-span-2">
                        <Label>Guest {index + 1} Name</Label>
                        <Input
                          value={guest.name}
                          data-testid={`testid-guest-${index + 1}-name-input`}
                          onChange={(e) => updateGuestDetail(index, 'name', e.target.value)}
                          placeholder="Enter guest name"
                          aria-label={`Guest ${index + 1} name`}
                        />
                      </div>
                      <div>
                        <Label>Age</Label>
                        <Input
                          type="number"
                          value={guest.age || ''}
                          data-testid={`testid-guest-${index + 1}-age-input`}
                          onChange={(e) =>
                            updateGuestDetail(index, 'age', Number(e.target.value))
                          }
                          placeholder="Age"
                          aria-label={`Guest ${index + 1} age`}
                        />
                      </div>
                      <div>
                        <Label>Contact Number</Label>
                        <Input
                          value={guest.contact}
                          data-testid={`testid-guest-${index + 1}-contact-input`}
                          onChange={(e) =>
                            updateGuestDetail(index, 'contact', e.target.value)
                          }
                          placeholder={PHONE_PLACEHOLDER}
                          aria-label={`Guest ${index + 1} contact`}
                        />
                      </div>
                    </div>
                  </Card>
                ))}
              </div>
            )}

            {/* Booking Summary */}
            <div className="bg-primary/5 p-4 rounded-lg space-y-2">
              <h4 className="font-semibold">Booking Summary</h4>
              <div className="text-sm space-y-1">
                <p>
                  <span className="text-muted-foreground">Room:</span>{' '}
                  <span className="font-medium">{selectedRoom?.name}</span>
                </p>
                <p>
                  <span className="text-muted-foreground">Check-in Date:</span>{' '}
                  <span className="font-medium">
                    {checkInDate && format(checkInDate, 'MMM dd, yyyy')}
                  </span>
                </p>
                <p>
                  <span className="text-muted-foreground">Check-out Date:</span>{' '}
                  <span className="font-medium">
                    {checkOutDate && format(checkOutDate, 'MMM dd, yyyy')}
                  </span>
                </p>
                <p>
                  <span className="text-muted-foreground">Beds:</span>{' '}
                  <span className="font-medium">
                    {selectedRoom?.allowSingleBed ? bedCount : selectedRoom?.capacity}
                  </span>
                </p>
                <p>
                  <span className="text-muted-foreground">For:</span>{' '}
                  <span className="font-medium capitalize">{bookingFor}</span>
                </p>
                <Separator className="my-2" />
                <p className="text-lg font-bold text-primary">
                  <span className="text-muted-foreground font-normal">Advance Payment:</span> ₹
                  {advanceAmount}
                </p>
                <p className="text-xs text-muted-foreground">
                  Remaining amount to be paid after stay
                </p>
              </div>
            </div>
          </div>

          <div className="flex gap-2 justify-end pt-4">
            <Button
              variant="outline"
              onClick={() => setBookingDialogOpen(false)}
              data-testid="testid-booking-modal-cancel-button"
            >
              Cancel
            </Button>
            <Button
              onClick={confirmBooking}
              disabled={isProcessingPayment}
              data-testid="testid-booking-proceed-payment-button"
            >
              {isProcessingPayment ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Opening Payment...
                </>
              ) : (
                'Proceed to Payment'
              )}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Cancel Booking Dialog */}
      <Dialog open={cancelDialogOpen} onOpenChange={(open: boolean) => setCancelDialogOpen(open)}>
        <DialogContent data-testid="testid-booking-cancel-modal">
          <DialogHeader>
            <DialogTitle>Cancel Booking</DialogTitle>
          </DialogHeader>
          {selectedBooking && (
            <div className="py-4">
              <p className="text-sm">
                <span className="text-muted-foreground">Room:</span>{' '}
                <span className="font-medium">{selectedBooking.roomName}</span>
              </p>
              <p className="text-sm">
                <span className="text-muted-foreground">Date:</span>{' '}
                <span className="font-medium">
                  {format(new Date(selectedBooking.checkInDate), 'MMM dd, yyyy')}
                </span>
              </p>
              <p className="text-sm text-muted-foreground mt-3">
                Are you sure you want to cancel this booking? This action cannot be undone.
              </p>
            </div>
          )}
          <div className="flex gap-2 justify-end pt-4">
            <Button
              variant="outline"
              onClick={() => setCancelDialogOpen(false)}
              data-testid="testid-cancel-booking-close-button"
            >
              No, Keep Booking
            </Button>
            <Button
              variant="destructive"
              onClick={confirmCancellation}
              data-testid="testid-confirm-booking-cancel-button"
            >
              Yes, Cancel Booking
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}