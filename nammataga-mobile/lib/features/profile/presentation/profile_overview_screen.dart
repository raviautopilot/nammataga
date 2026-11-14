import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../domain/member_profile.dart';
import '../providers/profile_provider.dart';

/// Screen 1: Profile Overview
/// Implements the executive deep emerald header banner,
/// membership summary card, and structured navigation list matching the reference design.
class ProfileOverviewScreen extends ConsumerWidget {
  const ProfileOverviewScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profileAsync = ref.watch(memberProfileProvider);
    final subStatusAsync = ref.watch(memberSubscriptionStatusProvider);
    final unreadCountAsync = ref.watch(unreadAnnouncementsCountProvider);
    final unreadCount = unreadCountAsync.value ?? 0;

    return Scaffold(
      backgroundColor: const Color(0xFFF9FAF7),
      body: RefreshIndicator(
        color: AppColors.primary,
        onRefresh: () async {
          HapticFeedback.lightImpact();
          ref.invalidate(memberProfileProvider);
          ref.invalidate(memberSubscriptionStatusProvider);
          ref.invalidate(memberAnnouncementsProvider);
          ref.invalidate(unreadAnnouncementsCountProvider);
        },
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(
            parent: BouncingScrollPhysics(),
          ),
          child: Column(
            children: [
              // 1. Cinematic Deep Emerald Header matching reference & other screen banners
              profileAsync.when(
                data: (profile) => _buildHeader(
                  context,
                  profile,
                ),
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
                      data: (profile) => _buildMembershipCard(
                        context,
                        profile,
                        subStatus: subStatusAsync.value,
                      ),
                      loading: () => _buildMembershipCardSkeleton(),
                      error: (e, s) => const SizedBox.shrink(),
                    ),
                    const SizedBox(height: 16),

                    // Profile Navigation List Card with Official Announcements
                    _buildNavigationList(context, unreadCount: unreadCount),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildHeader(
    BuildContext context,
    MemberProfile profile,
  ) {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        borderRadius: const BorderRadius.vertical(
          bottom: Radius.circular(30),
        ),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF04190E).withValues(alpha: 0.38),
            blurRadius: 22,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      clipBehavior: Clip.antiAlias,
      child: Stack(
        children: [
          // Multi-stop Emerald Brand Gradient matching Events, Resources, Grievance
          Positioned.fill(
            child: Container(
              decoration: const BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  colors: [
                    Color(0xFF04170D), // Deep rich forest obsidian
                    Color(0xFF0A2B1B), // Signature TAGA deep emerald
                    Color(0xFF14452F), // Forest green
                    Color(0xFF1D5A3D), // Vibrant highlight green
                  ],
                  stops: [0.0, 0.35, 0.72, 1.0],
                ),
              ),
            ),
          ),

          // Ambient Top-Right Radial Glow (exact dimensions matching Grievance)
          Positioned(
            top: -45,
            right: -35,
            width: 220,
            height: 220,
            child: Container(
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: RadialGradient(
                  colors: [
                    const Color(0xFF52B788).withValues(alpha: 0.28),
                    Colors.transparent,
                  ],
                ),
              ),
            ),
          ),

          // Ambient Bottom-Left Glow (exact dimensions matching Grievance)
          Positioned(
            bottom: -35,
            left: -25,
            width: 170,
            height: 170,
            child: Container(
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: RadialGradient(
                  colors: [
                    const Color(0xFF2D6A4F).withValues(alpha: 0.32),
                    Colors.transparent,
                  ],
                ),
              ),
            ),
          ),

          // Subtle Watermark Leaf / User Motif
          Positioned(
            right: -10,
            bottom: -15,
            child: IgnorePointer(
              child: Icon(
                LucideIcons.userRound,
                size: 135,
                color: Colors.white.withValues(alpha: 0.045),
              ),
            ),
          ),

          // Foreground Content (matching Grievance dimensions and spacing)
          SafeArea(
            bottom: false,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(20, 10, 20, 26),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Top Row: Back button, Executive Pill Badge, Settings Gear
                  Row(
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
                        scaleDownFactor: 0.90,
                        child: Container(
                          width: 38,
                          height: 38,
                          decoration: BoxDecoration(
                            color: Colors.white.withValues(alpha: 0.14),
                            shape: BoxShape.circle,
                            border: Border.all(
                              color: Colors.white.withValues(alpha: 0.25),
                              width: 1.1,
                            ),
                          ),
                          child: const Center(
                            child: Icon(
                              LucideIcons.arrowLeft,
                              color: Colors.white,
                              size: 18,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),

                      // Pill Badge matching Grievance screen branding pill
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 5,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.white.withValues(alpha: 0.12),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(
                            color: Colors.white.withValues(alpha: 0.20),
                          ),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(
                              LucideIcons.crown,
                              size: 12,
                              color: Color(0xFFFDE68A),
                            ),
                            const SizedBox(width: 6),
                            Text(
                              profile.membershipBadgeTitle,
                              style: const TextStyle(
                                fontSize: 11.5,
                                fontWeight: FontWeight.w600,
                                color: Colors.white,
                                letterSpacing: 0.3,
                              ),
                            ),
                          ],
                        ),
                      ),

                      const Spacer(),

                      // Settings Gear
                      BouncyButton(
                        onTap: () {
                          HapticFeedback.lightImpact();
                          context.push('/profile/settings');
                        },
                        scaleDownFactor: 0.90,
                        child: Container(
                          width: 38,
                          height: 38,
                          decoration: BoxDecoration(
                            color: Colors.white.withValues(alpha: 0.14),
                            shape: BoxShape.circle,
                            border: Border.all(
                              color: Colors.white.withValues(alpha: 0.25),
                              width: 1.1,
                            ),
                          ),
                          child: const Center(
                            child: Icon(
                              LucideIcons.settings,
                              color: Colors.white,
                              size: 18,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),

                  const SizedBox(height: 22),

                  // Category tag matching Grievance banner structure
                  Text(
                    'ASSOCIATION MEMBER PROFILE',
                    style: TextStyle(
                      color: const Color(0xFF95D5B2).withValues(alpha: 0.9),
                      fontSize: 11,
                      fontWeight: FontWeight.w700,
                      letterSpacing: 1.2,
                    ),
                  ),

                  const SizedBox(height: 6),

                  // Member Name (Title) matching Grievance typography & sizing
                  Text(
                    profile.fullNameWithInitial,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 26,
                      fontWeight: FontWeight.w800,
                      letterSpacing: -0.4,
                    ),
                  ),
                  const SizedBox(height: 6),

                  // TAGA ID & Subtitle matching Grievance typography & height
                  Row(
                    children: [
                      if (profile.tagaId.isNotEmpty) ...[
                        Text(
                          'TAGA ID: ${profile.tagaId}',
                          style: const TextStyle(
                            fontSize: 13.5,
                            color: Color(0xFFB7E4C7),
                            fontWeight: FontWeight.w600,
                            letterSpacing: 0.2,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Container(
                          width: 3.5,
                          height: 3.5,
                          decoration: const BoxDecoration(
                            shape: BoxShape.circle,
                            color: Color(0xFF74C69D),
                          ),
                        ),
                        const SizedBox(width: 8),
                      ],
                      Expanded(
                        child: Text(
                          _buildSubtitleText(profile),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            fontSize: 14,
                            color: Colors.white.withValues(alpha: 0.85),
                            height: 1.35,
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _buildSubtitleText(MemberProfile profile) {
    final parts = <String>[];
    if (profile.designation.isNotEmpty) {
      parts.add(profile.designation);
    }
    if (profile.recruitmentBatch.isNotEmpty) {
      parts.add('Batch ${profile.recruitmentBatch}');
    }
    if (parts.isEmpty && profile.workingDistrict.isNotEmpty) {
      parts.add(profile.workingDistrict);
    }
    if (parts.isEmpty) {
      return 'Together for a Stronger Community';
    }
    return parts.join(' • ');
  }

  String _formatValidDate(String? dateStr) {
    if (dateStr == null || dateStr.isEmpty) return '31 Mar 2027';
    try {
      final parts = dateStr.split('-');
      if (parts.length == 3) {
        final year = int.parse(parts[0]);
        final month = int.parse(parts[1]);
        final day = int.parse(parts[2]);
        const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
        final mStr = months[month - 1]; return '$day $mStr $year';
      }
    } catch (_) {}
    return dateStr;
  }

  Widget _buildMembershipCard(
    BuildContext context,
    MemberProfile profile, {
    Map<String, dynamic>? subStatus,
  }) {
    final isSubscribed = profile.isPaid ||
        profile.subscriptionActive ||
        (subStatus != null && subStatus['is_paid'] == true);

    final rawValidUntil = subStatus?['valid_until']?.toString() ??
        subStatus?['next_due_date']?.toString();
    final validTillFormatted = isSubscribed
        ? _formatValidDate(rawValidUntil)
        : '—';

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

          // Key Value Rows matching backend and subscription page data
          _buildSummaryRow(
            label: 'Membership Type',
            value: isSubscribed ? 'Subscriber' : 'General Member',
          ),
          const SizedBox(height: 10),
          _buildSummaryRow(
            label: 'Valid Till',
            value: validTillFormatted,
          ),
          const SizedBox(height: 10),
          _buildSummaryRow(
            label: 'Payment Status',
            value: isSubscribed ? 'Paid' : profile.paymentStatus,
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

  Widget _buildNavigationList(BuildContext context, {int unreadCount = 0}) {
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
            icon: LucideIcons.megaphone,
            title: 'Official Announcements',
            route: '/profile/announcements',
            badgeCount: unreadCount,
          ),
          const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
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
    int? badgeCount,
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
            if (badgeCount != null && badgeCount > 0) ...[
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(
                  color: const Color(0xFFEF4444),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  '$badgeCount new',
                  style: const TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
              ),
              const SizedBox(width: 8),
            ],
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
      height: 220,
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            Color(0xFF04170D),
            Color(0xFF0A2B1B),
            Color(0xFF14452F),
            Color(0xFF1D5A3D),
          ],
          stops: [0.0, 0.35, 0.72, 1.0],
        ),
        borderRadius: BorderRadius.vertical(bottom: Radius.circular(30)),
      ),
      child: const Center(
        child: CircularProgressIndicator(color: Colors.white),
      ),
    );
  }

  Widget _buildHeaderFallback(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 26),
      decoration: const BoxDecoration(
        color: Color(0xFF0D2C1B),
        borderRadius: BorderRadius.vertical(bottom: Radius.circular(30)),
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
