import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../domain/resource_category.dart';

/// Horizontally scrollable row of chips displaying resource categories.
///
/// Features:
/// - Distinct active chip styling (emerald green gradient, white bold text, glow shadow)
/// - Inactive chip styling (soft mint/neutral background, border, dark text)
/// - Contextual Lucide category icons matching department themes
/// - Smooth horizontal scrolling with [BouncingScrollPhysics]
class CategoryChipsBar extends StatelessWidget {
  final List<ResourceCategory> categories;
  final String? selectedCategoryId;
  final ValueChanged<String> onCategorySelected;
  final EdgeInsetsGeometry padding;

  const CategoryChipsBar({
    super.key,
    required this.categories,
    required this.selectedCategoryId,
    required this.onCategorySelected,
    this.padding = const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
  });

  /// Maps category ID / name to a distinctive Lucide icon.
  static IconData getCategoryIcon(String categoryId) {
    switch (categoryId.toLowerCase().trim()) {
      case 'establishment':
        return LucideIcons.building2;
      case 'scheme-gos':
      case 'schemes':
        return LucideIcons.landmark;
      case 'forms':
      case 'applications':
        return LucideIcons.fileSpreadsheet;
      case 'circulars':
        return LucideIcons.fileText;
      case 'links':
      case 'external-links':
        return LucideIcons.link2;
      default:
        return LucideIcons.folderOpen;
    }
  }

  @override
  Widget build(BuildContext context) {
    if (categories.isEmpty) {
      return const SizedBox.shrink();
    }

    return SizedBox(
      height: 48,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        physics: const BouncingScrollPhysics(),
        padding: padding,
        itemCount: categories.length,
        separatorBuilder: (context, index) => const SizedBox(width: 8),
        itemBuilder: (context, index) {
          final category = categories[index];
          final isSelected = category.id == selectedCategoryId;
          final icon = getCategoryIcon(category.id);

          return _CategoryChip(
            category: category,
            icon: icon,
            isSelected: isSelected,
            onTap: () {
              HapticFeedback.selectionClick();
              onCategorySelected(category.id);
            },
          );
        },
      ),
    );
  }
}

class _CategoryChip extends StatelessWidget {
  final ResourceCategory category;
  final IconData icon;
  final bool isSelected;
  final VoidCallback onTap;

  const _CategoryChip({
    required this.category,
    required this.icon,
    required this.isSelected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return BouncyButton(
      onTap: onTap,
      scaleDownFactor: 0.96,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 220),
        curve: Curves.easeOutCubic,
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        decoration: BoxDecoration(
          gradient: isSelected
              ? const LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  colors: [
                    Color(0xFF0F3622),
                    Color(0xFF1B4332),
                    Color(0xFF2D6A4F),
                  ],
                )
              : null,
          color: isSelected ? null : const Color(0xFFF6FAF7),
          borderRadius: BorderRadius.circular(24),
          border: Border.all(
            color: isSelected
                ? const Color(0xFF52B788).withValues(alpha: 0.6)
                : const Color(0xFFE2EBE5),
            width: isSelected ? 1.4 : 1.1,
          ),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: const Color(0xFF1B4332).withValues(alpha: 0.28),
                    blurRadius: 10,
                    offset: const Offset(0, 3),
                  ),
                ]
              : null,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              icon,
              size: 16,
              color: isSelected
                  ? const Color(0xFFB7E4C7)
                  : AppColors.textSecondary,
            ),
            const SizedBox(width: 8),
            Text(
              category.name,
              style: TextStyle(
                color: isSelected ? Colors.white : AppColors.textPrimary,
                fontSize: 13.5,
                fontWeight: isSelected ? FontWeight.w700 : FontWeight.w600,
                letterSpacing: 0.1,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
