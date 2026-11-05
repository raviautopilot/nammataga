import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../domain/resource_category.dart';

/// Full-width vertical category row for the NammaTAGA Resources directory.
///
/// Follows modern enterprise document-library design:
/// - Deep forest green primary typography
/// - Soft mint green container for the category icon
/// - Complete category name without truncation (natural text wrapping)
/// - Right-facing navigation arrow for tactile visual hierarchy
/// - Clean white surface with subtle borders and restrained shadows
class CategoryCard extends StatelessWidget {
  final ResourceCategory category;
  final VoidCallback onTap;

  const CategoryCard({
    super.key,
    required this.category,
    required this.onTap,
  });

  /// Maps category IDs to contextual department and governance icons.
  static IconData getCategoryIcon(String categoryId) {
    switch (categoryId.toLowerCase().trim()) {
      case 'scheme-gos':
      case 'schemes':
        return LucideIcons.landmark;
      case 'pay-gos':
      case 'pay':
        return LucideIcons.banknote;
      case 'establishment':
        return LucideIcons.building2;
      case 'leave-forms':
      case 'forms':
        return LucideIcons.calendarCheck;
      case 'taga-membership':
      case 'membership':
        return LucideIcons.users;
      case 'office-contacts':
      case 'contacts':
        return LucideIcons.phoneCall;
      case 'technical':
        return LucideIcons.cpu;
      case 'miscellaneous':
        return LucideIcons.layers;
      case 'links':
      case 'external-links':
        return LucideIcons.globe;
      default:
        return LucideIcons.folder;
    }
  }

  @override
  Widget build(BuildContext context) {
    final icon = getCategoryIcon(category.id);
    final hasSubcategories =
        category.subcategories != null && category.subcategories!.isNotEmpty;

    return BouncyButton(
      onTap: onTap,
      scaleDownFactor: 0.98,
      child: Container(
        width: double.infinity,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(
            color: const Color(0xFFE5E7EB),
            width: 1.0,
          ),
          boxShadow: [
            BoxShadow(
              color: const Color(0xFF1B4332).withValues(alpha: 0.04),
              blurRadius: 8,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 15),
        child: Row(
          children: [
            // Soft mint green icon container
            Container(
              width: 44,
              height: 44,
              decoration: BoxDecoration(
                color: const Color(0xFFE8F5E9),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Center(
                child: Icon(
                  icon,
                  size: 22,
                  color: AppColors.primary,
                ),
              ),
            ),
            const SizedBox(width: 14),

            // Complete category name with natural wrapping (no truncation)
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    category.name,
                    style: const TextStyle(
                      color: AppColors.textPrimary,
                      fontSize: 15.5,
                      fontWeight: FontWeight.w600,
                      height: 1.3,
                      letterSpacing: -0.1,
                    ),
                    softWrap: true,
                  ),
                  if (hasSubcategories) ...[
                    const SizedBox(height: 3),
                    Text(
                      category.subcategories!.join(' • '),
                      style: const TextStyle(
                        color: Color(0xFF52796F),
                        fontSize: 12,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const SizedBox(width: 8),

            // Subtle right-facing navigation arrow
            const Icon(
              LucideIcons.chevronRight,
              size: 20,
              color: Color(0xFF9CA3AF),
            ),
          ],
        ),
      ),
    );
  }
}
