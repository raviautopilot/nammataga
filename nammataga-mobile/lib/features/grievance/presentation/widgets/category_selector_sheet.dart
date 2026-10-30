import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../providers/grievance_provider.dart';

/// Modal bottom sheet enabling members to select a grievance category.
/// Follows NammaTAGA emerald design guidelines with frosted header,
/// rounded 28px corners, smooth tactile feedback, and comprehensive async state handling.
class CategorySelectorSheet extends ConsumerWidget {
  final String? selectedCategory;

  const CategorySelectorSheet({
    super.key,
    this.selectedCategory,
  });

  /// Displays the [CategorySelectorSheet] as a modal bottom sheet and returns
  /// the selected category [String], or `null` if dismissed.
  static Future<String?> show(
    BuildContext context, {
    String? selectedCategory,
  }) {
    return showModalBottomSheet<String>(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (context) => CategorySelectorSheet(
        selectedCategory: selectedCategory,
      ),
    );
  }

  /// Maps a category name to an evocative Lucide icon.
  IconData _getCategoryIcon(String category) {
    final lower = category.toLowerCase();
    if (lower.contains('pay') ||
        lower.contains('allowance') ||
        lower.contains('salary') ||
        lower.contains('wage')) {
      return LucideIcons.wallet;
    } else if (lower.contains('transfer') ||
        lower.contains('posting') ||
        lower.contains('station')) {
      return LucideIcons.arrowLeftRight;
    } else if (lower.contains('promotion') ||
        lower.contains('seniority') ||
        lower.contains('cadre') ||
        lower.contains('career')) {
      return LucideIcons.award;
    } else if (lower.contains('welfare') ||
        lower.contains('benefit') ||
        lower.contains('medical') ||
        lower.contains('health')) {
      return LucideIcons.heartHandshake;
    } else if (lower.contains('facility') ||
        lower.contains('office') ||
        lower.contains('infrastructure') ||
        lower.contains('equipment')) {
      return LucideIcons.building2;
    } else if (lower.contains('service') ||
        lower.contains('leave') ||
        lower.contains('pension')) {
      return LucideIcons.briefcase;
    }
    return LucideIcons.fileText;
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final categoriesAsync = ref.watch(grievanceCategoriesProvider);

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
              child: categoriesAsync.when(
                loading: () => _buildLoadingState(),
                error: (error, _) => _buildErrorState(ref, error),
                data: (categories) {
                  if (categories.isEmpty) {
                    return _buildEmptyState(ref);
                  }
                  return _buildCategoryList(context, categories);
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
                          'Select Category',
                          style: TextStyle(
                            fontSize: 19,
                            fontWeight: FontWeight.bold,
                            color: AppColors.textPrimary,
                            letterSpacing: -0.2,
                          ),
                        ),
                        SizedBox(height: 3),
                        Text(
                          'Choose the area related to your grievance',
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
      itemCount: 5,
      separatorBuilder: (_, _) => const SizedBox(height: 10),
      itemBuilder: (_, _) => const _CategoryShimmerItem(),
    );
  }

  /// Error state with retry action invoking [grievanceCategoriesProvider].
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
            'Unable to Load Categories',
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
                : 'Failed to retrieve grievance categories from server.',
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
              final _ = ref.refresh(grievanceCategoriesProvider);
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

  /// Polite empty state when backend returns 0 categories.
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
            'No Categories Available',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 6),
          const Text(
            'No grievance categories are currently listed. Please check back later.',
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
              final _ = ref.refresh(grievanceCategoriesProvider);
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

  /// Category item list with selection highlight.
  Widget _buildCategoryList(BuildContext context, List<String> categories) {
    return ListView.separated(
      shrinkWrap: true,
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(20, 14, 20, 24),
      itemCount: categories.length,
      separatorBuilder: (_, _) => const SizedBox(height: 10),
      itemBuilder: (context, index) {
        final category = categories[index];
        final isSelected = selectedCategory != null &&
            selectedCategory!.trim().toLowerCase() ==
                category.trim().toLowerCase();

        return BouncyButton(
          onTap: () {
            HapticFeedback.lightImpact();
            Navigator.of(context).pop(category);
          },
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 180),
            constraints: const BoxConstraints(minHeight: 52),
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
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
              children: [
                // Category icon container
                Container(
                  width: 38,
                  height: 38,
                  decoration: BoxDecoration(
                    color: isSelected
                        ? AppColors.primary.withValues(alpha: 0.12)
                        : AppColors.muted,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Icon(
                    _getCategoryIcon(category),
                    size: 19,
                    color: isSelected
                        ? AppColors.primary
                        : AppColors.textSecondary,
                  ),
                ),
                const SizedBox(width: 14),

                // Category name
                Expanded(
                  child: Text(
                    category,
                    style: TextStyle(
                      fontSize: 15,
                      fontWeight:
                          isSelected ? FontWeight.w700 : FontWeight.w500,
                      color: isSelected
                          ? AppColors.primaryDark
                          : AppColors.textPrimary,
                      letterSpacing: -0.1,
                    ),
                  ),
                ),

                // Selection checkmark or right chevron
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
                  Icon(
                    LucideIcons.chevronRight,
                    size: 18,
                    color: AppColors.textMuted.withValues(alpha: 0.45),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }
}

/// Shimmer skeleton row widget matching the layout of category tiles.
class _CategoryShimmerItem extends StatelessWidget {
  const _CategoryShimmerItem();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 54,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
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
            width: 32,
            height: 32,
            decoration: BoxDecoration(
              color: Colors.black.withValues(alpha: 0.05),
              borderRadius: BorderRadius.circular(8),
            ),
          ),
          const SizedBox(width: 14),
          Container(
            width: 150,
            height: 12,
            decoration: BoxDecoration(
              color: Colors.black.withValues(alpha: 0.06),
              borderRadius: BorderRadius.circular(6),
            ),
          ),
        ],
      ),
    );
  }
}
