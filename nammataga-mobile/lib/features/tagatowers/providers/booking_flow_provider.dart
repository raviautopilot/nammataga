import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../domain/booking_model.dart';
import '../domain/room_model.dart';
import '../presentation/theme/tagatowers_design_tokens.dart';

/// State for the ongoing multi-step booking wizard
class BookingFlowState {
  final Room? selectedRoom;
  final RoomAvailability? roomAvailability;
  final DateTime? checkInDate;
  final DateTime? checkOutDate;
  final String bookingFor; // 'self' | 'guest'
  final int bedCount;
  final String bookingGender; // 'male' | 'female'
  final String bookerPhone;
  final List<GuestDetail> guestDetails;
  final String? pendingBookingId;
  final BookingResponse? pendingBooking;
  final String? lastPaymentOrderId;
  final String? lastPaymentId;
  final String? lastPaymentSignature;

  const BookingFlowState({
    this.selectedRoom,
    this.roomAvailability,
    this.checkInDate,
    this.checkOutDate,
    this.bookingFor = 'self',
    this.bedCount = 1,
    this.bookingGender = 'male',
    this.bookerPhone = '+91 ',
    this.guestDetails = const [],
    this.pendingBookingId,
    this.pendingBooking,
    this.lastPaymentOrderId,
    this.lastPaymentId,
    this.lastPaymentSignature,
  });

  int get totalNights {
    if (checkInDate == null || checkOutDate == null) return 1;
    final diff = checkOutDate!.difference(checkInDate!).inDays;
    return diff > 0 ? diff : 1;
  }

  int get effectiveBeds {
    if (bookingFor == 'self') return 1;
    if (selectedRoom?.allowSingleBed ?? true) return bedCount;
    return selectedRoom?.capacity ?? 1;
  }

  int get advanceRatePerBed {
    if (selectedRoom == null) return TagaTowersDesignTokens.standardAdvance;
    return TagaTowersDesignTokens.getAdvanceRatePerBed(selectedRoom!.type);
  }

  int get advanceAmount => advanceRatePerBed * effectiveBeds;

  BookingFlowState copyWith({
    Room? selectedRoom,
    RoomAvailability? roomAvailability,
    DateTime? checkInDate,
    DateTime? checkOutDate,
    String? bookingFor,
    int? bedCount,
    String? bookingGender,
    String? bookerPhone,
    List<GuestDetail>? guestDetails,
    String? pendingBookingId,
    BookingResponse? pendingBooking,
    String? lastPaymentOrderId,
    String? lastPaymentId,
    String? lastPaymentSignature,
  }) {
    return BookingFlowState(
      selectedRoom: selectedRoom ?? this.selectedRoom,
      roomAvailability: roomAvailability ?? this.roomAvailability,
      checkInDate: checkInDate ?? this.checkInDate,
      checkOutDate: checkOutDate ?? this.checkOutDate,
      bookingFor: bookingFor ?? this.bookingFor,
      bedCount: bedCount ?? this.bedCount,
      bookingGender: bookingGender ?? this.bookingGender,
      bookerPhone: bookerPhone ?? this.bookerPhone,
      guestDetails: guestDetails ?? this.guestDetails,
      pendingBookingId: pendingBookingId ?? this.pendingBookingId,
      pendingBooking: pendingBooking ?? this.pendingBooking,
      lastPaymentOrderId: lastPaymentOrderId ?? this.lastPaymentOrderId,
      lastPaymentId: lastPaymentId ?? this.lastPaymentId,
      lastPaymentSignature: lastPaymentSignature ?? this.lastPaymentSignature,
    );
  }
}

class BookingFlowNotifier extends Notifier<BookingFlowState> {
  @override
  BookingFlowState build() {
    return BookingFlowState(
      checkInDate: DateTime.now(),
      checkOutDate: DateTime.now().add(const Duration(days: 1)),
    );
  }

  void setDates(DateTime checkIn, DateTime checkOut) {
    state = state.copyWith(checkInDate: checkIn, checkOutDate: checkOut);
  }

  void selectRoom(Room room, [RoomAvailability? availability]) {
    state = state.copyWith(
      selectedRoom: room,
      roomAvailability: availability,
      bedCount: 1,
      bookingFor: 'self',
      guestDetails: [const GuestDetail(name: '', age: 0, contact: '', gender: '')],
    );
  }

  void setBookingFor(String bookingFor) {
    state = state.copyWith(bookingFor: bookingFor, bedCount: 1);
  }

  void setBedCount(int count) {
    final currentList = List<GuestDetail>.from(state.guestDetails);
    if (count > currentList.length) {
      currentList.addAll(
        List.generate(
          count - currentList.length,
          (_) => const GuestDetail(name: '', age: 0, contact: '', gender: ''),
        ),
      );
    } else if (count < currentList.length) {
      currentList.removeRange(count, currentList.length);
    }
    state = state.copyWith(bedCount: count, guestDetails: currentList);
  }

  void setBookingGender(String gender) {
    state = state.copyWith(bookingGender: gender);
  }

  void setBookerPhone(String phone) {
    state = state.copyWith(bookerPhone: phone);
  }

  void setGuestDetails(List<GuestDetail> guests) {
    state = state.copyWith(guestDetails: guests);
  }

  void setPendingBookingId(String? id) {
    state = state.copyWith(pendingBookingId: id);
  }

  void setPendingBooking(BookingResponse? booking) {
    state = state.copyWith(
      pendingBooking: booking,
      pendingBookingId: booking?.id ?? state.pendingBookingId,
    );
  }

  void setPaymentInterrupted({
    required String bookingId,
    required String orderId,
    required String paymentId,
    required String signature,
  }) {
    state = state.copyWith(
      pendingBookingId: bookingId,
      lastPaymentOrderId: orderId,
      lastPaymentId: paymentId,
      lastPaymentSignature: signature,
    );
  }

  void reset() {
    state = BookingFlowState(
      checkInDate: DateTime.now(),
      checkOutDate: DateTime.now().add(const Duration(days: 1)),
    );
  }
}

final bookingFlowNotifierProvider =
    NotifierProvider<BookingFlowNotifier, BookingFlowState>(
  BookingFlowNotifier.new,
);
