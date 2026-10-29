import 'package:flutter/foundation.dart';

/// Centralized API endpoint configurations with runtime and compile-time environment selectors.
class ApiEndpoints {
  ApiEndpoints._();

  // Environment passed via --dart-define=API_ENV=local|dev|prod
  static const String environment = String.fromEnvironment(
    'API_ENV',
    defaultValue: 'local',
  );

  // Explicit URL passed via --dart-define=API_BASE_URL=...
  static const String definedBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: '',
  );

  /// Resolves the current base URL based on platform and environment
  static String get baseUrl {
    if (definedBaseUrl.isNotEmpty) {
      return definedBaseUrl.replaceAll(RegExp(r'/+$'), '');
    }

    switch (environment.toLowerCase()) {
      case 'prod':
      case 'production':
        return 'https://api.nammataga.com/api';
      case 'dev':
      case 'development':
        return 'https://devapi.nammataga.com/api';
      case 'local':
      default:
        if (kIsWeb) {
          return 'http://localhost:1801/api';
        } else if (defaultTargetPlatform == TargetPlatform.android) {
          // Android Emulator default loopback
          return 'http://10.0.2.2:1801/api';
        } else {
          return 'http://localhost:1801/api';
        }
    }
  }

  // Auth & Profile
  static String get login => '$baseUrl/member/login';
  static String get logout => '$baseUrl/member/logout';
  static String get forgotPassword => '$baseUrl/auth/member-forgot-password';
  static String get changePassword => '$baseUrl/member/change-password';
  static String get profile => '$baseUrl/member/profile';
  static String get editRequest => '$baseUrl/member/edit-request';

  // Membership & Districts
  static String get districts => '$baseUrl/membership/districts';
  static String get applyMembership => '$baseUrl/membership/apply';

  // Public / About
  static String get about => '$baseUrl/public/about';
  static String get aboutStats => '$baseUrl/public/about/stats';
  static String get aboutObjectives => '$baseUrl/public/about/objectives';
  static String get aboutContact => '$baseUrl/public/about/contact';
  static String get logo => '$baseUrl/logo';

  // Office Bearers
  static String get stateExecutive => '$baseUrl/office-bearers/state-executive';
  static String get bearerDistricts => '$baseUrl/office-bearers/districts';
  static String districtBearers(String district) =>
      '$baseUrl/office-bearers/district-office-bearers?district=${Uri.encodeComponent(district)}';

  // Events & Gallery
  static String get upcomingEvents => '$baseUrl/events/upcoming';
  static String get galleryYears => '$baseUrl/gallery/years';
  static String galleryByYear(int year) => '$baseUrl/gallery?year=$year';

  // Subscriptions & Razorpay
  static String get subscriptions => '$baseUrl/subscriptions';
  static String subscriptionStatus(String email) =>
      '$baseUrl/subscriptions/status?email=${Uri.encodeComponent(email)}';
  static String get createSubscriptionOrder =>
      '$baseUrl/subscriptions/create-order';
  static String get verifySubscriptionPayment =>
      '$baseUrl/subscriptions/verify-payment';

  // Notifications / Announcements
  static String get notifications => '$baseUrl/member/notifications';
  static String markNotificationRead(String id) =>
      '$baseUrl/member/notifications/$id/read';
  static String get unreadNotificationsCount =>
      '$baseUrl/member/notifications/unread/count';

  // Grievance
  static String get grievances => '$baseUrl/grievances';
  static String grievanceById(String id) => '$baseUrl/grievances/$id';
  static String get grievanceCategories => '$baseUrl/categories';
  static String get grievancePriorities => '$baseUrl/priorities';

  // Resources
  static String get resourceCategories => '$baseUrl/resources';
  static String documentsByCategory(String categoryId, [String? subcategory]) {
    final sub = subcategory != null
        ? '?subcategory=${Uri.encodeComponent(subcategory)}'
        : '';
    return '$baseUrl/resources/$categoryId$sub';
  }

  static String get externalLinks => '$baseUrl/resources/external-links';

  // TAGA Towers
  static String get towersRooms => '$baseUrl/towers/rooms';
  static String towersAvailability(String date) =>
      '$baseUrl/towers/availability?date=$date';
  static String towersAvailabilityRange(String start, String end) =>
      '$baseUrl/towers/availability-range?start=$start&end=$end';
  static String get towersBookings => '$baseUrl/towers/bookings';
  static String cancelTowersBooking(String id) =>
      '$baseUrl/towers/bookings/$id';
  static String get towersCreateOrder => '$baseUrl/towers/create-order';
  static String get towersVerifyPayment => '$baseUrl/towers/verify-payment';
}
