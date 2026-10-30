import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../providers/grievance_provider.dart';

/// Confirmation screen displayed ONLY after successful grievance submission.
///
/// Provides a reassuring, celebratory confirmation acknowledging receipt by the
/// NammaTAGA state executive committee, and a prominent call-to-action returning
/// the member to the home screen.
class GrievanceSuccessScreen extends ConsumerWidget {
  const GrievanceSuccessScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (didPop) return;
        _navigateHome(context, ref);
      },
      child: Scaffold(
        backgroundColor: AppColors.background,
        body: SafeArea(
          child: LayoutBuilder(
            builder: (context, constraints) {
              return SingleChildScrollView(
                physics: const BouncingScrollPhysics(),
                child: ConstrainedBox(
                  constraints: BoxConstraints(
                    minHeight: constraints.maxHeight,
                  ),
                  child: IntrinsicHeight(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 24.0,
                        vertical: 20.0,
                      ),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Spacer(flex: 1),

                          // 1. Animated Green Checkmark with Glowing Circular Badge
                          _buildAnimatedGlowingCheckBadge(),

                          const SizedBox(height: 32),

                          // 2. Heading
                          const Text(
                            'Grievance Submitted Successfully!',
                            textAlign: TextAlign.center,
                            style: TextStyle(
                              fontSize: 24,
                              fontWeight: FontWeight.w800,
                              color: AppColors.textPrimary,
                              letterSpacing: -0.4,
                              height: 1.25,
                            ),
                          )
                              .animate()
                              .fadeIn(duration: 450.ms, delay: 250.ms)
                              .slideY(
                                begin: 0.2,
                                end: 0,
                                curve: Curves.easeOutCubic,
                              ),

                          const SizedBox(height: 14),

                          // 3. Subtitle / Message
                          const Text(
                            'Thank you for reaching out to NammaTAGA. Your grievance petition has been delivered to the association state executive committee. Our team will review the details and follow up through your preferred response channel.',
                            textAlign: TextAlign.center,
                            style: TextStyle(
                              fontSize: 14.5,
                              color: AppColors.textSecondary,
                              height: 1.55,
                              fontWeight: FontWeight.w400,
                            ),
                          )
                              .animate()
                              .fadeIn(duration: 450.ms, delay: 350.ms)
                              .slideY(
                                begin: 0.15,
                                end: 0,
                                curve: Curves.easeOutCubic,
                              ),

                          const SizedBox(height: 32),

                          // 4. Key Confirmation Summary Card
                          _buildConfirmationSummaryCard()
                              .animate()
                              .fadeIn(duration: 500.ms, delay: 450.ms)
                              .slideY(
                                begin: 0.12,
                                end: 0,
                                curve: Curves.easeOutCubic,
                              ),

                          const Spacer(flex: 2),
                          const SizedBox(height: 24),

                          // 5. Prominent "Back to Home" Button
                          _buildBackToHomeButton(context, ref)
                              .animate()
                              .fadeIn(duration: 400.ms, delay: 550.ms)
                              .slideY(
                                begin: 0.1,
                                end: 0,
                                curve: Curves.easeOutCubic,
                              ),

                          const SizedBox(height: 12),
                        ],
                      ),
                    ),
                  ),
                ),
              );
            },
          ),
        ),
      ),
    );
  }

  /// Builds the glowing, multi-ring animated badge with elastic bounce micro-interaction.
  Widget _buildAnimatedGlowingCheckBadge() {
    return Center(
      child: Container(
        width: 128,
        height: 128,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          color: const Color(0xFFD8F3DC).withValues(alpha: 0.45),
          boxShadow: [
            BoxShadow(
              color: AppColors.primary.withValues(alpha: 0.16),
              blurRadius: 36,
              spreadRadius: 8,
              offset: const Offset(0, 8),
            ),
            BoxShadow(
              color: const Color(0xFF95D5B2).withValues(alpha: 0.3),
              blurRadius: 20,
              spreadRadius: 2,
            ),
          ],
        ),
        child: Center(
          child: Container(
            width: 96,
            height: 96,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: const Color(0xFFB7E4C7).withValues(alpha: 0.5),
              border: Border.all(
                color: const Color(0xFF74C69D).withValues(alpha: 0.4),
                width: 1.5,
              ),
            ),
            child: Center(
              child: Container(
                width: 72,
                height: 72,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: const LinearGradient(
                    colors: [
                      Color(0xFF40916C),
                      Color(0xFF2D6A4F),
                      Color(0xFF1B4332),
                    ],
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  boxShadow: [
                    BoxShadow(
                      color: AppColors.primaryDark.withValues(alpha: 0.35),
                      blurRadius: 14,
                      offset: const Offset(0, 6),
                    ),
                  ],
                ),
                child: Center(
                  child: const Icon(
                    LucideIcons.check,
                    color: Colors.white,
                    size: 38,
                  )
                      .animate(delay: 220.ms)
                      .scale(
                        duration: 450.ms,
                        curve: Curves.bounceOut,
                        begin: const Offset(0.0, 0.0),
                        end: const Offset(1.0, 1.0),
                      ),
                ),
              ),
            ),
          ),
        ),
      )
          .animate()
          .scale(
            duration: 700.ms,
            curve: Curves.elasticOut,
            begin: const Offset(0.3, 0.3),
            end: const Offset(1.0, 1.0),
          )
          .fadeIn(duration: 400.ms),
    );
  }

  /// Builds the clean emerald-tinted summary card with shield icon acknowledgment.
  Widget _buildConfirmationSummaryCard() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 18),
      decoration: BoxDecoration(
        color: const Color(0xFFF2F9F5),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: const Color(0xFFB7E4C7),
          width: 1.2,
        ),
        boxShadow: [
          BoxShadow(
            color: AppColors.primary.withValues(alpha: 0.05),
            blurRadius: 16,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header Row with Shield Icon and Delivered Badge
          Row(
            children: [
              Container(
                width: 42,
                height: 42,
                decoration: BoxDecoration(
                  color: AppColors.secondary,
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: const Color(0xFF95D5B2),
                    width: 1.2,
                  ),
                ),
                child: const Center(
                  child: Icon(
                    LucideIcons.shieldCheck,
                    size: 22,
                    color: AppColors.primary,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: const [
                    Text(
                      'Submission Acknowledgment',
                      style: TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.w700,
                        color: AppColors.textPrimary,
                        letterSpacing: -0.2,
                      ),
                    ),
                    SizedBox(height: 2),
                    Text(
                      'Official State Executive Receipt',
                      style: TextStyle(
                        fontSize: 12.5,
                        color: AppColors.textSecondary,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: const Color(0xFFD8F3DC),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(
                    color: const Color(0xFF95D5B2),
                    width: 0.8,
                  ),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Container(
                      width: 6,
                      height: 6,
                      decoration: const BoxDecoration(
                        color: AppColors.primary,
                        shape: BoxShape.circle,
                      ),
                    ),
                    const SizedBox(width: 5),
                    const Text(
                      'Delivered',
                      style: TextStyle(
                        fontSize: 11.5,
                        fontWeight: FontWeight.w700,
                        color: AppColors.primaryDark,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),

          const Padding(
            padding: EdgeInsets.symmetric(vertical: 14),
            child: Divider(
              height: 1,
              thickness: 1,
              color: Color(0x202D6A4F),
            ),
          ),

          // Reassurance bullet items
          _buildSummaryPoint(
            title: 'Executive Committee Notified',
            description:
                'Your petition has been logged directly with the association executive board.',
          ),
          const SizedBox(height: 10),
          _buildSummaryPoint(
            title: 'Prompt Evaluation',
            description:
                'Issues are reviewed diligently according to the TAGA member charter guidelines.',
          ),
          const SizedBox(height: 10),
          _buildSummaryPoint(
            title: 'Strictly Confidential',
            description:
                'Your submission is handled with utmost privacy and member discretion.',
          ),
        ],
      ),
    );
  }

  /// Builds an individual reassurance point row.
  Widget _buildSummaryPoint({
    required String title,
    required String description,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          margin: const EdgeInsets.only(top: 2),
          width: 18,
          height: 18,
          decoration: const BoxDecoration(
            color: Color(0xFFD8F3DC),
            shape: BoxShape.circle,
          ),
          child: const Center(
            child: Icon(
              LucideIcons.check,
              size: 11,
              color: AppColors.primary,
            ),
          ),
        ),
        const SizedBox(width: 10),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: const TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                ),
              ),
              const SizedBox(height: 1),
              Text(
                description,
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
    );
  }

  /// Builds the prominent "Back to Home" CTA button with BouncyButton and emerald gradient.
  Widget _buildBackToHomeButton(BuildContext context, WidgetRef ref) {
    return BouncyButton(
      onTap: () {
        HapticFeedback.lightImpact();
        _navigateHome(context, ref);
      },
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(vertical: 16),
        decoration: BoxDecoration(
          gradient: const LinearGradient(
            colors: [
              Color(0xFF1B4332),
              Color(0xFF2D6A4F),
              Color(0xFF40916C),
            ],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: AppColors.primary.withValues(alpha: 0.32),
              blurRadius: 16,
              offset: const Offset(0, 6),
            ),
          ],
        ),
        child: const Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              LucideIcons.home,
              color: Colors.white,
              size: 19,
            ),
            SizedBox(width: 10),
            Text(
              'Back to Home',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w700,
                color: Colors.white,
                letterSpacing: 0.2,
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Handles clean navigation back to home and resets submission state.
  void _navigateHome(BuildContext context, WidgetRef ref) {
    ref.read(grievanceSubmissionProvider.notifier).reset();
    context.go('/home');
  }
}
