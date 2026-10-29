library;

/// Model for member announcements / notifications from `GET /api/member/notifications`
class MemberNotification {
  final String id;
  final String title;
  final String message;
  final String priority;
  final bool isRead;
  final String createdAt;

  const MemberNotification({
    required this.id,
    required this.title,
    required this.message,
    required this.priority,
    required this.isRead,
    required this.createdAt,
  });

  String get formattedDate {
    if (createdAt.isEmpty) return 'Recent';
    try {
      final dt = DateTime.parse(createdAt);
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
      return createdAt;
    }
  }

  factory MemberNotification.fromJson(Map<String, dynamic> json) {
    return MemberNotification(
      id: (json['id'] ?? '').toString(),
      title: (json['title'] ?? 'Notification').toString().trim(),
      message: (json['message'] ?? '').toString().trim(),
      priority: (json['priority'] ?? 'normal').toString().trim(),
      isRead: json['is_read'] == true,
      createdAt: (json['created_at'] ?? '').toString().trim(),
    );
  }
}
