/// Domain models for Member Account Deletion workflows.
class DeletionEligibility {
  final bool canDelete;
  final bool hasActiveStay;
  final String? message;
  final String? reason;

  const DeletionEligibility({
    required this.canDelete,
    required this.hasActiveStay,
    this.message,
    this.reason,
  });

  factory DeletionEligibility.fromJson(Map<String, dynamic> json) {
    final eligible = json['eligible'] == true || json['canDelete'] == true;
    final reasonStr = json['reason']?.toString();
    final hasActiveStay = json['hasActiveStay'] == true ||
        reasonStr?.toLowerCase() == 'active_stay';

    return DeletionEligibility(
      canDelete: eligible && !hasActiveStay,
      hasActiveStay: hasActiveStay,
      message: json['message']?.toString(),
      reason: reasonStr,
    );
  }

  Map<String, dynamic> toJson() => {
    'canDelete': canDelete,
    'hasActiveStay': hasActiveStay,
    'message': message,
    'reason': reason,
  };
}

class AccountDeletionRequestStatus {
  final String status;
  final String? reason;
  final String? adminRemarks;
  final DateTime? createdAt;
  final DateTime? processedAt;
  final String? requestId;
  final bool hasPendingRequest;

  const AccountDeletionRequestStatus({
    required this.status,
    this.reason,
    this.adminRemarks,
    this.createdAt,
    this.processedAt,
    this.requestId,
    this.hasPendingRequest = false,
  });

  bool get isPending => status.toLowerCase() == 'pending' || hasPendingRequest;
  bool get isRejected => status.toLowerCase() == 'rejected';
  bool get isApproved => status.toLowerCase() == 'approved';
  bool get isNone => status.toLowerCase() == 'none' || status.isEmpty;

  factory AccountDeletionRequestStatus.fromJson(Map<String, dynamic> json) {
    final statusStr = json['status']?.toString().toLowerCase() ?? 'none';
    final hasPending =
        statusStr == 'pending' || json['hasPendingRequest'] == true;

    DateTime? parseDate(dynamic date) {
      if (date == null) return null;
      if (date is DateTime) return date;
      try {
        return DateTime.tryParse(date.toString());
      } catch (_) {
        return null;
      }
    }

    return AccountDeletionRequestStatus(
      status: statusStr,
      reason: json['reason']?.toString(),
      adminRemarks: json['remarks']?.toString() ??
          json['adminRemarks']?.toString() ??
          json['admin_remarks']?.toString(),
      createdAt: parseDate(
        json['requested_at'] ?? json['createdAt'] ?? json['created_at'],
      ),
      processedAt: parseDate(json['processed_at'] ?? json['processedAt']),
      requestId: json['request_id']?.toString() ?? json['id']?.toString(),
      hasPendingRequest: hasPending,
    );
  }

  Map<String, dynamic> toJson() => {
    'status': status,
    'reason': reason,
    'adminRemarks': adminRemarks,
    'createdAt': createdAt?.toIso8601String(),
    'processedAt': processedAt?.toIso8601String(),
    'requestId': requestId,
    'hasPendingRequest': hasPendingRequest,
  };
}
