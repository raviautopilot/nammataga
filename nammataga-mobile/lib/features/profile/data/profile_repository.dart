import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../../../core/constants/api_endpoints.dart';
import '../../../core/network/api_client.dart';
import '../../../core/storage/secure_storage_service.dart';
import '../domain/deletion_request_model.dart';
import '../domain/edit_request_model.dart';
import '../domain/member_notification_model.dart';
import '../domain/member_profile.dart';
import '../domain/subscription_plan.dart';

class ProfileRepository {
  final Dio _dio;

  ProfileRepository({Dio? dio}) : _dio = dio ?? ApiClient.instance;

  Map<String, dynamic>? _toMap(dynamic data) {
    if (data is Map<String, dynamic>) return data;
    if (data is Map) return Map<String, dynamic>.from(data);
    if (data is String) {
      try {
        final decoded = jsonDecode(data);
        if (decoded is Map<String, dynamic>) return decoded;
        if (decoded is Map) return Map<String, dynamic>.from(decoded);
      } catch (_) {}
    }
    return null;
  }

  /// Fetches authenticated member profile from `GET /api/member/profile`
  Future<MemberProfile> getProfile() async {
    try {
      final response = await _dio.get(ApiEndpoints.profile);
      final map = _toMap(response.data);
      if (map != null) {
        final userRaw = map['user'] ?? map;
        final userMap = _toMap(userRaw);
        if (userMap != null) {
          final profile = MemberProfile.fromJson(userMap);
          // Sync to local storage for offline resilience
          await StorageService.saveUserData(
            jsonEncode(userMap),
            role: profile.isPaid ? 'subscriber' : 'member',
          );
          return profile;
        }
      }
    } catch (e) {
      debugPrint('[ProfileRepository] Error fetching profile: $e');

      // Attempt reading cached profile from storage
      final cachedJson = await StorageService.getUserData();
      if (cachedJson != null && cachedJson.isNotEmpty) {
        try {
          final userMap = _toMap(jsonDecode(cachedJson));
          if (userMap != null) {
            return MemberProfile.fromJson(userMap);
          }
        } catch (_) {}
      }
      rethrow;
    }
    throw Exception('Failed to load profile data from server');
  }

  /// Fetches the official Tamil Nadu districts strictly from backend `GET /api/membership/districts`
  Future<List<String>> getDistricts() async {
    try {
      final response = await _dio.get(ApiEndpoints.districts);
      if (response.data is List) {
        return (response.data as List).map((e) => e.toString().trim()).toList();
      }
    } catch (e) {
      debugPrint('[ProfileRepository] Error fetching districts from API: $e');
      rethrow;
    }
    throw Exception('Failed to load districts from server');
  }

