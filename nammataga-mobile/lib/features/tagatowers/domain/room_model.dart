/// Domain models representing Room and RoomAvailability for TAGA Towers.
class Room {
  final String id;
  final String name;
  final String type; // 'apex-suite' | 'ac-room' | 'gents-dorm' | 'ladies-dorm'
  final int capacity;
  final bool allowSingleBed;
  final bool hide;

  const Room({
    required this.id,
    required this.name,
    required this.type,
    required this.capacity,
    required this.allowSingleBed,
    this.hide = false,
  });

  factory Room.fromJson(Map<String, dynamic> json) {
    return Room(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      type: json['type'] as String? ?? '',
      capacity: json['capacity'] as int? ?? 1,
      allowSingleBed: json['allowSingleBed'] as bool? ?? false,
      hide: json['hide'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'type': type,
    'capacity': capacity,
    'allowSingleBed': allowSingleBed,
    'hide': hide,
  };
}

class RoomAvailability {
  final Room room;
  final bool available;
  final int availableBeds;
  final String? genderRestriction; // 'male' | 'female' | null

  const RoomAvailability({
    required this.room,
    required this.available,
    required this.availableBeds,
    this.genderRestriction,
  });

  factory RoomAvailability.fromJson(Map<String, dynamic> json) {
    final roomMap = json['room'] as Map<String, dynamic>? ?? {};
    return RoomAvailability(
      room: Room.fromJson(roomMap),
      available: json['available'] as bool? ?? false,
      availableBeds: json['availableBeds'] as int? ?? 0,
      genderRestriction: json['genderRestriction'] as String?,
    );
  }

  Map<String, dynamic> toJson() => {
    'room': room.toJson(),
    'available': available,
    'availableBeds': availableBeds,
    if (genderRestriction != null) 'genderRestriction': genderRestriction,
  };
}

class DateRangeParams {
  final DateTime checkIn;
  final DateTime checkOut;

  const DateRangeParams({required this.checkIn, required this.checkOut});

  String get checkInFormatted =>
      '${checkIn.year}-${checkIn.month.toString().padLeft(2, '0')}-${checkIn.day.toString().padLeft(2, '0')}';

  String get checkOutFormatted =>
      '${checkOut.year}-${checkOut.month.toString().padLeft(2, '0')}-${checkOut.day.toString().padLeft(2, '0')}';

  int get totalNights => checkOut.difference(checkIn).inDays;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is DateRangeParams &&
          runtimeType == other.runtimeType &&
          checkIn.year == other.checkIn.year &&
          checkIn.month == other.checkIn.month &&
          checkIn.day == other.checkIn.day &&
          checkOut.year == other.checkOut.year &&
          checkOut.month == other.checkOut.month &&
          checkOut.day == other.checkOut.day;

  @override
  int get hashCode =>
      Object.hash(checkIn.year, checkIn.month, checkIn.day, checkOut.year, checkOut.month, checkOut.day);
}
