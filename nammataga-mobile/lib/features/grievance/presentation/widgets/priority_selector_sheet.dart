import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../domain/grievance_priority_model.dart';
import '../../providers/grievance_provider.dart';

/// Styling and icon configuration for priority tiers.
class _PriorityTierStyle {
  final Color textColor;
  final Color bgColor;
  final IconData icon;
  final String description;

  const _PriorityTierStyle({
    required this.textColor,
    required this.bgColor,
    required this.icon,
    required this.description,
  });
}

/// Modal bottom sheet enabling members to select a grievance priority.
/// Follows NammaTAGA emerald design guidelines with frosted header,
/// rounded 28px corners, color-coded indicator pills, tactile feedback,
/// and complete async state handling.
class PrioritySelectorSheet extends ConsumerWidget {
  final String? selectedValue;

  const PrioritySelectorSheet({
    super.key,
    this.selectedValue,
  });

  /// Displays the [PrioritySelectorSheet] as a modal bottom sheet and returns
  /// the selected [GrievancePriorityModel], or `null` if dismissed.
  static Future<GrievancePriorityModel?> show(
    BuildContext context, {
    String? selectedValue,
  }) {
    return showModalBottomSheet<GrievancePriorityModel>(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (context) => PrioritySelectorSheet(
        selectedValue: selectedValue,
      ),
    );
  }

