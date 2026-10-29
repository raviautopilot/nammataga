import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../domain/member_profile.dart';
import '../providers/profile_provider.dart';
import 'widgets/initials_avatar.dart';

/// Screen 1: Profile Overview
/// Implements the cinematic deep emerald header with watermark, circular initials avatar,
/// membership summary card, and structured navigation list matching the reference image.
class ProfileOverviewScreen extends ConsumerWidget {
  const ProfileOverviewScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profileAsync = ref.watch(memberProfileProvider);

    return Scaffold(
      backgroundColor: const Color(0xFFF9FAF7),
      body: RefreshIndicator(
        color: AppColors.primary,
        onRefresh: () async {
          HapticFeedback.lightImpact();
          ref.invalidate(memberProfileProvider);
          ref.invalidate(memberNotificationsProvider);
          ref.invalidate(unreadNotificationsCountProvider);
        },
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(
            parent: BouncingScrollPhysics(),
          ),
          child: Column(
            children: [
              // 1. Cinematic Deep Emerald Header matching reference
              profileAsync.when(
                data: (profile) => _buildHeader(context, profile),
                loading: () => _buildHeaderLoading(context),
                error: (e, s) => _buildHeaderFallback(context),
              ),

              // 2. Body Cards with generous spacing
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 36),
                child: Column(
                  children: [
                    // Membership Summary Card
                    profileAsync.when(
                      data: (profile) => _buildMembershipCard(context, profile),
                      loading: () => _buildMembershipCardSkeleton(),
                      error: (e, s) => const SizedBox.shrink(),
                    ),
                    const SizedBox(height: 16),

                    // Profile Navigation List Card
                    _buildNavigationList(context),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildHeader(BuildContext context, MemberProfile profile) {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            Color(0xFF061A0F),
            Color(0xFF0D2C1B),
            Color(0xFF1B4332),
            Color(0xFF265942),
          ],
        ),
        borderRadius: const BorderRadius.vertical(
          bottom: Radius.circular(32),
        ),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF061A0F).withValues(alpha: 0.3),
            blurRadius: 18,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: SafeArea(
        bottom: false,
        child: Stack(
          children: [
            // Watermark silhouette in top right
            Positioned(
              right: -16,
              top: -12,
              child: Icon(
                LucideIcons.userRound,
                size: 140,
                color: Colors.white.withValues(alpha: 0.05),
              ),
            ),

            Padding(
              padding: const EdgeInsets.fromLTRB(16, 10, 16, 26),
              child: Column(
                children: [
                  // Top Row: Back button, Screen Title, Settings Gear
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      BouncyButton(
                        onTap: () {
                          HapticFeedback.lightImpact();
                          if (context.canPop()) {
                            context.pop();
                          } else {
                            context.go('/home');
                          }
                        },
                        child: Container(
                          width: 38,
                          height: 38,
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            color: Colors.white.withValues(alpha: 0.12),
                            border: Border.all(
                              color: Colors.white.withValues(alpha: 0.2),
                              width: 1,
                            ),
                          ),
                          child: const Icon(
                            LucideIcons.arrowLeft,
                            color: Colors.white,
                            size: 19,
                          ),
                        ),
                      ),
                      const Text(
                        'My Profile',
                        style: TextStyle(
                          fontSize: 18.5,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                          letterSpacing: -0.3,
                        ),
                      ),
                      BouncyButton(
                        onTap: () {
                          HapticFeedback.lightImpact();
                          context.push('/profile/settings');
                        },
                        child: Container(
                          width: 38,
                          height: 38,
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            color: Colors.white.withValues(alpha: 0.12),
                            border: Border.all(
                              color: Colors.white.withValues(alpha: 0.2),
                              width: 1,
                            ),
                          ),
                          child: const Icon(
                            LucideIcons.settings,
                            color: Colors.white,
                            size: 19,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 18),

                  // Circular Initials Avatar (Strictly initials - NO photo upload)
                  InitialsAvatar(
                    initials: profile.avatarInitials,
                    size: 84,
                    fontSize: 28,
                  ),
                  const SizedBox(height: 14),

                  // Member Name
                  Text(
                    profile.fullNameWithInitial,
                    textAlign: TextAlign.center,
                    style: const TextStyle(
                      fontSize: 21,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                      letterSpacing: -0.3,
                    ),
                  ),
                  const SizedBox(height: 4),

                  // TAGA ID
                  if (profile.tagaId.isNotEmpty)
                    Text(
                      'TAGA ID: ${profile.tagaId}',
                      style: const TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFFB7E4C7),
                        letterSpacing: 0.3,
                      ),
                    ),
                  const SizedBox(height: 12),

                  // Subscriber Crown Badge Pill
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 14,
                      vertical: 5.5,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFEF3C7),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(
                        color: const Color(0xFFFDE68A),
                        width: 1.1,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withValues(alpha: 0.08),
                          blurRadius: 8,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(
                          LucideIcons.crown,
                          size: 13.5,
                          color: Color(0xFFD97706),
                        ),
                        const SizedBox(width: 6),
                        Text(
                          profile.membershipBadgeTitle,
                          style: const TextStyle(
                            fontSize: 12,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF92400E),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 12),

                  // Association Slogan / Motto matching reference
                  const Text(
                    'Together for a Stronger Community',
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      fontSize: 12,
                      fontStyle: FontStyle.italic,
                      fontWeight: FontWeight.w500,
                      color: Color(0xFFD8F3DC),
                      letterSpacing: 0.2,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMembershipCard(BuildContext context, MemberProfile profile) {
    final isSubscribed = profile.isPaid || profile.subscriptionActive;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
        boxShadow: const [
          BoxShadow(
            color: Color(0x061B4332),
            blurRadius: 16,
            offset: Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Row(
                  children: [
                    Container(
                      width: 38,
                      height: 38,
                      decoration: BoxDecoration(
                        color: const Color(0xFFFEF3C7),
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(
                          color: const Color(0xFFFDE68A),
                          width: 1,
                        ),
                      ),
                      child: const Icon(
                        LucideIcons.crown,
                        size: 19,
                        color: Color(0xFFD97706),
                      ),
                    ),
                    const SizedBox(width: 10),
                    const Flexible(
                      child: Text(
                        'Membership Status',
                        overflow: TextOverflow.ellipsis,
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.bold,
                          color: AppColors.textPrimary,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 4,
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
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Container(
                      width: 6,
                      height: 6,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: isSubscribed
                            ? const Color(0xFF2D6A4F)
                            : const Color(0xFFD97706),
                      ),
                    ),
                    const SizedBox(width: 5),
                    Text(
                      isSubscribed ? 'Active' : 'Pending',
                      style: TextStyle(
                        fontSize: 11.5,
                        fontWeight: FontWeight.bold,
                        color: isSubscribed
                            ? const Color(0xFF2D6A4F)
                            : const Color(0xFF92400E),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          const Divider(height: 1, thickness: 1, color: Color(0xFFEFF4F0)),
          const SizedBox(height: 14),

          // Key Value Rows matching reference
          _buildSummaryRow(
            label: 'Membership Type',
            value: isSubscribed ? 'Subscriber' : 'General Member',
          ),
          const SizedBox(height: 10),
          _buildSummaryRow(
            label: 'Valid Till',
            value: isSubscribed ? '31 Dec 2026' : '—',
          ),
          const SizedBox(height: 10),
          _buildSummaryRow(
            label: 'Payment Status',
            value: profile.paymentStatus,
            hasCheckmark: isSubscribed,
          ),
          const SizedBox(height: 16),

          // Primary Manage Membership Button matching reference
          BouncyButton(
            onTap: () {
              HapticFeedback.lightImpact();
              context.push('/profile/subscriptions');
            },
            child: Container(
              width: double.infinity,
              height: 48,
              decoration: BoxDecoration(
                color: const Color(0xFF0D2C1B),
                borderRadius: BorderRadius.circular(14),
                boxShadow: [
                  BoxShadow(
                    color: const Color(0xFF0D2C1B).withValues(alpha: 0.25),
                    blurRadius: 10,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              child: const Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    LucideIcons.crown,
                    size: 15,
                    color: Color(0xFFFDE68A),
                  ),
                  SizedBox(width: 8),
                  Flexible(
                    child: Text(
                      'Manage Membership & Dues',
                      overflow: TextOverflow.ellipsis,
                      maxLines: 1,
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                        letterSpacing: 0.1,
                      ),
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
    );
  }

  Widget _buildSummaryRow({
    required String label,
    required String value,
    bool hasCheckmark = false,
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
          mainAxisSize: MainAxisSize.min,
          children: [
            if (hasCheckmark) ...[
              const Icon(
                LucideIcons.check,
                size: 14,
                color: Color(0xFF2D6A4F),
              ),
              const SizedBox(width: 4),
            ],
            Text(
              value,
              style: TextStyle(
                fontSize: 13.5,
                fontWeight: FontWeight.w600,
                color: hasCheckmark
                    ? const Color(0xFF2D6A4F)
                    : AppColors.textPrimary,
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildNavigationList(BuildContext context) {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
        boxShadow: const [
          BoxShadow(
            color: Color(0x061B4332),
            blurRadius: 16,
            offset: Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: [
          _buildNavItem(
            context,
            icon: LucideIcons.user,
            title: 'Personal Information',
            route: '/profile/personal',
          ),
          const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
          _buildNavItem(
            context,
            icon: LucideIcons.phone,
            title: 'Contact Details',
            route: '/profile/contact',
          ),
          const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
          _buildNavItem(
            context,
            icon: LucideIcons.briefcase,
            title: 'Employment Details',
            route: '/profile/employment',
          ),
          const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
          _buildNavItem(
            context,
            icon: LucideIcons.mapPin,
            title: 'Address Details',
            route: '/profile/address',
          ),
          const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
          _buildNavItem(
            context,
            icon: LucideIcons.fileText,
            title: 'Official Details',
            route: '/profile/official',
          ),
          const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
          _buildNavItem(
            context,
            icon: LucideIcons.crown,
            title: 'Membership Status & Benefits',
            route: '/profile/membership',
          ),
          const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
          _buildNavItem(
            context,
            icon: LucideIcons.creditCard,
            title: 'Subscription Plans & Dues',
            route: '/profile/subscriptions',
          ),
          const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
          _buildNavItem(
            context,
            icon: LucideIcons.filePenLine,
            title: 'Request Profile Edit',
            route: '/profile/edit-request',
          ),
        ],
      ),
    );
  }

  Widget _buildNavItem(
    BuildContext context, {
    required IconData icon,
    required String title,
    required String route,
  }) {
    return BouncyButton(
      onTap: () {
        HapticFeedback.lightImpact();
        context.push(route);
      },
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        child: Row(
          children: [
            Container(
              width: 38,
              height: 38,
              decoration: BoxDecoration(
                color: const Color(0xFFE8F5EE),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(
                  color: const Color(0xFFCBE3D5),
                  width: 1,
                ),
              ),
              child: Icon(icon, size: 18, color: AppColors.primary),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Text(
                title,
                style: const TextStyle(
                  fontSize: 14.5,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                  letterSpacing: -0.1,
                ),
              ),
            ),
            const Icon(
              LucideIcons.chevronRight,
              size: 18,
              color: Color(0xFF9CA3AF),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHeaderLoading(BuildContext context) {
    return Container(
      width: double.infinity,
      height: 260,
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            Color(0xFF061A0F),
            Color(0xFF0D2C1B),
            Color(0xFF1B4332),
          ],
        ),
        borderRadius: BorderRadius.vertical(bottom: Radius.circular(32)),
      ),
      child: const Center(
        child: CircularProgressIndicator(color: Colors.white),
      ),
    );
  }

  Widget _buildHeaderFallback(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(24),
      decoration: const BoxDecoration(
        color: Color(0xFF0D2C1B),
        borderRadius: BorderRadius.vertical(bottom: Radius.circular(32)),
      ),
      child: const SafeArea(
        child: Center(
          child: Text(
            'Unable to load profile data',
            style: TextStyle(color: Colors.white),
          ),
        ),
      ),
    );
  }

  Widget _buildMembershipCardSkeleton() {
    return Container(
      height: 140,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFE2EBE5)),
      ),
      child: const Center(
        child: CircularProgressIndicator(color: AppColors.primary),
      ),
    );
  }
}
