import 'dart:convert';
import 'package:dio/dio.dart';
import '../../../core/constants/api_endpoints.dart';
import '../../../core/network/api_client.dart';
import '../../../core/storage/secure_storage_service.dart';
import '../domain/member_session.dart';

class AuthRepository {
  final Dio _dio;

  AuthRepository({Dio? dio}) : _dio = dio ?? ApiClient.instance;

  /// Performs member login using email or TBF number and password
  Future<MemberSession> login({
    required String identifier,
    required String password,
    bool rememberMe = false,
  }) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.login,
        data: {'email': identifier.trim(), 'password': password},
      );

      final data = response.data is Map<String, dynamic>
          ? response.data as Map<String, dynamic>
          : jsonDecode(response.data.toString()) as Map<String, dynamic>;

      // Check if user must change temporary password first
      final bool forceChange = data['forceChangePassword'] == true;
      if (forceChange) {
        return MemberSession(
          forceChangePassword: true,
          tempEmail: (data['email'] ?? identifier).toString(),
        );
      }

      final String? token = data['token'];
      final String role = data['role'] ?? 'member';
      final userData = data['user'] as Map<String, dynamic>?;

      final bool isPaid =
          userData != null &&
          (userData['isPaid'] == true ||
              userData['subscription_active'] == true);

      if (token != null && token.isNotEmpty) {
        final int? expiresIn = data['expires_in'];
        await StorageService.saveToken(token, expiresInSeconds: expiresIn);

        if (userData != null) {
          await StorageService.saveUserData(jsonEncode(userData), role: role);
        }
      }

      if (rememberMe) {
        await StorageService.setRememberEmail(identifier.trim());
      } else {
        await StorageService.setRememberEmail('');
      }

      await StorageService.setGuestMode(false);

      return MemberSession(
        token: token,
        role: role,
        isGuest: false,
        isPaid: isPaid,
        user: userData,
      );
    } on DioException catch (e) {
      final errorMsg = _extractErrorMessage(e);
      throw Exception(errorMsg);
    } catch (e) {
      throw Exception(e.toString().replaceAll('Exception: ', ''));
    }
  }

  /// Request password reset via registered email or TBF number
  Future<String> requestPasswordReset({
    String? email,
    String? tbfNumber,
  }) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.forgotPassword,
        data: {
          if (email != null && email.isNotEmpty) 'email': email.trim(),
          if (tbfNumber != null && tbfNumber.isNotEmpty)
            'tbfNumber': tbfNumber.trim(),
        },
      );

      final data = response.data is Map<String, dynamic>
          ? response.data as Map<String, dynamic>
          : <String, dynamic>{};

      return data['message'] ??
          'Password reset instructions sent successfully.';
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e));
    }
  }

  /// Change temporary password on first login or on demand
  Future<String> changePassword({
    required String email,
    required String oldPassword,
    required String newPassword,
    required String confirmPassword,
  }) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.changePassword,
        data: {
          'email': email.trim(),
          'oldPassword': oldPassword,
          'newPassword': newPassword,
          'confirmPassword': confirmPassword,
        },
      );

      final data = response.data is Map<String, dynamic>
          ? response.data as Map<String, dynamic>
          : <String, dynamic>{};

      return data['message'] ??
          'Password changed successfully. Please sign in with your new password.';
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e));
    }
  }

  /// Logout and wipe stored session tokens
  Future<void> logout() async {
    try {
      await _dio.post(ApiEndpoints.logout);
    } catch (_) {
      // Best-effort remote logout
    } finally {
      await StorageService.clearAuth();
      await StorageService.setGuestMode(false);
    }
  }

  String _extractErrorMessage(DioException e) {
    if (e.response?.data != null) {
      final data = e.response!.data;
      if (data is Map<String, dynamic>) {
        return data['error'] ?? data['message'] ?? 'Authentication failed.';
      } else if (data is String) {
        return data;
      }
    }
    if (e.type == DioExceptionType.connectionTimeout ||
        e.type == DioExceptionType.receiveTimeout) {
      return 'Connection timed out. Please check your internet connection.';
    }
    return e.message ?? 'An unexpected network error occurred.';
  }
}
