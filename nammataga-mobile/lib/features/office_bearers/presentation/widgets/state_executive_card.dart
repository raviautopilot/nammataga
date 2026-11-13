import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../domain/state_executive_member.dart';
import 'contact_actions_helper.dart';

/// Premium, bespoke card displaying a State Executive member in a rich, rectangular portrait layout.
/// Strictly follows NammaTAGA's emerald design system:
/// - Rectangular portrait photograph with subtle depth, rounded corners and gold/emerald ring badge
/// - Clean typography with high-contrast text and dual-tone designation pill
/// - Qualification, Experience, and Location with contextual pastel iconography
/// - Tactile BouncyButton quick-contact dials (Call & WhatsApp)
/// - Clean "View Profile" pill indicator
class StateExecutiveCard extends StatelessWidget {
  final StateExecutiveMember member;
  final VoidCallback onTap;

  const StateExecutiveCard({
    super.key,
    required this.member,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final imageUrl = member.resolvedImageUrl;
    final hasPhone = ContactActionsHelper.hasValidPhone(member.phone);

    return BouncyButton(
      onTap: onTap,
      scaleDownFactor: 0.985,
      child: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(18),
          border: Border.all(
            color: const Color(0xFFE2EBE5),
            width: 1.1,
          ),
          boxShadow: [
            BoxShadow(
              color: const Color(0xFF04190E).withValues(alpha: 0.05),
              blurRadius: 14,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        padding: const EdgeInsets.all(14),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // STRICT REQUIREMENT: Rectangular portrait photograph (NEVER circular)
            Stack(
              clipBehavior: Clip.none,
              children: [
                Container(
                  width: 98,
                  height: 132,
                  decoration: BoxDecoration(
                    color: const Color(0xFFE8F5EE),
                    borderRadius: BorderRadius.circular(14),
                    border: Border.all(
                      color: const Color(0xFFC8E6C9),
                      width: 1.6,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF1B4332).withValues(alpha: 0.12),
                        blurRadius: 8,
                        offset: const Offset(0, 3),
                      ),
                    ],
                  ),
                  child: ClipRRect(
                    borderRadius: BorderRadius.circular(12.4),
                    child: imageUrl != null && imageUrl.isNotEmpty
                        ? CachedNetworkImage(
                            imageUrl: imageUrl,
                            fit: BoxFit.cover,
                            placeholder: (context, url) => Container(
                              color: const Color(0xFFE8F5EE),
                              child: const Center(
                                child: SizedBox(
                                  width: 20,
                                  height: 20,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                    color: AppColors.primary,
                                  ),
                                ),
                              ),
                            ),
                            errorWidget: (context, url, error) =>
                                _buildInitialsFallback(),
                          )
                        : _buildInitialsFallback(),
                  ),
                ),
                // Corner seal badge matching web leadership award
                Positioned(
                  bottom: -4,
                  right: -4,
                  child: Container(
                    padding: const EdgeInsets.all(4.5),
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [Color(0xFF2D6A4F), Color(0xFF1B4332)],
                      ),
                      shape: BoxShape.circle,
                      border: Border.all(color: Colors.white, width: 1.8),
                      boxShadow: const [
                        BoxShadow(
                          color: Color(0x33000000),
                          blurRadius: 5,
                          offset: Offset(0, 2),
                        ),
                      ],
                    ),
                    child: const Icon(
                      LucideIcons.award,
                      size: 11,
                      color: Color(0xFFB7E4C7),
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(width: 15),

            // Content Column
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Full Name
                  Text(
                    member.name,
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w800,
                      color: AppColors.textPrimary,
                      letterSpacing: -0.3,
                      height: 1.22,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 5),

                  // Designation Tag
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8.5,
                      vertical: 3.5,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFFE8F5EE),
                      borderRadius: BorderRadius.circular(6),
                      border: Border.all(
                        color: const Color(0xFF74C69D).withValues(alpha: 0.5),
                        width: 0.9,
                      ),
                    ),
                    child: Text(
                      member.designation,
                      style: const TextStyle(
                        fontSize: 11.5,
                        fontWeight: FontWeight.w700,
                        color: AppColors.primaryDark,
                        letterSpacing: -0.1,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  const SizedBox(height: 9),

                  // Qualification
                  if (member.qualification.isNotEmpty)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 4),
                      child: Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(3),
                            decoration: BoxDecoration(
                              color: const Color(0xFFF1F8E9),
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: const Icon(
                              LucideIcons.graduationCap,
                              size: 12,
                              color: AppColors.primary,
                            ),
                          ),
                          const SizedBox(width: 6),
                          Expanded(
                            child: Text(
                              member.qualification,
                              style: const TextStyle(
                                fontSize: 12,
                                color: Color(0xFF334155),
                                fontWeight: FontWeight.w500,
                              ),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        ],
                      ),
                    ),

                  // Location & Experience
                  Padding(
                    padding: const EdgeInsets.only(bottom: 10),
                    child: Row(
                      children: [
                        if (member.location.isNotEmpty) ...[
                          Container(
                            padding: const EdgeInsets.all(3),
                            decoration: BoxDecoration(
                              color: const Color(0xFFF1F8E9),
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: const Icon(
                              LucideIcons.mapPin,
                              size: 12,
                              color: AppColors.primary,
                            ),
                          ),
                          const SizedBox(width: 5),
                          Flexible(
                            child: Text(
                              member.location,
                              style: const TextStyle(
                                fontSize: 12,
                                color: Color(0xFF334155),
                                fontWeight: FontWeight.w500,
                              ),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        ],
                        if (member.location.isNotEmpty && member.experience > 0)
                          const Padding(
                            padding: EdgeInsets.symmetric(horizontal: 6),
                            child: Text(
                              '•',
                              style: TextStyle(
                                color: Color(0xFF94A3B8),
                                fontSize: 12,
                              ),
                            ),
                          ),
                        if (member.experience > 0)
                          Text(
                            '${member.experience} yrs exp',
                            style: const TextStyle(
                              fontSize: 11.5,
                              color: Color(0xFF64748B),
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                      ],
                    ),
                  ),

                  // Quick Action Bar: Call, WhatsApp, View Profile
                  Row(
                    children: [
                      if (hasPhone) ...[
                        _buildMicroContactButton(
                          icon: LucideIcons.phone,
                          color: AppColors.primary,
                          bgColor: const Color(0xFFE8F5EE),
                          tooltip: 'Call',
                          onTap: () {
                            HapticFeedback.lightImpact();
                            ContactActionsHelper.makePhoneCall(
                              context,
                              member.phone,
                            );
                          },
                        ),
                        const SizedBox(width: 6),
                        _buildMicroContactButton(
                          icon: LucideIcons.messageSquare,
                          color: const Color(0xFF1E7E34),
                          bgColor: const Color(0xFFE8F8EE),
                          tooltip: 'WhatsApp',
                          onTap: () {
                            HapticFeedback.lightImpact();
                            ContactActionsHelper.openWhatsApp(
                              context,
                              member.phone,
                            );
                          },
                        ),
                      ],
                      const Spacer(),
                      // Profile pill button
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 9,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          color: const Color(0xFFF8FAF9),
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(
                            color: const Color(0xFFE2EBE5),
                          ),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: const [
                            Text(
                              'Profile',
                              style: TextStyle(
                                fontSize: 11.5,
                                fontWeight: FontWeight.w700,
                                color: AppColors.primary,
                              ),
                            ),
                            SizedBox(width: 3),
                            Icon(
                              LucideIcons.chevronRight,
                              size: 13,
                              color: AppColors.primary,
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMicroContactButton({
    required IconData icon,
    required Color color,
    required Color bgColor,
    required String tooltip,
    required VoidCallback onTap,
  }) {
    return BouncyButton(
      onTap: onTap,
      scaleDownFactor: 0.90,
      child: Container(
        width: 30,
        height: 30,
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

  Widget _buildInitialsFallback() {
    return Container(
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [Color(0xFFE8F5EE), Color(0xFFD8F3DC)],
        ),
      ),
      child: Center(
        child: Text(
          member.initials,
          style: const TextStyle(
            fontSize: 28,
            fontWeight: FontWeight.w800,
            color: AppColors.primaryDark,
            letterSpacing: 0.5,
          ),
        ),
      ),
    );
  }
}
