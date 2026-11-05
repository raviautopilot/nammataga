import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';

/// Horizontally scrollable row of subcategory filter chips.
///
/// Features:
/// - First item is "All", clearing the subcategory filter (`null`)
/// - Subsequent items represent subcategories (e.g., "Central", "State")
/// - Distinct active state highlighting with emerald tint and check icon
/// - Smooth horizontal scrolling with [BouncingScrollPhysics]
class SubcategoryChipsBar extends StatelessWidget {
  final List<String> subcategories;
  final String? selectedSubcategory;
  final ValueChanged<String?> onSubcategorySelected;
  final EdgeInsetsGeometry padding;

  const SubcategoryChipsBar({
    super.key,
    required this.subcategories,
    required this.selectedSubcategory,
    required this.onSubcategorySelected,
    this.padding = const EdgeInsets.symmetric(horizontal: 16, vertical: 2),
  });

  @override
  Widget build(BuildContext context) {
    if (subcategories.isEmpty) {
      return const SizedBox.shrink();
    }

    // Build the list including "All" as first option
    final items = <String?>[null, ...subcategories];

    return SizedBox(
      height: 48,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        physics: const BouncingScrollPhysics(),
        padding: padding,
        itemCount: items.length,
        separatorBuilder: (context, index) => const SizedBox(width: 8),
        itemBuilder: (context, index) {
          final item = items[index];
          final isSelected = item == null
              ? (selectedSubcategory == null || selectedSubcategory!.isEmpty)
              : selectedSubcategory == item;
          final label = item ?? 'All';

          return _SubcategoryChip(
            label: label,
            isSelected: isSelected,
            onTap: () {
              HapticFeedback.selectionClick();
              onSubcategorySelected(item);
            },
          );
        },
      ),
    );
  }
}

class _SubcategoryChip extends StatelessWidget {
  final String label;
  final bool isSelected;
  final VoidCallback onTap;

  const _SubcategoryChip({
    required this.label,
    required this.isSelected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return BouncyButton(
      onTap: onTap,
      scaleDownFactor: 0.96,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeOutCubic,
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        decoration: BoxDecoration(
          color: isSelected
              ? AppColors.primary
              : const Color(0xFFF1F8E9),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: isSelected
                ? AppColors.primary
                : const Color(0xFFD8F3DC),
            width: 1.1,
          ),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: AppColors.primary.withValues(alpha: 0.18),
                    blurRadius: 4,
                    offset: const Offset(0, 1),
                  ),
                ]
              : null,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            if (isSelected) ...[
              const Icon(
                LucideIcons.check,
                size: 13,
                color: Colors.white,
              ),
              const SizedBox(width: 5),
            ],
            Text(
              label,
              style: TextStyle(
                color: isSelected ? Colors.white : AppColors.primaryDark,
                fontSize: 13,
                fontWeight: isSelected ? FontWeight.w700 : FontWeight.w600,
                letterSpacing: 0.2,
                height: 1.2,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
