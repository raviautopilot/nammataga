library;

// Domain models representing Booking, GuestDetail, and Booking Lifecycle for TAGA Towers.

class GuestDetail {
  final String name;
  final int age;
  final String contact;
  final String gender; // 'male' | 'female' | ''

  const GuestDetail({
    required this.name,
    required this.age,
    required this.contact,
    required this.gender,
  });

  factory GuestDetail.fromJson(Map<String, dynamic> json) {
    return GuestDetail(
      name: json['name'] as String? ?? '',
      age: json['age'] as int? ?? 0,
      contact: json['contact'] as String? ?? '',
      gender: json['gender'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
    'name': name,
    'age': age,
    'contact': contact,
    'gender': gender,
  };

  GuestDetail copyWith({
    String? name,
    int? age,
    String? contact,
    String? gender,
  }) {
    return GuestDetail(
      name: name ?? this.name,
      age: age ?? this.age,
      contact: contact ?? this.contact,
      gender: gender ?? this.gender,
    );
  }
}

class CreateBookingRequest {
  final String roomId;
  final String checkInDate; // YYYY-MM-DD
  final String checkOutDate; // YYYY-MM-DD
  final String bookerPhone;
  final String bookingFor; // 'self' | 'guest'
  final int bedCount;
  final String? gender; // 'male' | 'female' | 'mixed'
  final List<GuestDetail>? guestDetails;
  final String? upiId;

  const CreateBookingRequest({
    required this.roomId,
    required this.checkInDate,
    required this.checkOutDate,
    required this.bookerPhone,
    required this.bookingFor,
    required this.bedCount,
    this.gender,
    this.guestDetails,
    this.upiId,
  });

  Map<String, dynamic> toJson() => {
    'roomId': roomId,
    'checkInDate': checkInDate,
    'checkOutDate': checkOutDate,
    'bookerPhone': bookerPhone,
    'bookingFor': bookingFor,
    'bedCount': bedCount,
    if (gender != null && gender!.isNotEmpty) 'gender': gender,
    if (guestDetails != null && guestDetails!.isNotEmpty)
      'guestDetails': guestDetails!.map((g) => g.toJson()).toList(),
    if (upiId != null) 'upiId': upiId,
  };
}

class BookingResponse {
  final String id;
  final String roomId;
  final String roomName;
  final String checkInDate;
  final String checkOutDate;
  final String bookerName;
  final String bookerId;
  final String? bookerPhone;
  final String bookingFor; // 'self' | 'guest'
  final int bedCount;
  final String? gender;
  final List<GuestDetail> guestDetails;
  final String paymentStatus; // 'pending' | 'confirmed' | 'cancelled' | 'refunded'
  final int advanceAmount;
  final String bookingStatus; // 'upcoming' | 'active' | 'completed' | 'cancelled'
  final String? razorpayOrderId;
  final String? razorpayPaymentId;

  const BookingResponse({
    required this.id,
    required this.roomId,
    required this.roomName,
    required this.checkInDate,
    required this.checkOutDate,
    required this.bookerName,
    required this.bookerId,
    this.bookerPhone,
    required this.bookingFor,
    required this.bedCount,
    this.gender,
    this.guestDetails = const [],
    required this.paymentStatus,
    required this.advanceAmount,
    required this.bookingStatus,
    this.razorpayOrderId,
    this.razorpayPaymentId,
  });

  factory BookingResponse.fromJson(Map<String, dynamic> json) {
    final guestsRaw = json['guestDetails'] as List<dynamic>? ?? [];
    return BookingResponse(
      id: json['id'] as String? ?? '',
      roomId: json['roomId'] as String? ?? '',
      roomName: json['roomName'] as String? ?? 'Room',
      checkInDate: json['checkInDate'] as String? ?? '',
      checkOutDate: json['checkOutDate'] as String? ?? '',
      bookerName: json['bookerName'] as String? ?? 'Member',
      bookerId: json['bookerId'] as String? ?? '',
      bookerPhone: (json['bookerPhone'] ??
              json['booker_phone'] ??
              json['phone'] ??
              json['mobileNumber'] ??
              json['mobile_number'] ??
              json['contact']) as String?,
      bookingFor: json['bookingFor'] as String? ?? 'self',
      bedCount: json['bedCount'] as int? ?? 1,
      gender: json['gender'] as String?,
      guestDetails: guestsRaw
          .map((g) => GuestDetail.fromJson(g as Map<String, dynamic>))
          .toList(),
      paymentStatus: json['paymentStatus'] as String? ?? 'pending',
      advanceAmount: json['advanceAmount'] as int? ?? 0,
      bookingStatus: json['bookingStatus'] as String? ?? 'upcoming',
      razorpayOrderId: json['razorpayOrderId'] as String?,
      razorpayPaymentId: json['razorpayPaymentId'] as String?,
    );
  }

  BookingResponse copyWith({
    String? id,
    String? roomId,
    String? roomName,
    String? checkInDate,
    String? checkOutDate,
    String? bookerName,
    String? bookerId,
    String? bookerPhone,
    String? bookingFor,
    int? bedCount,
    String? gender,
    List<GuestDetail>? guestDetails,
    String? paymentStatus,
    int? advanceAmount,
    String? bookingStatus,
    String? razorpayOrderId,
    String? razorpayPaymentId,
  }) {
    return BookingResponse(
      id: id ?? this.id,
      roomId: roomId ?? this.roomId,
      roomName: roomName ?? this.roomName,
      checkInDate: checkInDate ?? this.checkInDate,
      checkOutDate: checkOutDate ?? this.checkOutDate,
      bookerName: bookerName ?? this.bookerName,
      bookerId: bookerId ?? this.bookerId,
      bookerPhone: bookerPhone ?? this.bookerPhone,
      bookingFor: bookingFor ?? this.bookingFor,
      bedCount: bedCount ?? this.bedCount,
      gender: gender ?? this.gender,
      guestDetails: guestDetails ?? this.guestDetails,
      paymentStatus: paymentStatus ?? this.paymentStatus,
      advanceAmount: advanceAmount ?? this.advanceAmount,
      bookingStatus: bookingStatus ?? this.bookingStatus,
      razorpayOrderId: razorpayOrderId ?? this.razorpayOrderId,
      razorpayPaymentId: razorpayPaymentId ?? this.razorpayPaymentId,
    );
  }

  bool get isConfirmed => paymentStatus.toLowerCase() == 'confirmed';
  bool get isPending => paymentStatus.toLowerCase() == 'pending';
  bool get isCancelled =>
      paymentStatus.toLowerCase() == 'cancelled' ||
      bookingStatus.toLowerCase() == 'cancelled';
  bool get isUpcoming => bookingStatus.toLowerCase() == 'upcoming';
  bool get isActive => bookingStatus.toLowerCase() == 'active';
  bool get isCompleted => bookingStatus.toLowerCase() == 'completed';

  /// Whether cancellation is eligible:
  /// Cancellation is only permitted strictly before the check-in date
  /// and only if the booking is not already cancelled.
  bool get isCancellable {

    if (paymentStatus == 'cancelled' || bookingStatus == 'cancelled') return false;
    try {
      final checkIn = DateTime.parse(checkInDate);
      final today = DateTime.now();
      final todayDate = DateTime(today.year, today.month, today.day);
      final checkInOnly = DateTime(checkIn.year, checkIn.month, checkIn.day);
      return todayDate.isBefore(checkInOnly);
    } catch (_) {
      return false;
    }
  }
}
