import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../theme/tagatowers_design_tokens.dart';

/// Descriptor for each step within the booking wizard.
class BookingStepItem {
  final int stepIndex;
  final String title;
  final String shortTitle;
  final IconData icon;

  const BookingStepItem({
    required this.stepIndex,
    required this.title,
    required this.shortTitle,
    required this.icon,
  });
}

/// An elegant wizard header for TAGA Towers multi-step booking.
///
/// Displays:
/// 1. Room & Beds → 2. Guest Info → 3. Review → 4. Pay
/// with completed checkmarks, active ring glows, animated transitions,
/// and clickable previous steps.
class BookingStepProgressBar extends StatelessWidget {
  /// 1-based index representing the current step (1 to [steps.length])
  final int currentStep;

  /// Custom step list. Defaults to the standard 4 TAGA Towers booking steps.
  final List<BookingStepItem> steps;

  /// Optional callback invoked when a previously completed step or active step is tapped.
  final ValueChanged<int>? onStepTapped;

  /// Whether to display the compact progress tracker (useful for bottom sheets or small dialogs).
  final bool isCompact;

  const BookingStepProgressBar({
    super.key,
    required this.currentStep,
    this.steps = defaultSteps,
    this.onStepTapped,
    this.isCompact = false,
  });

  /// Standard TAGA Towers 4-step wizard sequence
  static const List<BookingStepItem> defaultSteps = [
    BookingStepItem(
      stepIndex: 1,
      title: 'Room & Beds',
      shortTitle: 'Rooms',
      icon: LucideIcons.bedDouble,
    ),
    BookingStepItem(
      stepIndex: 2,
      title: 'Guest Info',
      shortTitle: 'Guests',
      icon: LucideIcons.users,
    ),
    BookingStepItem(
      stepIndex: 3,
      title: 'Review',
      shortTitle: 'Review',
      icon: LucideIcons.clipboardCheck,
    ),
    BookingStepItem(
      stepIndex: 4,
      title: 'Pay',
      shortTitle: 'Payment',
      icon: LucideIcons.creditCard,
    ),
  ];

  @override
  Widget build(BuildContext context) {
    final clampedStep = currentStep.clamp(1, steps.length);
    final progressPercent = ((clampedStep - 1) / (steps.length - 1)).clamp(0.0, 1.0);

    return Container(
      width: double.infinity,
      padding: EdgeInsets.symmetric(
        horizontal: isCompact ? 12 : 16,
        vertical: isCompact ? 10 : 14,
      ),
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 16,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Header summary: Current step label and progress badge
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                    decoration: BoxDecoration(
                      color: TagaTowersDesignTokens.softMint,
                      borderRadius: BorderRadius.circular(999),
                      border: Border.all(
                        color: TagaTowersDesignTokens.mintHighlight,
                        width: 1,
                      ),
                    ),
                    child: Text(
                      'STEP $clampedStep OF ${steps.length}',
                      style: const TextStyle(
                        fontSize: 10.5,
                        fontWeight: FontWeight.w800,
                        letterSpacing: 0.6,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(
                    steps[clampedStep - 1].title,
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                      letterSpacing: -0.2,
                    ),
                  ),
                ],
              ),
              Text(
                '${(progressPercent * 100).toInt()}%',
                style: const TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),

          // Wizard Indicator Circles and Connecting Bars
          _buildIndicatorTrack(context, clampedStep),

          if (!isCompact) ...[
            const SizedBox(height: 8),
            // Step Labels Row
            _buildLabelsRow(context, clampedStep),
          ],
        ],
      ),
    );
  }

  /// Builds the horizontal track of circles and lines
  Widget _buildIndicatorTrack(BuildContext context, int activeStep) {
    final List<Widget> children = [];

    for (int i = 0; i < steps.length; i++) {
      final step = steps[i];
      final isCompleted = step.stepIndex < activeStep;
      final isActive = step.stepIndex == activeStep;

      // Add Circle Node
      children.add(
        _StepCircleNode(
          step: step,
          isCompleted: isCompleted,
          isActive: isActive,
          onTap: (onStepTapped != null && step.stepIndex <= activeStep)
              ? () {
                  HapticFeedback.selectionClick();
                  onStepTapped!(step.stepIndex);
                }
              : null,
        ),
      );

      // Add connecting line if not the last step
      if (i < steps.length - 1) {
        final isLineCompleted = step.stepIndex < activeStep;
        children.add(
          Expanded(
            child: _AnimatedProgressBarLine(
              isCompleted: isLineCompleted,
            ),
          ),
        );
      }
    }

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: children,
      ),
    );
  }

  /// Builds the aligned labels beneath each step
  Widget _buildLabelsRow(BuildContext context, int activeStep) {
    return Row(
      children: steps.map((step) {
        final isCompleted = step.stepIndex < activeStep;
        final isActive = step.stepIndex == activeStep;

        final Color textColor;
        final FontWeight fontWeight;

        if (isActive) {
          textColor = TagaTowersDesignTokens.deepEmerald;
          fontWeight = FontWeight.bold;
        } else if (isCompleted) {
          textColor = TagaTowersDesignTokens.primaryGreen;
          fontWeight = FontWeight.w600;
        } else {
          textColor = const Color(0xFF8A9E93);
          fontWeight = FontWeight.w500;
        }

        return Expanded(
          child: GestureDetector(
            behavior: HitTestBehavior.opaque,
            onTap: (onStepTapped != null && step.stepIndex <= activeStep)
                ? () {
                    HapticFeedback.selectionClick();
                    onStepTapped!(step.stepIndex);
                  }
                : null,
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 2),
              child: Text(
                step.shortTitle,
                textAlign: TextAlign.center,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: fontWeight,
                  color: textColor,
                  letterSpacing: -0.1,
                ),
              ),
            ),
          ),
        );
      }).toList(),
    );
  }
}

