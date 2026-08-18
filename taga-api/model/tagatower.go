package model

import "time"

// Room represents a bookable room in TAGA Towers
type Room struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Type           RoomType `json:"type"`
	Capacity       int      `json:"capacity"`
	AllowSingleBed bool     `json:"allowSingleBed"`
}

type RoomType string

const (
	RoomTypeApexSuite  RoomType = "apex-suite"
	RoomTypeACRoom     RoomType = "ac-room"
	RoomTypeGentsDorm  RoomType = "gents-dorm"
	RoomTypeLadiesDorm RoomType = "ladies-dorm"
)

// Booking represents a room booking
type Booking struct {
	ID                string        `json:"id"`
	RoomID            string        `json:"roomId"`
	CheckInDate       time.Time     `json:"checkInDate"`
	CheckOutDate      time.Time     `json:"checkOutDate"`
	BookerName        string        `json:"bookerName"`
	BookerID          string        `json:"bookerId"`
	BookerPhone       string        `json:"bookerPhone"`
	BookingFor        BookingFor    `json:"bookingFor"`
	BedCount          int           `json:"bedCount"`
	Gender            Gender        `json:"gender,omitempty"`
	GuestDetails      []GuestDetail `json:"guestDetails,omitempty"`
	PaymentStatus     PaymentStatus `json:"paymentStatus"`
	UpiID             string        `json:"upiId,omitempty"`
	AdvanceAmount     int           `json:"advanceAmount"`
	CreatedAt         time.Time     `json:"createdAt"`
	RazorpayOrderID   string        `json:"razorpayOrderId"`
	RazorpayPaymentID string        `json:"razorpayPaymentId"`
}

type BookingFor string

const (
	BookingForSelf  BookingFor = "self"
	BookingForGuest BookingFor = "guest"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentConfirmed PaymentStatus = "confirmed"
	PaymentCancelled PaymentStatus = "cancelled"
	PaymentRefunded  PaymentStatus = "refunded"
)

// GuestDetail holds details of a guest being booked for
type GuestDetail struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Contact string `json:"contact"`
	Gender  Gender `json:"gender,omitempty"`
}

// CreateBookingRequest is the payload for POST /towers/bookings
type CreateBookingRequest struct {
	RoomID       string        `json:"roomId" binding:"required"`
	CheckInDate  string        `json:"checkInDate" binding:"required"`
	CheckOutDate string        `json:"checkOutDate" binding:"required"` // "2006-01-02"
	BookerPhone  string        `json:"bookerPhone" binding:"required"`
	BookingFor   BookingFor    `json:"bookingFor" binding:"required"`
	BedCount     int           `json:"bedCount" binding:"required"`
	Gender       Gender        `json:"gender,omitempty"`
	GuestDetails []GuestDetail `json:"guestDetails,omitempty"`
	UpiID        string        `json:"upiId,omitempty"`
}

// RoomAvailability is returned for GET /towers/availability
type RoomAvailability struct {
	Room              Room   `json:"room"`
	Available         bool   `json:"available"`
	AvailableBeds     int    `json:"availableBeds"`
	GenderRestriction Gender `json:"genderRestriction,omitempty"`
}

// BookingResponse is the outbound shape for a booking
type BookingResponse struct {
	ID            string        `json:"id"`
	RoomID        string        `json:"roomId"`
	RoomName      string        `json:"roomName"`
	CheckInDate   string        `json:"checkInDate"`
	CheckOutDate  string        `json:"checkOutDate"`
	BookerName    string        `json:"bookerName"`
	BookerID      string        `json:"bookerId"`
	BookingFor    BookingFor    `json:"bookingFor"`
	BedCount      int           `json:"bedCount"`
	Gender        Gender        `json:"gender,omitempty"`
	GuestDetails  []GuestDetail `json:"guestDetails,omitempty"`
	PaymentStatus PaymentStatus `json:"paymentStatus"`
	AdvanceAmount int           `json:"advanceAmount"`
	// ✅ NEW: Computed booking status (upcoming, active, completed, cancelled)
	BookingStatus string `json:"bookingStatus"`
}
