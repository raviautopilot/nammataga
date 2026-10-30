import 'package:intl/intl.dart';

/// Domain model representing a member's submitted grievance petition.
class GrievanceModel {
  final String id;
  final String subject;
  final String category;
  final String priority;
  final String description;
  final String? contactPhone;
  final String preferredResponse;
  final String status;
  final DateTime? submittedDate;
  final String? memberName;
  final String? memberEmail;
  final String? assignedTo;
  final DateTime? lastUpdate;

  const GrievanceModel({
    required this.id,
    required this.subject,
    required this.category,
    required this.priority,
    required this.description,
    this.contactPhone,
    required this.preferredResponse,
    required this.status,
    this.submittedDate,
    this.memberName,
    this.memberEmail,
    this.assignedTo,
    this.lastUpdate,
  });

  /// Factory constructor to parse grievance JSON payloads from backend.
  factory GrievanceModel.fromJson(Map<String, dynamic> json) {
    return GrievanceModel(
      id: json['id']?.toString() ?? '',
      subject: json['subject']?.toString() ?? '',
      category: json['category']?.toString() ?? '',
      priority: json['priority']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      contactPhone: json['contactPhone']?.toString(),
      preferredResponse: json['preferredResponse']?.toString() ?? 'Email',
      status: json['status']?.toString() ?? 'Pending',
      submittedDate: _parseDate(json['submittedDate']?.toString()),
      memberName: json['memberName']?.toString(),
      memberEmail: json['memberEmail']?.toString(),
      assignedTo: json['assignedTo']?.toString(),
      lastUpdate: _parseDate(json['lastUpdate']?.toString()),
    );
  }

  /// Parses date across various backend formats (ISO-8601, RFC3339, standard date formats).
  static DateTime? _parseDate(String? raw) {
    if (raw == null || raw.trim().isEmpty) return null;
    final trimmed = raw.trim();

    final direct = DateTime.tryParse(trimmed);
    if (direct != null) return direct;

    const formats = [
      'yyyy-MM-dd HH:mm:ss',
      'yyyy-MM-dd HH:mm',
      'yyyy-MM-dd',
      'dd-MM-yyyy HH:mm',
      'dd-MM-yyyy',
      'dd/MM/yyyy HH:mm',
      'dd/MM/yyyy',
      'd MMM yyyy, hh:mm a',
      'd MMM yyyy',
    ];

    for (final pattern in formats) {
      try {
        return DateFormat(pattern).parseLoose(trimmed);
      } catch (_) {}
    }

    return null;
  }

  /// Serializes grievance model back to JSON map.
  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'subject': subject,
      'category': category,
      'priority': priority,
      'description': description,
      if (contactPhone != null) 'contactPhone': contactPhone,
      'preferredResponse': preferredResponse,
      'status': status,
      if (submittedDate != null)
        'submittedDate': submittedDate!.toIso8601String(),
      if (memberName != null) 'memberName': memberName,
      if (memberEmail != null) 'memberEmail': memberEmail,
      if (assignedTo != null) 'assignedTo': assignedTo,
      if (lastUpdate != null) 'lastUpdate': lastUpdate!.toIso8601String(),
    };
  }

  /// Formatted date and time string for display (e.g., "30 Oct 2026, 02:30 PM").
  String get formattedDate {
    if (submittedDate == null) return '';
    return DateFormat('dd MMM yyyy, hh:mm a').format(submittedDate!);
  }

  /// Formatted date only (e.g., "30 Oct 2026").
  String get formattedDateOnly {
    if (submittedDate == null) return '';
    return DateFormat('dd MMM yyyy').format(submittedDate!);
  }

  /// Indicates if the grievance is still awaiting review or action.
  bool get isPending => status.toLowerCase() == 'pending';

  /// Indicates if the grievance has been resolved or closed.
  bool get isResolved =>
      status.toLowerCase() == 'resolved' ||
      status.toLowerCase() == 'closed' ||
      status.toLowerCase() == 'read';

  GrievanceModel copyWith({
    String? id,
    String? subject,
    String? category,
    String? priority,
    String? description,
    String? contactPhone,
    String? preferredResponse,
    String? status,
    DateTime? submittedDate,
    String? memberName,
    String? memberEmail,
    String? assignedTo,
    DateTime? lastUpdate,
  }) {
    return GrievanceModel(
      id: id ?? this.id,
      subject: subject ?? this.subject,
      category: category ?? this.category,
      priority: priority ?? this.priority,
      description: description ?? this.description,
      contactPhone: contactPhone ?? this.contactPhone,
      preferredResponse: preferredResponse ?? this.preferredResponse,
      status: status ?? this.status,
      submittedDate: submittedDate ?? this.submittedDate,
      memberName: memberName ?? this.memberName,
      memberEmail: memberEmail ?? this.memberEmail,
      assignedTo: assignedTo ?? this.assignedTo,
      lastUpdate: lastUpdate ?? this.lastUpdate,
    );
  }

  @override
  String toString() =>
      'GrievanceModel(id: $id, subject: $subject, status: $status, category: $category)';

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is GrievanceModel &&
          runtimeType == other.runtimeType &&
          id == other.id;

  @override
  int get hashCode => id.hashCode;
}
