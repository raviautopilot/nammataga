import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../providers/profile_provider.dart';
import 'widgets/profile_subpage_scaffold.dart';

class MembershipDetailsScreen extends ConsumerWidget {
  const MembershipDetailsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profileAsync = ref.watch(memberProfileProvider);

    return ProfileSubpageScaffold(
      title: 'Membership Details',
      watermarkIcon: LucideIcons.crown,
      showEditNotice: false,
      body: profileAsync.when(
        data: (profile) {
          final isSubscribed = profile.isPaid || profile.subscriptionActive;

          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Main Membership Card
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(
                    color: const Color(0xFFE2EBE5),
                    width: 1.1,
                  ),
                  boxShadow: const [
                    BoxShadow(
                      color: Color(0x061B4332),
                      blurRadius: 14,
                      offset: Offset(0, 4),
                    ),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Header Row: Crown Icon & Status
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          width: 44,
                          height: 44,
                          decoration: BoxDecoration(
                            color: const Color(0xFFFEF3C7),
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: const Color(0xFFFDE68A),
                              width: 1.2,
                            ),
                          ),
                          child: const Icon(
                            LucideIcons.crown,
                            color: Color(0xFFD97706),
                            size: 22,
                          ),
                        ),
                        const SizedBox(width: 14),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                mainAxisAlignment:
                                    MainAxisAlignment.spaceBetween,
                                children: [
                                  Text(
                                    profile.membershipBadgeTitle,
                                    style: const TextStyle(
                                      fontSize: 16.5,
                                      fontWeight: FontWeight.bold,
                                      color: AppColors.textPrimary,
                                    ),
                                  ),
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 3,
                                    ),
                                    decoration: BoxDecoration(
                                      color: isSubscribed
                                          ? const Color(0xFFE8F5EE)
                                          : const Color(0xFFFEF3C7),
                                      borderRadius: BorderRadius.circular(12),
                                      border: Border.all(
                                        color: isSubscribed
                                            ? const Color(0xFF52B788)
                                            : const Color(0xFFF59E0B),
                                        width: 1,
                                      ),
                                    ),
                                    child: Text(
                                      isSubscribed ? 'Active' : 'Pending',
                                      style: TextStyle(
                                        fontSize: 11,
                                        fontWeight: FontWeight.bold,
                                        color: isSubscribed
                                            ? AppColors.primary
                                            : const Color(0xFFB45309),
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 4),
                              Text(
                                isSubscribed
                                    ? 'Your membership is active and in good standing.'
                                    : 'Subscribe to access full association member privileges.',
                                style: const TextStyle(
                                  fontSize: 12,
                                  color: AppColors.textSecondary,
                                  height: 1.35,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 18),
                    const Divider(height: 1, color: Color(0xFFEFF4F0)),
                    const SizedBox(height: 14),

                    // Details
                    _buildStatusRow(
                      label: 'Membership Type',
                      value: isSubscribed ? 'Subscriber' : 'General Member',
                    ),
                    const SizedBox(height: 12),
                    _buildStatusRow(
                      label: 'Payment Status',
                      value: profile.paymentStatus,
                      isPositive: isSubscribed,
                    ),
                    const SizedBox(height: 18),

                    // Manage Button
                    BouncyButton(
                      onTap: () {
                        context.push('/profile/subscriptions');
                      },
                      child: Container(
                        width: double.infinity,
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        decoration: BoxDecoration(
                          color: AppColors.primaryDark,
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: const Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              LucideIcons.creditCard,
                              size: 15,
                              color: Color(0xFFFDE68A),
                            ),
                            SizedBox(width: 8),
                            Text(
                              'View Subscription Plans & Pay',
                              style: TextStyle(
                                fontSize: 13.5,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                              ),
                            ),
                            SizedBox(width: 6),
                            Icon(
                              LucideIcons.arrowRight,
                              size: 15,
                              color: Colors.white,
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 20),

              // Subscriber Benefits Card
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(
                    color: const Color(0xFFE2EBE5),
                    width: 1.1,
                  ),
                  boxShadow: const [
                    BoxShadow(
                      color: Color(0x061B4332),
                      blurRadius: 14,
                      offset: Offset(0, 4),
                    ),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Row(
                      children: [
                        Icon(
                          LucideIcons.shieldCheck,
                          size: 18,
                          color: AppColors.primary,
                        ),
                        SizedBox(width: 8),
                        Text(
                          'Subscriber Benefits',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: AppColors.textPrimary,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    _buildBenefitItem('Access to exclusive resources & bylaws'),
                    _buildBenefitItem(
                      'Priority event and conference registrations',
                    ),
                    _buildBenefitItem(
                      'Direct grievance petitions to state leadership',
                    ),
                    _buildBenefitItem('TAGA Towers room and dorm bookings'),
                    _buildBenefitItem(
                      'Association bulletins & community updates',
                    ),
                  ],
                ),
              ),
            ],
          );
        },
        loading: () => const Center(
          child: Padding(
            padding: EdgeInsets.all(40),
            child: CircularProgressIndicator(color: AppColors.primary),
          ),
        ),
        error: (error, stack) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Text(
              'Unable to load membership details.\n$error',
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: AppColors.destructive,
                fontSize: 13,
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildStatusRow({
    required String label,
    required String value,
    bool? isPositive,
  }) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 13,
            color: Color(0xFF5C7164),
            fontWeight: FontWeight.w500,
          ),
        ),
        Row(
          children: [
            if (isPositive != null) ...[
              Icon(
                isPositive ? LucideIcons.circleCheck : LucideIcons.circleAlert,
                size: 14,
                color: isPositive ? AppColors.primary : const Color(0xFFD97706),
              ),
              const SizedBox(width: 5),
            ],
            Text(
              value,
              style: const TextStyle(
                fontSize: 13.5,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildBenefitItem(String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(3),
            decoration: const BoxDecoration(
              color: Color(0xFFE8F5EE),
              shape: BoxShape.circle,
            ),
            child: const Icon(
              LucideIcons.check,
              size: 13,
              color: AppColors.primary,
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              text,
              style: const TextStyle(
                fontSize: 13,
                color: AppColors.textSecondary,
                height: 1.35,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
