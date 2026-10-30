import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../../profile/providers/profile_provider.dart';

/// Premium introductory screen for the NammaTAGA Grievance Redressal feature.
///
/// Features a bespoke emerald gradient header with a watermark motif,
/// attractive civic and balance scale visual elements with ambient glow,
/// structured member benefit cards, membership eligibility evaluation,
/// and smooth entrance animations.
class GrievanceIntroScreen extends ConsumerWidget {
  const GrievanceIntroScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    // Check membership eligibility across profile and auth session providers
    final isPaid = ref.watch(memberProfileProvider).value?.isPaid ??
        ref.watch(authProvider).value?.isPaid ??
        false;

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAF7),
      body: RefreshIndicator(
        color: AppColors.primary,
        backgroundColor: Colors.white,
        onRefresh: () async {
          HapticFeedback.lightImpact();
          ref.invalidate(memberProfileProvider);
          ref.invalidate(authProvider);
        },
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(
            parent: BouncingScrollPhysics(),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // 1. Bespoke NammaTAGA Emerald Gradient Header
              _buildHeader(context),

              // 2. Content Body
              Padding(
                padding: const EdgeInsets.fromLTRB(20, 22, 20, 24),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // Hero Visual Emblem (Shield, Scales, Message Heart)
                    _buildHeroVisual()
                        .animate()
                        .fadeIn(
                          duration: const Duration(milliseconds: 600),
                          delay: const Duration(milliseconds: 150),
                        )
                        .scale(
                          begin: const Offset(0.88, 0.88),
                          end: const Offset(1, 1),
                          curve: Curves.easeOutBack,
                        ),

                    const SizedBox(height: 24),

                    // Section Header
                    Text(
                      'WHY CHOOSE TAGA REDRESSAL',
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        letterSpacing: 1.1,
                        color: AppColors.textSecondary.withValues(alpha: 0.8),
                      ),
                    ).animate().fadeIn(
                          duration: const Duration(milliseconds: 400),
                          delay: const Duration(milliseconds: 200),
                        ),

                    const SizedBox(height: 12),

                    // Benefit Card 1: Share Your Issue
                    _buildBenefitCard(
                      icon: LucideIcons.send,
                      title: 'Share Your Issue',
                      description:
                          'Direct channel to association office bearers and committee for rapid review.',
                      iconBg: const Color(0xFFE8F5E9),
                      iconColor: const Color(0xFF2D6A4F),
                    )
                        .animate()
                        .fadeIn(
                          duration: const Duration(milliseconds: 450),
                          delay: const Duration(milliseconds: 250),
                        )
                        .slideY(
                          begin: 0.12,
                          end: 0,
                          curve: Curves.easeOut,
                        ),

                    const SizedBox(height: 12),

                    // Benefit Card 2: For a Better Community
                    _buildBenefitCard(
                      icon: LucideIcons.users,
                      title: 'For a Better Community',
                      description:
                          'Constructive resolution of department, posting, pay, and service matters.',
                      iconBg: const Color(0xFFE0F2FE),
                      iconColor: const Color(0xFF0369A1),
                    )
                        .animate()
                        .fadeIn(
                          duration: const Duration(milliseconds: 450),
                          delay: const Duration(milliseconds: 350),
                        )
                        .slideY(
                          begin: 0.12,
                          end: 0,
                          curve: Curves.easeOut,
                        ),

                    const SizedBox(height: 12),

                    // Benefit Card 3: Available for Paid Members
                    _buildBenefitCard(
                      icon: LucideIcons.crown,
                      title: 'Available for Paid Members',
                      description:
                          'Dedicated redressal priority and tracking for subscribed NammaTAGA members.',
                      iconBg: const Color(0xFFFEF3C7),
                      iconColor: const Color(0xFFB45309),
                      trailing: isPaid
                          ? _buildUnlockedPill()
                          : _buildSubscribersOnlyPill(),
                    )
                        .animate()
                        .fadeIn(
                          duration: const Duration(milliseconds: 450),
                          delay: const Duration(milliseconds: 450),
                        )
                        .slideY(
                          begin: 0.12,
                          end: 0,
                          curve: Curves.easeOut,
                        ),

                    const SizedBox(height: 18),

                    // Eligibility Context Banner
                    _buildEligibilityBanner(isPaid)
                        .animate()
                        .fadeIn(
                          duration: const Duration(milliseconds: 450),
                          delay: const Duration(milliseconds: 520),
                        )
                        .slideY(
                          begin: 0.12,
                          end: 0,
                          curve: Curves.easeOut,
                        ),

                    const SizedBox(height: 16),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
      bottomNavigationBar: _buildBottomBar(context, isPaid),
    );
  }

  /// Builds the bespoke NammaTAGA multi-stop emerald gradient header
  /// with 30px curved bottom radius, subtle leaf watermark, back button,
  /// branding title, and soft ambient radial glows.
  Widget _buildHeader(BuildContext context) {
    return Container(
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
          // Multi-stop Emerald Gradient (#04170D -> #0A2B1B -> #14452F -> #1D5A3D)
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

          // Ambient Radial Light Glow (Top-Right)
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

          // Ambient Soft Radial Glow (Bottom-Left)
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

          // Subtle Leaf Watermark Motif
          Positioned(
            right: -10,
            bottom: -15,
            child: IgnorePointer(
              child: Icon(
                LucideIcons.leaf,
                size: 135,
                color: Colors.white.withValues(alpha: 0.045),
              ),
            ),
          ),

          // Header Foreground Content
          SafeArea(
            bottom: false,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(20, 10, 20, 26),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Top Row: Back Button & Branding Pill
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
                              size: 18,
                              color: Colors.white,
                            ),
                          ),
                        ),
                      ),
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
                            Container(
                              width: 6,
                              height: 6,
                              decoration: const BoxDecoration(
                                color: Color(0xFF52B788),
                                shape: BoxShape.circle,
                              ),
                            ),
                            const SizedBox(width: 6),
                            const Text(
                              'NammaTAGA Redressal',
                              style: TextStyle(
                                color: Colors.white,
                                fontSize: 11.5,
                                fontWeight: FontWeight.w600,
                                letterSpacing: 0.3,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ).animate().fadeIn(
                        duration: const Duration(milliseconds: 400),
                      ),

                  const SizedBox(height: 22),

                  // Category tag
                  Text(
                    'ASSOCIATION MEMBER CARE',
                    style: TextStyle(
                      color: const Color(0xFF95D5B2).withValues(alpha: 0.9),
                      fontSize: 11,
                      fontWeight: FontWeight.w700,
                      letterSpacing: 1.2,
                    ),
                  ).animate().fadeIn(
                        duration: const Duration(milliseconds: 450),
                        delay: const Duration(milliseconds: 80),
                      ),

                  const SizedBox(height: 6),

                  // Title: "Grievance Redressal"
                  const Text(
                    'Grievance Redressal',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 26,
                      fontWeight: FontWeight.w800,
                      letterSpacing: -0.4,
                    ),
                  ).animate().fadeIn(
                        duration: const Duration(milliseconds: 500),
                        delay: const Duration(milliseconds: 100),
                      ),

                  const SizedBox(height: 6),

                  // Subtitle: "Raise your concern. We are here to listen."
                  Text(
                    'Raise your concern. We are here to listen.',
                    style: TextStyle(
                      color: Colors.white.withValues(alpha: 0.85),
                      fontSize: 14,
                      height: 1.35,
                    ),
                  ).animate().fadeIn(
                        duration: const Duration(milliseconds: 500),
                        delay: const Duration(milliseconds: 140),
                      ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// Builds the attractive hero visual badge incorporating civic shield,
  /// balance scales, and message heart badges with ambient emerald glow and soft lighting.
  Widget _buildHeroVisual() {
    return Center(
      child: SizedBox(
        width: 220,
        height: 156,
        child: Stack(
          alignment: Alignment.center,
          children: [
            // Ambient emerald glow background
            Container(
              width: 170,
              height: 170,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: RadialGradient(
                  colors: [
                    const Color(0xFF52B788).withValues(alpha: 0.22),
                    const Color(0xFF2D6A4F).withValues(alpha: 0.08),
                    Colors.transparent,
                  ],
                  stops: const [0.0, 0.55, 1.0],
                ),
              ),
            ),

            // Soft outer halo ring
            Container(
              width: 136,
              height: 136,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                border: Border.all(
                  color: const Color(0xFF74C69D).withValues(alpha: 0.22),
                  width: 1.4,
                ),
              ),
            ),

            // Central Medallion: Balance Scales
            Container(
              width: 96,
              height: 96,
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  colors: [
                    Color(0xFFFFFFFF),
                    Color(0xFFF1F8E9),
                    Color(0xFFD8F3DC),
                  ],
                ),
                borderRadius: BorderRadius.circular(28),
                border: Border.all(
                  color: const Color(0xFF95D5B2).withValues(alpha: 0.65),
                  width: 1.6,
                ),
                boxShadow: [
                  BoxShadow(
                    color: const Color(0xFF1B4332).withValues(alpha: 0.12),
                    blurRadius: 18,
                    offset: const Offset(0, 8),
                  ),
                  BoxShadow(
                    color: const Color(0xFF52B788).withValues(alpha: 0.16),
                    blurRadius: 10,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              child: const Center(
                child: Icon(
                  LucideIcons.scale,
                  size: 46,
                  color: Color(0xFF1B4332),
                ),
              ),
            ),

            // Civic Shield Badge (Floating Top-Right)
            Positioned(
              top: 8,
              right: 26,
              child: Container(
                width: 42,
                height: 42,
                decoration: BoxDecoration(
                  color: Colors.white,
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: const Color(0xFF52B788).withValues(alpha: 0.4),
                    width: 1.4,
                  ),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xFF1B4332).withValues(alpha: 0.10),
                      blurRadius: 10,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: const Center(
                  child: Icon(
                    LucideIcons.shieldCheck,
                    size: 21,
                    color: Color(0xFF2D6A4F),
                  ),
                ),
              ),
            ),

            // Message Heart Badge (Floating Bottom-Left)
            Positioned(
              bottom: 8,
              left: 26,
              child: Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [
                      Color(0xFFFFFFFF),
                      Color(0xFFE8F5E9),
                    ],
                  ),
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: const Color(0xFF74C69D).withValues(alpha: 0.45),
                    width: 1.4,
                  ),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xFF1B4332).withValues(alpha: 0.10),
                      blurRadius: 10,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: const Center(
                  child: Icon(
                    LucideIcons.messageSquareHeart,
                    size: 19,
                    color: Color(0xFF2D6A4F),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Builds a compact benefit card with leading icon, title, description,
  /// and optional trailing status pill.
  Widget _buildBenefitCard({
    required IconData icon,
    required String title,
    required String description,
    required Color iconBg,
    required Color iconColor,
    Widget? trailing,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: const Color(0xFFE2EBE5),
          width: 1.1,
        ),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF1B4332).withValues(alpha: 0.04),
            blurRadius: 12,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 42,
            height: 42,
            decoration: BoxDecoration(
              color: iconBg,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: iconColor.withValues(alpha: 0.18),
                width: 1,
              ),
            ),
            child: Center(
              child: Icon(
                icon,
                size: 20,
                color: iconColor,
              ),
            ),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        title,
                        style: const TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w700,
                          color: Color(0xFF1B4332),
                          letterSpacing: -0.2,
                        ),
                      ),
                    ),
                    ?trailing,
                  ],
                ),
                const SizedBox(height: 4),
                Text(
                  description,
                  style: const TextStyle(
                    fontSize: 12.5,
                    color: Color(0xFF52796F),
                    height: 1.35,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// Small status pill shown when membership is active / unlocked.
  Widget _buildUnlockedPill() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: const Color(0xFFD8F3DC),
        borderRadius: BorderRadius.circular(10),
      ),
      child: const Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            LucideIcons.check,
            size: 11,
            color: Color(0xFF2D6A4F),
          ),
          SizedBox(width: 3),
          Text(
            'Unlocked',
            style: TextStyle(
              fontSize: 10.5,
              fontWeight: FontWeight.bold,
              color: Color(0xFF2D6A4F),
            ),
          ),
        ],
      ),
    );
  }

  /// Small status pill shown when subscription is required.
  Widget _buildSubscribersOnlyPill() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: const Color(0xFFFEF3C7),
        borderRadius: BorderRadius.circular(10),
      ),
      child: const Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            LucideIcons.lock,
            size: 10,
            color: Color(0xFFB45309),
          ),
          SizedBox(width: 3),
          Text(
            'Subscribers Only',
            style: TextStyle(
              fontSize: 10,
              fontWeight: FontWeight.bold,
              color: Color(0xFFB45309),
            ),
          ),
        ],
      ),
    );
  }

  /// Banner providing contextual clarity regarding the user's membership eligibility.
  Widget _buildEligibilityBanner(bool isPaid) {
    if (isPaid) {
      return Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 11),
        decoration: BoxDecoration(
          color: const Color(0xFFD8F3DC).withValues(alpha: 0.65),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: const Color(0xFF74C69D).withValues(alpha: 0.5),
            width: 1,
          ),
        ),
        child: const Row(
          children: [
            Icon(
              LucideIcons.badgeCheck,
              size: 18,
              color: Color(0xFF2D6A4F),
            ),
            SizedBox(width: 10),
            Expanded(
              child: Text(
                'Membership Active: You have full access to submit and track grievances.',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF1B4332),
                  height: 1.35,
                ),
              ),
            ),
          ],
        ),
      );
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 11),
      decoration: BoxDecoration(
        color: const Color(0xFFFEF3C7).withValues(alpha: 0.75),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: const Color(0xFFF59E0B).withValues(alpha: 0.45),
          width: 1,
        ),
      ),
      child: const Row(
        children: [
          Icon(
            LucideIcons.lock,
            size: 17,
            color: Color(0xFFB45309),
          ),
          SizedBox(width: 10),
          Expanded(
            child: Text(
              'Subscription Required: Grievance redressal is an exclusive feature for active TAGA members.',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: Color(0xFF92400E),
                height: 1.35,
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// Builds the bottom navigation area with the prominent "Raise Grievance" button
  /// and optional subscription navigation fallback for non-paid users.
  Widget _buildBottomBar(BuildContext context, bool isPaid) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border(
          top: BorderSide(
            color: const Color(0xFFE2EBE5).withValues(alpha: 0.9),
            width: 1.0,
          ),
        ),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF1B4332).withValues(alpha: 0.06),
            blurRadius: 16,
            offset: const Offset(0, -4),
          ),
        ],
      ),
      padding: const EdgeInsets.fromLTRB(20, 12, 20, 12),
      child: SafeArea(
        top: false,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            BouncyButton(
              onTap: () {
                HapticFeedback.lightImpact();
                if (isPaid) {
                  context.push('/grievance/form');
                } else {
                  context.push('/grievance/restricted');
                }
              },
              child: Container(
                width: double.infinity,
                height: 52,
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [
                      Color(0xFF0D2C1B),
                      Color(0xFF1B4332),
                    ],
                  ),
                  borderRadius: BorderRadius.circular(16),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xFF0D2C1B).withValues(alpha: 0.35),
                      blurRadius: 16,
                      offset: const Offset(0, 6),
                    ),
                  ],
                  border: Border.all(
                    color: const Color(0xFF40916C).withValues(alpha: 0.4),
                    width: 1.1,
                  ),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(
                      isPaid ? LucideIcons.messageSquarePlus : LucideIcons.lock,
                      size: 20,
                      color: Colors.white,
                    ),
                    const SizedBox(width: 10),
                    const Text(
                      'Raise Grievance',
                      style: TextStyle(
                        color: Colors.white,
                        fontSize: 16,
                        fontWeight: FontWeight.w700,
                        letterSpacing: 0.3,
                      ),
                    ),
                    const SizedBox(width: 8),
                    const Icon(
                      LucideIcons.arrowRight,
                      size: 18,
                      color: Colors.white,
                    ),
                  ],
                ),
              ),
            ),
            if (!isPaid) ...[
              const SizedBox(height: 8),
              BouncyButton(
                onTap: () {
                  HapticFeedback.lightImpact();
                  context.push('/profile/subscriptions');
                },
                child: const Padding(
                  padding: EdgeInsets.symmetric(vertical: 4),
                  child: Text(
                    'Explore Membership Subscriptions',
                    style: TextStyle(
                      fontSize: 12.5,
                      fontWeight: FontWeight.w600,
                      color: AppColors.primary,
                      decoration: TextDecoration.underline,
                    ),
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    )
        .animate()
        .fadeIn(
          duration: const Duration(milliseconds: 450),
          delay: const Duration(milliseconds: 580),
        )
        .slideY(
          begin: 0.18,
          end: 0,
        );
  }
}