  /// Resolves the color-coded theme configuration for a given priority.
  /// - Urgent: Soft Red (`#D4183D`, bg `#FFEBEB`)
  /// - High: Amber/Orange (`#E07A5F`, bg `#FFF3EB`)
  /// - Medium: Forest Emerald (`#2D6A4F`, bg `#E8F5E9`)
  /// - Low: Slate Gray (`#6B7280`, bg `#F3F4F6`)
  _PriorityTierStyle _getPriorityStyle(String key) {
    final normalized = key.toLowerCase().trim();
    if (normalized.contains('urgent')) {
      return const _PriorityTierStyle(
        textColor: Color(0xFFD4183D),
        bgColor: Color(0xFFFFEBEB),
        icon: LucideIcons.flame,
        description: 'Critical, immediate administrative intervention required',
      );
    } else if (normalized.contains('high')) {
      return const _PriorityTierStyle(
        textColor: Color(0xFFE07A5F),
        bgColor: Color(0xFFFFF3EB),
        icon: LucideIcons.arrowUp,
        description: 'Important matter needing prompt review and action',
      );
    } else if (normalized.contains('medium')) {
      return const _PriorityTierStyle(
        textColor: Color(0xFF2D6A4F),
        bgColor: Color(0xFFE8F5E9),
        icon: LucideIcons.activity,
        description: 'Standard grievance processed within normal timelines',
      );
    } else {
      return const _PriorityTierStyle(
        textColor: Color(0xFF6B7280),
        bgColor: Color(0xFFF3F4F6),
        icon: LucideIcons.arrowDown,
        description: 'General inquiry, feedback, or non-urgent request',
      );
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final prioritiesAsync = ref.watch(grievancePrioritiesProvider);

    return Container(
      constraints: BoxConstraints(
        maxHeight: MediaQuery.of(context).size.height * 0.82,
      ),
      decoration: const BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
        boxShadow: [
          BoxShadow(
            color: Color(0x331B4332),
            blurRadius: 32,
            offset: Offset(0, -6),
          ),
        ],
      ),
      child: SafeArea(
        top: false,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Frosted Header with Drag Handle
            _buildFrostedHeader(context),

            // Content Area
            Flexible(
              child: prioritiesAsync.when(
                loading: () => _buildLoadingState(),
                error: (error, _) => _buildErrorState(ref, error),
                data: (priorities) {
                  if (priorities.isEmpty) {
                    return _buildEmptyState(ref);
                  }
                  return _buildPriorityList(context, priorities);
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Frosted header with drag handle, title, and close button.
  Widget _buildFrostedHeader(BuildContext context) {
    return ClipRRect(
      borderRadius: const BorderRadius.vertical(top: Radius.circular(28)),
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 12, sigmaY: 12),
        child: Container(
          padding: const EdgeInsets.fromLTRB(20, 12, 16, 16),
          decoration: BoxDecoration(
            color: AppColors.surface.withValues(alpha: 0.92),
            border: const Border(
              bottom: BorderSide(color: AppColors.borderLight, width: 1),
            ),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Subtle Drag Handle
              Container(
                width: 44,
                height: 4.5,
                margin: const EdgeInsets.only(bottom: 14),
                decoration: BoxDecoration(
                  color: AppColors.textMuted.withValues(alpha: 0.28),
                  borderRadius: BorderRadius.circular(10),
                ),
              ),

              // Title, Subtitle & Dismiss button
              Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  const Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Select Priority',
                          style: TextStyle(
                            fontSize: 19,
                            fontWeight: FontWeight.bold,
                            color: AppColors.textPrimary,
                            letterSpacing: -0.2,
                          ),
                        ),
                        SizedBox(height: 3),
                        Text(
                          'Indicate the urgency level of this petition',
                          style: TextStyle(
                            fontSize: 13,
                            color: AppColors.textSecondary,
                          ),
                        ),
                      ],
                    ),
                  ),
                  BouncyButton(
                    onTap: () {
                      HapticFeedback.lightImpact();
                      Navigator.of(context).pop();
                    },
                    child: Container(
                      width: 36,
                      height: 36,
                      decoration: BoxDecoration(
                        color: AppColors.muted,
                        shape: BoxShape.circle,
                        border: Border.all(color: AppColors.borderLight),
                      ),
                      child: const Icon(
                        LucideIcons.x,
                        size: 18,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  /// Shimmer skeleton loading state.
  Widget _buildLoadingState() {
    return ListView.separated(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(20, 16, 20, 24),
      itemCount: 4,
      separatorBuilder: (_, _) => const SizedBox(height: 12),
      itemBuilder: (_, _) => const _PriorityShimmerItem(),
    );
  }

  /// Error state with retry action invoking [grievancePrioritiesProvider].
  Widget _buildErrorState(WidgetRef ref, Object error) {
    final cleanMessage =
        error.toString().replaceFirst(RegExp(r'^Exception:\s*'), '');

    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 36),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 56,
            height: 56,
            decoration: BoxDecoration(
              color: const Color(0xFFFFEBEB),
              shape: BoxShape.circle,
              border: Border.all(color: const Color(0xFFFFCDD2)),
            ),
            child: const Icon(
              LucideIcons.alertCircle,
              color: AppColors.destructive,
              size: 26,
            ),
          ),
          const SizedBox(height: 16),
          const Text(
            'Unable to Load Priorities',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 6),
          Text(
            cleanMessage.isNotEmpty
                ? cleanMessage
                : 'Failed to retrieve priority levels from server.',
            textAlign: TextAlign.center,
            style: const TextStyle(
              fontSize: 13,
              color: AppColors.textSecondary,
              height: 1.4,
            ),
          ),
          const SizedBox(height: 22),
          BouncyButton(
            onTap: () {
              HapticFeedback.lightImpact();
              final _ = ref.refresh(grievancePrioritiesProvider);
            },
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 22, vertical: 12),
              decoration: BoxDecoration(
                color: AppColors.primary,
                borderRadius: BorderRadius.circular(14),
                boxShadow: [
                  BoxShadow(
                    color: AppColors.primary.withValues(alpha: 0.25),
                    blurRadius: 10,
                    offset: const Offset(0, 3),
                  ),
                ],
              ),
              child: const Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(LucideIcons.rotateCcw, color: Colors.white, size: 16),
                  SizedBox(width: 8),
                  Text(
                    'Retry',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// Polite empty state when backend returns 0 priorities.
  Widget _buildEmptyState(WidgetRef ref) {
    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 36),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 56,
            height: 56,
            decoration: BoxDecoration(
              color: AppColors.muted,
              shape: BoxShape.circle,
              border: Border.all(color: AppColors.borderLight),
            ),
            child: const Icon(
              LucideIcons.inbox,
              color: AppColors.textSecondary,
              size: 26,
            ),
          ),
          const SizedBox(height: 16),
          const Text(
            'No Priorities Available',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 6),
          const Text(
            'No grievance priority levels are configured. Please try again later.',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 13,
              color: AppColors.textSecondary,
              height: 1.4,
            ),
          ),
          const SizedBox(height: 22),
          BouncyButton(
            onTap: () {
              HapticFeedback.lightImpact();
              final _ = ref.refresh(grievancePrioritiesProvider);
            },
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 11),
              decoration: BoxDecoration(
                color: AppColors.muted,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: AppColors.border),
              ),
              child: const Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(LucideIcons.rotateCcw,
                      color: AppColors.primary, size: 15),
                  SizedBox(width: 8),
                  Text(
                    'Refresh',
                    style: TextStyle(
                      color: AppColors.primary,
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// List of priorities with color-coded pills and visual selection highlight.
  Widget _buildPriorityList(
      BuildContext context, List<GrievancePriorityModel> priorities) {
    return ListView.separated(
      shrinkWrap: true,
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(20, 14, 20, 24),
      itemCount: priorities.length,
      separatorBuilder: (_, _) => const SizedBox(height: 12),
      itemBuilder: (context, index) {
        final priority = priorities[index];
        final isSelected = selectedValue != null &&
            (selectedValue!.trim().toLowerCase() ==
                    priority.value.trim().toLowerCase() ||
                selectedValue!.trim().toLowerCase() ==
                    priority.label.trim().toLowerCase());

        final style = _getPriorityStyle(
            priority.value.isNotEmpty ? priority.value : priority.label);

        return BouncyButton(
          onTap: () {
            HapticFeedback.lightImpact();
            Navigator.of(context).pop(priority);
          },
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 180),
            constraints: const BoxConstraints(minHeight: 56),
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: isSelected ? const Color(0xFFE8F5E9) : AppColors.surface,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: isSelected ? AppColors.primary : AppColors.border,
                width: isSelected ? 1.6 : 1.0,
              ),
              boxShadow: isSelected
                  ? [
                      BoxShadow(
                        color: AppColors.primary.withValues(alpha: 0.12),
                        blurRadius: 10,
                        offset: const Offset(0, 3),
                      ),
                    ]
                  : [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.02),
                        blurRadius: 4,
                        offset: const Offset(0, 2),
                      ),
                    ],
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                // Priority icon badge in priority tone
                Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: style.bgColor,
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: style.textColor.withValues(alpha: 0.25),
                      width: 1,
                    ),
                  ),
                  child: Icon(
                    style.icon,
                    size: 20,
                    color: style.textColor,
                  ),
                ),
                const SizedBox(width: 14),

                // Main priority details
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Row(
                        children: [
                          Text(
                            priority.label,
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: isSelected
                                  ? FontWeight.w700
                                  : FontWeight.w600,
                              color: isSelected
                                  ? AppColors.primaryDark
                                  : AppColors.textPrimary,
                              letterSpacing: -0.2,
                            ),
                          ),
                          const SizedBox(width: 8),

                          // Color-coded Indicator Pill
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8,
                              vertical: 2.5,
                            ),
                            decoration: BoxDecoration(
                              color: style.bgColor,
                              borderRadius: BorderRadius.circular(20),
                              border: Border.all(
                                color: style.textColor.withValues(alpha: 0.3),
                                width: 0.8,
                              ),
                            ),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Container(
                                  width: 6,
                                  height: 6,
                                  decoration: BoxDecoration(
                                    color: style.textColor,
                                    shape: BoxShape.circle,
                                  ),
                                ),
                                const SizedBox(width: 5),
                                Text(
                                  priority.value.toUpperCase(),
                                  style: TextStyle(
                                    fontSize: 10.5,
                                    fontWeight: FontWeight.w700,
                                    color: style.textColor,
                                    letterSpacing: 0.4,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 3),
                      Text(
                        style.description,
                        style: TextStyle(
                          fontSize: 12.5,
                          color: isSelected
                              ? AppColors.textSecondary
                              : AppColors.textMuted,
                          height: 1.3,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 10),

                // Selection checkmark or radio ring
                if (isSelected)
                  Container(
                    width: 26,
                    height: 26,
                    decoration: const BoxDecoration(
                      color: AppColors.primary,
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(
                      LucideIcons.check,
                      color: Colors.white,
                      size: 16,
                    ),
                  )
                else
                  Container(
                    width: 22,
                    height: 22,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: AppColors.border,
                        width: 1.5,
                      ),
                    ),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }
}

/// Shimmer skeleton row widget matching the layout of priority tiles.
class _PriorityShimmerItem extends StatelessWidget {
  const _PriorityShimmerItem();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 64,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppColors.borderLight),
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            Color(0xFFF6FAF7),
            Color(0xFFE8EFEA),
            Color(0xFFF6FAF7),
          ],
        ),
      ),
      child: Row(
        children: [
          Container(
            width: 38,
            height: 38,
            decoration: BoxDecoration(
              color: Colors.black.withValues(alpha: 0.05),
              borderRadius: BorderRadius.circular(12),
            ),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  width: 110,
                  height: 13,
                  decoration: BoxDecoration(
                    color: Colors.black.withValues(alpha: 0.06),
                    borderRadius: BorderRadius.circular(6),
                  ),
                ),
                const SizedBox(height: 6),
                Container(
                  width: 180,
                  height: 10,
                  decoration: BoxDecoration(
                    color: Colors.black.withValues(alpha: 0.04),
                    borderRadius: BorderRadius.circular(5),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
