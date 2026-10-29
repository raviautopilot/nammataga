import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../../../core/constants/api_endpoints.dart';
import '../../../core/network/api_client.dart';
import '../../../core/storage/secure_storage_service.dart';
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
  Future<List<MemberNotification>> getNotifications() async {
    try {
      final response = await _dio.get(ApiEndpoints.notifications);
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

  /// Marks a notification as read via `PUT /api/member/notifications/:id/read`
  Future<void> markNotificationRead(String id) async {
    try {
      await _dio.put(ApiEndpoints.markNotificationRead(id));
    } catch (e) {
      debugPrint('[ProfileRepository] Error marking notification read: $e');
    }
  }

  /// Fetches unread notifications count from `GET /api/member/notifications/unread/count`
  Future<int> getUnreadCount() async {
    try {
      final response = await _dio.get(ApiEndpoints.unreadNotificationsCount);
      if (response.data is Map && response.data['count'] != null) {
        return int.tryParse(response.data['count'].toString()) ?? 0;
      }
    } catch (_) {}
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
}
