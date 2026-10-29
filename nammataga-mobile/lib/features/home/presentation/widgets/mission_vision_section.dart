import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../providers/home_provider.dart';

class MissionVisionSection extends ConsumerWidget {
  const MissionVisionSection({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final aboutAsync = ref.watch(aboutInfoProvider);

    return aboutAsync.when(
      data: (about) {
        final hasMission = about.mission.isNotEmpty;
        final hasVision = about.vision.isNotEmpty;

        if (!hasMission && !hasVision) return const SizedBox.shrink();

        return Column(
          children: [
            if (hasMission) ...[
              _buildCard(
                title: 'Our Mission',
                description: about.mission,
                icon: LucideIcons.target,
                badgeColor: const Color(0xFFE8F5EE),
                iconColor: AppColors.primary,
              ),
              if (hasVision) const SizedBox(height: 12),
            ],
            if (hasVision)
              _buildCard(
                title: 'Our Vision',
                description: about.vision,
                icon: LucideIcons.heart,
                badgeColor: const Color(0xFFF1F8E9),
                iconColor: const Color(0xFF2D6A4F),
              ),
          ],
        );
      },
      loading: () => const SizedBox.shrink(),
      error: (error, stack) => const SizedBox.shrink(),
    );
  }

  Widget _buildCard({
    required String title,
    required String description,
    required IconData icon,
    required Color badgeColor,
    required Color iconColor,
  }) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
        boxShadow: const [
          BoxShadow(
            color: Color(0x061B4332),
            blurRadius: 10,
            offset: Offset(0, 3),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(7),
                decoration: BoxDecoration(
                  color: badgeColor,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(icon, color: iconColor, size: 18),
              ),
              const SizedBox(width: 10),
              Text(
                title,
                style: const TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: AppColors.primary,
                  letterSpacing: -0.2,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Text(
            description,
            style: const TextStyle(
              fontSize: 13,
              color: AppColors.textSecondary,
              height: 1.5,
            ),
          ),
        ],
      ),
    );
  }
}
