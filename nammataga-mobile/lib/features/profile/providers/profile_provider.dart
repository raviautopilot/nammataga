import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../auth/providers/auth_provider.dart';
import '../data/profile_repository.dart';
import '../domain/deletion_request_model.dart';
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

/// FutureProvider for association announcements (active announcements)
final memberAnnouncementsProvider =
    FutureProvider.autoDispose<List<MemberNotification>>((ref) async {
      final repo = ref.watch(profileRepositoryProvider);
      final session = ref.watch(authProvider).value;
      final profile = ref.watch(memberProfileProvider).value;
      final memberId = (profile?.id.isNotEmpty == true)
          ? profile!.id
          : ((session?.memberId.isNotEmpty == true)
              ? session!.memberId
              : session?.memberEmail);
      return repo.getNotifications(memberId: memberId);
    });

/// Compatibility provider: returns 0 as unread tracking is eliminated
final unreadAnnouncementsCountProvider = FutureProvider.autoDispose<int>((
  ref,
) async {
  return 0;
});

/// Aliases for backward compatibility
final memberNotificationsProvider = memberAnnouncementsProvider;
final unreadNotificationsCountProvider = unreadAnnouncementsCountProvider;

/// FutureProvider for checking account deletion eligibility
final accountDeletionEligibilityProvider =
    FutureProvider.autoDispose<DeletionEligibility>((ref) async {
  final repo = ref.watch(profileRepositoryProvider);
  return repo.getAccountDeletionEligibility();
});

/// FutureProvider for checking current account deletion status
final accountDeletionStatusProvider =
    FutureProvider.autoDispose<AccountDeletionRequestStatus>((ref) async {
  final repo = ref.watch(profileRepositoryProvider);
  return repo.getAccountDeletionStatus();
});
