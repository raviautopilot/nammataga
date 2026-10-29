import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../domain/member_profile.dart';
import '../providers/profile_provider.dart';
import 'widgets/initials_avatar.dart';

class ProfileOverviewScreen extends ConsumerWidget {
  const ProfileOverviewScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessionAsync = ref.watch(authProvider);
    final session = sessionAsync.value;
    final isAuthenticated = session != null && session.isAuthenticated;

    if (!isAuthenticated) {
      return _buildGuestScreen(context);
    }

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
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 32),
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
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            Color(0xFF092014),
            Color(0xFF133924),
            Color(0xFF1B4332),
            Color(0xFF2D6A4F),
          ],
        ),
        borderRadius: BorderRadius.vertical(bottom: Radius.circular(28)),
      ),
      child: SafeArea(
        bottom: false,
        child: Stack(
          children: [
            // Architectural guide lines / ambient glow
            Positioned(
              right: -20,
              top: -20,
              child: Container(
                width: 140,
                height: 140,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: RadialGradient(
                    colors: [
                      const Color(0xFF52B788).withValues(alpha: 0.2),
                      Colors.transparent,
                    ],
                  ),
                ),
              ),
            ),

            Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
              child: Column(
                children: [
                  // Top Row: Back button, Screen Title, Settings Gear
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      BouncyButton(
                        onTap: () {
                          if (context.canPop()) {
                            context.pop();
                          } else {
                            context.go('/home');
                          }
                        },
                        child: Container(
                          padding: const EdgeInsets.all(8),
                          child: const Icon(
                            LucideIcons.arrowLeft,
                            color: Colors.white,
                            size: 22,
                          ),
                        ),
                      ),
                      const Text(
                        'My Profile',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                          letterSpacing: -0.2,
                        ),
                      ),
                      BouncyButton(
                        onTap: () {
                          context.push('/profile/settings');
                        },
                        child: Container(
                          padding: const EdgeInsets.all(8),
                          child: const Icon(
                            LucideIcons.settings,
                            color: Colors.white,
                            size: 22,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),

                  // Circular Initials Avatar (Strictly initials - NO image upload)
                  InitialsAvatar(
                    initials: profile.avatarInitials,
                    size: 80,
                    fontSize: 28,
                  ),
                  const SizedBox(height: 12),

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
                  const SizedBox(height: 3),

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
                  const SizedBox(height: 10),

                  // Subscriber Crown Badge Pill
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 5,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFEF3C7),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(
                        color: const Color(0xFFFDE68A),
                        width: 1.1,
                      ),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(
                          LucideIcons.crown,
                          size: 13,
                          color: Color(0xFFD97706),
                        ),
                        const SizedBox(width: 6),
                        Text(
                          profile.membershipBadgeTitle,
                          style: const TextStyle(
                            fontSize: 11.5,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF92400E),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 10),

                  // Association Motto
                  const Text(
                    '"Together for a Stronger Community"',
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      fontSize: 12,
                      fontStyle: FontStyle.italic,
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
            blurRadius: 14,
            offset: Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Expanded(
                child: Row(
                  children: [
                    Icon(LucideIcons.crown, size: 17, color: Color(0xFFD97706)),
                    SizedBox(width: 8),
                    Flexible(
                      child: Text(
                        'Membership Status',
                        style: TextStyle(
                          fontSize: 14.5,
                          fontWeight: FontWeight.bold,
                          color: AppColors.textPrimary,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 8,
                  vertical: 2.5,
                ),
                decoration: BoxDecoration(
                  color: isSubscribed
                      ? const Color(0xFFE8F5EE)
                      : const Color(0xFFFEF3C7),
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(
                    color: isSubscribed
                        ? const Color(0xFF52B788)
                        : const Color(0xFFF59E0B),
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
          const SizedBox(height: 14),
          const Divider(height: 1, color: Color(0xFFEFF4F0)),
          const SizedBox(height: 12),
          _buildSummaryRow(
            'Membership Type',
            isSubscribed ? 'Subscriber' : 'General Member',
          ),
          const SizedBox(height: 8),
          _buildSummaryRow(
            'Payment Status',
            profile.paymentStatus,
            isPositive: isSubscribed,
          ),
          const SizedBox(height: 14),
          BouncyButton(
            onTap: () {
              context.push('/profile/subscriptions');
            },
            child: Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(vertical: 10),
              decoration: BoxDecoration(
                color: AppColors.primaryDark,
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(LucideIcons.creditCard, size: 14, color: Color(0xFFFDE68A)),
                  SizedBox(width: 6),
                  Text(
                    'Manage Membership & Dues',
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                  SizedBox(width: 5),
                  Icon(LucideIcons.arrowRight, size: 14, color: Colors.white),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildNavigationList(BuildContext context) {
    final navItems = [
      _NavItem(
        title: 'Subscription Plans & Dues',
        icon: LucideIcons.creditCard,
        route: '/profile/subscriptions',
      ),
      _NavItem(
        title: 'Membership Status & Benefits',
        icon: LucideIcons.crown,
        route: '/profile/membership',
      ),
      _NavItem(
        title: 'Personal Information',
        icon: LucideIcons.user,
        route: '/profile/personal',
      ),
      _NavItem(
        title: 'Contact Details',
        icon: LucideIcons.phone,
        route: '/profile/contact',
      ),
      _NavItem(
        title: 'Employment Details',
        icon: LucideIcons.briefcase,
        route: '/profile/employment',
      ),
      _NavItem(
        title: 'Address Details',
        icon: LucideIcons.mapPin,
        route: '/profile/address',
      ),
      _NavItem(
        title: 'Official Details',
        icon: LucideIcons.fileBadge,
        route: '/profile/official',
      ),
      _NavItem(
        title: 'Request Profile Edit',
        icon: LucideIcons.filePenLine,
        route: '/profile/edit-request',
      ),
    ];

    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
        boxShadow: const [
          BoxShadow(
            color: Color(0x061B4332),
            blurRadius: 14,
            offset: Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: List.generate(navItems.length, (index) {
          final item = navItems[index];
          final isLast = index == navItems.length - 1;

          return Column(
            children: [
              BouncyButton(
                onTap: () => context.push(item.route),
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 14,
                  ),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(7),
                        decoration: BoxDecoration(
                          color: const Color(0xFFE8F5EE),
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: Icon(
                          item.icon,
                          size: 17,
                          color: AppColors.primary,
                        ),
                      ),
                      const SizedBox(width: 14),
                      Expanded(
                        child: Text(
                          item.title,
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: AppColors.textPrimary,
                          ),
                        ),
                      ),
                      const Icon(
                        LucideIcons.chevronRight,
                        size: 17,
                        color: Color(0xFF9CA3AF),
                      ),
                    ],
                  ),
                ),
              ),
              if (!isLast)
                const Divider(height: 1, indent: 54, color: Color(0xFFEFF4F0)),
            ],
          );
        }),
      ),
    );
  }

  Widget _buildSummaryRow(String label, String value, {bool? isPositive}) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 12.5,
            color: Color(0xFF5C7164),
            fontWeight: FontWeight.w500,
          ),
        ),
        Row(
          children: [
            if (isPositive != null) ...[
              Icon(
                isPositive ? LucideIcons.circleCheck : LucideIcons.circleAlert,
                size: 13,
                color: isPositive ? AppColors.primary : const Color(0xFFD97706),
              ),
              const SizedBox(width: 5),
            ],
            Text(
              value.isNotEmpty ? value : '—',
              style: const TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildHeaderLoading(BuildContext context) {
    return Container(
      width: double.infinity,
      height: 240,
      decoration: const BoxDecoration(
        color: Color(0xFF092014),
        borderRadius: BorderRadius.vertical(bottom: Radius.circular(28)),
      ),
      child: const Center(
        child: CircularProgressIndicator(color: AppColors.primary),
      ),
    );
  }

  Widget _buildHeaderFallback(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(24),
      decoration: const BoxDecoration(
        color: Color(0xFF092014),
        borderRadius: BorderRadius.vertical(bottom: Radius.circular(28)),
      ),
      child: const Center(
        child: Text(
          'My Profile',
          style: TextStyle(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: Colors.white,
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
    );
  }

  Widget _buildGuestScreen(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF9FAF7),
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text('Member Profile'),
        elevation: 0,
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 72,
                height: 72,
                decoration: const BoxDecoration(
                  color: Color(0xFFE8F5EE),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  LucideIcons.lock,
                  color: AppColors.primary,
                  size: 32,
                ),
              ),
              const SizedBox(height: 16),
              const Text(
                'Member Access Required',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
              const SizedBox(height: 6),
              const Text(
                'Profile details, government records, and membership options are exclusively available to verified TAGA members.',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 13,
                  color: AppColors.textSecondary,
                  height: 1.45,
                ),
              ),
              const SizedBox(height: 24),
              BouncyButton(
                onTap: () => context.go('/login'),
                child: Container(
                  width: double.infinity,
                  height: 48,
                  decoration: BoxDecoration(
                    color: AppColors.primary,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Center(
                    child: Text(
                      'Sign In to Your Account',
                      style: TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.bold,
                        fontSize: 14.5,
                      ),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _NavItem {
  final String title;
  final IconData icon;
  final String route;

  const _NavItem({
    required this.title,
    required this.icon,
    required this.route,
  });
}
