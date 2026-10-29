import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../auth/providers/auth_provider.dart';
import '../data/profile_repository.dart';
import '../domain/member_notification_model.dart';
import '../domain/member_profile.dart';
import '../domain/subscription_plan.dart';

final profileRepositoryProvider = Provider<ProfileRepository>((ref) {
  return ProfileRepository();
});

/// FutureProvider that fetches the latest MemberProfile
final memberProfileProvider = FutureProvider.autoDispose<MemberProfile>((
  ref,
) async {
  final repo = ref.watch(profileRepositoryProvider);
  return repo.getProfile();
});

/// FutureProvider that fetches the 38 TN administrative districts
final districtsListProvider = FutureProvider<List<String>>((ref) async {
  final repo = ref.watch(profileRepositoryProvider);
  return repo.getDistricts();
});

/// FutureProvider that fetches all available subscription plans
final subscriptionPlansProvider =
    FutureProvider.autoDispose<List<SubscriptionPlan>>((ref) async {
      final repo = ref.watch(profileRepositoryProvider);
      return repo.getSubscriptions();
    });

/// FutureProvider for the current member's subscription status
final memberSubscriptionStatusProvider =
    FutureProvider.autoDispose<Map<String, dynamic>?>((ref) async {
      final repo = ref.watch(profileRepositoryProvider);
      final session = ref.watch(authProvider).value;
      final email = session?.memberEmail ?? '';
      if (email.isEmpty) return null;
      return repo.getMemberSubscriptionStatus(email);
    });

/// FutureProvider for the list of plan IDs already paid by the member
final memberPaidSubscriptionsProvider =
    FutureProvider.autoDispose<List<String>>((ref) async {
      final repo = ref.watch(profileRepositoryProvider);
      final session = ref.watch(authProvider).value;
      final email = session?.memberEmail ?? '';
      if (email.isEmpty) return [];
      return repo.getMemberPaidSubscriptions(email);
    });

/// FutureProvider for association announcements (replaces legacy notifications)
final memberAnnouncementsProvider =
    FutureProvider.autoDispose<List<MemberNotification>>((ref) async {
      final repo = ref.watch(profileRepositoryProvider);
      return repo.getNotifications();
    });

/// FutureProvider for unread announcements count
final unreadAnnouncementsCountProvider = FutureProvider.autoDispose<int>((
  ref,
) async {
  final repo = ref.watch(profileRepositoryProvider);
  return repo.getUnreadCount();
});

/// Aliases for backward compatibility
final memberNotificationsProvider = memberAnnouncementsProvider;
final unreadNotificationsCountProvider = unreadAnnouncementsCountProvider;
