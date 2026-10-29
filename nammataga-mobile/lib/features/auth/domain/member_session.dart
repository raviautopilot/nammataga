/// Represents the authenticated member or guest session state.
class MemberSession {
  final String? token;
  final String? role;
  final bool isGuest;
  final bool isPaid;
  final Map<String, dynamic>? user;
  final bool forceChangePassword;
  final String? tempEmail;

  const MemberSession({
    this.token,
    this.role,
    this.isGuest = false,
    this.isPaid = false,
    this.user,
    this.forceChangePassword = false,
    this.tempEmail,
  });

  bool get isAuthenticated => token != null && token!.isNotEmpty && !isGuest;

  String get displayName {
    if (isGuest) return 'Guest Visitor';
    if (user != null) {
      final name = user!['name'] ?? user!['username'] ?? '';
      if (name.isNotEmpty) return name.toString();
    }
    return 'Member';
  }

  String get memberEmail {
    if (user != null) {
      return (user!['emailId'] ?? user!['email'] ?? '').toString();
    }
    return tempEmail ?? '';
  }

  MemberSession copyWith({
    String? token,
    String? role,
    bool? isGuest,
    bool? isPaid,
    Map<String, dynamic>? user,
    bool? forceChangePassword,
    String? tempEmail,
  }) {
    return MemberSession(
      token: token ?? this.token,
      role: role ?? this.role,
      isGuest: isGuest ?? this.isGuest,
      isPaid: isPaid ?? this.isPaid,
      user: user ?? this.user,
      forceChangePassword: forceChangePassword ?? this.forceChangePassword,
      tempEmail: tempEmail ?? this.tempEmail,
    );
  }

  factory MemberSession.guest() {
    return const MemberSession(isGuest: true, role: 'guest', isPaid: false);
  }

  factory MemberSession.unauthenticated() {
    return const MemberSession(isGuest: false, isPaid: false);
  }
}
