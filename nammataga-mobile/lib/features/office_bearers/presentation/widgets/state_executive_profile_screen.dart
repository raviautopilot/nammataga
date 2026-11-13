import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../domain/state_executive_member.dart';
import 'contact_actions_helper.dart';

/// Full-screen profile view for a State Executive Committee member.
/// Features:
/// - Rich emerald gradient header with back navigation
/// - Centered portrait photograph card with gold accent ring and subtle shadow
/// - Official Tamil association context banner
/// - High-contrast action bar for Call & WhatsApp
/// - Detailed credential cards (Designation, Qualification, Experience, Location, Bio)
class StateExecutiveProfileScreen extends StatelessWidget {
  final StateExecutiveMember member;

  const StateExecutiveProfileScreen({
    super.key,
    required this.member,
  });

  @override
  Widget build(BuildContext context) {
    final imageUrl = member.resolvedImageUrl;
    final hasPhone = ContactActionsHelper.hasValidPhone(member.phone);

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAF7),
      body: CustomScrollView(
        physics: const BouncingScrollPhysics(),
        slivers: [
          // App Bar with Emerald Header & Back Button
          SliverAppBar(
            expandedHeight: 280,
            pinned: true,
            elevation: 0,
            backgroundColor: const Color(0xFF0D281E),
            title: const Text(
              'Executive Profile',
              style: TextStyle(
                fontSize: 16.5,
                fontWeight: FontWeight.bold,
                color: Colors.white,
              ),
            ),
            centerTitle: true,
            leading: Padding(
              padding: const EdgeInsets.all(8.0),
              child: Semantics(
                label: 'Back',
                button: true,
                child: Tooltip(
                  message: 'Back',
                  child: BouncyButton(
                    onTap: () {
                      HapticFeedback.lightImpact();
                      Navigator.of(context).pop();
                    },
                    child: Container(
                      decoration: BoxDecoration(
                        color: Colors.black.withValues(alpha: 0.35),
                        shape: BoxShape.circle,
                        border: Border.all(
                          color: Colors.white.withValues(alpha: 0.25),
                          width: 1,
                        ),
                      ),
                      child: const Center(
                        child: Icon(
                          LucideIcons.arrowLeft,
                          color: Colors.white,
                          size: 18,
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ),
            flexibleSpace: FlexibleSpaceBar(
              background: Stack(
                fit: StackFit.expand,
                children: [
                  // Emerald Gradient
                  Container(
                    decoration: const BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          Color(0xFF04170D),
                          Color(0xFF0A2B1B),
                          Color(0xFF14452F),
                          Color(0xFF1D5A3D),
                        ],
                      ),
                    ),
                  ),
                  // Ambient glow
                  Positioned(
                    top: -40,
                    right: -30,
                    width: 200,
                    height: 200,
                    child: Container(
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        gradient: RadialGradient(
                          colors: [
                            const Color(0xFF52B788).withValues(alpha: 0.3),
                            Colors.transparent,
                          ],
                        ),
                      ),
                    ),
                  ),
                  // Watermark
                  Positioned(
                    right: 10,
                    bottom: 20,
                    child: Icon(
                      LucideIcons.award,
                      size: 140,
                      color: Colors.white.withValues(alpha: 0.05),
                    ),
                  ),
                  // Portrait photograph centerpiece
                  Positioned(
                    bottom: 20,
                    left: 0,
                    right: 0,
                    child: Center(
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Container(
                            width: 130,
                            height: 168,
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(16),
                              border: Border.all(
                                color: const Color(0xFFD8F3DC),
                                width: 3.5,
                              ),
                              boxShadow: const [
                                BoxShadow(
                                  color: Color(0x3D000000),
                                  blurRadius: 18,
                                  offset: Offset(0, 8),
                                ),
                              ],
                            ),
                            child: ClipRRect(
                              borderRadius: BorderRadius.circular(12),
                              child: imageUrl != null && imageUrl.isNotEmpty
                                  ? CachedNetworkImage(
                                      imageUrl: imageUrl,
                                      fit: BoxFit.cover,
                                      placeholder: (context, url) => Container(
                                        color: const Color(0xFFE8F5EE),
                                        child: const Center(
                                          child: SizedBox(
                                            width: 24,
                                            height: 24,
                                            child: CircularProgressIndicator(
                                              strokeWidth: 2,
                                              color: AppColors.primary,
                                            ),
                                          ),
                                        ),
                                      ),
                                      errorWidget: (context, url, error) =>
                                          _buildFallback(),
                                    )
                                  : _buildFallback(),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),

          // Body Content
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(20, 16, 20, 36),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Full Name
                  Text(
                    member.name,
                    textAlign: TextAlign.center,
                    style: const TextStyle(
                      fontSize: 22,
                      fontWeight: FontWeight.w800,
                      color: AppColors.textPrimary,
                      letterSpacing: -0.4,
                    ),
                  ),
                  const SizedBox(height: 6),

                  // Official Designation Pill
                  Center(
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 14,
                        vertical: 5,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFFE8F5EE),
                        borderRadius: BorderRadius.circular(20),
                        border: Border.all(
                          color: const Color(0xFF74C69D).withValues(alpha: 0.5),
                        ),
                      ),
                      child: Text(
                        member.designation,
                        style: const TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w700,
                          color: AppColors.primaryDark,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 20),

                  // Call & WhatsApp Action Buttons
                  if (hasPhone)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 22),
                      child: Row(
                        children: [
                          Expanded(
                            child: BouncyButton(
                              onTap: () {
                                HapticFeedback.lightImpact();
                                ContactActionsHelper.makePhoneCall(
                                  context,
                                  member.phone,
                                );
                              },
                              child: Container(
                                padding:
                                    const EdgeInsets.symmetric(vertical: 13),
                                decoration: BoxDecoration(
                                  gradient: const LinearGradient(
                                    colors: [
                                      Color(0xFF2D6A4F),
                                      Color(0xFF1B4332),
                                    ],
                                  ),
                                  borderRadius: BorderRadius.circular(12),
                                  boxShadow: const [
                                    BoxShadow(
                                      color: Color(0x1F1B4332),
                                      blurRadius: 8,
                                      offset: Offset(0, 3),
                                    ),
                                  ],
                                ),
                                child: Row(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: const [
                                    Icon(
                                      LucideIcons.phone,
                                      size: 16,
                                      color: Colors.white,
                                    ),
                                    SizedBox(width: 8),
                                    Text(
                                      'Call Official',
                                      style: TextStyle(
                                        fontSize: 14,
                                        fontWeight: FontWeight.w700,
                                        color: Colors.white,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: BouncyButton(
                              onTap: () {
                                HapticFeedback.lightImpact();
                                ContactActionsHelper.openWhatsApp(
                                  context,
                                  member.phone,
                                );
                              },
                              child: Container(
                                padding:
                                    const EdgeInsets.symmetric(vertical: 13),
                                decoration: BoxDecoration(
                                  color: Colors.white,
                                  borderRadius: BorderRadius.circular(12),
                                  border: Border.all(
                                    color: const Color(0xFF25D366),
                                    width: 1.5,
                                  ),
                                  boxShadow: const [
                                    BoxShadow(
                                      color: Color(0x0F25D366),
                                      blurRadius: 8,
                                      offset: Offset(0, 3),
                                    ),
                                  ],
                                ),
                                child: Row(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: const [
                                    Icon(
                                      LucideIcons.messageSquare,
                                      size: 16,
                                      color: Color(0xFF1E7E34),
                                    ),
                                    SizedBox(width: 8),
                                    Text(
                                      'WhatsApp',
                                      style: TextStyle(
                                        fontSize: 14,
                                        fontWeight: FontWeight.w700,
                                        color: Color(0xFF1E7E34),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),

                  // Credential Card
                  Container(
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(18),
                      border: Border.all(
                        color: const Color(0xFFE2EBE5),
                        width: 1.1,
                      ),
                      boxShadow: const [
                        BoxShadow(
                          color: Color(0x061B4332),
                          blurRadius: 12,
                          offset: Offset(0, 3),
                        ),
                      ],
                    ),
                    child: Column(
                      children: [
                        if (member.qualification.isNotEmpty)
                          _buildInfoRow(
                            icon: LucideIcons.graduationCap,
                            label: 'Educational Qualification',
                            value: member.qualification,
                          ),
                        if (member.experience > 0)
                          _buildInfoRow(
                            icon: LucideIcons.briefcase,
                            label: 'Professional Experience',
                            value: '${member.experience} Years in Agriculture',
                          ),
                        if (member.location.isNotEmpty)
                          _buildInfoRow(
                            icon: LucideIcons.mapPin,
                            label: 'Home District / Location',
                            value: member.location,
                          ),
                        if (member.phone.isNotEmpty)
                          _buildInfoRow(
                            icon: LucideIcons.phone,
                            label: 'Official Contact',
                            value: member.phone,
                            isLast: member.description.isEmpty,
                          ),
                        if (member.description.isNotEmpty)
                          _buildInfoRow(
                            icon: LucideIcons.info,
                            label: 'Association Background',
                            value: member.description,
                            isLast: true,
                          ),
                      ],
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

  Widget _buildFallback() {
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
            fontSize: 38,
            fontWeight: FontWeight.bold,
            color: AppColors.primary,
          ),
        ),
      ),
    );
  }

  Widget _buildInfoRow({
    required IconData icon,
    required String label,
    required String value,
    bool isLast = false,
  }) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: const Color(0xFFE8F5EE),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(icon, size: 16, color: AppColors.primary),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      label,
                      style: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: AppColors.textSecondary,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      value,
                      style: const TextStyle(
                        fontSize: 14.5,
                        fontWeight: FontWeight.w700,
                        color: AppColors.textPrimary,
                        height: 1.35,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
        if (!isLast)
          const Divider(height: 1, thickness: 1, color: Color(0xFFF1F5F2)),
      ],
    );
  }
}
