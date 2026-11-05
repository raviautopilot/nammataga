import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../tagatowers/presentation/theme/tagatowers_design_tokens.dart';

/// Dedicated access guard screen displayed when an unpaid member or guest opens Resources.
///
/// Faithfully reproduces the website's access-control messaging and aesthetic from
/// `taga-web/src/components/Resources.tsx`.
class ResourcesAccessRestrictedScreen extends StatelessWidget {
  /// Custom callback triggered when the user taps "Manage Subscription".
  final VoidCallback? onUpgrade;

  /// Custom callback triggered when the user taps "Back to Home" or the top back button.
  final VoidCallback? onBackToHome;

  /// Whether to render the custom scaffold AppBar. Defaults to true.
  final bool showAppBar;

  const ResourcesAccessRestrictedScreen({
    super.key,
    this.onUpgrade,
    this.onBackToHome,
    this.showAppBar = true,
  });

  /// Key benefits exclusive to paid subscribers.
  static const List<String> benefitsList = [
    'Access to official Tamil Nadu Agricultural Engineering GOs',
    'Scheme guidelines, department orders, and job charts',
    'Direct download of official leave and application formats',
  ];

  void _handleBack(BuildContext context) {
    if (onBackToHome != null) {
      onBackToHome!();
      return;
    }

    try {
      if (context.canPop()) {
        context.pop();
        return;
      }
    } catch (_) {
      // GoRouter not mounted in unit testing context
    }

    try {
      context.go('/home');
    } catch (_) {
      Navigator.of(context).maybePop();
    }
  }

  void _handleUpgrade(BuildContext context) {
    if (onUpgrade != null) {
      onUpgrade!();
      return;
    }

    try {
      context.push('/profile/subscriptions');
    } catch (_) {
      Navigator.of(context).pushNamed('/profile/subscriptions');
    }
  }