/// Circular indicator node representing a specific wizard step
class _StepCircleNode extends StatelessWidget {
  final BookingStepItem step;
  final bool isCompleted;
  final bool isActive;
  final VoidCallback? onTap;

  const _StepCircleNode({
    required this.step,
    required this.isCompleted,
    required this.isActive,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    Widget content;

    if (isCompleted) {
      // Completed Step: Filled Green with Checkmark
      content = AnimatedContainer(
        duration: const Duration(milliseconds: 250),
        curve: Curves.easeInOut,
        width: 32,
        height: 32,
        decoration: BoxDecoration(
          color: TagaTowersDesignTokens.primaryGreen,
          shape: BoxShape.circle,
          boxShadow: [
            BoxShadow(
              color: TagaTowersDesignTokens.primaryGreen.withValues(alpha: 0.35),
              blurRadius: 6,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: const Center(
          child: Icon(
            LucideIcons.check,
            size: 16,
            color: Colors.white,
          ),
        ),
      );
    } else if (isActive) {
      // Active Step: Pulsing halo with deep emerald background & icon or number
      content = AnimatedContainer(
        duration: const Duration(milliseconds: 250),
        curve: Curves.easeInOut,
        width: 34,
        height: 34,
        decoration: BoxDecoration(
          color: TagaTowersDesignTokens.deepEmerald,
          shape: BoxShape.circle,
          border: Border.all(
            color: TagaTowersDesignTokens.softMint,
            width: 2.5,
          ),
          boxShadow: [
            BoxShadow(
              color: TagaTowersDesignTokens.primaryGreen.withValues(alpha: 0.4),
              blurRadius: 10,
              spreadRadius: 1.5,
              offset: const Offset(0, 3),
            ),
          ],
        ),
        child: Center(
          child: Text(
            '${step.stepIndex}',
            style: const TextStyle(
              color: Colors.white,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
      );
    } else {
      // Upcoming Step: Muted circle with clean border
      content = Container(
        width: 28,
        height: 28,
        decoration: BoxDecoration(
          color: const Color(0xFFF4F7F5),
          shape: BoxShape.circle,
          border: Border.all(
            color: const Color(0xFFD3E0D8),
            width: 1.4,
          ),
        ),
        child: Center(
          child: Text(
            '${step.stepIndex}',
            style: const TextStyle(
              color: Color(0xFF8A9E93),
              fontSize: 12,
              fontWeight: FontWeight.w600,
            ),
          ),
        ),
      );
    }

    if (onTap != null) {
      return BouncyButton(
        onTap: onTap,
        child: SizedBox(
          width: 36,
          height: 36,
          child: Center(child: content),
        ),
      );
    }

    return SizedBox(
      width: 36,
      height: 36,
      child: Center(child: content),
    );
  }
}

/// Smooth connecting horizontal line between step nodes
class _AnimatedProgressBarLine extends StatelessWidget {
  final bool isCompleted;

  const _AnimatedProgressBarLine({
    required this.isCompleted,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 3.5,
      margin: const EdgeInsets.symmetric(horizontal: 4),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(999),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeInOut,
          decoration: BoxDecoration(
            color: isCompleted
                ? TagaTowersDesignTokens.primaryGreen
                : const Color(0xFFE2EBE5),
            borderRadius: BorderRadius.circular(999),
          ),
        ),
      ),
    );
  }
}
