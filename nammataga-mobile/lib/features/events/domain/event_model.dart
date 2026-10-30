import 'package:intl/intl.dart';

/// Domain model representing a NammaTAGA association event.
class EventModel {
  final String id;
  final String title;
  final String dateRaw;
  final DateTime? date;
  final String location;
  final String description;
  final int attendees;
  final String status;

  const EventModel({
    required this.id,
    required this.title,
    required this.dateRaw,
    this.date,
    required this.location,
    required this.description,
    this.attendees = 0,
    required this.status,
  });

  factory EventModel.fromJson(Map<String, dynamic> json) {
    final rawDate = json['date'] as String? ?? '';
    final parsedDate = _parseDate(rawDate);

    return EventModel(
      id: json['id']?.toString() ?? '',
      title: json['title'] as String? ?? 'Untitled Event',
      dateRaw: rawDate,
      date: parsedDate,
      location: json['location'] as String? ?? 'Venue TBA',
      description: json['description'] as String? ?? '',
      attendees: json['attendees'] as int? ?? 0,
      status: json['status'] as String? ?? 'upcoming',
    );
  }

  /// Parses date across various backend formats (ISO-8601, 'yyyy-MM-dd HH:mm', 'yyyy-MM-dd', etc.)
  static DateTime? _parseDate(String raw) {
    if (raw.trim().isEmpty) return null;
    final trimmed = raw.trim();

    // 1. Direct ISO-8601 / standard DateTime parser
    final direct = DateTime.tryParse(trimmed);
    if (direct != null) return direct;

    // 2. Common format patterns
    final formats = [
      'yyyy-MM-dd HH:mm',
      'yyyy-MM-dd HH:mm:ss',
      'yyyy/MM/dd HH:mm',
      'yyyy/MM/dd',
      'dd-MM-yyyy HH:mm',
      'dd-MM-yyyy',
      'dd/MM/yyyy HH:mm',
      'dd/MM/yyyy',
      'MM/dd/yyyy HH:mm',
      'MM/dd/yyyy',
      'd MMM yyyy HH:mm',
      'd MMM yyyy',
    ];

    for (final pattern in formats) {
      try {
        return DateFormat(pattern).parseLoose(trimmed);
      } catch (_) {}
    }

    return null;
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'date': dateRaw,
      'location': location,
      'description': description,
      'attendees': attendees,
      'status': status,
    };
  }

  /// Day of the month e.g. "28"
  String get formattedDay {
    if (date != null) {
      return DateFormat('dd').format(date!);
    }
    return '--';
  }

  /// Short month e.g. "NOV"
  String get formattedMonth {
    if (date != null) {
      return DateFormat('MMM').format(date!).toUpperCase();
    }
    return 'TBA';
  }

  /// Formatted date string e.g. "Sat, 28 Nov 2026"
  String get formattedDate {
    if (date != null) {
      return DateFormat('EEE, d MMM yyyy').format(date!);
    }
    return dateRaw.isNotEmpty ? dateRaw : 'Date to be announced';
  }

  /// Formatted time e.g. "09:30 AM" or "All Day" if no time provided
  String get formattedTime {
    if (date != null) {
      if (dateRaw.contains(':')) {
        return DateFormat('hh:mm a').format(date!);
      }
      return 'All Day';
    }
    return 'Time TBA';
  }

  /// Days remaining until event (0 for Today, > 0 for future dates, null if passed or date is null)
  int? get daysRemaining {
    if (date == null) return null;
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final eventDay = DateTime(date!.year, date!.month, date!.day);
    final difference = eventDay.difference(today).inDays;
    return difference >= 0 ? difference : null;
  }

  /// Whether this event is scheduled for today or a future date.
  /// Any event whose date has already passed or whose status is completed/past/cancelled is excluded.
  bool get isUpcoming {
    final lowerStatus = status.toLowerCase().trim();
    if (lowerStatus == 'completed' ||
        lowerStatus == 'finished' ||
        lowerStatus == 'past' ||
        lowerStatus == 'cancelled') {
      return false;
    }

    if (date == null) {
      // If date is absent, only keep if status explicitly indicates upcoming
      return lowerStatus == 'upcoming';
    }

    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final eventDay = DateTime(date!.year, date!.month, date!.day);

    // If the event calendar day is before today, it has finished / passed
    if (eventDay.isBefore(today)) {
      return false;
    }

    return true;
  }

  /// Whether the event date has already passed.
  bool get isPast => !isUpcoming;

  /// Inferred category label based on title keywords
  String get category {
    final lower = title.toLowerCase();
    if (lower.contains('conference') || lower.contains('convention')) {
      return 'State Conference';
    } else if (lower.contains('workshop') || lower.contains('training')) {
      return 'Workshop';
    } else if (lower.contains('summit') || lower.contains('conclave')) {
      return 'Leadership Summit';
    } else if (lower.contains('meeting') || lower.contains('agm') || lower.contains('gbm')) {
      return 'General Meeting';
    }
    return 'Association Event';
  }
}