  @override
  Widget build(BuildContext context) {
    final content = SafeArea(
      child: Center(
        child: SingleChildScrollView(
          physics: const BouncingScrollPhysics(),
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 20),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Top Hero Banner (matches taga-web Resources banner overlay)
                _buildHeroBanner(),
                const SizedBox(height: 18),

                // Explanatory Yellow/Amber Card (matches taga-web locked Card)
                _buildExplanatoryCard(),
                const SizedBox(height: 18),

                // Subscriber Privileges Pill List
                _buildBenefitsCard(),
                const SizedBox(height: 24),

                // Primary & Secondary Buttons
                _buildActionButtons(context),
              ],
            ),
          ),
        ),
      ),
    );

    if (!showAppBar) {
      return content;
    }

    return Scaffold(
      backgroundColor: const Color(0xFFF9FAF7),
      appBar: PreferredSize(
        preferredSize: const Size.fromHeight(60),
        child: Container(
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [
                Color(0xFF061A0F),
                Color(0xFF0D2C1B),
                Color(0xFF1B4332),
              ],
            ),
            borderRadius: const BorderRadius.vertical(
              bottom: Radius.circular(20),
            ),
            boxShadow: [
              BoxShadow(
                color: const Color(0xFF061A0F).withValues(alpha: 0.25),
                blurRadius: 10,
                offset: const Offset(0, 3),
              ),
            ],
          ),
          child: SafeArea(
            bottom: false,
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
              child: Row(
                children: [
                  BouncyButton(
                    onTap: () {
                      HapticFeedback.lightImpact();
                      _handleBack(context);
                    },
                    child: Container(
                      width: 38,
                      height: 38,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: Colors.white.withValues(alpha: 0.12),
                        border: Border.all(
                          color: Colors.white.withValues(alpha: 0.22),
                          width: 1,
                        ),
                      ),
                      child: const Center(
                        child: Icon(
                          LucideIcons.arrowLeft,
                          color: Colors.white,
                          size: 19,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 14),
                  const Text(
                    'Resources Center',
                    style: TextStyle(
                      fontSize: 17,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                      letterSpacing: -0.2,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
      body: content,
    );
  }

  /// Top hero banner card mimicking the web green-overlay banner image.
  Widget _buildHeroBanner() {
    return Container(
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            Color(0xFF0B2418),
            Color(0xFF1B4332),
            Color(0xFF2D6A4F),
          ],
        ),
        borderRadius: BorderRadius.circular(22),
        boxShadow: const [
          BoxShadow(
            color: Color(0x181B4332),
            blurRadius: 20,
            offset: Offset(0, 8),
          ),
        ],
      ),
      child: Stack(
        children: [
          // Decorative leaf/document watermark in background
          Positioned(
            right: -16,
            bottom: -16,
            child: Icon(
              LucideIcons.fileText,
              size: 130,
              color: Colors.white.withValues(alpha: 0.05),
            ),
          ),

          // Main banner text content
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 22, vertical: 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Badge: "For Paid Members Only"
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 10,
                    vertical: 5,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xFF40916C),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(
                      color: Colors.white.withValues(alpha: 0.25),
                      width: 1,
                    ),
                  ),
                  child: const Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        LucideIcons.fileText,
                        color: Colors.white,
                        size: 13,
                      ),
                      SizedBox(width: 6),
                      Text(
                        'For Paid Members Only',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 11.5,
                          fontWeight: FontWeight.bold,
                          letterSpacing: 0.2,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 14),

                // Heading
                const Text(
                  'Resources Center',
                  style: TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                    letterSpacing: -0.4,
                  ),
                ),
                const SizedBox(height: 6),

                // Subtitle
                Text(
                  'Official documents, government orders, and agricultural circulars',
                  style: TextStyle(
                    fontSize: 13,
                    color: Colors.white.withValues(alpha: 0.88),
                    height: 1.4,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// Explanatory yellow/amber card exactly matching taga-web styling.
  Widget _buildExplanatoryCard() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFBEB), // bg-yellow-50
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: const Color(0xFFFCD34D), // border-yellow-200
          width: 1.3,
        ),
        boxShadow: const [
          BoxShadow(
            color: Color(0x0C000000),
            blurRadius: 16,
            offset: Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: [
          // Amber Alert Icon
          Container(
            width: 64,
            height: 64,
            decoration: BoxDecoration(
              color: const Color(0xFFFEF3C7),
              shape: BoxShape.circle,
              border: Border.all(
                color: const Color(0xFFF59E0B).withValues(alpha: 0.4),
                width: 1.4,
              ),
            ),
            child: const Center(
              child: Icon(
                LucideIcons.alertCircle,
                size: 32,
                color: Color(0xFFD97706),
              ),
            ),
          ),
          const SizedBox(height: 14),

          // Title: "Subscriber Access Only"
          const Text(
            'Subscriber Access Only',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 19,
              fontWeight: FontWeight.bold,
              color: Color(0xFF78350F),
              letterSpacing: -0.2,
            ),
          ),
          const SizedBox(height: 10),

          // Message
          const Text(
            'Access to TAGA official resources and government orders is reserved for active paid subscribers. Please complete or renew your annual subscription to view resources.',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 13.5,
              height: 1.5,
              color: Color(0xFF92400E),
            ),
          ),
        ],
      ),
    );
  }

  /// Exclusive benefits list card.
  Widget _buildBenefitsCard() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 18),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: TagaTowersDesignTokens.subtleBorder,
          width: 1.2,
        ),
        boxShadow: TagaTowersDesignTokens.softShadow,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Padding(
            padding: EdgeInsets.only(left: 4, bottom: 12),
            child: Text(
              'Included with Active Subscription:',
              style: TextStyle(
                fontSize: 13.5,
                fontWeight: FontWeight.w700,
                color: AppColors.primaryDark,
                letterSpacing: -0.1,
              ),
            ),
          ),
          ...benefitsList.map((benefit) {
            return Padding(
              padding: const EdgeInsets.only(bottom: 9),
              child: Container(
                width: double.infinity,
                padding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 10,
                ),
                decoration: BoxDecoration(
                  color: const Color(0xFFF1F8E9),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: const Color(0xFFB7E4C7).withValues(alpha: 0.7),
                    width: 1.1,
                  ),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Padding(
                      padding: EdgeInsets.only(top: 2),
                      child: Icon(
                        LucideIcons.check,
                        color: AppColors.primary,
                        size: 16,
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        benefit,
                        style: const TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                          color: AppColors.textPrimary,
                          height: 1.35,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            );
          }),
        ],
      ),
    );
  }

  /// Action buttons: Manage Subscription (primary) and Back to Home (secondary).
  Widget _buildActionButtons(BuildContext context) {
    return Column(
      children: [
        // Primary: Manage Subscription
        BouncyButton(
          onTap: () {
            HapticFeedback.mediumImpact();
            _handleUpgrade(context);
          },
          child: Container(
            width: double.infinity,
            height: 50,
            decoration: BoxDecoration(
              color: AppColors.primary,
              borderRadius: BorderRadius.circular(14),
              boxShadow: [
                BoxShadow(
                  color: AppColors.primary.withValues(alpha: 0.28),
                  blurRadius: 10,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            child: const Center(
              child: Row(
                mainAxisSize: MainAxisSize.min,
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    LucideIcons.crown,
                    color: Colors.white,
                    size: 18,
                  ),
                  SizedBox(width: 8),
                  Text(
                    'Manage Subscription',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 15,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 0.1,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
        const SizedBox(height: 10),

        // Secondary: Back to Home
        BouncyButton(
          onTap: () {
            HapticFeedback.lightImpact();
            _handleBack(context);
          },
          child: Container(
            width: double.infinity,
            height: 46,
            decoration: BoxDecoration(
              color: Colors.transparent,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(
                color: AppColors.border,
                width: 1.2,
              ),
            ),
            child: const Center(
              child: Row(
                mainAxisSize: MainAxisSize.min,
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    LucideIcons.arrowLeft,
                    color: AppColors.primaryDark,
                    size: 17,
                  ),
                  SizedBox(width: 8),
                  Text(
                    'Back to Home',
                    style: TextStyle(
                      color: AppColors.primaryDark,
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}