  /// Submits profile change request to `POST /api/member/edit-request`
  Future<Map<String, dynamic>> submitEditRequest(
    EditRequestPayload payload,
  ) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.editRequest,
        data: payload.toJson(),
      );
      final map = _toMap(response.data);
      return map ?? {'success': true};
    } catch (e) {
      debugPrint('[ProfileRepository] Error submitting edit request: $e');
      if (e is DioException) {
        final errMap = _toMap(e.response?.data);
        final errMsg = errMap?['error'] ?? errMap?['message'] ?? e.message;
        throw Exception(errMsg);
      }
      rethrow;
    }
  }

  /// Fetches notifications for authenticated member from `GET /api/member/notifications`
  Future<List<MemberNotification>> getNotifications({String? memberId}) async {
    try {
      final response = await _dio.get(
        ApiEndpoints.notifications,
        queryParameters: memberId != null && memberId.isNotEmpty
            ? {'member_id': memberId}
            : null,
      );
      if (response.data is List) {
        return (response.data as List)
            .map((e) => MemberNotification.fromJson(e as Map<String, dynamic>))
            .toList();
      }
    } catch (e) {
      debugPrint('[ProfileRepository] Error fetching notifications: $e');
    }
    return [];
  }

  /// Stubbed out as read/unread tracking is eliminated. Always a no-op.
  Future<void> markNotificationRead(String id) async {
    // Read/unread tracking is eliminated.
  }

  /// Stubbed out as read/unread tracking is eliminated. Always returns 0.
  Future<int> getUnreadCount({String? memberId}) async {
    return 0;
  }

  /// Fetches all available subscription plans and contribution options from `GET /api/subscriptions`
  Future<List<SubscriptionPlan>> getSubscriptions() async {
    try {
      final response = await _dio.get(ApiEndpoints.subscriptions);
      final map = _toMap(response.data);
      if (map != null && map['data'] is List) {
        return (map['data'] as List)
            .map((item) => SubscriptionPlan.fromJson(item as Map<String, dynamic>))
            .toList();
      }
    } catch (e) {
      debugPrint('[ProfileRepository] Error fetching subscription plans: $e');
    }
    return [];
  }

  /// Fetches member subscription status from `GET /api/subscriptions/status?email=...`
  Future<Map<String, dynamic>?> getMemberSubscriptionStatus(String email) async {
    try {
      final response = await _dio.get(ApiEndpoints.subscriptionStatus(email));
      return _toMap(response.data);
    } catch (e) {
      debugPrint('[ProfileRepository] Error fetching subscription status: $e');
    }
    return null;
  }

  /// Fetches the IDs of plans already paid by the member from `GET /api/subscriptions/member-paid?email=...`
  Future<List<String>> getMemberPaidSubscriptions(String email) async {
    try {
      final response = await _dio.get(
        '${ApiEndpoints.baseUrl}/subscriptions/member-paid?email=${Uri.encodeComponent(email)}',
      );
      if (response.data is List) {
        return (response.data as List).map((e) => e.toString()).toList();
      }
    } catch (e) {
      debugPrint('[ProfileRepository] Error fetching paid subscriptions: $e');
    }
    return [];
  }

  /// Creates a Razorpay order from `POST /api/subscriptions/create-order`
  Future<RazorpayOrder> createSubscriptionOrder({
    required String subscriptionId,
    required int amountInPaise,
    required String email,
    Map<String, dynamic>? notes,
  }) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.createSubscriptionOrder,
        data: {
          'subscriptionId': subscriptionId,
          'amount': amountInPaise,
          'email': email,
          'notes': ?notes,
        },
      );
      final map = _toMap(response.data);
      if (map == null) throw Exception('Empty response from create-order');
      return RazorpayOrder.fromJson(map);
    } catch (e) {
      debugPrint('[ProfileRepository] Error creating Razorpay order: $e');
      if (e is DioException) {
        final errMap = _toMap(e.response?.data);
        final errMsg = errMap?['error'] ?? errMap?['message'] ?? e.message;
        throw Exception(errMsg);
      }
      rethrow;
    }
  }

  /// Verifies a Razorpay payment via `POST /api/subscriptions/verify-payment`
  Future<Map<String, dynamic>> verifySubscriptionPayment({
    required String subscriptionId,
    required String orderId,
    required String paymentId,
    required String signature,
    required String email,
    required int amountInPaise,
  }) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.verifySubscriptionPayment,
        data: {
          'subscriptionId': subscriptionId,
          'razorpay_order_id': orderId,
          'razorpay_payment_id': paymentId,
          'razorpay_signature': signature,
          'email': email,
          'amount': amountInPaise,
        },
      );
      final map = _toMap(response.data);
      return map ?? {'success': true};
    } catch (e) {
      debugPrint('[ProfileRepository] Error verifying Razorpay payment: $e');
      if (e is DioException) {
        final errMap = _toMap(e.response?.data);
        final errMsg = errMap?['error'] ?? errMap?['message'] ?? e.message;
        throw Exception(errMsg);
      }
      rethrow;
    }
  }

  /// Fetches account deletion eligibility from `GET /api/member/account-deletion/eligibility`
  Future<DeletionEligibility> getAccountDeletionEligibility() async {
    try {
      final response = await _dio.get(
        ApiEndpoints.accountDeletionEligibility,
        options: Options(
          sendTimeout: const Duration(seconds: 5),
          receiveTimeout: const Duration(seconds: 5),
        ),
      );
      final map = _toMap(response.data);
      if (map != null) {
        return DeletionEligibility.fromJson(map);
      }
      return const DeletionEligibility(canDelete: true, hasActiveStay: false);
    } catch (e) {
      debugPrint('[ProfileRepository] Error fetching deletion eligibility: $e');
      if (e is DioException) {
        final errMap = _toMap(e.response?.data);
        final errMsg = errMap?['error'] ?? errMap?['message'] ?? e.message;
        final reason = errMap?['reason']?.toString();
        if (reason == 'active_stay' || errMsg?.toString().contains('TAGA Towers') == true) {
          return DeletionEligibility(
            canDelete: false,
            hasActiveStay: true,
            message: errMsg?.toString(),
            reason: 'active_stay',
          );
        }
      }
      return const DeletionEligibility(canDelete: true, hasActiveStay: false);
    }
  }

  /// Fetches account deletion status from `GET /api/member/account-deletion/status`
  Future<AccountDeletionRequestStatus> getAccountDeletionStatus() async {
    try {
      final response = await _dio.get(
        ApiEndpoints.accountDeletionStatus,
        options: Options(
          sendTimeout: const Duration(seconds: 5),
          receiveTimeout: const Duration(seconds: 5),
        ),
      );
      final map = _toMap(response.data);
      if (map != null) {
        return AccountDeletionRequestStatus.fromJson(map);
      }
      return const AccountDeletionRequestStatus(status: 'none');
    } catch (e) {
      debugPrint('[ProfileRepository] Error fetching deletion status: $e');
      return const AccountDeletionRequestStatus(status: 'none');
    }
  }

  /// Submits an account deletion request to `POST /api/member/account-deletion`
  Future<Map<String, dynamic>> submitAccountDeletionRequest({
    required String password,
    String? reason,
  }) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.accountDeletion,
        data: {
          'password': password,
          if (reason != null && reason.trim().isNotEmpty)
            'reason': reason.trim(),
        },
      );
      final map = _toMap(response.data);
      return map ??
          {
            'success': true,
            'message': 'Account deletion request submitted successfully',
          };
    } catch (e) {
      debugPrint('[ProfileRepository] Error submitting deletion request: $e');
      if (e is DioException) {
        final errMap = _toMap(e.response?.data);
        final errMsg = errMap?['error'] ?? errMap?['message'] ?? e.message;
        throw Exception(errMsg);
      }
      rethrow;
    }
  }
}
