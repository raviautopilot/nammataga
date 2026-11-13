import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../domain/district_office_bearer.dart';
import 'contact_actions_helper.dart';

/// Structured table-like directory row for a District Office Bearer.
/// STRICT REQUIREMENT: No photos, no avatars, no circular images, no image placeholders.
/// Polished with NammaTAGA styling:
/// - Clear typography, prominent status/designation chip
/// - Tactile Call & WhatsApp micro-actions
/// - Elegant alternating soft green-tint zebra striping
/// - Responsive text wrapping with natural overflow prevention
class DistrictBearerRow extends StatelessWidget {
  final DistrictOfficeBearer bearer;
  final VoidCallback onTap;
  final bool isEven;
  final bool isLast;

  const DistrictBearerRow({
    super.key,
    required this.bearer,
    required this.onTap,
    this.isEven = false,
    this.isLast = false,
  });

  @override
  Widget build(BuildContext context) {
    final hasPhone = ContactActionsHelper.hasValidPhone(bearer.contact);

    return BouncyButton(
      onTap: onTap,
      scaleDownFactor: 0.99,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 13),
        decoration: BoxDecoration(
          color: isEven ? const Color(0xFFF9FCFA) : Colors.white,
          border: isLast
              ? null
              : const Border(
                  bottom: BorderSide(
                    color: Color(0xFFEBF1ED),
                    width: 1,
                  ),
                ),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            // Column 1: Member Name & Designation Badge
            Expanded(
              flex: 6,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    bearer.name,
                    style: const TextStyle(
                      fontSize: 14.5,
                      fontWeight: FontWeight.w700,
                      color: AppColors.textPrimary,
                      letterSpacing: -0.2,
                      height: 1.25,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 4),
                  // Title / Designation Tag
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 3,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFFE8F5EE),
                      borderRadius: BorderRadius.circular(5),
                      border: Border.all(
                        color: const Color(0xFF74C69D).withValues(alpha: 0.45),
                        width: 0.8,
                      ),
                    ),
                    child: Text(
                      bearer.title,
                      style: const TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        color: AppColors.primaryDark,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 8),

            // Column 2: Quick Contact Micro-Actions
            if (hasPhone) ...[
              _buildMicroButton(
                icon: LucideIcons.phone,
                color: AppColors.primary,
                bgColor: const Color(0xFFE8F5EE),
                onTap: () {
                  HapticFeedback.lightImpact();
                  ContactActionsHelper.makePhoneCall(
                    context,
                    bearer.contact,
                  );
                },
              ),
              const SizedBox(width: 6),
              _buildMicroButton(
                icon: LucideIcons.messageSquare,
                color: const Color(0xFF1E7E34),
                bgColor: const Color(0xFFE8F8EE),
                onTap: () {
                  HapticFeedback.lightImpact();
                  ContactActionsHelper.openWhatsApp(
                    context,
                    bearer.contact,
                  );
                },
              ),
            ],

            const SizedBox(width: 8),

            // Profile navigation indicator
            const Icon(
              LucideIcons.chevronRight,
              size: 17,
              color: Color(0xFF94A3B8),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMicroButton({
    required IconData icon,
    required Color color,
    required Color bgColor,
    required VoidCallback onTap,
  }) {
    return BouncyButton(
      onTap: onTap,
      scaleDownFactor: 0.88,
      child: Container(
        width: 32,
        height: 32,
        decoration: BoxDecoration(
          color: bgColor,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: color.withValues(alpha: 0.25),
            width: 1,
          ),
        ),
        child: Center(
          child: Icon(
            icon,
            size: 14,
            color: color,
          ),
        ),
      ),
    );
  }
}
