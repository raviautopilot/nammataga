library;

/// Model for member announcements / notifications from `GET /api/member/notifications`
class MemberNotification {
  final String id;
  final String title;
  final String message;
  final String priority;
  final bool isRead;
  final String publishedAt;

  const MemberNotification({
    required this.id,
    required this.title,
    required this.message,
    required this.priority,
    this.isRead = false,
    String? publishedAt,
    String? createdAt,
  }) : publishedAt = publishedAt ?? createdAt ?? '';

  /// Backward compatibility getter
  String get createdAt => publishedAt;

  String get formattedDate {
    final dateStr = publishedAt.isNotEmpty ? publishedAt : createdAt;
    if (dateStr.isEmpty) return 'Recent';
    try {
      final dt = DateTime.parse(dateStr);
      const months = [
        'Jan',
        'Feb',
        'Mar',
        'Apr',
        'May',
        'Jun',
        'Jul',
        'Aug',
        'Sep',
        'Oct',
        'Nov',
        'Dec',
      ];
      return '${months[dt.month - 1]} ${dt.day}, ${dt.year}';
    } catch (_) {
      return dateStr;
    }
  }

  factory MemberNotification.fromJson(Map<String, dynamic> json) {
    final rawDate = (json['published_at'] ??
            json['publishedAt'] ??
            json['created_at'] ??
            json['createdAt'] ??
            json['sent_at'] ??
            json['sentAt'] ??
            '')
        .toString()
        .trim();

    return MemberNotification(
      id: (json['id'] ?? '').toString(),
      title: (json['title'] ?? 'Notification').toString().trim(),
      message: (json['message'] ?? '').toString().trim(),
      priority: (json['priority'] ?? 'normal').toString().trim(),
      isRead: json['is_read'] == true || json['isRead'] == true,
      publishedAt: rawDate,
    );
  }
}
