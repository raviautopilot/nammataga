package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"taga-api/model"
)

func TestCreateBooking_DateValidation(t *testing.T) {
	// 1. Check-out date before check-in date
	reqCheckOutBefore := model.CreateBookingRequest{
		RoomID:       "gents-dorm",
		CheckInDate:  "2027-01-10",
		CheckOutDate: "2027-01-05",
		BookerPhone:  "9876543210",
		BookingFor:   model.BookingForSelf,
		BedCount:     1,
	}
	_, err := CreateBooking(reqCheckOutBefore, "Test User", "TAGA001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check-out date must be after check-in date")

	// 2. Check-out date equal to check-in date (0 days stay)
	reqSameDate := model.CreateBookingRequest{
		RoomID:       "gents-dorm",
		CheckInDate:  "2027-01-10",
		CheckOutDate: "2027-01-10",
		BookerPhone:  "9876543210",
		BookingFor:   model.BookingForSelf,
		BedCount:     1,
	}
	_, err = CreateBooking(reqSameDate, "Test User", "TAGA001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check-out date must be after check-in date")

	// 3. Past date
	reqPastDate := model.CreateBookingRequest{
		RoomID:       "gents-dorm",
		CheckInDate:  "2020-01-01",
		CheckOutDate: "2020-01-05",
		BookerPhone:  "9876543210",
		BookingFor:   model.BookingForSelf,
		BedCount:     1,
	}
	_, err = CreateBooking(reqPastDate, "Test User", "TAGA001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot book dates in the past")

	// 4. Stay duration > 10 days
	reqLongStay := model.CreateBookingRequest{
		RoomID:       "gents-dorm",
		CheckInDate:  "2027-01-01",
		CheckOutDate: "2027-01-15", // 14 days
		BookerPhone:  "9876543210",
		BookingFor:   model.BookingForSelf,
		BedCount:     1,
	}
	_, err = CreateBooking(reqLongStay, "Test User", "TAGA001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum stay duration is 10 days")

	// 5. Invalid date strings
	reqInvalidDate := model.CreateBookingRequest{
		RoomID:       "gents-dorm",
		CheckInDate:  "invalid-date",
		CheckOutDate: "2027-01-05",
		BookerPhone:  "9876543210",
		BookingFor:   model.BookingForSelf,
		BedCount:     1,
	}
	_, err = CreateBooking(reqInvalidDate, "Test User", "TAGA001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid check-in date")

	_ = time.Now()
}
