import 'dart:convert';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/storage/secure_storage_service.dart';
import '../data/auth_repository.dart';
import '../domain/member_session.dart';

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository();
});

class AuthNotifier extends AsyncNotifier<MemberSession> {
  @override
  Future<MemberSession> build() async {
    return _restoreSession();
  }

  Future<MemberSession> _restoreSession() async {
    final isValid = await StorageService.isTokenValid();
    if (isValid) {
      final token = await StorageService.getToken();
      final rawUser = await StorageService.getUserData();

      Map<String, dynamic>? user;
      bool isPaid = false;
      if (rawUser != null && rawUser.isNotEmpty) {
        try {
          user = jsonDecode(rawUser) as Map<String, dynamic>;
          isPaid =
              user['isPaid'] == true ||
              user['subscription_active'] == true ||
              (user['payment_status'] as String? ?? '').toLowerCase() == 'paid';
        } catch (_) {}
      }

      return MemberSession(
        token: token,
        role: 'member',
        isPaid: isPaid,
        user: user,
        isGuest: false,
      );
    }

    // Unauthenticated: Guest browsing is disallowed
    await StorageService.setGuestMode(false);
    return MemberSession.unauthenticated();
  }

  /// Sign in member with credentials
  Future<MemberSession> login({
    required String identifier,
    required String password,
    bool rememberMe = false,
  }) async {
    state = const AsyncValue.loading();
    final repository = ref.read(authRepositoryProvider);

    state = await AsyncValue.guard(() async {
      return await repository.login(
        identifier: identifier,
        password: password,
        rememberMe: rememberMe,
      );
    });

    return state.requireValue;
  }

  /// Submit password recovery
  Future<String> requestPasswordReset({
    String? email,
    String? tbfNumber,
  }) async {
    final repository = ref.read(authRepositoryProvider);
    return await repository.requestPasswordReset(
      email: email,
      tbfNumber: tbfNumber,
    );
  }

  /// Change password
  Future<String> changePassword({
    required String email,
    required String oldPassword,
    required String newPassword,
    required String confirmPassword,
  }) async {
    final repository = ref.read(authRepositoryProvider);
    return await repository.changePassword(
      email: email,
      oldPassword: oldPassword,
      newPassword: newPassword,
      confirmPassword: confirmPassword,
    );
  }

  /// Sign out
  Future<void> logout() async {
    state = const AsyncValue.loading();
    final repository = ref.read(authRepositoryProvider);
    try {
      await repository.logout();
    } finally {
      state = AsyncValue.data(MemberSession.unauthenticated());
    }
  }
}

final authProvider = AsyncNotifierProvider<AuthNotifier, MemberSession>(
  AuthNotifier.new,
);
