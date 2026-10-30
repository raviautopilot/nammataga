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
    DateTime? parsedDate;
    if (rawDate.isNotEmpty) {
      try {
        // Tries standard API format 'yyyy-MM-dd HH:mm'
        parsedDate = DateFormat('yyyy-MM-dd HH:mm').parseLoose(rawDate);
      } catch (_) {
        try {
          parsedDate = DateTime.tryParse(rawDate);
        } catch (_) {}
      }
    }

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

  /// Formatted time e.g. "09:30 AM"
  String get formattedTime {
    if (date != null) {
      return DateFormat('hh:mm a').format(date!);
    }
    return 'Time TBA';
  }

  /// Days remaining until event
  int? get daysRemaining {
    if (date == null) return null;
    final now = DateTime.now();
    final difference = date!.difference(now).inDays;
    return difference >= 0 ? difference : null;
  }

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
