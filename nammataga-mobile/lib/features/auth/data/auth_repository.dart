import 'dart:convert';
import 'package:dio/dio.dart';
import '../../../core/constants/api_endpoints.dart';
import '../../../core/network/api_client.dart';
import '../../../core/storage/secure_storage_service.dart';
import '../domain/auth_exceptions.dart';
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

      final data = response.data is Map
          ? response.data as Map
          : jsonDecode(response.data.toString()) as Map;

      // Check if user must change temporary password first (can occur with 200 or 403)
      final bool forceChange = data['forceChangePassword'] == true ||
          (data['message'] ?? '').toString().toLowerCase().contains('password change required');
      if (forceChange) {
        return MemberSession(
          forceChangePassword: true,
          tempEmail: (data['email'] ?? identifier).toString(),
        );
      }

      if (response.statusCode == 403) {
        throw const DeletionPendingException();
      }

      if (data['status'] == 'deletion_pending') {
        throw const DeletionPendingException();
      }

      // Explicit error verification in JSON body
      if (data['error'] != null && data['error'].toString().isNotEmpty) {
        final errStr = data['error'].toString();
        if (errStr.toLowerCase().contains('deletion') ||
            errStr.toLowerCase().contains('administrator review') ||
            errStr.toLowerCase().contains('pending')) {
          throw const DeletionPendingException();
        }
        throw Exception(errStr);
      }

      final String? token = data['token'];
      if (token == null || token.toString().trim().isEmpty) {
        throw Exception('Invalid credentials');
      }

      final userData = data['user'] as Map<String, dynamic>?;
      if (userData == null) {
        throw Exception('Invalid credentials');
      }

      final String role = data['role'] ?? 'member';
      final bool isPaid =
          userData['isPaid'] == true ||
          userData['subscription_active'] == true ||
          (userData['payment_status'] as String? ?? '').toLowerCase() == 'paid';

      final int? expiresIn = data['expires_in'];
      await StorageService.saveToken(token, expiresInSeconds: expiresIn);
      await StorageService.saveUserData(jsonEncode(userData), role: role);

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
    } on DeletionPendingException {
      rethrow;
    } on DioException catch (e) {
      final resData = e.response?.data;
      Map? mapData;
      if (resData is Map) {
        mapData = resData;
      } else if (resData is String && resData.isNotEmpty) {
        try {
          final decoded = jsonDecode(resData);
          if (decoded is Map) {
            mapData = decoded;
          }
        } catch (_) {}
      }

      // Check if temporary password change required on 403 or error response
      if (mapData != null) {
        final bool isForceChange = mapData['forceChangePassword'] == true ||
            (mapData['message'] ?? '').toString().toLowerCase().contains('password change required');
        if (isForceChange) {
          return MemberSession(
            forceChangePassword: true,
            tempEmail: (mapData['email'] ?? identifier).toString(),
          );
        }
      }

      if (e.response?.statusCode == 403) {
        // If 403 is received without forceChangePassword, check if it's deletion pending
        throw const DeletionPendingException();
      }
      final data = e.response?.data;
      if (data is Map) {
        final errText = (data['error'] ?? data['message'] ?? '').toString().toLowerCase();
        if (data['status'] == 'deletion_pending' ||
            errText.contains('deletion') ||
            errText.contains('administrator review') ||
            errText.contains('pending')) {
          throw const DeletionPendingException();
        }
      } else if (data is String) {
        final errText = data.toLowerCase();
        if (errText.contains('deletion') ||
            errText.contains('administrator review') ||
            errText.contains('pending') ||
            errText.contains('deletion_pending')) {
          throw const DeletionPendingException();
        }
      }
      final errorMsg = _extractErrorMessage(e);
      if (errorMsg.toLowerCase().contains('deletion') ||
          errorMsg.toLowerCase().contains('administrator review') ||
          errorMsg.toLowerCase().contains('pending')) {
        throw const DeletionPendingException();
      }
      throw Exception(errorMsg);
    } catch (e) {
      if (e is DeletionPendingException) rethrow;
      final errStr = e.toString().toLowerCase();
      if (errStr.contains('deletion') ||
          errStr.contains('administrator review') ||
          errStr.contains('pending')) {
        throw const DeletionPendingException();
      }
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
      if (data is Map) {
        final err = data['error'] ?? data['message'];
        if (err != null && err.toString().isNotEmpty) {
          return err.toString();
        }
      } else if (data is String && data.isNotEmpty) {
        try {
          final decoded = jsonDecode(data);
          if (decoded is Map) {
            final err = decoded['error'] ?? decoded['message'];
            if (err != null && err.toString().isNotEmpty) {
              return err.toString();
            }
          }
        } catch (_) {
          return data;
        }
      }
    }
    if (e.response?.statusCode == 401) {
      return 'Invalid credentials';
    }
    if (e.response?.statusCode == 403) {
      return 'Your account deletion request is awaiting administrator review. You cannot access your account while the request is pending.';
    }
    if (e.type == DioExceptionType.connectionTimeout ||
        e.type == DioExceptionType.receiveTimeout) {
      return 'Connection timed out. Please check your internet connection.';
    }
    return e.message ?? 'Invalid credentials';
  }
}
